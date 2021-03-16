FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s"

FROM scratch AS prod
COPY --from=builder /app/server /app/server
USER 1200 
ENTRYPOINT ["/app/server"]
