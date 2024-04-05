FROM golang:1.22.1-alpine3.19 AS builder

COPY . /auth_service/

WORKDIR /auth_service/

RUN go build -o ./.bin/app ./cmd/app/main.go

FROM alpine:latest

COPY --from=0 /auth_service/.bin/app .
COPY --from=0 /auth_service/config config/

EXPOSE 82
EXPOSE 83

CMD ["./app"]