FROM golang:1.25.5 AS builder
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN make build

FROM alpine:3.23.2
WORKDIR /app
COPY --from=builder /app/bin/ca-service ./ca-service
EXPOSE 8443
ENTRYPOINT ["./ca-service"]
