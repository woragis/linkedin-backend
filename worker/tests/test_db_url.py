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


def test_railway_host_adds_sslmode_without_env(monkeypatch):
    monkeypatch.delenv("RAILWAY_ENVIRONMENT", raising=False)
    monkeypatch.delenv("RAILWAY_SERVICE_NAME", raising=False)
    url = normalize_database_url("postgresql://u:p@postgres.railway.internal:5432/db")
    assert "sslmode=require" in url


def test_database_sslmode_override(monkeypatch):
    monkeypatch.setenv("DATABASE_SSLMODE", "prefer")
    url = normalize_database_url("postgresql://u:p@postgres.railway.internal:5432/db")
    assert "sslmode=prefer" in url


def test_no_sslmode_on_localhost(monkeypatch):
    monkeypatch.delenv("RAILWAY_ENVIRONMENT", raising=False)
    monkeypatch.delenv("RAILWAY_SERVICE_NAME", raising=False)
    url = normalize_database_url("postgresql://u:p@localhost:5432/db?sslmode=disable")
    assert url.endswith("sslmode=disable")
