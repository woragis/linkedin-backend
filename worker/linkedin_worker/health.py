"""Lightweight HTTP health server for Railway / orchestrators."""

from __future__ import annotations

import logging
import threading
from http.server import BaseHTTPRequestHandler, HTTPServer

from linkedin_worker import settings

log = logging.getLogger("linkedin-worker.health")


class _Handler(BaseHTTPRequestHandler):
    role: str = "worker"

    def do_GET(self) -> None:
        if self.path in ("/health", "/"):
            body = f'{{"status":"ok","role":"{self.role}"}}'.encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
            return
        self.send_response(404)
        self.end_headers()

    def log_message(self, _format: str, *_args: object) -> None:
        return


def start_health_server(role: str) -> None:
    if not settings.WORKER_HEALTH_ENABLED:
        return

    port = settings.WORKER_HEALTH_PORT
    handler = type("HealthHandler", (_Handler,), {"role": role})

    def _serve() -> None:
        server = HTTPServer(("0.0.0.0", port), handler)
        log.info("health server listening on :%s role=%s", port, role)
        server.serve_forever()

    thread = threading.Thread(target=_serve, daemon=True, name="health")
    thread.start()
