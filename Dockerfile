FROM golang:1.11-alpine

COPY . /go/src/github.com/flavio/collage
WORKDIR /go/src/github.com/flavio/collage
RUN go build


FROM alpine
WORKDIR /app
RUN adduser -h /app -D web
RUN apk --update upgrade && \
    apk add ca-certificates sudo && \
    rm -rf /var/cache/apk/*
# Allow web user to update system certificates
RUN echo "web ALL=(ALL) NOPASSWD: /usr/sbin/update-ca-certificates" >> /etc/sudoers

COPY --from=0 /go/src/github.com/flavio/collage/collage /app/
ADD docker/init /init

## Cannot use the --chown option of COPY because it's not supported by
## Docker Hub's automated builds :/
RUN chown -R web:web *
ENTRYPOINT ["/init"]
EXPOSE 5000
USER web
