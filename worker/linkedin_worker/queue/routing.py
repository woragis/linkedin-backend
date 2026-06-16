"""Route outbox job types to dedicated Redis queues per worker service."""

from __future__ import annotations

from linkedin_worker import settings

INDEXER_JOBS = frozenset({"search.index_profile", "search.index_post"})
GRAPH_JOBS = frozenset({"graph.recompute_user"})
REALTIME_JOBS = frozenset({"analytics.process_event", "notifications.send"})

ROLE_JOB_TYPES: dict[str, frozenset[str]] = {
    "realtime": REALTIME_JOBS,
    "indexer": INDEXER_JOBS,
    "graph": GRAPH_JOBS,
}


def queue_for_job(job_type: str) -> str:
    if job_type in INDEXER_JOBS:
        return settings.REDIS_QUEUE_INDEXER
    if job_type in GRAPH_JOBS:
        return settings.REDIS_QUEUE_GRAPH
    return settings.REDIS_QUEUE_REALTIME


def queues_for_role(role: str) -> list[str]:
    if role == "all":
        return [
            settings.REDIS_QUEUE_REALTIME,
            settings.REDIS_QUEUE_INDEXER,
            settings.REDIS_QUEUE_GRAPH,
        ]
    if role in ROLE_JOB_TYPES:
        key = {
            "realtime": settings.REDIS_QUEUE_REALTIME,
            "indexer": settings.REDIS_QUEUE_INDEXER,
            "graph": settings.REDIS_QUEUE_GRAPH,
        }[role]
        return [key]
    return []
