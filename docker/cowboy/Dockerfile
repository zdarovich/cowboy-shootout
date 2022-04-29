FROM golang:1.17-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o cowboy cmd/cowboy/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/cowboy .
ENTRYPOINT [ "/app/cowboy" ]