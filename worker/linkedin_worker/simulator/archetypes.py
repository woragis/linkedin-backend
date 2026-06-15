from __future__ import annotations

import random
from typing import Any

ARCHETYPES: dict[str, dict[str, Any]] = {
    "programmer": {
        "interests": ["go", "python", "devops", "backend", "apis"],
        "skills": ["Go", "Python", "PostgreSQL", "Docker", "Redis"],
        "headlines": [
            "Software Engineer · Backend",
            "Full Stack Developer",
            "Platform Engineer",
        ],
        "schools": ["UNIPe", "UFPE", "UFPB"],
        "companies": ["Lokra", "TechCorp", "Startup Hub"],
        "titles": ["Backend Developer", "Software Engineer", "DevOps Engineer"],
        "activity_mu": 0.6,
        "active_hours": (20, 2),
        "template_topic": "tech",
    },
    "fitness": {
        "interests": ["saúde", "treino", "nutrição", "corrida", "wellness"],
        "skills": ["Personal Training", "Nutrição", "Yoga", "CrossFit"],
        "headlines": [
            "Personal Trainer",
            "Nutricionista esportiva",
            "Coach de bem-estar",
        ],
        "schools": ["UFPE", "Estácio", "Anhanguera"],
        "companies": ["Academia Fit", "Studio Wellness", "Freelance"],
        "titles": ["Personal Trainer", "Nutricionista", "Coach"],
        "activity_mu": 0.7,
        "active_hours": (6, 9),
        "template_topic": "fitness",
    },
    "student": {
        "interests": ["estudos", "estágio", "tecnologia", "carreira", "projetos"],
        "skills": ["JavaScript", "Git", "SQL", "Comunicação"],
        "headlines": [
            "Estudante de Ciência da Computação",
            "Estagiário em TI",
            "Graduando · UNIPe",
        ],
        "schools": ["UNIPe", "UFPE", "CESAR School"],
        "companies": ["Estágio", "Projeto Acadêmico"],
        "titles": ["Estagiário", "Monitor", "Bolsista"],
        "activity_mu": 0.8,
        "active_hours": (14, 23),
        "template_topic": "career",
    },
    "entrepreneur": {
        "interests": ["negócios", "startups", "vendas", "produto", "growth"],
        "skills": ["Product", "Vendas", "Marketing", "Pitch"],
        "headlines": [
            "Founder · SaaS",
            "Empreendedor digital",
            "Head de Growth",
        ],
        "schools": ["FGV", "Insper", "UNIPe"],
        "companies": ["Minha Startup", "Venture Lab", "Consultoria"],
        "titles": ["CEO", "Co-founder", "Growth Lead"],
        "activity_mu": 0.5,
        "active_hours": (8, 18),
        "template_topic": "business",
    },
    "recruiter": {
        "interests": ["vagas", "rh", "carreira", "talentos", "hunting"],
        "skills": ["Recrutamento", "LinkedIn", "Employer Branding", "RH"],
        "headlines": [
            "Tech Recruiter",
            "Talent Acquisition",
            "RH · Tecnologia",
        ],
        "schools": ["UNIPe", "FGV", "Estácio"],
        "companies": ["RH Tech", "Consultoria RH", "Lokra"],
        "titles": ["Recruiter", "Talent Partner", "Analista de RH"],
        "activity_mu": 0.4,
        "active_hours": (9, 17),
        "template_topic": "career",
    },
    "designer": {
        "interests": ["ux", "figma", "design", "ui", "acessibilidade"],
        "skills": ["Figma", "UX Research", "UI Design", "Prototipagem"],
        "headlines": [
            "UX/UI Designer",
            "Product Designer",
            "Designer de interfaces",
        ],
        "schools": ["CESAR School", "UFPE", "UNIPe"],
        "companies": ["Agência Digital", "Product Studio", "Freelance"],
        "titles": ["UX Designer", "UI Designer", "Product Designer"],
        "activity_mu": 0.55,
        "active_hours": (10, 19),
        "template_topic": "design",
    },
    "data_scientist": {
        "interests": ["estatística", "ml", "python", "dados", "analytics"],
        "skills": ["Python", "Machine Learning", "SQL", "Statistics", "Pandas"],
        "headlines": [
            "Data Scientist",
            "ML Engineer",
            "Analista de Dados",
        ],
        "schools": ["UNIPe", "UFPE", "UFRJ"],
        "companies": ["Lokra", "DataLab", "Consultoria Analytics"],
        "titles": ["Data Scientist", "ML Engineer", "Analytics Engineer"],
        "activity_mu": 0.65,
        "active_hours": (9, 22),
        "template_topic": "tech",
    },
}

ARCHETYPE_NAMES = tuple(ARCHETYPES.keys())


def pick_archetype(rng: random.Random) -> str:
    return rng.choice(ARCHETYPE_NAMES)


def sample_traits(rng: random.Random, archetype: str) -> tuple[float, float, list[str]]:
    spec = ARCHETYPES[archetype]
    activity = _clamp(rng.gauss(spec["activity_mu"], 0.12))
    extraversion = _clamp(rng.gauss(0.5, 0.18))
    interests = list(spec["interests"])
    rng.shuffle(interests)
    count = rng.randint(2, min(4, len(interests)))
    return extraversion, activity, interests[:count]


def profile_fields(rng: random.Random, archetype: str) -> dict[str, Any]:
    spec = ARCHETYPES[archetype]
    return {
        "headline": rng.choice(spec["headlines"]),
        "skills": rng.sample(spec["skills"], k=min(rng.randint(2, 4), len(spec["skills"]))),
        "school": rng.choice(spec["schools"]),
        "company": rng.choice(spec["companies"]),
        "title": rng.choice(spec["titles"]),
        "template_topic": spec["template_topic"],
    }


def _clamp(value: float, low: float = 0.05, high: float = 0.95) -> float:
    return max(low, min(high, value))
