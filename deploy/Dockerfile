FROM golang:alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o main .

FROM alpine
RUN adduser -S -D -H -h /app appuser
RUN mkdir -p /app/storage/migrations
RUN mkdir -p /app/config
RUN mkdir -p /app/app/rbac
RUN mkdir -p /app/static
RUN chown -R appuser: /app
USER appuser
COPY --from=builder /build/main /app/
COPY ./config /app/config
COPY ./storage/migrations /app/storage/migrations
COPY ./app/rbac /app/app/rbac
COPY ./static /app/static
WORKDIR /app

CMD ["./main"]

