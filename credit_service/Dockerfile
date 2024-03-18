FROM golang:1.22.1-alpine3.19 AS builder

COPY . /credit_service

WORKDIR /credit_service

RUN go mod download
RUN go build -o ./.bin/app ./cmd/app/main.go

FROM alpine:latest

WORKDIR /credit_service

COPY --from=builder /credit_service/.bin/app .
COPY ./config/local.yml /credit_service/config/

EXPOSE 80

CMD ["./app"]