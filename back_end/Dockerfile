# Build a small static binary first so the runtime image can stay minimal.
FROM golang:1.25-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/visual-agent ./cmd/visual-agent

# The runtime image only needs CA certificates plus the compiled binary and a
# writable directory for persisted graph JSON files.
FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /app/data/graphs

WORKDIR /app

COPY --from=build /out/visual-agent /app/visual-agent

ENV VISUAL_AGENT_SERVER_ADDR=0.0.0.0:8080

VOLUME ["/app/data/graphs"]

EXPOSE 8080

CMD ["/app/visual-agent", "serve"]
