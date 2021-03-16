FROM golang AS builder
WORKDIR /app
COPY go.mod . 
COPY go.sum .
RUN go mod download
COPY . .
RUN go test -v -race ./...

FROM builder AS dev
ARG CMD
RUN go build --race "./cmd/${CMD}"

FROM builder AS prodBuilder
ARG CMD
RUN CGO_ENABLED=0 go build -ldflags="-w -s" "./cmd/${CMD}"

FROM scratch AS prod
ARG CMD
COPY --from=prodBuilder "/app/${CMD}" "/app/${CMD}"
USER 1200