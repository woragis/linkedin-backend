from unittest.mock import MagicMock, patch

import linkedin_worker.connections as connections


def test_connect_db_retries_then_succeeds():
    mock_conn = MagicMock()
    side_effects = [OSError("refused"), mock_conn]

    with patch.object(connections.psycopg, "connect", side_effect=side_effects) as connect:
        with patch.object(connections.time, "sleep"):
            conn = connections.connect_db()

    assert conn is mock_conn
    mock_conn.execute.assert_called_once_with("SELECT 1")
    assert connect.call_count == 2


def test_connect_redis_pings():
    mock_client = MagicMock()
    with patch.object(connections.redis, "from_url", return_value=mock_client):
        client = connections.connect_redis()
    assert client is mock_client
    mock_client.ping.assert_called_once()
