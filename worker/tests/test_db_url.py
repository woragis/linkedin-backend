import os

from linkedin_worker.db_url import normalize_database_url


def test_postgres_scheme_normalized():
    url = "postgres://u:p@host:5432/db"
    assert normalize_database_url(url) == "postgresql://u:p@host:5432/db"


def test_railway_adds_sslmode(monkeypatch):
    monkeypatch.setenv("RAILWAY_ENVIRONMENT", "production")
    url = normalize_database_url("postgresql://u:p@host:5432/db")
    assert "sslmode=require" in url


def test_existing_sslmode_preserved(monkeypatch):
    monkeypatch.setenv("RAILWAY_ENVIRONMENT", "production")
    url = normalize_database_url("postgresql://u:p@host:5432/db?sslmode=disable")
    assert "sslmode=disable" in url
    assert url.count("sslmode=") == 1


def test_no_sslmode_off_railway(monkeypatch):
    monkeypatch.delenv("RAILWAY_ENVIRONMENT", raising=False)
    url = normalize_database_url("postgresql://u:p@host:5432/db?sslmode=disable")
    assert url.endswith("sslmode=disable")
