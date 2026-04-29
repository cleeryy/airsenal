---
description: Fast web fuzzer for content discovery and parameter fuzzing
tags: [web, fuzzing, recon, pentest]
---

# ffuf

Fuzz keyword: FUZZ (can be placed anywhere in URL, header, or body)

## Directory / Path Discovery
    ffuf -w /usr/share/wordlists/dirb/common.txt -u http://target/FUZZ
    ffuf -w wordlist.txt -u http://target/FUZZ -e .php,.html,.txt
    ffuf -w wordlist.txt -u http://target/FUZZ -recursion -recursion-depth 2

## Virtual Host Discovery
    ffuf -w subdomains.txt -u http://target -H "Host: FUZZ.target.com" -fs 0

## Parameter Fuzzing
    ffuf -w params.txt -u http://target/page?FUZZ=value         # GET param name
    ffuf -w values.txt -u http://target/page?id=FUZZ            # GET param value
    ffuf -w wordlist.txt -u http://target/ -X POST -d "user=FUZZ" -H "Content-Type: application/x-www-form-urlencoded"

## Multiple Fuzz Points
    ffuf -w users.txt:FUZZUSER -w passwords.txt:FUZZPASS \
         -u http://target/login \
         -X POST -d "user=FUZZUSER&pass=FUZZPASS" \
         -H "Content-Type: application/x-www-form-urlencoded"

## Filtering Results
    ffuf -w wordlist.txt -u http://target/FUZZ -fc 403,404       # filter status codes
    ffuf -w wordlist.txt -u http://target/FUZZ -fs 1234          # filter by response size
    ffuf -w wordlist.txt -u http://target/FUZZ -fw 10            # filter by word count
    ffuf -w wordlist.txt -u http://target/FUZZ -fl 42            # filter by line count
    ffuf -w wordlist.txt -u http://target/FUZZ -fr "Not found"   # filter by regex

## Matching Results
    ffuf -w wordlist.txt -u http://target/FUZZ -mc 200,301,302   # match status codes
    ffuf -w wordlist.txt -u http://target/FUZZ -ms 1234          # match size
    ffuf -w wordlist.txt -u http://target/FUZZ -mr "admin"       # match regex

## Proxy & Headers
    ffuf -w wordlist.txt -u http://target/FUZZ -x http://127.0.0.1:8080   # proxy (Burp)
    ffuf -w wordlist.txt -u http://target/FUZZ -H "Cookie: session=abc"
    ffuf -w wordlist.txt -u http://target/FUZZ -b "session=abc123"

## Rate & Throttle
    ffuf -w wordlist.txt -u http://target/FUZZ -t 50             # 50 threads (default)
    ffuf -w wordlist.txt -u http://target/FUZZ -p 0.1            # 0.1s between requests

## Output
    ffuf -w wordlist.txt -u http://target/FUZZ -o results.json -of json
    ffuf -w wordlist.txt -u http://target/FUZZ -o results.html -of html
