FROM golang:1.15-stretch as base
WORKDIR /build/
COPY . .
RUN ["make"]

FROM alpine:latest
WORKDIR /opt/guardian
COPY --from=base /build/guardian /opt/guardian/bin/guardian
RUN ["apk", "update"]
EXPOSE 8080

# glibc compatibility library, since go binaries 
# don't work well with musl libc that alpine uses
RUN ["apk", "add", "libc6-compat"] 
ENTRYPOINT ["/opt/guardian/bin/guardian"]