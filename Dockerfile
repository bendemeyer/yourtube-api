FROM golang:1.26-alpine

ENV PORT="8080"
ENV SOCKET=""
ENV PGDSN=""

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o yourtube ./cmd

CMD ./yourtube -port "$PORT" -socket "$SOCKET" -dsn "$PGDSN"