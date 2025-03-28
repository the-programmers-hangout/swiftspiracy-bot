FROM golang:alpine AS builder
RUN apk --update add ca-certificates
WORKDIR /app
COPY . ./
RUN go mod tidy
ENV DISCORD_BOT_TOKEN=""
ENV GIT_COMMIT=""
ENV BUILD_DATE=""
RUN go build -ldflags="-s -w -X main.CommitHash=$GIT_COMMIT -X main.BuildDate=$BUILD_DATE" -o bin/swiftspiracybot cmd/bot/main.go
ENTRYPOINT ["bin/swiftspiracybot"]
