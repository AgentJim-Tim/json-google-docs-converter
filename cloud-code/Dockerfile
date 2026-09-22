FROM node:22-bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends git bash ca-certificates curl python3 make g++ && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY app.part.* /tmp/
RUN cat /tmp/app.part.* | base64 -d > /tmp/app.tar.gz \
 && tar -xzf /tmp/app.tar.gz -C /tmp \
 && cp -a /tmp/chat-cloud-code/. /app/ \
 && rm -rf /tmp/app.part.* /tmp/app.tar.gz /tmp/chat-cloud-code
RUN npm install --omit=dev
ENV NODE_ENV=production
ENV HOST=0.0.0.0
EXPOSE 3333
CMD ["node","src/server.mjs"]