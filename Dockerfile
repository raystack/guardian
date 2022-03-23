FROM alpine:3.13

RUN apk add curl
RUN apk add openssl

RUN apk add --no-cache ca-certificates && update-ca-certificates

RUN curl --output /usr/local/share/ca-certificates/SectigoRSADomainValidationSecureServerCA.crt http://crt.sectigo.com/SectigoRSADomainValidationSecureServerCA.crt

RUN openssl x509 -inform DER -in /usr/local/share/ca-certificates/SectigoRSADomainValidationSecureServerCA.crt -out /usr/local/share/ca-certificates/SectigoRSADomainValidationSecureServerCA.pem -text
RUN update-ca-certificates

COPY guardian .

EXPOSE 8080
ENTRYPOINT ["./guardian"]
