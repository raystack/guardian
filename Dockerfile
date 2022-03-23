FROM alpine:3.13

RUN apk add --no-cache ca-certificates && update-ca-certificates

# Add intermediate certs to connect to OpenIDM
RUN curl --output /usr/local/share/ca-certificates/temp.crt http://crt.sectigo.com/SectigoRSADomainValidationSecure
ServerCA.crt

COPY guardian .

EXPOSE 8080
ENTRYPOINT ["./guardian"]
