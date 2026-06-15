FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY linkedin_worker ./linkedin_worker

ENV PYTHONUNBUFFERED=1
ENV WORKER_ROLE=all

CMD ["python", "-m", "linkedin_worker.main"]
