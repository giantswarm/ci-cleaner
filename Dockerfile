FROM alpine:3.8

COPY ./ci-cleaner /ci-cleaner

ENTRYPOINT ["/ci-cleaner"]
