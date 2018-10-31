FROM alpine:3.8

RUN apk --no-cache add ca-certificates

COPY ./ci-cleaner /ci-cleaner

ENTRYPOINT ["/ci-cleaner"]
