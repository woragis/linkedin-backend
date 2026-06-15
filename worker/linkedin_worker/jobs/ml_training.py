"""Weekly ML model training for connection acceptance."""

from __future__ import annotations

import json
import logging
import pickle
import uuid
from pathlib import Path

import psycopg

log = logging.getLogger("linkedin-worker.ml")

MODEL_DIR = Path("/tmp/linkedin_models")
FEATURE_NAMES = ["mutual", "same_school", "shared_skills", "same_company"]


def _features(conn: psycopg.Connection, requester: str, addressee: str) -> list[float]:
    mutual = conn.execute(
        """
        WITH rp AS (
          SELECT CASE WHEN requester_id = %s::uuid THEN addressee_id ELSE requester_id END AS peer
          FROM connections WHERE status = 'accepted' AND %s::uuid IN (requester_id, addressee_id)
        ),
        tp AS (
          SELECT CASE WHEN requester_id = %s::uuid THEN addressee_id ELSE requester_id END AS peer
          FROM connections WHERE status = 'accepted' AND %s::uuid IN (requester_id, addressee_id)
        )
        SELECT COUNT(*)::int FROM rp JOIN tp ON rp.peer = tp.peer
        """,
        (requester, requester, addressee, addressee),
    ).fetchone()
    same_school = conn.execute(
        """
        SELECT 1 FROM educations e1 JOIN educations e2 ON e1.institution_id = e2.institution_id
        WHERE e1.user_id = %s::uuid AND e2.user_id = %s::uuid LIMIT 1
        """,
        (requester, addressee),
    ).fetchone()
    skills = conn.execute(
        """
        SELECT COUNT(*)::int FROM user_skills u1
        JOIN user_skills u2 ON u1.skill_id = u2.skill_id
        WHERE u1.user_id = %s::uuid AND u2.user_id = %s::uuid
        """,
        (requester, addressee),
    ).fetchone()
    same_co = conn.execute(
        """
        SELECT 1 FROM experiences e1 JOIN experiences e2 ON e1.company_id = e2.company_id
        WHERE e1.user_id = %s::uuid AND e2.user_id = %s::uuid LIMIT 1
        """,
        (requester, addressee),
    ).fetchone()
    return [
        float(mutual[0]) if mutual else 0.0,
        1.0 if same_school else 0.0,
        float(skills[0]) if skills else 0.0,
        1.0 if same_co else 0.0,
    ]


def run_batch(conn: psycopg.Connection) -> None:
    log.info("ml_training batch started")
    rows = conn.execute(
        "SELECT requester_id::text, addressee_id::text, status FROM connections"
    ).fetchall()
    if len(rows) < 5:
        log.info("ml_training skipped: insufficient data")
        return

    X: list[list[float]] = []
    y: list[int] = []
    for req, addr, status in rows:
        X.append(_features(conn, req, addr))
        y.append(1 if status == "accepted" else 0)

    try:
        from sklearn.linear_model import LogisticRegression
        from sklearn.metrics import roc_auc_score
        from sklearn.model_selection import train_test_split
    except ImportError:
        log.warning("sklearn not installed, skipping ml training")
        return

    if len(set(y)) < 2:
        log.info("ml_training skipped: single class only")
        return

    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.25, random_state=42)
    model = LogisticRegression(max_iter=500)
    model.fit(X_train, y_train)
    auc = 0.5
    if len(set(y_test)) > 1:
        proba = model.predict_proba(X_test)[:, 1]
        auc = float(roc_auc_score(y_test, proba))

    MODEL_DIR.mkdir(parents=True, exist_ok=True)
    version = uuid.uuid4().hex[:8]
    path = MODEL_DIR / f"connection_acceptance_{version}.pkl"
    with path.open("wb") as f:
        pickle.dump({"model": model, "features": FEATURE_NAMES}, f)

    conn.execute("UPDATE model_versions SET is_active = false WHERE model_name = 'connection_acceptance'")
    conn.execute(
        """
        INSERT INTO model_versions (model_name, version, metrics, artifact_path, is_active, trained_at)
        VALUES ('connection_acceptance', %s, %s::jsonb, %s, true, now())
        """,
        (version, json.dumps({"roc_auc": auc, "samples": len(y)}), str(path)),
    )
    conn.commit()
    log.info("ml_training batch finished version=%s auc=%.3f", version, auc)
