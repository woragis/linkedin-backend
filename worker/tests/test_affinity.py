from linkedin_worker.affinity import compute_affinity_score


def test_zero_affinity():
    score, reasons = compute_affinity_score()
    assert score == 0.0
    assert reasons == []


def test_mutual_connections():
    score, reasons = compute_affinity_score(mutual_connections=3)
    assert score > 0
    assert "3 mutual connections" in reasons


def test_full_overlap():
    score, reasons = compute_affinity_score(
        mutual_connections=5,
        same_school=True,
        shared_skills=3,
        same_company=True,
        graduation_cohort=True,
    )
    assert score >= 0.79
    assert len(reasons) == 5


def test_shared_skills_capped():
    low, _ = compute_affinity_score(shared_skills=3)
    high, _ = compute_affinity_score(shared_skills=10)
    assert low == high


def test_threshold_filter():
    score, _ = compute_affinity_score(same_company=True)
    assert score > 0.05
