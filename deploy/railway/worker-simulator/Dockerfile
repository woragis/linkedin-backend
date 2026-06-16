# Build context: repo root (linkedin/backend). Canonical layout matches worker/Dockerfile.
FROM python:3.12-slim

WORKDIR /app

ARG WORKER_DEPS=simulator
ARG WORKER_ROLE=simulator

COPY worker/requirements-base.txt worker/requirements-ml.txt worker/requirements-simulator.txt ./

RUN if [ "$WORKER_DEPS" = "ml" ]; then       pip install --no-cache-dir -r requirements-ml.txt;     elif [ "$WORKER_DEPS" = "simulator" ]; then       pip install --no-cache-dir -r requirements-simulator.txt;     else       pip install --no-cache-dir -r requirements-base.txt;     fi

COPY worker/linkedin_worker ./linkedin_worker

ENV PYTHONUNBUFFERED=1 \
    WORKER_HEALTH_ENABLED=1 \
    WORKER_ROLE=${WORKER_ROLE}
ENV SIMULATOR_ENABLED=1 \
    SIMULATOR_AGENT_COUNT=500 \
    SIMULATOR_PHASE=auto \
    SIMULATOR_ENQUEUE_SEARCH=0 \
    SIMULATOR_METRICS_ENABLED=0

CMD ["python", "-m", "linkedin_worker.main"]
