from __future__ import annotations

import random

from linkedin_worker.simulator.agent import Agent
from linkedin_worker.simulator.content import llm
from linkedin_worker.simulator.content.templates import pick_comment_body, pick_post_body


def post_body(rng: random.Random, agent: Agent, topic: str) -> str:
    if llm.llm_enabled():
        generated = llm.generate_post_body(
            llm.PersonaContext(
                full_name=agent.full_name or "Profissional",
                headline=agent.headline or "",
                archetype=agent.archetype,
                topic=topic,
                interests=agent.interests,
            )
        )
        if generated:
            return generated
    return pick_post_body(rng, topic)


def comment_body(
    rng: random.Random,
    agent: Agent,
    topic: str,
    *,
    post_body_text: str = "",
    parent_comment: str | None = None,
) -> str:
    if llm.llm_enabled() and post_body_text:
        generated = llm.generate_comment_body(
            llm.PersonaContext(
                full_name=agent.full_name or "Profissional",
                headline=agent.headline or "",
                archetype=agent.archetype,
                topic=topic,
                interests=agent.interests,
            ),
            post_body=post_body_text,
            parent_comment=parent_comment,
        )
        if generated:
            return generated
    return pick_comment_body(rng, topic)
