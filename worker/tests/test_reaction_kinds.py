"""Reaction kinds accepted by the API must match the database constraint."""

from __future__ import annotations

import re
from pathlib import Path

API_REACTION_KINDS = frozenset(
    {"like", "celebrate", "support", "insightful", "love", "funny"}
)

REPO_ROOT = Path(__file__).resolve().parents[2]
MIGRATION = REPO_ROOT / "migrations" / "000006_content_reactions.sql"
SERVICE_GO = REPO_ROOT / "server" / "internal" / "post" / "service" / "service.go"


def _kinds_from_migration(text: str) -> frozenset[str]:
    match = re.search(
        r"content_reactions_kind_valid CHECK \(\s*kind IN \(([^)]+)\)",
        text,
        re.DOTALL,
    )
    assert match, "content_reactions_kind_valid CHECK not found in migration"
    return frozenset(re.findall(r"'(\w+)'", match.group(1)))


def _kinds_from_service(text: str) -> frozenset[str]:
    start = text.find("var validReactionKinds")
    assert start != -1, "validReactionKinds map not found in service.go"
    end = text.find("\n}", start)
    block = text[start:end]
    return frozenset(re.findall(r'"(\w+)"', block))


def test_reaction_kinds_count():
    assert len(API_REACTION_KINDS) == 6


def test_reaction_kinds_match_migration():
    text = MIGRATION.read_text(encoding="utf-8")
    db_kinds = _kinds_from_migration(text)
    assert db_kinds == API_REACTION_KINDS


def test_reaction_kinds_match_go_service():
    text = SERVICE_GO.read_text(encoding="utf-8")
    go_kinds = _kinds_from_service(text)
    assert go_kinds == API_REACTION_KINDS
