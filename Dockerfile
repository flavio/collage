FROM golang:1.8-alpine

COPY . /go/src/github.com/flavio/collage
WORKDIR /go/src/github.com/flavio/collage
RUN go build


FROM alpine
WORKDIR /app
RUN adduser -h /app -D web
COPY --from=0 /go/src/github.com/flavio/collage/collage /app/

## Cannot use the --chown option of COPY because it's not supported by
## Docker Hub's automated builds :/
RUN chown -R web:web *
USER web
ENTRYPOINT ["./collage"]
EXPOSE 5000
