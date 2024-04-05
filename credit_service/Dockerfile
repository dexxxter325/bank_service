FROM golang:1.22.1-alpine3.19 AS builder

COPY . /credit_service/

WORKDIR /credit_service/

RUN go build -o ./.bin/app ./cmd/app/main.go

FROM alpine:latest

COPY --from=0 /credit_service/.bin/app .
COPY --from=0 /credit_service/config config/

EXPOSE 81

CMD ["./app"]
