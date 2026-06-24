from __future__ import annotations

import logging
import math
import random
from uuid import UUID, uuid4

import psycopg

from linkedin_worker import settings
from linkedin_worker.jobs.graph import run_batch
from linkedin_worker.simulator import demographics
from linkedin_worker.simulator.bootstrap import password_hash
from linkedin_worker.simulator.db import count_simulator_agents, load_existing_slugs

log = logging.getLogger("linkedin-worker.simulator.graph_bootstrap")

_GRAPH_ARCHETYPE = "graph_lab"


def count_graph_lab_agents(conn: psycopg.Connection) -> int:
    row = conn.execute(
        "SELECT COUNT(*)::int FROM simulator_agents WHERE archetype = %s",
        (_GRAPH_ARCHETYPE,),
    ).fetchone()
    return int(row[0]) if row else 0


def sample_degrees(
    n: int,
    rng: random.Random,
    *,
    mean: float,
    min_degree: int,
    max_degree: int,
) -> list[int]:
    degrees: list[int] = []
    for _ in range(n):
        value = int(rng.lognormvariate(math.log(max(mean, 1.0)), 0.55))
        degrees.append(max(min_degree, min(max_degree, value)))
    return degrees


def chung_lu_edges(
    degrees: list[int],
    rng: random.Random,
) -> list[tuple[int, int]]:
    n = len(degrees)
    total = sum(degrees)
    if n < 2 or total <= 0:
        return []

    edges: list[tuple[int, int]] = []
    for i in range(n):
        for j in range(i + 1, n):
            probability = min(degrees[i] * degrees[j] / total, 1.0)
            if rng.random() < probability:
                edges.append((i, j))
    return edges


def bootstrap_graph(conn: psycopg.Connection) -> int:
    existing = count_graph_lab_agents(conn)
    target = settings.simulator_target_count()
    if existing >= target:
        log.info("graph bootstrap skipped existing=%s target=%s", existing, target)
        return 0

    remaining = target - existing
    rng = random.Random(settings.SIMULATOR_SEED + existing)
    taken_slugs = load_existing_slugs(conn)
    pwd_hash = password_hash()
    commit_every = max(1, settings.SIMULATOR_BOOTSTRAP_COMMIT_EVERY)

    log.info(
        "graph bootstrap starting remaining=%s target=%s mean_degree=%s",
        remaining,
        target,
        settings.SIMULATOR_GRAPH_MEAN_DEGREE,
    )

    user_ids: list[UUID] = []
    for index in range(remaining):
        user_id = _create_graph_user(conn, rng, taken_slugs, pwd_hash, existing + index)
        user_ids.append(user_id)
        if (index + 1) % commit_every == 0:
            conn.commit()
            log.info("graph bootstrap users progress=%s/%s", index + 1, remaining)

    conn.commit()
    log.info("graph bootstrap users created=%s", len(user_ids))

    if len(user_ids) >= 2:
        degrees = sample_degrees(
            len(user_ids),
            rng,
            mean=settings.SIMULATOR_GRAPH_MEAN_DEGREE,
            min_degree=settings.SIMULATOR_GRAPH_MIN_DEGREE,
            max_degree=settings.SIMULATOR_GRAPH_MAX_DEGREE,
        )
        edge_pairs = chung_lu_edges(degrees, rng)
        _bulk_insert_connections(conn, user_ids, edge_pairs)
        conn.commit()
        log.info(
            "graph bootstrap edges=%s avg_degree_target=%.1f",
            len(edge_pairs),
            settings.SIMULATOR_GRAPH_MEAN_DEGREE,
        )

    run_batch(conn)
    total = count_graph_lab_agents(conn)
    log.info("graph bootstrap complete total_agents=%s", total)
    return len(user_ids)


def _create_graph_user(
    conn: psycopg.Connection,
    rng: random.Random,
    taken_slugs: set[str],
    pwd_hash: str,
    rng_offset: int,
) -> UUID:
    gender = demographics.pick_gender(rng)
    city = demographics.pick_city(rng)
    user_id = uuid4()
    full_name = demographics.sample_name(rng, gender)
    slug = demographics.ensure_unique_slug(demographics.slug_from_name(full_name), taken_slugs)
    taken_slugs.add(slug)
    email = f"graph-{user_id}@sim.local"
    headline = rng.choice(
        (
            "Pesquisador em redes complexas",
            "Estudante de Ciência da Computação",
            "Analista de dados",
            "Professor universitário",
            "Engenheiro de software",
        )
    )

    conn.execute(
        "INSERT INTO users (id, email, password_hash) VALUES (%s, %s, %s)",
        (user_id, email, pwd_hash),
    )
    conn.execute(
        """
        INSERT INTO profiles (user_id, slug, full_name, headline, location)
        VALUES (%s, %s, %s, %s, %s)
        """,
        (user_id, slug, full_name, headline, city.name),
    )
    conn.execute(
        """
        INSERT INTO simulator_agents (
            user_id, archetype, age, gender, city, latitude, longitude,
            extraversion, activity_level, interests, markov_state, rng_offset
        )
        VALUES (%s, %s, %s, %s, %s, %s, %s, 0.1, 0.1, '[]'::jsonb, 'offline', %s)
        """,
        (
            user_id,
            _GRAPH_ARCHETYPE,
            25,
            gender,
            city.name,
            city.latitude,
            city.longitude,
            rng_offset,
        ),
    )
    return user_id


def _bulk_insert_connections(
    conn: psycopg.Connection,
    user_ids: list[UUID],
    edge_pairs: list[tuple[int, int]],
) -> None:
    if not edge_pairs:
        return

    with conn.cursor() as cur:
        with cur.copy(
            "COPY connections (id, requester_id, addressee_id, status) FROM STDIN"
        ) as copy:
            for i, j in edge_pairs:
                a, b = user_ids[i], user_ids[j]
                if a < b:
                    requester, addressee = a, b
                else:
                    requester, addressee = b, a
                copy.write_row((uuid4(), requester, addressee, "accepted"))
