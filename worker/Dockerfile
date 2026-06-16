FROM python:3.12-slim

WORKDIR /app

# base | ml | simulator | dev  (default: base — realtime, indexer, graph, batch)
ARG WORKER_DEPS=base
ARG WORKER_ROLE=realtime

COPY requirements-base.txt requirements-ml.txt requirements-simulator.txt ./

RUN if [ "$WORKER_DEPS" = "ml" ]; then \
      pip install --no-cache-dir -r requirements-ml.txt; \
    elif [ "$WORKER_DEPS" = "simulator" ]; then \
      pip install --no-cache-dir -r requirements-simulator.txt; \
    elif [ "$WORKER_DEPS" = "dev" ]; then \
      pip install --no-cache-dir -r requirements.txt; \
    else \
      pip install --no-cache-dir -r requirements-base.txt; \
    fi

COPY linkedin_worker ./linkedin_worker

ENV PYTHONUNBUFFERED=1
ENV WORKER_ROLE=${WORKER_ROLE}
ENV WORKER_HEALTH_ENABLED=1

CMD ["python", "-m", "linkedin_worker.main"]
