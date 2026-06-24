import os

from linkedin_worker.db_url import normalize_database_url

_DATABASE_URL_RAW = os.getenv(
    "DATABASE_URL",
    "postgres://linkedin:linkedin@postgres:5432/linkedin?sslmode=disable",
)
DATABASE_URL = normalize_database_url(_DATABASE_URL_RAW)
REDIS_URL = os.getenv("REDIS_URL", "redis://redis:6379/0")
REDIS_QUEUE_KEY = os.getenv("REDIS_QUEUE_KEY", "linkedin:jobs")
REDIS_QUEUE_REALTIME = os.getenv("REDIS_QUEUE_REALTIME", REDIS_QUEUE_KEY)
REDIS_QUEUE_INDEXER = os.getenv("REDIS_QUEUE_INDEXER", "linkedin:jobs:indexer")
REDIS_QUEUE_GRAPH = os.getenv("REDIS_QUEUE_GRAPH", "linkedin:jobs:graph")

WORKER_ROLE = os.getenv("WORKER_ROLE", "all").strip().lower()
WORKER_HEALTH_ENABLED = os.getenv("WORKER_HEALTH_ENABLED", "1").lower() in ("1", "true", "yes")
WORKER_HEALTH_PORT = int(os.getenv("WORKER_HEALTH_PORT", os.getenv("PORT", "8081")))
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

# Simulator worker (WORKER_ROLE=simulator)
SIMULATOR_ENABLED = os.getenv("SIMULATOR_ENABLED", "1").lower() in ("1", "true", "yes")
SIMULATOR_AGENT_COUNT = int(os.getenv("SIMULATOR_AGENT_COUNT", "2000"))
SIMULATOR_SEED = int(os.getenv("SIMULATOR_SEED", "42"))
SIMULATOR_TICK_SEC = float(os.getenv("SIMULATOR_TICK_SEC", "1"))
SIMULATOR_PHASE = os.getenv("SIMULATOR_PHASE", "auto").strip().lower()
SIMULATOR_BATCH_SIZE = int(os.getenv("SIMULATOR_BATCH_SIZE", "50"))
SIMULATOR_OUTBOX_EVERY = int(os.getenv("SIMULATOR_OUTBOX_EVERY", "100"))
SIMULATOR_TIMEZONE = os.getenv("SIMULATOR_TIMEZONE", "America/Recife")
SIMULATOR_BOOTSTRAP_COMMIT_EVERY = int(os.getenv("SIMULATOR_BOOTSTRAP_COMMIT_EVERY", "100"))
SIMULATOR_ENQUEUE_SEARCH = os.getenv("SIMULATOR_ENQUEUE_SEARCH", "0").lower() in ("1", "true", "yes")
SIMULATOR_METRICS_ENABLED = os.getenv("SIMULATOR_METRICS_ENABLED", "1").lower() in ("1", "true", "yes")
SIMULATOR_METRICS_PORT = int(os.getenv("SIMULATOR_METRICS_PORT", "9100"))
SIMULATOR_GLOBAL_POSTS_POOL = int(os.getenv("SIMULATOR_GLOBAL_POSTS_POOL", "200"))
SIMULATOR_MODE = os.getenv("SIMULATOR_MODE", "full").strip().lower()
SIMULATOR_VOLUME_AGENT_COUNT = int(
    os.getenv("SIMULATOR_VOLUME_AGENT_COUNT", os.getenv("SIMULATOR_AGENT_COUNT", "1000"))
)
SIMULATOR_GRAPH_MEAN_DEGREE = float(os.getenv("SIMULATOR_GRAPH_MEAN_DEGREE", "80"))
SIMULATOR_GRAPH_MIN_DEGREE = int(os.getenv("SIMULATOR_GRAPH_MIN_DEGREE", "10"))
SIMULATOR_GRAPH_MAX_DEGREE = int(os.getenv("SIMULATOR_GRAPH_MAX_DEGREE", "300"))


def simulator_target_count() -> int:
    if SIMULATOR_MODE == "graph_only":
        return SIMULATOR_VOLUME_AGENT_COUNT
    return SIMULATOR_AGENT_COUNT


# LLM (R4 — live simulator + e2e)
SIMULATOR_LLM = os.getenv("SIMULATOR_LLM", "0").lower() in ("1", "true", "yes")
LLM_PROVIDER = os.getenv("LLM_PROVIDER", "openai").strip().lower()
LLM_MODEL = os.getenv("LLM_MODEL", "gpt-4o-mini")
OPENAI_API_KEY = os.getenv("OPENAI_API_KEY", "")
SIMULATOR_LLM_MAX_PER_MIN = int(os.getenv("SIMULATOR_LLM_MAX_PER_MIN", "12"))
SIMULATOR_LLM_MAX_PER_DAY = int(os.getenv("SIMULATOR_LLM_MAX_PER_DAY", "800"))
LLM_TIMEOUT_SEC = float(os.getenv("LLM_TIMEOUT_SEC", "30"))
DATABASE_URL_LIVE = normalize_database_url(os.getenv("DATABASE_URL_LIVE", ""))
