FROM node:22-bookworm

ENV NODE_PATH=/usr/local/lib/node_modules
ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright

RUN npm install --prefix /opt/taskboard-deps --no-audit --no-fund \
      react@18.3.1 react-dom@18.3.1 express@5.3.0 vite@8.3.4 @vitejs/plugin-react@6.1.2 \
    && npm install --global playwright@1.52.0 \
    && playwright install --with-deps chromium \
    && apt-get update \
    && apt-get install -y --no-install-recommends python3 \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home --uid 10001 --shell /bin/bash agent \
    && mkdir -p /work/candidate /logs/verifier \
    && chown agent:agent /work/candidate /logs/verifier

USER agent
WORKDIR /work/candidate
CMD ["sleep", "infinity"]
