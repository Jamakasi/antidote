##
## скомпилить antidote из исходников с правками
##
FROM golang:alpine3.15 AS build

RUN apk add --update \
    git \
    make \
    && rm -rf /var/cache/apk/* \
    && mkdir /work \
    && cd /work \
    && git clone https://github.com/Jamakasi/antidote
RUN cd /work/antidote \
    && go build . \
    && chmod +x antidote


##
## взять бинарник и положить в чистый контейнер
##
FROM alpine:latest

RUN apk add --update \
    curl \
    bash \
    && rm -rf /var/cache/apk/* \
    && mkdir /data /app
COPY --from=build /work/antidote/antidote /app/antidote
#RUN chmod +x /app/coredns
#WORKDIR /app
#EXPOSE 53 53/udp

ENTRYPOINT ["/app/antidote", "-config", "/data/antidote.json", "-listen", "0.0.0.0:53"]