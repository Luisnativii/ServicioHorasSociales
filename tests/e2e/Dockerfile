FROM node:22.23.4-bookworm-slim
WORKDIR /runner
ENV PLAYWRIGHT_BROWSERS_PATH=/opt/playwright
RUN npm install --no-audit --no-fund --save-exact playwright@1.64.1 \
 && npx playwright install --with-deps --only-shell chromium \
 && apt-get update && apt-get install -y --no-install-recommends unzip \
 && rm -rf /var/lib/apt/lists/*
COPY run.cjs .
ENTRYPOINT ["node", "/runner/run.cjs"]
