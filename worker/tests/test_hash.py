from linkedin_worker.experiment_hash import hash_index


def test_hash_index_deterministic():
    uid = "550e8400-e29b-41d4-a716-446655440000"
    assert hash_index(uid) == hash_index(uid)


def test_hash_index_non_negative():
    assert hash_index("") >= 0
    assert hash_index("ana-silva") >= 0


def test_variant_distribution():
    variants = ["chronological", "ranked"]
    counts = {v: 0 for v in variants}
    for i in range(200):
        uid = f"{i:08x}-{i*3:04x}-{i*7:04x}-{i*11:04x}-{i*13:012x}"
        idx = hash_index(uid) % len(variants)
        counts[variants[idx]] += 1
    assert counts["chronological"] > 0
    assert counts["ranked"] > 0
