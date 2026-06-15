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


def test_simple_affinity_geo_and_interests():
    from linkedin_worker.simulator.agent import Agent
    from linkedin_worker.simulator.scoring import simple_affinity
    from uuid import uuid4

    a = Agent(
        user_id=uuid4(),
        archetype="programmer",
        age=28,
        gender="M",
        city="Recife",
        latitude=-8.0476,
        longitude=-34.8770,
        extraversion=0.5,
        activity_level=0.6,
        interests=["go", "python"],
    )
    near = Agent(
        user_id=uuid4(),
        archetype="programmer",
        age=29,
        gender="F",
        city="Recife",
        latitude=-8.05,
        longitude=-34.88,
        extraversion=0.5,
        activity_level=0.6,
        interests=["go", "python", "devops"],
    )
    far = Agent(
        user_id=uuid4(),
        archetype="fitness",
        age=40,
        gender="M",
        city="São Paulo",
        latitude=-23.55,
        longitude=-46.63,
        extraversion=0.5,
        activity_level=0.6,
        interests=["treino", "nutrição"],
    )
    assert simple_affinity(a, near) > simple_affinity(a, far)


def test_choose_action_excludes_accept_without_pending():
    from linkedin_worker.simulator.agent import Agent
    from linkedin_worker.simulator.scoring import choose_action
    from uuid import uuid4

    agent = Agent(
        user_id=uuid4(),
        archetype="student",
        age=22,
        gender="F",
        city="Recife",
        latitude=-8.0,
        longitude=-34.8,
        extraversion=0.6,
        activity_level=0.7,
        interests=["estudos"],
    )
    rng = random.Random(0)
    actions = {choose_action(rng, agent, can_accept=False) for _ in range(100)}
    assert "accept" not in actions


def test_is_active_hour_wraparound():
    from linkedin_worker.simulator.archetypes import is_active_hour

    assert is_active_hour("programmer", 21)
    assert is_active_hour("programmer", 1)
    assert not is_active_hour("programmer", 12)
    assert is_active_hour("fitness", 7)
    assert not is_active_hour("fitness", 14)


def test_markov_offline_outside_active_window():
    from uuid import uuid4

    from linkedin_worker.simulator.agent import Agent
    from linkedin_worker.simulator.markov import step

    agent = Agent(
        user_id=uuid4(),
        archetype="programmer",
        age=28,
        gender="M",
        city="Recife",
        latitude=-8.0,
        longitude=-34.8,
        extraversion=0.5,
        activity_level=0.6,
        interests=["go"],
        markov_state="offline",
    )
    rng = random.Random(99)
    # programmer active 20-02; hour 12 should rarely wake
    result = step(agent, 12, rng)
    assert result.state == "offline"
    assert result.action is None


def test_markov_wake_emits_session():
    from uuid import uuid4

    from linkedin_worker.simulator.agent import Agent
    from linkedin_worker.simulator.markov import step

    agent = Agent(
        user_id=uuid4(),
        archetype="programmer",
        age=28,
        gender="M",
        city="Recife",
        latitude=-8.0,
        longitude=-34.8,
        extraversion=0.5,
        activity_level=0.9,
        interests=["go"],
        markov_state="offline",
    )
    rng = random.Random(1)
    result = step(agent, 21, rng)
    assert result.state == "browsing"
    assert result.session_start is True


def test_fitness_post_probability_higher_in_window():
    from linkedin_worker.simulator.archetypes import post_transition_probability

    class Stub:
        archetype = "fitness"
        activity_level = 0.7

    in_window = post_transition_probability(Stub(), 7)
    out_window = post_transition_probability(Stub(), 15)
    assert in_window > out_window


