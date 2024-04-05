FROM golang:1.22.1-alpine3.19 AS builder

COPY . /credit_service

WORKDIR /credit_service

RUN go mod download
RUN go build -o ./.bin/app ./cmd/app/main.go

FROM alpine:latest

WORKDIR /credit_service

COPY --from=builder /credit_service/.bin/app .
COPY ./config/local.yml /credit_service/config/

EXPOSE 81

CMD ["./app"]


FROM golang:1.22.1-alpine3.19 AS builder

COPY . /auth_service


WORKDIR /auth_service

#RUN go mod init bank/auth_service
RUN go build -o ./.bin/app ./cmd/app/main.go

FROM alpine:latest

WORKDIR /auth_service

COPY --from=builder /auth_service/.bin/app .
COPY --from=builder /auth_service/config/local.yml /auth_service/config/local.yml
COPY --from=builder /auth_service/.env /auth_service/.env

EXPOSE 82
EXPOSE 83

CMD ["./app"]