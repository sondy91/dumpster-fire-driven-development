FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy dependency files
COPY requirements.txt pyproject.toml uv.lock ./

# Install Python dependencies
RUN pip3 install uv && \
    uv venv .venv && \
    uv pip install -r requirements.txt

# Create non-root user
RUN addgroup --system app && \
    adduser --system --ingroup app --home /home/app app

# Copy application code
COPY ./src/ ./

# Set ownership
RUN chown -R app:app /app

# Switch to non-root user
USER app

# Configure environment
ENV PATH="/app/.venv/bin:${PATH}"
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
ENV REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000", "--app-dir", "."]
