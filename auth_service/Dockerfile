FROM golang:1.22.1-alpine3.19 AS builder

COPY . /auth_service

WORKDIR /auth_service

RUN go build -o ./.bin/app ./cmd/app/main.go

FROM alpine:latest

WORKDIR /auth_service

COPY --from=builder /auth_service/.bin/app .
COPY ./config/local.yml /auth_service/config/local.yml
COPY ./.env /auth_service/.env

EXPOSE 82
EXPOSE 83

CMD ["./app"]