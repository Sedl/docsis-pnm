FROM golang:1.17-alpine3.14

LABEL maintainer="Stephan Sedlmeier <stephan2048@gmail.com>"

WORKDIR /go/src/app
COPY . .

RUN cd cmd/docsis-pnm && go build -v

FROM alpine:3.14

RUN mkdir /etc/docsis-pnm

COPY --from=0 /go/src/app/cmd/docsis-pnm/docsis-pnm /usr/local/bin/docsis-pnm
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
COPY docsis-pnm.toml.example /etc/docsis-pnm/docsis-pnm.toml

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]


EXPOSE 8080
