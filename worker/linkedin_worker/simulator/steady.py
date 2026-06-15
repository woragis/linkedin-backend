from __future__ import annotations

import logging
import random
from collections import Counter

import psycopg

from linkedin_worker import settings
from linkedin_worker.simulator import archetypes
from linkedin_worker.simulator.actions.comments import comment_on_post
from linkedin_worker.simulator.actions.connections import accept_connection, request_connection
from linkedin_worker.simulator.actions.events import post_viewed
from linkedin_worker.simulator.actions.posts import create_post
from linkedin_worker.simulator.actions.reactions import like_post
from linkedin_worker.simulator.agent import Agent
from linkedin_worker.simulator.content.templates import pick_post_body
from linkedin_worker.simulator.db import load_agents, pending_requester_for_addressee, sample_recent_posts
from linkedin_worker.simulator.scoring import choose_action, pick_target

log = logging.getLogger("linkedin-worker.simulator.steady")


def run_tick(
    conn: psycopg.Connection,
    agents: list[Agent],
    rng: random.Random,
) -> tuple[int, Counter[str]]:
    if not agents:
        return 0, Counter()

    batch_size = min(settings.SIMULATOR_BATCH_SIZE, len(agents))
    batch = rng.sample(agents, batch_size)
    counts: Counter[str] = Counter()
    posts_cache: dict[UUID, list[tuple[UUID, UUID]]] = {}

    for agent in batch:
        pending = pending_requester_for_addressee(conn, agent.user_id)
        action = choose_action(rng, agent, can_accept=pending is not None)
        topic = archetypes.ARCHETYPES.get(agent.archetype, {}).get("template_topic", "tech")

        if action == "post":
            body = pick_post_body(rng, topic)
            create_post(conn, agent.user_id, body)
            counts["post"] += 1
            continue

        if action == "accept" and pending:
            if accept_connection(conn, agent.user_id, pending):
                counts["accept"] += 1
            continue

        if action == "connect":
            target = pick_target(agent, agents, rng)
            if target and request_connection(conn, agent.user_id, target.user_id):
                if rng.random() < 0.35:
                    accept_connection(conn, target.user_id, agent.user_id)
                    counts["accept"] += 1
                counts["connect"] += 1
            continue

        if agent.user_id not in posts_cache:
            posts_cache[agent.user_id] = sample_recent_posts(conn, agent.user_id)
        posts = posts_cache[agent.user_id]
        if not posts:
            body = pick_post_body(rng, topic)
            create_post(conn, agent.user_id, body)
            counts["post"] += 1
            continue

        post_id, _author_id = rng.choice(posts)

        if action == "like":
            if like_post(conn, agent.user_id, post_id):
                counts["like"] += 1
        elif action == "comment":
            if comment_on_post(conn, agent.user_id, post_id, rng, topic):
                counts["comment"] += 1
        else:
            post_viewed(conn, agent.user_id, post_id)
            counts["view"] += 1

    conn.commit()
    total = sum(counts.values())
    return total, counts
