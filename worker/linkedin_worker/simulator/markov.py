from __future__ import annotations

import random

from linkedin_worker.simulator import archetypes

VALID_STATES = frozenset({"offline", "browsing", "reading"})

# Transient states sampled as transitions; mapped to actions immediately.
_ACTION_STATES: dict[str, str] = {
    "posting": "post",
    "liking": "like",
    "commenting": "comment",
    "connecting": "connect",
    "viewing": "view",
}


class MarkovStep:
    __slots__ = ("state", "action", "session_start")

    def __init__(self, state: str, action: str | None = None, *, session_start: bool = False) -> None:
        self.state = state
        self.action = action
        self.session_start = session_start


def step(agent: archetypes.AgentLike, hour: int, rng: random.Random) -> MarkovStep:
    """Advance one Markov step for an agent at the given local hour."""
    state = agent.markov_state if agent.markov_state in VALID_STATES else "offline"
    active = archetypes.is_active_hour(agent.archetype, hour)

    if state == "offline":
        return _from_offline(agent, hour, active, rng)

    if state == "browsing":
        return _from_browsing(agent, hour, rng)

    if state == "reading":
        return _from_reading(agent, rng)

    return MarkovStep("offline")


def _from_offline(agent: archetypes.AgentLike, hour: int, active: bool, rng: random.Random) -> MarkovStep:
    p_wake = archetypes.wake_probability(agent, hour, active)
    if rng.random() >= p_wake:
        return MarkovStep("offline")
    return MarkovStep("browsing", session_start=True)


def _from_browsing(agent: archetypes.AgentLike, hour: int, rng: random.Random) -> MarkovStep:
    p_post = archetypes.post_transition_probability(agent, hour)
    p_connect = archetypes.connect_transition_probability(agent, hour)
    p_read = 0.38
    p_view = 0.14
    p_offline = 0.08
    remainder = max(0.0, 1.0 - (p_post + p_connect + p_read + p_view + p_offline))
    options = [
        ("reading", p_read),
        ("posting", p_post),
        ("connecting", p_connect),
        ("viewing", p_view),
        ("offline", p_offline),
        ("browsing", remainder),
    ]
    return _resolve_transition(options, rng)


def _from_reading(agent: archetypes.AgentLike, rng: random.Random) -> MarkovStep:
    p_like = 0.42
    p_comment = 0.14 * max(0.5, agent.extraversion)
    p_view = 0.22
    p_browse = 0.14
    p_offline = 0.08
    remainder = max(0.0, 1.0 - (p_like + p_comment + p_view + p_browse + p_offline))
    options = [
        ("liking", p_like),
        ("commenting", p_comment),
        ("viewing", p_view),
        ("browsing", p_browse),
        ("offline", p_offline),
        ("reading", remainder),
    ]
    return _resolve_transition(options, rng)


def _resolve_transition(options: list[tuple[str, float]], rng: random.Random) -> MarkovStep:
    labels, weights = zip(*options)
    total = sum(weights)
    if total <= 0:
        return MarkovStep("browsing")
    norm = [w / total for w in weights]
    chosen = rng.choices(labels, weights=norm, k=1)[0]
    action = _ACTION_STATES.get(chosen)
    if action:
        after = "offline" if rng.random() < 0.12 else "browsing"
        return MarkovStep(after, action=action)
    return MarkovStep(chosen)
