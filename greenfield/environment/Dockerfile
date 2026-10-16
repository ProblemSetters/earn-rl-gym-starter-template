FROM golang:1.22-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends python3 \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home --uid 10001 --shell /bin/bash agent \
    && mkdir -p /work/candidate /logs/verifier \
    && chown agent:agent /work/candidate /logs/verifier

USER agent
WORKDIR /work/candidate
CMD ["sleep", "infinity"]
