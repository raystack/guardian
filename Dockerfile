FROM alpine:3.13

COPY guardian .

EXPOSE 8080
ENTRYPOINT ["./guardian"]
