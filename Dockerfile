FROM golang:1.14-alpine3.12

LABEL maintainer="Stephan Sedlmeier <stephan2048@gmail.com>"

WORKDIR /go/src/app
COPY . .

RUN cd cmd/docsis-pnm && go build -v

FROM alpine:3.12

RUN mkdir /etc/docsis-pnm

COPY --from=0 /go/src/app/cmd/docsis-pnm/docsis-pnm /usr/local/bin/docsis-pnm
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
COPY docsis-pnm.toml.example /etc/docsis-pnm/docsis-pnm.toml

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]


EXPOSE 8080
