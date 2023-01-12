FROM golang:1.19-alpine3.17

LABEL maintainer="Stephan Sedlmeier <stephan2048@gmail.com>"

WORKDIR /go/src/app
COPY . .

RUN cd cmd/docsis-pnm && go build -v

FROM alpine:3.17

RUN mkdir /etc/docsis-pnm

COPY --from=0 /go/src/app/cmd/docsis-pnm/docsis-pnm /usr/local/bin/docsis-pnm
COPY docsis-pnm.toml.example /etc/docsis-pnm/docsis-pnm.toml

ENTRYPOINT ["/usr/local/bin/docsis-pnm"]
CMD ["--config=/etc/docsis-pnm/docsis-pnm.toml", "run"]

EXPOSE 8080
