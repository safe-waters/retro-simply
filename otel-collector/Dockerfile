FROM otel/opentelemetry-collector:latest
COPY otel-collector-config.yml /etc/otel-collector-config.yml
CMD ["--config=/etc/otel-collector-config.yml"]
USER 1200