from __future__ import annotations

import logging
import threading

from prometheus_client import Counter, Gauge, Histogram, start_http_server

from linkedin_worker import settings

log = logging.getLogger("linkedin-worker.simulator.metrics")

ACTIONS_TOTAL = Counter(
    "simulator_actions_total",
    "Total simulator actions executed",
    ["type"],
)
TICK_DURATION = Histogram(
    "simulator_tick_duration_seconds",
    "Simulator tick wall time",
    buckets=(0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0),
)
AGENTS_TOTAL = Gauge("simulator_agents_total", "Total simulator agents loaded")
AGENTS_ONLINE = Gauge(
    "simulator_agents_online",
    "Agents in browsing or reading Markov state",
)
EVENTS_LAST_TICK = Gauge("simulator_events_last_tick", "Actions committed in the last tick")

_server_started = False
_server_lock = threading.Lock()


def start_metrics_server() -> None:
    global _server_started
    if not settings.SIMULATOR_METRICS_ENABLED:
        return
    with _server_lock:
        if _server_started:
            return
        port = settings.SIMULATOR_METRICS_PORT
        start_http_server(port)
        _server_started = True
        log.info("prometheus metrics listening on :%s/metrics", port)


def record_tick(actions: int, breakdown: dict[str, int], duration_sec: float, agents: list) -> None:
    if not settings.SIMULATOR_METRICS_ENABLED:
        return
    TICK_DURATION.observe(duration_sec)
    EVENTS_LAST_TICK.set(actions)
    AGENTS_TOTAL.set(len(agents))
    online = sum(1 for a in agents if a.markov_state in ("browsing", "reading"))
    AGENTS_ONLINE.set(online)
    for action_type, count in breakdown.items():
        ACTIONS_TOTAL.labels(type=action_type).inc(count)
