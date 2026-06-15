-- Connection requests: requester is always the initiator (drop canonical pair ordering).

ALTER TABLE connections DROP CONSTRAINT IF EXISTS connections_canonical_pair;
ALTER TABLE connections DROP CONSTRAINT IF EXISTS connections_requester_id_addressee_id_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_connections_pair
    ON connections (LEAST(requester_id, addressee_id), GREATEST(requester_id, addressee_id));
