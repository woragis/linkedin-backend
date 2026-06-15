from __future__ import annotations

import math
import random
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from linkedin_worker.simulator.agent import Agent


def sigmoid(x: float) -> float:
    return 1.0 / (1.0 + math.exp(-x))


def jaccard(a: set[str], b: set[str]) -> float:
    if not a and not b:
        return 0.0
    union = a | b
    if not union:
        return 0.0
    return len(a & b) / len(union)


def geo_score(distance_km: float, sigma: float = 120.0) -> float:
    return math.exp(-(distance_km ** 2) / (2 * sigma ** 2))


def age_score(age_delta: int, sigma: float = 8.0) -> float:
    return math.exp(-(age_delta ** 2) / (2 * sigma ** 2))


def affinity_score(
    *,
    interests_a: set[str],
    interests_b: set[str],
    mutual_connections: int = 0,
    degree_a: int = 1,
    distance_km: float = 500.0,
    age_delta: int = 10,
    popularity_b: int = 0,
) -> float:
    """Weighted affinity between two agents (calibrável via pesos)."""
    w_interests = 0.30
    w_mutual = 0.25
    w_geo = 0.20
    w_age = 0.15
    w_pop = 0.10

    mutual_norm = mutual_connections / max(1, degree_a)
    pop_term = math.log1p(popularity_b)

    return (
        w_interests * jaccard(interests_a, interests_b)
        + w_mutual * mutual_norm
        + w_geo * geo_score(distance_km)
        + w_age * age_score(age_delta)
        + w_pop * pop_term
    )


def haversine_km(lat1: float, lon1: float, lat2: float, lon2: float) -> float:
    radius = 6371.0
    phi1, phi2 = math.radians(lat1), math.radians(lat2)
    dphi = math.radians(lat2 - lat1)
    dlambda = math.radians(lon2 - lon1)
    a = math.sin(dphi / 2) ** 2 + math.cos(phi1) * math.cos(phi2) * math.sin(dlambda / 2) ** 2
    return 2 * radius * math.asin(math.sqrt(a))


def agent_distance_km(a: Agent, b: Agent, default_km: float = 500.0) -> float:
    if a.latitude is None or a.longitude is None or b.latitude is None or b.longitude is None:
        return default_km
    return haversine_km(a.latitude, a.longitude, b.latitude, b.longitude)


def simple_affinity(a: Agent, b: Agent) -> float:
    """S2: interests + geo only."""
    dist = agent_distance_km(a, b)
    return 0.6 * jaccard(set(a.interests), set(b.interests)) + 0.4 * geo_score(dist)


def pick_target(agent: Agent, candidates: list[Agent], rng: random.Random) -> Agent | None:
    pool = [c for c in candidates if c.user_id != agent.user_id]
    if not pool:
        return None
    ranked = sorted(pool, key=lambda c: simple_affinity(agent, c), reverse=True)
    top = ranked[: min(20, len(ranked))]
    weights = [max(simple_affinity(agent, c), 0.01) for c in top]
    return rng.choices(top, weights=weights, k=1)[0]


ACTION_WEIGHTS: dict[str, float] = {
    "post": 0.15,
    "like": 0.30,
    "comment": 0.12,
    "view": 0.28,
    "connect": 0.10,
    "accept": 0.05,
}


def choose_action(rng: random.Random, agent: Agent, *, can_accept: bool) -> str:
    weights = dict(ACTION_WEIGHTS)
    if not can_accept:
        weights.pop("accept", None)
    scale = max(0.2, agent.activity_level)
    weighted: dict[str, float] = {}
    for key, base in weights.items():
        value = base * scale
        if key == "comment":
            value *= max(0.5, agent.extraversion)
        if key == "connect":
            value *= max(0.5, agent.extraversion)
        weighted[key] = value
    labels = list(weighted.keys())
    probs = [weighted[k] for k in labels]
    total = sum(probs)
    return rng.choices(labels, weights=[p / total for p in probs], k=1)[0]
