FROM node:22-bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends git bash ca-certificates curl python3 make g++ && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY payload.00 payload.01 payload.02 payload.03 payload.04 payload.05 payload.06 /tmp/
RUN cat /tmp/payload.00 /tmp/payload.01 /tmp/payload.02 /tmp/payload.03 /tmp/payload.04 /tmp/payload.05 /tmp/payload.06 \
 | base64 -d > /tmp/app.tar.gz \
 && echo "bddf37b917dc2de99b54473a68d4119a6c5c53a424c8af981898129923ca5743  /tmp/app.tar.gz" | sha256sum -c - \
 && tar -xzf /tmp/app.tar.gz -C /tmp \
 && cp -a /tmp/chat-cloud-code/. /app/ \
 && rm -f /tmp/payload.* /tmp/app.tar.gz \
 && rm -rf /tmp/chat-cloud-code
RUN python3 - <<'PY'
from pathlib import Path
p=Path('/app/src/server.mjs')
s=p.read_text()
old='const runner = TEST_MODE ? new TestRunner(config) : new DockerRunner(config);\nif (!TEST_MODE) await runner.verify();'
new='const MANAGED_CONTAINER = process.env.CLOUDCODE_RUNNER_MODE === "managed-container";\nconst runner = (TEST_MODE || MANAGED_CONTAINER) ? new TestRunner(config) : new DockerRunner(config);\nif (!TEST_MODE) await runner.verify();'
if old not in s: raise SystemExit('runner patch target missing')
s=s.replace(old,new)
old2='runner: TEST_MODE ? "test-local" : "docker-sandbox", runner_image: RUNNER_IMAGE,'
new2='runner: TEST_MODE ? "test-local" : (MANAGED_CONTAINER ? "managed-container" : "docker-sandbox"), runner_image: MANAGED_CONTAINER ? null : RUNNER_IMAGE,'
if old2 not in s: raise SystemExit('status patch target missing')
s=s.replace(old2,new2)
old3='"Runner commands execute in an isolated container with network disabled and only the selected workspace mounted writable.",'
new3='MANAGED_CONTAINER ? "Runner commands execute inside the isolated Railway service container; the Railway service boundary provides the outer isolation instead of nested Docker." : "Runner commands execute in an isolated container with network disabled and only the selected workspace mounted writable.",'
if old3 not in s: raise SystemExit('instructions patch target missing')
s=s.replace(old3,new3)
p.write_text(s)
PY
RUN npm install --omit=dev
ENV NODE_ENV=production
ENV HOST=0.0.0.0
EXPOSE 3333
CMD ["node","src/server.mjs"]