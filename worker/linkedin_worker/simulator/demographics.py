from __future__ import annotations

import random
import re
from dataclasses import dataclass

_INVALID_SLUG_CHARS = re.compile(r"[^a-z0-9]+")

CITIES: tuple[tuple[str, float, float], ...] = (
    ("Recife", -8.0476, -34.8770),
    ("Olinda", -8.0089, -34.8553),
    ("São Paulo", -23.5505, -46.6333),
    ("Rio de Janeiro", -22.9068, -43.1729),
    ("Belo Horizonte", -19.9167, -43.9345),
    ("João Pessoa", -7.1195, -34.8450),
)

GENDER_WEIGHTS: tuple[tuple[str, float], ...] = (
    ("F", 0.48),
    ("M", 0.48),
    ("other", 0.04),
)

FIRST_NAMES: dict[str, tuple[str, ...]] = {
    "F": (
        "Ana", "Beatriz", "Carla", "Daniela", "Elisa", "Fernanda", "Gabriela",
        "Helena", "Isabela", "Juliana", "Larissa", "Mariana", "Natália", "Paula",
    ),
    "M": (
        "Bruno", "Carlos", "Diego", "Eduardo", "Felipe", "Gustavo", "Henrique",
        "Igor", "João", "Lucas", "Marcos", "Pedro", "Rafael", "Thiago",
    ),
    "other": (
        "Alex", "Jordan", "Sam", "Taylor", "Riley", "Casey",
    ),
}

LAST_NAMES: tuple[str, ...] = (
    "Silva", "Santos", "Oliveira", "Souza", "Lima", "Costa", "Pereira",
    "Ferreira", "Almeida", "Ribeiro", "Carvalho", "Gomes", "Martins", "Rocha",
    "Mendes", "Nunes", "Dias", "Barbosa", "Araújo", "Cavalcanti",
)


@dataclass(frozen=True)
class City:
    name: str
    latitude: float
    longitude: float


def pick_gender(rng: random.Random) -> str:
    labels, weights = zip(*GENDER_WEIGHTS)
    return rng.choices(labels, weights=weights, k=1)[0]


def pick_city(rng: random.Random) -> City:
    name, lat, lon = rng.choice(CITIES)
    return City(name=name, latitude=lat, longitude=lon)


def sample_age(rng: random.Random) -> int:
    if rng.random() < 0.35:
        return int(max(18, min(65, rng.gauss(22, 2))))
    return int(max(18, min(65, rng.gauss(34, 8))))


def birth_year_from_age(age: int, reference_year: int = 2026) -> int:
    return reference_year - age


def sample_name(rng: random.Random, gender: str) -> str:
    pool = FIRST_NAMES.get(gender, FIRST_NAMES["other"])
    return f"{rng.choice(pool)} {rng.choice(LAST_NAMES)}"


def slug_from_name(name: str) -> str:
    slug = _INVALID_SLUG_CHARS.sub("-", name.strip().lower())
    slug = slug.strip("-")
    return slug or "user"


def ensure_unique_slug(base: str, taken: set[str]) -> str:
    if base not in taken:
        return base
    for i in range(2, 102):
        candidate = f"{base}-{i}"
        if candidate not in taken:
            return candidate
    raise RuntimeError(f"could not allocate unique slug for {base!r}")
