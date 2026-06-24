from __future__ import annotations

import json
import logging
import time
import urllib.error
import urllib.request
from dataclasses import dataclass

from linkedin_worker import settings

log = logging.getLogger("linkedin-worker.simulator.llm")

_TEMPLATE_MARKERS = (
    "Migrando mais um serviço para Go",
    "Alguém mais usando Redis como fila",
    "Treino de pernas feito",
    "Buscando estágio em tecnologia",
    "Validamos a hipótese com 20 entrevistas",
    "Prototipando fluxo de onboarding",
    "Concordo — já vi isso em produção",
)


@dataclass
class PersonaContext:
    full_name: str
    headline: str
    archetype: str
    topic: str
    interests: list[str]


class LLMRateLimiter:
    def __init__(self, max_per_min: int) -> None:
        self._max = max(1, max_per_min)
        self._window: list[float] = []

    def allow(self) -> bool:
        now = time.monotonic()
        self._window = [t for t in self._window if now - t < 60.0]
        if len(self._window) >= self._max:
            return False
        self._window.append(now)
        return True


_limiter = LLMRateLimiter(settings.SIMULATOR_LLM_MAX_PER_MIN)


def llm_enabled() -> bool:
    return settings.SIMULATOR_LLM and bool(settings.OPENAI_API_KEY.strip())


def is_template_text(text: str) -> bool:
    body = text.strip()
    return any(marker in body for marker in _TEMPLATE_MARKERS)


def _chat(prompt: str, *, max_tokens: int = 180) -> str | None:
    if not llm_enabled():
        return None
    if not _limiter.allow():
        log.warning("llm rate limit reached")
        return None

    payload = {
        "model": settings.LLM_MODEL,
        "messages": [
            {
                "role": "system",
                "content": (
                    "Você escreve posts e comentários curtos para uma rede social "
                    "profissional brasileira (estilo LinkedIn). Responda só com o texto "
                    "final em português do Brasil, sem aspas, sem markdown, máximo 280 caracteres."
                ),
            },
            {"role": "user", "content": prompt},
        ],
        "max_tokens": max_tokens,
        "temperature": 0.85,
    }
    req = urllib.request.Request(
        "https://api.openai.com/v1/chat/completions",
        data=json.dumps(payload).encode(),
        headers={
            "Authorization": f"Bearer {settings.OPENAI_API_KEY}",
            "Content-Type": "application/json",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=settings.LLM_TIMEOUT_SEC) as res:
            data = json.loads(res.read().decode())
        text = data["choices"][0]["message"]["content"].strip()
        if not text or is_template_text(text):
            return None
        return text[:500]
    except (urllib.error.URLError, KeyError, json.JSONDecodeError, IndexError) as exc:
        log.warning("llm request failed: %s", exc)
        return None


def generate_post_body(persona: PersonaContext) -> str | None:
    interests = ", ".join(persona.interests[:4]) or persona.topic
    prompt = (
        f"Escreva um post autêntico para {persona.full_name} ({persona.headline}). "
        f"Arquétipo: {persona.archetype}. Interesses: {interests}. "
        f"Tema: {persona.topic}. Tom natural, 1-3 frases."
    )
    return _chat(prompt)


def generate_comment_body(
    persona: PersonaContext,
    *,
    post_body: str,
    parent_comment: str | None = None,
) -> str | None:
    if parent_comment:
        prompt = (
            f"{persona.full_name} ({persona.headline}) responde em thread. "
            f"Post original: {post_body[:200]}. Comentário pai: {parent_comment[:200]}. "
            "Escreva uma resposta curta e cordial."
        )
    else:
        prompt = (
            f"{persona.full_name} ({persona.headline}) comenta um post. "
            f"Conteúdo do post: {post_body[:240]}. Comentário curto e relevante."
        )
    return _chat(prompt, max_tokens=120)
