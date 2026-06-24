import random

from linkedin_worker.simulator.graph_bootstrap import chung_lu_edges, sample_degrees


def test_sample_degrees_bounded():
    rng = random.Random(7)
    degrees = sample_degrees(200, rng, mean=80, min_degree=10, max_degree=300)
    assert len(degrees) == 200
    assert all(10 <= d <= 300 for d in degrees)


def test_chung_lu_produces_edges():
    rng = random.Random(99)
    degrees = [50, 40, 30, 20, 10]
    edges = chung_lu_edges(degrees, rng)
    assert all(0 <= i < j < len(degrees) for i, j in edges)


def test_chung_lu_empty_for_single_node():
    rng = random.Random(1)
    assert chung_lu_edges([100], rng) == []
