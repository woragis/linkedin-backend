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
from linkedin_worker.simulator.content.templates import pick_post_body
from linkedin_worker.simulator.db import (
    load_agents,
    pending_requester_for_addressee,
    sample_recent_posts,
    update_markov_state,
)
from linkedin_worker.simulator.markov import MarkovStep, step as markov_step
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
    posts_cache: dict[UUID, list[tuple[UUID, UUID]]] = {}

    for agent in batch:
        pending = pending_requester_for_addressee(conn, agent.user_id)
        if pending and rng.random() < 0.20 + 0.25 * agent.extraversion:
            if accept_connection(conn, agent.user_id, pending):
                counts["accept"] += 1
                update_markov_state(conn, agent.user_id, "browsing")
                agent.markov_state = "browsing"
            continue

        result = markov_step(agent, hour, rng)
        if result.session_start:
            session_start(conn, agent.user_id)
            counts["session"] += 1

        agent.markov_state = result.state
        update_markov_state(conn, agent.user_id, result.state)

        if not result.action:
            continue

        topic = archetypes.ARCHETYPES.get(agent.archetype, {}).get("template_topic", "tech")
        if _execute_action(conn, agent, agents, rng, result.action, topic, posts_cache, counts):
            pass

    conn.commit()
    return sum(counts.values()), counts


def _execute_action(
    conn: psycopg.Connection,
    agent: Agent,
    agents: list[Agent],
    rng: random.Random,
    action: str,
    topic: str,
    posts_cache: dict[UUID, list[tuple[UUID, UUID]]],
    counts: Counter[str],
) -> bool:
    if action == "post":
        body = pick_post_body(rng, topic)
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

    if agent.user_id not in posts_cache:
        posts_cache[agent.user_id] = sample_recent_posts(conn, agent.user_id)
    posts = posts_cache[agent.user_id]
    if not posts:
        body = pick_post_body(rng, topic)
        create_post(conn, agent.user_id, body)
        counts["post"] += 1
        return True

    post_id, _author_id = rng.choice(posts)

    if action == "like":
        if like_post(conn, agent.user_id, post_id):
            counts["like"] += 1
            return True
    elif action == "comment":
        if comment_on_post(conn, agent.user_id, post_id, rng, topic):
            counts["comment"] += 1
            return True
    elif action == "view":
        post_viewed(conn, agent.user_id, post_id)
        counts["view"] += 1
        return True
    return False
