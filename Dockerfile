FROM golang:alpine AS build

RUN apk add --no-cache -U make git

RUN export CGO_ENABLED=0 &&\
    make &&\
    cp cmd/registry/registry /
RUN cp etc/registry.sample.conf /registry.conf

FROM golang:alpine

RUN apk add -U git

COPY --from=build /registry /registry
COPY --from=build /registry.conf /etc/registry.conf

VOLUME /go

EXPOSE 8000

ENTRYPOINT ["/registry -config /etc/registry.conf"]
CMD []
