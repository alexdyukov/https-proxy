# https-proxy
HTTPS proxy with basic auth

## Parameters and their defaults
- ENCODED_HEADER=''

Encoded part of Proxy-Authorization header after Basic. For `Proxy-Authorization: Basic FAKE` it would be `ENCODED_HEADER=FAKE`. There is no point to fully support auth by user/pass

- SSL_CERT='example.crt'

Filesystem's path to .crt file of certificate in x509 format. Dockerized working dir is /proxy

- SSL_KEY='example.key'

Filesystem's path to .key file of certificate in x509 format. Dockerized working dir is /proxy

- LISTEN_ADDRESS=':8080'

Host:Port which service should listen for incoming requests. ':8080' means any address on port 8080

- TIMEOUT='3s'

Timeout for proxy incoming connections. Outgoing requests does not have timeouts and managed by client request. If you drop connection to proxy - proxy automaticaly drop outgoing connection to target host 

## Usage

```
# 1. create selfsigned certificates
openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes -keyout example.key -out example.crt -subj "/CN=localhost"

# 2. run proxy
# 2.1 by go command
ENCODED_HEADER=$(echo -n 'user:pass' | base64) go run main.go &

# 2.2 or in docker
docker run -d --rm --name https-proxy -e ENCODED_HEADER=$(echo -n 'user:pass' | base64) -v ./example.crt:/proxy/example.crt -v ./example.key:/proxy/example.key -p 8080:8080 ghcr.io/alexdyukov/https-proxy

# 3. test proxy
curl -I --proxy-cacert ./example.crt --proxy-basic --proxy-user user:pass -x https://localhost:8080 https://google.com
```
Curl output should be something like this
```
$ curl -I --proxy-cacert ./example.crt --proxy-basic --proxy-user user:pass -x https://localhost:8080 https://google.com
HTTP/1.1 200 OK
Date: Mon, 23 Sep 2024 07:02:43 GMT
Transfer-Encoding: chunked

HTTP/2 301
location: https://www.google.com/
content-type: text/html; charset=UTF-8
content-security-policy-report-only: object-src 'none';base-uri 'self';script-src 'nonce-9RWaw4LDFWC71NVo9PJ5QA' 'strict-dynamic' 'report-sample' 'unsafe-eval' 'unsafe-inline' https: http:;report-uri https://csp.withgoogle.com/csp/gws/other-hp
date: Mon, 23 Sep 2024 07:02:40 GMT
expires: Wed, 23 Oct 2024 07:02:40 GMT
cache-control: public, max-age=2592000
server: gws
content-length: 220
x-xss-protection: 0
x-frame-options: SAMEORIGIN
alt-svc: h3=":443"; ma=2592000,h3-29=":443"; ma=2592000
```

## License

MIT licensed. See the included LICENSE file for details.