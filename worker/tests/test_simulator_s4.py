def test_catalog_cache_reuses_ids():
    from uuid import uuid4

    from linkedin_worker.simulator.bootstrap_cache import CatalogCache

    cache = CatalogCache()
    calls: list[tuple[str, str, str]] = []

    class FakeConn:
        pass

    conn = FakeConn()
    uid = uuid4()

    def fake_ensure(conn, table, name, slug):
        calls.append((table, name, slug))
        return uid

    import linkedin_worker.simulator.bootstrap_cache as mod

    original = mod.ensure_catalog_entity
    mod.ensure_catalog_entity = fake_ensure
    try:
        a = cache.institution(conn, "UNIPe", "unipe")
        b = cache.institution(conn, "UNIPe", "unipe")
        assert a == b
        assert len(calls) == 1
    finally:
        mod.ensure_catalog_entity = original


def test_record_tick_increments_counter():
    from uuid import uuid4

    from linkedin_worker.simulator.agent import Agent
    from linkedin_worker.simulator.metrics import ACTIONS_TOTAL, record_tick

    before = ACTIONS_TOTAL.labels(type="view")._value.get()  # noqa: SLF001
    agent = Agent(
        user_id=uuid4(),
        archetype="student",
        age=22,
        gender="F",
        city="Recife",
        latitude=-8.0,
        longitude=-34.8,
        extraversion=0.5,
        activity_level=0.6,
        interests=["estudos"],
        markov_state="browsing",
    )
    record_tick(2, {"view": 2}, 0.05, [agent])
    after = ACTIONS_TOTAL.labels(type="view")._value.get()  # noqa: SLF001
    assert after >= before + 2
