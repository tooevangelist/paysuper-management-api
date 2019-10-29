FROM alpine:3.10

WORKDIR /app

RUN apk --no-cache add tzdata ca-certificates mailcap \
    && rm -rf /tmp/* \
    && rm -rf /var/cache/apk/*

COPY ./build/docker/app/files /
COPY ./api api
COPY ./assets assets
COPY ./configs configs
COPY ./bin bin

RUN chmod +x /docker/bin/* \
    && chmod +x bin

ENTRYPOINT ["/docker/bin/entrypoint.sh"]
