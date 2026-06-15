import random

from linkedin_worker.simulator import archetypes, demographics
from linkedin_worker.simulator.scoring import affinity_score, jaccard, sigmoid


def test_pick_archetype_is_valid():
    rng = random.Random(42)
    for _ in range(50):
        name = archetypes.pick_archetype(rng)
        assert name in archetypes.ARCHETYPE_NAMES


def test_sample_traits_bounded():
    rng = random.Random(1)
    extraversion, activity, interests = archetypes.sample_traits(rng, "programmer")
    assert 0.05 <= extraversion <= 0.95
    assert 0.05 <= activity <= 0.95
    assert 2 <= len(interests) <= 4


def test_slug_from_name():
    assert demographics.slug_from_name("Ana Silva") == "ana-silva"
    assert demographics.slug_from_name("  ") == "user"


def test_ensure_unique_slug():
    taken = {"ana-silva", "ana-silva-2"}
    assert demographics.ensure_unique_slug("ana-silva", taken) == "ana-silva-3"


def test_jaccard():
    assert jaccard({"go", "python"}, {"python", "rust"}) == 1 / 3


def test_sigmoid_bounds():
    assert 0.0 < sigmoid(-2) < sigmoid(2) < 1.0


def test_affinity_increases_with_overlap():
    low = affinity_score(interests_a={"go"}, interests_b={"python"})
    high = affinity_score(interests_a={"go", "python"}, interests_b={"go", "python"})
    assert high > low
