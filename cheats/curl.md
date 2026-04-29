---
description: Transfer data with URLs — the Swiss Army knife of HTTP clients
tags: [http, web, api, transfer]
---

# curl

## Basic Requests
    curl <url>                            # GET request
    curl -X POST <url>                    # POST request
    curl -X PUT <url>                     # PUT request
    curl -X DELETE <url>                  # DELETE request

## Headers
    curl -H "Accept: application/json" <url>
    curl -H "Authorization: Bearer <token>" <url>
    curl -H "Content-Type: application/json" -d '{"key":"val"}' <url>
    curl -I <url>                         # HEAD — print response headers only
    curl -v <url>                         # verbose (show request + response headers)

## Request Body
    curl -d "param=value&other=val" <url>       # form POST
    curl -d @data.json -H "Content-Type: application/json" <url>  # POST from file
    curl -F "file=@upload.png" <url>            # multipart file upload

## Authentication
    curl -u user:pass <url>               # basic auth
    curl -H "Authorization: Bearer <token>" <url>
    curl --cookie "session=abc123" <url>
    curl -c cookies.txt -b cookies.txt <url>    # save & send cookies

## Output & Download
    curl -o file.html <url>               # save to file
    curl -O <url>                         # save with remote filename
    curl -L <url>                         # follow redirects
    curl -s <url>                         # silent (no progress)
    curl -w "%{http_code}" -o /dev/null <url>   # print status code only

## TLS / SSL
    curl -k <url>                         # skip cert verification (insecure)
    curl --cacert ca.pem <url>            # custom CA
    curl --cert client.pem --key key.pem <url>  # mTLS

## Proxy
    curl -x http://proxy:8080 <url>
    curl --socks5 127.0.0.1:9050 <url>   # through Tor/SOCKS5

## Useful Combinations
    curl -sL <url> | jq .                             # pretty-print JSON
    curl -s -o /dev/null -w "%{time_total}\n" <url>  # measure latency
    curl -H "Host: internal.app" http://<ip>/         # virtual host testing
