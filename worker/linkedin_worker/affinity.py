"""Pure affinity scoring helpers (no database)."""

from __future__ import annotations

WEIGHTS = {
    "mutual_connections": 0.35,
    "same_school": 0.15,
    "shared_skills": 0.12,
    "same_company": 0.10,
    "graduation_cohort": 0.08,
}


def compute_affinity_score(
    *,
    mutual_connections: int = 0,
    same_school: bool = False,
    shared_skills: int = 0,
    same_company: bool = False,
    graduation_cohort: bool = False,
) -> tuple[float, list[str]]:
    reasons: list[str] = []
    score = 0.0

    if mutual_connections > 0:
        score += WEIGHTS["mutual_connections"] * min(mutual_connections / 5.0, 1.0)
        reasons.append(f"{mutual_connections} mutual connections")

    if same_school:
        score += WEIGHTS["same_school"]
        reasons.append("same school")

    if shared_skills > 0:
        score += WEIGHTS["shared_skills"] * min(shared_skills / 3.0, 1.0)
        reasons.append(f"{shared_skills} shared skills")

    if same_company:
        score += WEIGHTS["same_company"]
        reasons.append("same company")

    if graduation_cohort:
        score += WEIGHTS["graduation_cohort"]
        reasons.append("similar graduation period")

    return score, reasons
