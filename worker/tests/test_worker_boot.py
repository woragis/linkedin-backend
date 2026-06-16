import subprocess
import sys
import threading
from pathlib import Path
from unittest.mock import MagicMock, patch

import linkedin_worker.main as worker_main


def test_main_starts_health_before_role_runner(monkeypatch):
    monkeypatch.setattr(worker_main.settings, "WORKER_ROLE", "ml")
    order: list[str] = []

    def fake_health(role: str) -> None:
        order.append(f"health:{role}")

    def fake_ml() -> None:
        order.append("ml")

    with patch.object(worker_main, "start_health_server", side_effect=fake_health):
        with patch.object(worker_main, "run_ml", side_effect=fake_ml):
            worker_main.main()

    assert order == ["health:ml", "ml"]


def test_base_worker_module_loads_without_simulator_deps():
    worker_dir = Path(__file__).resolve().parents[1]
    result = subprocess.run(
        [sys.executable, "-c", "import linkedin_worker.main"],
        cwd=worker_dir,
        capture_output=True,
        text=True,
        timeout=30,
    )
    assert result.returncode == 0, result.stderr


def test_ml_role_uses_connect_db(monkeypatch):
    monkeypatch.setattr(worker_main.settings, "WORKER_ROLE", "ml")
    mock_conn = MagicMock()
    stop = threading.Event()

    def fake_start_ml(conn):
        assert conn is mock_conn
        stop.set()

    with patch.object(worker_main, "connect_db", return_value=mock_conn):
        with patch.object(worker_main.batch_scheduler, "start_ml", side_effect=fake_start_ml):
            with patch.object(worker_main, "start_health_server"):
                worker_main.main()

    assert stop.is_set()
