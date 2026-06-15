from __future__ import annotations

import random

TEMPLATES: dict[str, list[str]] = {
    "tech": [
        "Migrando mais um serviço para Go. Latência caiu bastante.",
        "Alguém mais usando Redis como fila além de cache?",
        "Deploy com zero downtime: o que funcionou pra vocês?",
        "Testes de integração salvaram o sprint de novo.",
    ],
    "fitness": [
        "Treino de pernas feito. PR no agachamento!",
        "Dica: hidratação antes das 10h faz diferença.",
        "Corrida matinal na orla — energia o dia inteiro.",
        "Descanso também é treino. Semana de deload.",
    ],
    "career": [
        "Buscando estágio em tecnologia. Aberto a dicas!",
        "Primeira entrevista técnica da semana. Wish me luck.",
        "Networking > cold apply na maioria dos casos.",
        "Compartilhando o que aprendi no último projeto acadêmico.",
    ],
    "business": [
        "Validamos a hipótese com 20 entrevistas esta semana.",
        "Pivot ou persistir? Dúvida clássica de founder.",
        "MRR subiu 12% — foco em retenção fez diferença.",
        "Contratamos o primeiro engenheiro. Marco importante.",
    ],
    "design": [
        "Prototipando fluxo de onboarding no Figma.",
        "Teste de usabilidade revelou 3 fricções óbvias.",
        "Design system atualizado — consistência em escala.",
        "Acessibilidade não é feature extra, é requisito.",
    ],
}


def pick_post_body(rng: random.Random, topic: str) -> str:
    pool = TEMPLATES.get(topic) or TEMPLATES["tech"]
    return rng.choice(pool)
