FROM alpine:3.13

COPY guardian /usr/bin/guardian

EXPOSE 8080
CMD ["guardian"]
