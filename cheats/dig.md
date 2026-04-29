---
description: DNS lookup utility — query name servers and inspect DNS records
tags: [dns, network, recon]
---

# dig

## Basic Queries
    dig example.com                       # A record (default)
    dig example.com A                     # explicit A record
    dig example.com AAAA                  # IPv6 record
    dig example.com MX                    # mail exchanger
    dig example.com NS                    # name servers
    dig example.com TXT                   # TXT records (SPF, DKIM, etc.)
    dig example.com SOA                   # Start of Authority
    dig example.com CNAME                 # canonical name
    dig example.com ANY                   # all records (not always supported)

## Reverse DNS
    dig -x 1.2.3.4                        # PTR record for IP

## Specify Name Server
    dig @8.8.8.8 example.com              # query Google's DNS
    dig @1.1.1.1 example.com             # query Cloudflare's DNS
    dig @dc.example.com example.com       # query internal DNS server

## Short & Clean Output
    dig example.com +short                # answer only
    dig example.com A +noall +answer      # answer section only
    dig example.com MX +short | sort -n  # sorted MX records

## Zone Transfer (AXFR)
    dig @ns1.example.com example.com AXFR          # attempt zone transfer
    dig @ns1.example.com example.com AXFR +nostats # cleaner output

## Tracing & Iteration
    dig example.com +trace                # full resolution trace from root
    dig example.com +nssearch             # search all authoritative NS

## DNSSEC
    dig example.com +dnssec               # request DNSSEC records
    dig example.com DNSKEY                # DNSSEC public key
    dig example.com DS                    # delegation signer

## Subdomain Enumeration Helpers
    for s in www mail ftp vpn api admin dev; do
        dig +short $s.example.com
    done

## Common Recon Sequence
    dig example.com NS +short             # find authoritative NS
    dig @ns1.example.com example.com AXFR # attempt zone transfer
    dig example.com MX +short            # mail servers
    dig example.com TXT +short           # SPF / DMARC / verification tokens
    dig -x $(dig example.com +short)     # reverse lookup on resolved IP
