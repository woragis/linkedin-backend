from unittest.mock import MagicMock, patch

from linkedin_worker.queue import relay


def test_kafka_producer_not_shadowed_by_function_name():
    """Regression: _kafka_producer() used to shadow the module global."""
    mock_producer = MagicMock()
    mock_kafka = MagicMock()
    mock_kafka.KafkaProducer.return_value = mock_producer

    relay._kafka_producer_client = None
    with patch.object(relay.settings, "KAFKA_BROKERS", "broker:9092"):
        with patch.dict("sys.modules", {"kafka": mock_kafka}):
            producer = relay._get_kafka_producer()

    assert producer is mock_producer
    mock_producer.send.assert_not_called()

    relay._kafka_producer_client = None
