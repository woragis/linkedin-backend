#!/usr/bin/env python3
"""Validate synthetic network metrics after running worker-simulator.

Usage:
  DATABASE_URL=postgres://... python scripts/simulator_validate.py
  python scripts/simulator_validate.py --plot   # requires matplotlib

Checks: agent count, event rate, degree distribution, homophily (city).
"""

from __future__ import annotations

import argparse
import math
import os
import sys
from collections import Counter

import psycopg

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "postgres://linkedin:linkedin@127.0.0.1:5432/linkedin?sslmode=disable",
)


def connect() -> psycopg.Connection:
    return psycopg.connect(DATABASE_URL)


def count_agents(conn: psycopg.Connection) -> int:
    row = conn.execute("SELECT COUNT(*)::int FROM simulator_agents").fetchone()
    return int(row[0]) if row else 0


def count_events(conn: psycopg.Connection) -> int:
    row = conn.execute("SELECT COUNT(*)::int FROM events").fetchone()
    return int(row[0]) if row else 0


def events_last_hour(conn: psycopg.Connection) -> int:
    row = conn.execute(
        """
        SELECT COUNT(*)::int FROM events
        WHERE created_at >= now() - interval '1 hour'
        """
    ).fetchone()
    return int(row[0]) if row else 0


def degree_distribution(conn: psycopg.Connection) -> list[int]:
    rows = conn.execute(
        """
        WITH edges AS (
            SELECT requester_id AS user_id FROM connections WHERE status = 'accepted'
            UNION ALL
            SELECT addressee_id FROM connections WHERE status = 'accepted'
        ),
        degrees AS (
            SELECT user_id, COUNT(*)::int AS deg
            FROM edges
            GROUP BY user_id
        )
        SELECT deg FROM degrees ORDER BY deg
        """
    ).fetchall()
    return [int(r[0]) for r in rows]


def homophily_city(conn: psycopg.Connection) -> tuple[float, float, int]:
    row = conn.execute(
        """
        WITH sim_pairs AS (
            SELECT
                c.id,
                sa1.city AS city_a,
                sa2.city AS city_b
            FROM connections c
            JOIN simulator_agents sa1 ON sa1.user_id = c.requester_id
            JOIN simulator_agents sa2 ON sa2.user_id = c.addressee_id
            WHERE c.status = 'accepted'
        ),
        cities AS (
            SELECT city, COUNT(*)::float AS cnt
            FROM simulator_agents
            GROUP BY city
        ),
        total_agents AS (
            SELECT COUNT(*)::float AS n FROM simulator_agents
        )
        SELECT
            AVG(CASE WHEN city_a = city_b THEN 1.0 ELSE 0.0 END) AS same_city_rate,
            (
                SELECT SUM((cnt / n) * (cnt / n)) FROM cities, total_agents
            ) AS random_baseline,
            COUNT(*)::int AS pairs
        FROM sim_pairs
        """
    ).fetchone()
    if not row or row[2] == 0:
        return 0.0, 0.0, 0
    return float(row[0]), float(row[1] or 0), int(row[2])


def graph_size(conn: psycopg.Connection) -> tuple[int, int]:
    users = conn.execute("SELECT COUNT(*)::int FROM users").fetchone()[0]
    edges = conn.execute(
        "SELECT COUNT(*)::int FROM connections WHERE status = 'accepted'"
    ).fetchone()[0]
    return int(users), int(edges)


def modularity_estimate(conn: psycopg.Connection) -> float | None:
    try:
        import networkx as nx
        from networkx.algorithms.community import modularity
    except ImportError:
        return None

    rows = conn.execute(
        """
        SELECT requester_id, addressee_id
        FROM connections
        WHERE status = 'accepted'
        """
    ).fetchall()
    if len(rows) < 10:
        return None

    graph = nx.Graph()
    for requester_id, addressee_id in rows:
        graph.add_edge(str(requester_id), str(addressee_id))

    city_rows = conn.execute("SELECT user_id, city FROM simulator_agents").fetchall()
    partition: dict[str, str] = {str(uid): city for uid, city in city_rows}
    groups: dict[str, set[str]] = {}
    for node in graph.nodes:
        label = partition.get(node, "unknown")
        groups.setdefault(label, set()).add(node)

    if len(groups) < 2:
        return None
    return float(modularity(graph, list(groups.values())))


def power_law_alpha(degrees: list[int]) -> float | None:
    positive = [d for d in degrees if d > 0]
    if len(positive) < 20:
        return None
    logs = [math.log(d) for d in positive]
    return -1.0 - (sum(logs) / len(logs))


def maybe_plot(degrees: list[int]) -> None:
    try:
        import matplotlib.pyplot as plt
    except ImportError:
        print("matplotlib not installed; skip --plot")
        return

    hist = Counter(degrees)
    xs = sorted(hist.keys())
    ys = [hist[x] for x in xs]
    plt.figure(figsize=(8, 5))
    plt.loglog(xs, ys, "o", alpha=0.7)
    plt.xlabel("Degree")
    plt.ylabel("Count")
    plt.title("Degree distribution (log-log)")
    plt.grid(True, which="both", alpha=0.3)
    out = "simulator_degree_distribution.png"
    plt.savefig(out, dpi=120)
    print(f"plot saved: {out}")


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate simulator synthetic network")
    parser.add_argument("--plot", action="store_true", help="Save degree distribution plot")
    args = parser.parse_args()

    conn = connect()
    agents = count_agents(conn)
    events = count_events(conn)
    events_h = events_last_hour(conn)
    users, edges = graph_size(conn)
    degrees = degree_distribution(conn)
    same_city, baseline, pairs = homophily_city(conn)
    alpha = power_law_alpha(degrees)
    modularity = modularity_estimate(conn)

    print("=== Simulator validation ===")
    print(f"simulator_agents:     {agents}")
    print(f"users (total):        {users}")
    print(f"accepted connections: {edges}")
    print(f"events (total):       {events}")
    print(f"events (last hour):   {events_h}")
    print(f"events/min (approx):  {events_h / 60:.1f}")

    if degrees:
        print(f"degree mean:          {sum(degrees) / len(degrees):.2f}")
        print(f"degree max:           {max(degrees)}")
        if alpha is not None:
            print(f"power-law alpha est:  {alpha:.2f}  (closer to -2..-3 is typical)")

    if pairs > 0:
        print(f"homophily (city):     {same_city:.3f} vs baseline {baseline:.3f} ({pairs} pairs)")
        lift = same_city - baseline
        print(f"homophily lift:       {lift:+.3f}")

    if modularity is not None:
        print(f"modularity (city):    {modularity:.3f}")

    ok = agents >= 500 and events >= 1000
    print(f"\nacceptance (soft):    {'PASS' if ok else 'NEEDS MORE DATA'}")
    conn.close()

    if args.plot and degrees:
        maybe_plot(degrees)

    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
