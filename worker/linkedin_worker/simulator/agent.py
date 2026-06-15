from __future__ import annotations

from dataclasses import dataclass, field
from uuid import UUID


@dataclass
class Agent:
    user_id: UUID
    archetype: str
    age: int
    gender: str
    city: str
    latitude: float | None
    longitude: float | None
    extraversion: float
    activity_level: float
    interests: list[str] = field(default_factory=list)
    markov_state: str = "offline"
    rng_offset: int = 0
    full_name: str = ""
    slug: str = ""
    headline: str = ""
    location: str = ""
    birth_year: int = 1990
