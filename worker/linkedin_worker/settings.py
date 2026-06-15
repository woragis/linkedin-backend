import os

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "postgres://linkedin:linkedin@postgres:5432/linkedin?sslmode=disable",
)
REDIS_URL = os.getenv("REDIS_URL", "redis://redis:6379/0")
REDIS_QUEUE_KEY = os.getenv("REDIS_QUEUE_KEY", "linkedin:jobs")

WORKER_ROLE = os.getenv("WORKER_ROLE", "all").strip().lower()
OUTBOX_POLL_INTERVAL_SEC = float(os.getenv("OUTBOX_POLL_INTERVAL_SEC", "2"))
BATCH_ENABLED = os.getenv("BATCH_ENABLED", "1").lower() in ("1", "true", "yes")

ELASTICSEARCH_URL = os.getenv("ELASTICSEARCH_URL", "")

KAFKA_BROKERS = os.getenv("KAFKA_BROKERS", "")
KAFKA_TOPIC = os.getenv("KAFKA_TOPIC", "linkedin.jobs")

# Cron expressions (batch worker)
BATCH_CRON_GRAPH = os.getenv("BATCH_CRON_GRAPH", "0 */6 * * *")
BATCH_CRON_RECOMMENDATIONS = os.getenv("BATCH_CRON_RECOMMENDATIONS", "30 */6 * * *")
BATCH_CRON_FEED_RANKING = os.getenv("BATCH_CRON_FEED_RANKING", "0 * * * *")
BATCH_CRON_CHURN = os.getenv("BATCH_CRON_CHURN", "0 3 * * *")
BATCH_CRON_ANALYTICS = os.getenv("BATCH_CRON_ANALYTICS", "15 * * * *")
BATCH_CRON_ML_TRAINING = os.getenv("BATCH_CRON_ML_TRAINING", "0 4 * * 0")
