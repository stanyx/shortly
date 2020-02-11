FROM golang:alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o main .

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
RUN mkdir -p /app/storage/migrations
RUN mkdir -p /app/config
COPY --from=builder /build/main /app/
COPY ./config /app/config
COPY ./storage/migrations /app/storage/migrations
WORKDIR /app

CMD ["./main"]

