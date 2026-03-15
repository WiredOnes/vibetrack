FROM golang:1.25.1-alpine AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN go generate .
RUN CGO_ENABLED=0 go build -o /cub ./cub/cmd/cub
RUN chmod +x /cub

FROM alpine:3.21
COPY --from=builder /cub ./cub
ENTRYPOINT ["./cub"]