from __future__ import annotations

import logging
import random
from collections import Counter
from datetime import datetime
from uuid import UUID
from zoneinfo import ZoneInfo

import psycopg

from linkedin_worker import settings
from linkedin_worker.simulator import archetypes
from linkedin_worker.simulator.actions.comments import comment_on_post
from linkedin_worker.simulator.actions.connections import accept_connection, request_connection
from linkedin_worker.simulator.actions.events import post_viewed
from linkedin_worker.simulator.actions.posts import create_post, session_start
from linkedin_worker.simulator.actions.reactions import like_post
from linkedin_worker.simulator.agent import Agent
from linkedin_worker.simulator.content.generator import comment_body, post_body
from linkedin_worker.simulator.db import (
    batch_update_markov_states,
    load_global_recent_posts,
    pending_requester_for_addressee,
)
from linkedin_worker.simulator.markov import step as markov_step
from linkedin_worker.simulator.scoring import pick_target

log = logging.getLogger("linkedin-worker.simulator.steady")


def current_hour() -> int:
    tz = ZoneInfo(settings.SIMULATOR_TIMEZONE)
    return datetime.now(tz).hour


def run_tick(
    conn: psycopg.Connection,
    agents: list[Agent],
    rng: random.Random,
) -> tuple[int, Counter[str]]:
    if not agents:
        return 0, Counter()

    hour = current_hour()
    batch_size = min(settings.SIMULATOR_BATCH_SIZE, len(agents))
    batch = rng.sample(agents, batch_size)
    counts: Counter[str] = Counter()
    markov_updates: list[tuple[str, UUID]] = []
    global_posts = load_global_recent_posts(conn, limit=settings.SIMULATOR_GLOBAL_POSTS_POOL)

    for agent in batch:
        pending = pending_requester_for_addressee(conn, agent.user_id)
        if pending and rng.random() < 0.20 + 0.25 * agent.extraversion:
            if accept_connection(conn, agent.user_id, pending):
                counts["accept"] += 1
                agent.markov_state = "browsing"
                markov_updates.append(("browsing", agent.user_id))
            continue

        result = markov_step(agent, hour, rng)
        if result.session_start:
            session_start(conn, agent.user_id)
            counts["session"] += 1

        agent.markov_state = result.state
        markov_updates.append((result.state, agent.user_id))

        if not result.action:
            continue

        topic = archetypes.ARCHETYPES.get(agent.archetype, {}).get("template_topic", "tech")
        _execute_action(conn, agent, agents, rng, result.action, topic, global_posts, counts)

    batch_update_markov_states(conn, markov_updates)
    conn.commit()
    return sum(counts.values()), counts


def _posts_for_agent(
    global_posts: list[tuple[UUID, UUID]],
    agent_id: UUID,
) -> list[tuple[UUID, UUID]]:
    return [(post_id, author_id) for post_id, author_id in global_posts if author_id != agent_id]


def _load_post_body(conn: psycopg.Connection, post_id: UUID) -> str:
    row = conn.execute(
        "SELECT body FROM posts WHERE id = %s AND deleted_at IS NULL",
        (post_id,),
    ).fetchone()
    return row[0] if row else ""


def _execute_action(
    conn: psycopg.Connection,
    agent: Agent,
    agents: list[Agent],
    rng: random.Random,
    action: str,
    topic: str,
    global_posts: list[tuple[UUID, UUID]],
    counts: Counter[str],
) -> bool:
    if action == "post":
        body = post_body(rng, agent, topic)
        create_post(conn, agent.user_id, body)
        counts["post"] += 1
        return True

    if action == "connect":
        target = pick_target(agent, agents, rng)
        if target and request_connection(conn, agent.user_id, target.user_id):
            if rng.random() < 0.35:
                accept_connection(conn, target.user_id, agent.user_id)
                counts["accept"] += 1
            counts["connect"] += 1
            return True
        return False

    posts = _posts_for_agent(global_posts, agent.user_id)
    if not posts:
        body = post_body(rng, agent, topic)
        create_post(conn, agent.user_id, body)
        counts["post"] += 1
        return True

    post_id, _author_id = rng.choice(posts)

    if action == "like":
        if like_post(conn, agent.user_id, post_id):
            counts["like"] += 1
            return True
    elif action == "comment":
        post_text = _load_post_body(conn, post_id)
        if comment_on_post(conn, agent.user_id, post_id, rng, topic, agent=agent, post_body_text=post_text):
            counts["comment"] += 1
            return True
    elif action == "view":
        post_viewed(conn, agent.user_id, post_id)
        counts["view"] += 1
        return True
    return False
