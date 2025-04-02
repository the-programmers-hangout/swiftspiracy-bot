FROM golang:alpine AS builder
RUN apk --update add ca-certificates git make bash
WORKDIR /app
COPY . ./
RUN go mod tidy
RUN make bot
ENTRYPOINT ["bin/bot"]
