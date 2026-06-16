from linkedin_worker import settings
from linkedin_worker.queue.routing import (
    GRAPH_JOBS,
    INDEXER_JOBS,
    queue_for_job,
    queues_for_role,
)


def test_queue_for_indexer_jobs():
    assert queue_for_job("search.index_profile") == settings.REDIS_QUEUE_INDEXER
    assert queue_for_job("search.index_post") == settings.REDIS_QUEUE_INDEXER


def test_queue_for_graph_jobs():
    assert queue_for_job("graph.recompute_user") == settings.REDIS_QUEUE_GRAPH


def test_queue_for_realtime_jobs():
    assert queue_for_job("analytics.process_event") == settings.REDIS_QUEUE_REALTIME
    assert queue_for_job("notifications.send") == settings.REDIS_QUEUE_REALTIME


def test_queues_for_role_indexer():
    assert queues_for_role("indexer") == [settings.REDIS_QUEUE_INDEXER]


def test_queues_for_role_all():
    queues = queues_for_role("all")
    assert settings.REDIS_QUEUE_REALTIME in queues
    assert settings.REDIS_QUEUE_INDEXER in queues
    assert settings.REDIS_QUEUE_GRAPH in queues


def test_job_sets_disjoint():
    assert INDEXER_JOBS.isdisjoint(GRAPH_JOBS)
