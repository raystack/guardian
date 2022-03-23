FROM alpine:3.13

RUN apk add --no-cache ca-certificates && update-ca-certificates

RUN curl --output /usr/local/share/ca-certificates/SectigoRSADomainValidationSecureServerCA.crt http://crt.sectigo.com/SectigoRSADomainValidationSecureServerCA.crt

COPY guardian .

EXPOSE 8080
ENTRYPOINT ["./guardian"]
