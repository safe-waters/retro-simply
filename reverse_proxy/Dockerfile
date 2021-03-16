FROM caddy:alpine AS builder
WORKDIR /app
COPY . .

FROM caddy:alpine AS dev
COPY --from=builder /app/Caddyfile.dev /etc/caddy/Caddyfile
RUN chown -R 1200:1200 /config /data
USER 1200

FROM caddy:alpine AS prod
COPY --from=builder /app/Caddyfile.prod /etc/caddy/Caddyfile
RUN chown -R 1200:1200 /config /data
USER 1200