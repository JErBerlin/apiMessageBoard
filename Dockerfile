FROM golang:1.12 as go-builder

WORKDIR /app

COPY . .
ARG GOPROXY=https://goproxy.io

RUN go mod download \
    && pwd && ls \
    && go install github.com/gobuffalo/packr/packr && \
    CGO_ENABLED=0 packr build -o back_message_board


FROM alpine:latest as prod
COPY --from=go-builder /app/back_message_board /app/back_message_board

WORKDIR /app

ENTRYPOINT ["/app/back_message_board"]
