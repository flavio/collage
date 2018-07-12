FROM golang:1.8-alpine

COPY . /go/src/github.com/flavio/collage
WORKDIR /go/src/github.com/flavio/collage
RUN go build


FROM alpine
WORKDIR /app
RUN adduser -h /app -D web
RUN apk --update upgrade && \
    apk add ca-certificates su-exec && \
    rm -rf /var/cache/apk/*
COPY --from=0 /go/src/github.com/flavio/collage/collage /app/
ADD docker/init /init

## Cannot use the --chown option of COPY because it's not supported by
## Docker Hub's automated builds :/
RUN chown -R web:web *
ENTRYPOINT ["/init"]
EXPOSE 5000
