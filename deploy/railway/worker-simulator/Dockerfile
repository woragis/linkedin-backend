FROM python:3.12-slim

WORKDIR /app/worker

COPY worker/requirements-base.txt worker/requirements-ml.txt worker/requirements-simulator.txt ./
RUN pip install --no-cache-dir -r requirements-simulator.txt

COPY worker/linkedin_worker ./linkedin_worker

ENV PYTHONUNBUFFERED=1 \
    WORKER_HEALTH_ENABLED=1 \
    WORKER_ROLE=simulator \
    SIMULATOR_ENABLED=1 \
    SIMULATOR_AGENT_COUNT=2000 \
    SIMULATOR_PHASE=auto \
    SIMULATOR_ENQUEUE_SEARCH=0 \
    SIMULATOR_METRICS_ENABLED=1

CMD ["python", "-m", "linkedin_worker.main"]
