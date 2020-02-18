FROM golang:1.12 as go-builder

RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go mod download
RUN go build -o back_message_board ./...
CMD ["/app/back_message_board"]
