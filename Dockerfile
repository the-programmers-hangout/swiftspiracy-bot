FROM golang:alpine AS builder
RUN apk --update add ca-certificates git make bash

WORKDIR /app
COPY . ./

RUN go mod tidy
RUN make bot

FROM alpine
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bin/bot /bot

ENTRYPOINT ["/bot"]
