"""Graph metrics: PageRank, centrality, communities."""

from __future__ import annotations

import logging
from collections import defaultdict

import psycopg

log = logging.getLogger("linkedin-worker.graph")


def _pagerank(adj: dict[str, set[str]], damping: float = 0.85, iters: int = 30) -> dict[str, float]:
    nodes = list(adj.keys())
    if not nodes:
        return {}
    n = len(nodes)
    rank = {node: 1.0 / n for node in nodes}
    for _ in range(iters):
        new_rank = {node: (1 - damping) / n for node in nodes}
        for node in nodes:
            neighbors = adj[node]
            if not neighbors:
                share = rank[node] / n
                for target in nodes:
                    new_rank[target] += damping * share
            else:
                share = rank[node] / len(neighbors)
                for nb in neighbors:
                    new_rank[nb] += damping * share
        rank = new_rank
    return rank


def _communities(adj: dict[str, set[str]]) -> dict[str, int]:
    visited: set[str] = set()
    communities: dict[str, int] = {}
    cid = 0
    for start in adj:
        if start in visited:
            continue
        stack = [start]
        while stack:
            node = stack.pop()
            if node in visited:
                continue
            visited.add(node)
            communities[node] = cid
            stack.extend(adj[node] - visited)
        cid += 1
    return communities


def run_batch(conn: psycopg.Connection) -> None:
    log.info("graph batch started")
    rows = conn.execute(
        "SELECT requester_id::text, addressee_id::text FROM connections WHERE status = 'accepted'"
    ).fetchall()

    adj: dict[str, set[str]] = defaultdict(set)
    for a, b in rows:
        adj[a].add(b)
        adj[b].add(a)

  # include isolated users with profiles
    profile_rows = conn.execute("SELECT user_id::text FROM profiles").fetchall()
    for (uid,) in profile_rows:
        adj.setdefault(uid, set())

    pr = _pagerank(adj)
    communities = _communities(adj)

    for uid, neighbors in adj.items():
        degree = len(neighbors)
        pagerank = pr.get(uid, 0.0)
        community_id = communities.get(uid, 0)
        conn.execute(
            """
            INSERT INTO user_graph_metrics
              (user_id, pagerank, degree, in_degree, out_degree, community_id, computed_at)
            VALUES (%s::uuid, %s, %s, %s, %s, %s, now())
            ON CONFLICT (user_id) DO UPDATE SET
              pagerank = EXCLUDED.pagerank,
              degree = EXCLUDED.degree,
              in_degree = EXCLUDED.in_degree,
              out_degree = EXCLUDED.out_degree,
              community_id = EXCLUDED.community_id,
              computed_at = now()
            """,
            (uid, pagerank, degree, degree, degree, community_id),
        )
    conn.commit()
    log.info("graph batch finished nodes=%d", len(adj))
