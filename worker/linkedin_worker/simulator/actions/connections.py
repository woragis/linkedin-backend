from __future__ import annotations

from uuid import UUID

import psycopg

from linkedin_worker.simulator.db import insert_event


def request_connection(
    conn: psycopg.Connection,
    requester_id: UUID,
    addressee_id: UUID,
) -> UUID | None:
    if requester_id == addressee_id:
        return None
    row = conn.execute(
        """
        INSERT INTO connections (requester_id, addressee_id, status)
        SELECT %s, %s, 'pending'
        WHERE NOT EXISTS (
            SELECT 1 FROM connections c
            WHERE LEAST(c.requester_id, c.addressee_id) = LEAST(%s::uuid, %s::uuid)
              AND GREATEST(c.requester_id, c.addressee_id) = GREATEST(%s::uuid, %s::uuid)
        )
        RETURNING id
        """,
        (requester_id, addressee_id, requester_id, addressee_id, requester_id, addressee_id),
    ).fetchone()
    if not row:
        return None
    connection_id = row[0]
    insert_event(
        conn,
        requester_id,
        "connection_requested",
        {"connection_id": str(connection_id), "target_user_id": str(addressee_id)},
    )
    return connection_id


def accept_connection(
    conn: psycopg.Connection,
    addressee_id: UUID,
    requester_id: UUID,
) -> UUID | None:
    row = conn.execute(
        """
        UPDATE connections
        SET status = 'accepted', updated_at = now()
        WHERE addressee_id = %s
          AND requester_id = %s
          AND status = 'pending'
        RETURNING id
        """,
        (addressee_id, requester_id),
    ).fetchone()
    if not row:
        return None
    connection_id = row[0]
    insert_event(
        conn,
        addressee_id,
        "connection_accepted",
        {"connection_id": str(connection_id), "requester_user_id": str(requester_id)},
    )
    return connection_id
