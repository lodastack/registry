FROM golang:1.17 AS build

COPY . /src/project
WORKDIR /src/project

RUN export CGO_ENABLED=0 &&\
    export GOPROXY=https://goproxy.io &&\
    make &&\
    cp cmd/registry/registry /registry &&\
    cp etc/registry.sample.conf /registry.conf

FROM debian:10
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=build /registry /registry
COPY --from=build /registry.conf /etc/registry.conf

EXPOSE 8000

CMD ["/registry", "-config", "/etc/registry.conf"]
