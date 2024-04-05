FROM golang:1.22.1-alpine3.19 AS builder

COPY . bank/auth_service/

WORKDIR bank/auth_service/

#RUN go mod init bank/auth_service
RUN go build -o ./.bin/app ./cmd/app/main.go

FROM alpine:latest

#WORKDIR /auth_service/

COPY --from=0 bank/auth_service/.bin/app .
COPY --from=0 bank/auth_service/config config/
#COPY --from=0 bank/auth_service/.env .env

EXPOSE 82
EXPOSE 83

CMD ["./app"]