from __future__ import annotations

import functools
import json
import logging
import random
import re
from uuid import uuid4

import bcrypt
import psycopg

from linkedin_worker import settings
from linkedin_worker.simulator import archetypes, demographics
from linkedin_worker.simulator.actions.posts import create_bootstrap_post, session_start
from linkedin_worker.simulator.db import (
    PASSWORD_PLAIN,
    count_simulator_agents,
    ensure_catalog_entity,
    ensure_skill,
    enqueue_outbox,
    insert_event,
    load_existing_slugs,
)

log = logging.getLogger("linkedin-worker.simulator.bootstrap")

_INVALID_SLUG_CHARS = re.compile(r"[^a-z0-9]+")
_BCRYPT_COST = 12


@functools.lru_cache(maxsize=1)
def password_hash() -> str:
    digest = bcrypt.hashpw(PASSWORD_PLAIN.encode(), bcrypt.gensalt(rounds=_BCRYPT_COST))
    return digest.decode()


def _slugify(value: str) -> str:
    slug = _INVALID_SLUG_CHARS.sub("-", value.strip().lower())
    slug = slug.strip("-")
    return slug or "entity"


def bootstrap_agents(conn: psycopg.Connection) -> int:
    existing = count_simulator_agents(conn)
    target = settings.SIMULATOR_AGENT_COUNT
    if existing >= target:
        log.info("bootstrap skipped existing=%s target=%s", existing, target)
        return 0

    remaining = target - existing
    rng = random.Random(settings.SIMULATOR_SEED + existing)
    taken_slugs = load_existing_slugs(conn)
    pwd_hash = password_hash()
    created = 0
    writes_since_outbox = 0

    log.info("bootstrap starting remaining=%s target=%s", remaining, target)

    for index in range(remaining):
        _create_agent(conn, rng, taken_slugs, pwd_hash, existing + index)
        created += 1
        writes_since_outbox += 1

        if writes_since_outbox >= settings.SIMULATOR_OUTBOX_EVERY:
            conn.commit()
            writes_since_outbox = 0
        elif created % 100 == 0:
            conn.commit()
            log.info("bootstrap progress created=%s/%s", created, remaining)

    conn.commit()
    log.info("bootstrap complete created=%s total=%s", created, count_simulator_agents(conn))
    return created


def _create_agent(
    conn: psycopg.Connection,
    rng: random.Random,
    taken_slugs: set[str],
    pwd_hash: str,
    rng_offset: int,
) -> None:
    archetype = archetypes.pick_archetype(rng)
    gender = demographics.pick_gender(rng)
    city = demographics.pick_city(rng)
    age = demographics.sample_age(rng)
    birth_year = demographics.birth_year_from_age(age)
    extraversion, activity_level, interests = archetypes.sample_traits(rng, archetype)
    profile = archetypes.profile_fields(rng, archetype)

    user_id = uuid4()
    full_name = demographics.sample_name(rng, gender)
    base_slug = demographics.slug_from_name(full_name)
    slug = demographics.ensure_unique_slug(base_slug, taken_slugs)
    taken_slugs.add(slug)

    email = f"sim-{user_id}@sim.local"
    location = city.name
    headline = profile["headline"]

    conn.execute(
        """
        INSERT INTO users (id, email, password_hash)
        VALUES (%s, %s, %s)
        """,
        (user_id, email, pwd_hash),
    )
    conn.execute(
        """
        INSERT INTO profiles (user_id, slug, full_name, headline, location, birth_year)
        VALUES (%s, %s, %s, %s, %s, %s)
        """,
        (user_id, slug, full_name, headline, location, birth_year),
    )
    conn.execute(
        """
        INSERT INTO simulator_agents (
            user_id, archetype, age, gender, city, latitude, longitude,
            extraversion, activity_level, interests, markov_state, rng_offset
        )
        VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s::jsonb, 'offline', %s)
        """,
        (
            user_id,
            archetype,
            age,
            gender,
            city.name,
            city.latitude,
            city.longitude,
            extraversion,
            activity_level,
            json.dumps(interests),
            rng_offset,
        ),
    )

    school_slug = _slugify(profile["school"])
    institution_id = ensure_catalog_entity(conn, "institutions", profile["school"], school_slug)
    conn.execute(
        """
        INSERT INTO educations (user_id, institution_id, field_of_study, degree, start_year, end_year)
        VALUES (%s, %s, %s, %s, %s, %s)
        """,
        (user_id, institution_id, "Ciência da Computação", "Bacharelado", birth_year + 18, birth_year + 22),
    )

    company_slug = _slugify(profile["company"])
    company_id = ensure_catalog_entity(conn, "companies", profile["company"], company_slug)
    conn.execute(
        """
        INSERT INTO experiences (user_id, company_id, title, start_year, is_current)
        VALUES (%s, %s, %s, %s, true)
        """,
        (user_id, company_id, profile["title"], birth_year + 23),
    )

    for skill_name in profile["skills"]:
        skill_slug = _slugify(skill_name)
        skill_id = ensure_skill(conn, skill_name, skill_slug)
        conn.execute(
            """
            INSERT INTO user_skills (user_id, skill_id)
            VALUES (%s, %s)
            ON CONFLICT DO NOTHING
            """,
            (user_id, skill_id),
        )

    session_start(conn, user_id)
    post_id = create_bootstrap_post(conn, user_id, rng, profile["template_topic"])
    enqueue_outbox(conn, "search.index_profile", {"user_id": str(user_id)})
    enqueue_outbox(conn, "search.index_post", {"post_id": str(post_id)})
    insert_event(conn, user_id, "profile_created", {"slug": slug, "source": "simulator"})
