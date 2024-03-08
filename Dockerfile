FROM golang:1.22-alpine AS BuilStage

ENV CGO_ENABLED 0

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY src/ ./src/

RUN go build -C src/main -o /app/p2p_aggregator

FROM alpine:latest

WORKDIR /app

COPY --from=BuilStage /app/p2p_aggregator ./p2p_aggregator

ENTRYPOINT [ "./p2p_aggregator" ]
