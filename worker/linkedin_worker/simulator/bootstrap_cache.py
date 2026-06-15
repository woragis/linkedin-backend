from __future__ import annotations

from uuid import UUID

import psycopg

from linkedin_worker.simulator.db import ensure_catalog_entity, ensure_skill


class CatalogCache:
    """In-memory cache for institutions, companies, and skills during bootstrap."""

    def __init__(self) -> None:
        self._ids: dict[tuple[str, str], UUID] = {}

    def institution(self, conn: psycopg.Connection, name: str, slug: str) -> UUID:
        return self._entity(conn, "institutions", name, slug)

    def company(self, conn: psycopg.Connection, name: str, slug: str) -> UUID:
        return self._entity(conn, "companies", name, slug)

    def skill(self, conn: psycopg.Connection, name: str, slug: str) -> UUID:
        key = ("skills", slug)
        if key in self._ids:
            return self._ids[key]
        skill_id = ensure_skill(conn, name, slug)
        self._ids[key] = skill_id
        return skill_id

    def _entity(self, conn: psycopg.Connection, table: str, name: str, slug: str) -> UUID:
        key = (table, slug)
        if key in self._ids:
            return self._ids[key]
        entity_id = ensure_catalog_entity(conn, table, name, slug)
        self._ids[key] = entity_id
        return entity_id
