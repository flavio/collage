FROM golang:1.11-alpine

COPY . /go/src/github.com/flavio/collage
WORKDIR /go/src/github.com/flavio/collage
RUN go build


FROM alpine
RUN adduser -h /app -D web

COPY --from=0 /go/src/github.com/flavio/collage/collage /app/

## Cannot use the --chown option of COPY because it's not supported by
## Docker Hub's automated builds :/
WORKDIR /app
RUN chown -R web:web *

ENTRYPOINT ["/app/collage"]
EXPOSE 5000
USER web
