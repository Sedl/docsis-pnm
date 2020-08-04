#!/bin/sh

cat <<EOF >/etc/nginx/conf.d/default.conf
proxy_cache_path /tmp/cache keys_zone=cache:10m levels=1:2 inactive=86400s max_size=100m;

server {
	listen 80 default_server;
	listen [::]:80 default_server;

  gzip on;
  gzip_vary on;
  gzip_types text/plain application/json;
  gunzip on;

  access_log /proc/self/fd/1 main;
  error_log /proc/self/fd/2 warn;

  proxy_cache cache;
  proxy_cache_lock on;
  proxy_cache_valid 200 3600s;
  proxy_cache_use_stale updating;
  add_header X-Cache-Status \$upstream_cache_status;

  location / {
    proxy_http_version 1.1; # Always upgrade to HTTP/1.1
    proxy_set_header Connection ""; # Enable keepalives
    proxy_set_header Accept-Encoding ""; # Optimize encoding
    proxy_pass ${BACKEND};
	}
}
EOF

exec "$@"