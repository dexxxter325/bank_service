FROM golang:1.22.1-alpine3.19 AS builder

COPY . .

WORKDIR /auth_service

RUN go mod download
RUN go build -o ./.bin/app ./cmd/app/main.go

FROM alpine:latest

WORKDIR /auth_service

COPY --from=builder /auth_service/.bin/app .
COPY --from=builder /auth_service/config/local.yml .
COPY --from=builder /auth_service/.env .

EXPOSE 82
EXPOSE 83

CMD ["./app"]