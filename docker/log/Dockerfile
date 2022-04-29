FROM golang:1.17-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o log cmd/log/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/log .
ENTRYPOINT [ "/app/log" ]