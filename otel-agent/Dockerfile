FROM otel/opentelemetry-collector:latest
COPY otel-agent-config.yml /etc/otel-agent-config.yml
CMD ["--config=/etc/otel-agent-config.yml"]
USER 1200