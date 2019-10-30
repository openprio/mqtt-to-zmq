FROM golang as builder
RUN mkdir /go/src/mqtt-to-websocket
WORKDIR /go/src/mqtt-to-websocket
COPY . .

RUN go get
RUN CGO_ENABLED=0 go build -o /go/bin/mqtt-to-websocket

FROM alpine
COPY --from=builder /go/bin/mqtt-to-websocket /app/mqtt-to-websocket
WORKDIR /app
CMD ["/app/mqtt-to-websocket"]
