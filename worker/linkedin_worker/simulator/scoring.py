from __future__ import annotations

import math


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
