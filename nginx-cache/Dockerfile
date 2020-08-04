FROM alpine:3.12

RUN apk add --update --no-cache nginx &&\
    mkdir /run/nginx

COPY entrypoint.sh /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

CMD ["/usr/sbin/nginx", "-g", "daemon off;"]