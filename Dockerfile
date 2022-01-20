FROM alpine:3.13

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY guardian .

EXPOSE 8080
ENTRYPOINT ["./guardian"]
