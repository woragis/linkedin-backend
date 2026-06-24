from __future__ import annotations

import logging
import random
import signal
import threading
import time

import psycopg

from linkedin_worker import settings
from linkedin_worker.connections import connect_db
from linkedin_worker.simulator.bootstrap import bootstrap_agents
from linkedin_worker.simulator.db import count_simulator_agents, load_agents
from linkedin_worker.simulator.graph_bootstrap import bootstrap_graph, count_graph_lab_agents
from linkedin_worker.simulator.metrics import record_tick, start_metrics_server
from linkedin_worker.simulator.steady import run_tick

log = logging.getLogger("linkedin-worker.simulator")

_stop = threading.Event()


def _handle_signal(signum: int, _frame: object) -> None:
    log.info("shutdown signal=%s", signum)
    _stop.set()


def run_simulator() -> None:
    if not settings.SIMULATOR_ENABLED:
        log.info("simulator disabled (SIMULATOR_ENABLED=0); sleeping")
        threading.Event().wait()
        return

    signal.signal(signal.SIGTERM, _handle_signal)
    signal.signal(signal.SIGINT, _handle_signal)

    # Prometheus on a second port breaks some PaaS health probes — optional.
    try:
        start_metrics_server()
    except OSError as exc:
        log.warning("simulator metrics disabled: %s", exc)

    conn = connect_db()
    conn.autocommit = False

    phase = settings.SIMULATOR_PHASE
    graph_only = settings.SIMULATOR_MODE == "graph_only"
    target = settings.simulator_target_count()
    log.info(
        "simulator starting phase=%s mode=%s agents_target=%s seed=%s",
        phase,
        settings.SIMULATOR_MODE,
        target,
        settings.SIMULATOR_SEED,
    )

    if graph_only:
        if phase in ("bootstrap", "auto"):
            created = bootstrap_graph(conn)
            if created:
                log.info(
                    "graph bootstrap created=%s total_graph_agents=%s",
                    created,
                    count_graph_lab_agents(conn),
                )
            if phase == "bootstrap":
                log.info("graph bootstrap phase complete; exiting")
                conn.close()
                return
        elif phase == "steady":
            total = count_graph_lab_agents(conn)
            if total < target:
                log.warning(
                    "graph steady with insufficient agents have=%s want=%s; run bootstrap first",
                    total,
                    target,
                )
        log.info("graph_only mode: steady social loop disabled; sleeping")
        threading.Event().wait()
        conn.close()
        return

    if phase in ("bootstrap", "auto"):
        created = bootstrap_agents(conn)
        if created:
            log.info("bootstrap created=%s total_agents=%s", created, count_simulator_agents(conn))
        if phase == "bootstrap":
            log.info("bootstrap phase complete; exiting steady loop")
            conn.close()
            return
    elif phase == "steady":
        total = count_simulator_agents(conn)
        if total < settings.SIMULATOR_AGENT_COUNT:
            log.warning(
                "steady phase with insufficient agents have=%s want=%s; run bootstrap first",
                total,
                settings.SIMULATOR_AGENT_COUNT,
            )

    _steady_loop(conn)


def _steady_loop(conn: psycopg.Connection) -> None:
    tick = 0
    rng = random.Random(settings.SIMULATOR_SEED)
    agents = load_agents(conn)
    log.info(
        "steady loop started tick_sec=%s batch_size=%s agents=%s markov=S3 scale=S4",
        settings.SIMULATOR_TICK_SEC,
        settings.SIMULATOR_BATCH_SIZE,
        len(agents),
    )
    while not _stop.is_set():
        tick += 1
        if tick % 60 == 1:
            agents = load_agents(conn)
        started = time.perf_counter()
        actions, breakdown = run_tick(conn, agents, rng)
        elapsed = time.perf_counter() - started
        record_tick(actions, dict(breakdown), elapsed, agents)
        log.info(
            "tick=%s agents=%s actions=%s duration_sec=%.3f breakdown=%s",
            tick,
            len(agents),
            actions,
            elapsed,
            dict(breakdown),
        )
        if _stop.wait(settings.SIMULATOR_TICK_SEC):
            break
    conn.close()
    log.info("simulator stopped")
