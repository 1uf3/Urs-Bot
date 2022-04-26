FROM golang:latest as builder

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

COPY ./ /bot
WORKDIR /bot
RUN go build -o urs-bot

# --

FROM alpine:3.11
COPY --from=builder /bot/urs-bot /bin

ENTRYPOINT ["/bin/urs-bot"]
