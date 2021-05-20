FROM jaegertracing/all-in-one:latest
RUN mkdir /badger && chown -R 1200:1200 /badger

# https://github.com/jaegertracing/jaeger/issues/1543#issuecomment-495777421
ENV BADGER_EPHEMERAL=false
ENV SPAN_STORAGE_TYPE=badger
ENV BADGER_DIRECTORY_VALUE=/badger/data
ENV BADGER_DIRECTORY_KEY=/badger/key
USER 1200