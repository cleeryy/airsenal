---
description: Network exploration tool and security/port scanner
tags: [network, scanning, recon, pentest]
---

# nmap

## Host Discovery
    nmap -sn 192.168.1.0/24              # ping sweep, no port scan
    nmap -Pn <target>                     # skip host discovery
    nmap -PS22,80,443 <target>            # TCP SYN ping on specific ports

## Port Scanning
    nmap <target>                         # top 1000 ports (default)
    nmap -p- <target>                     # all 65535 ports
    nmap -p 22,80,443 <target>            # specific ports
    nmap -p 1-1024 <target>               # port range
    nmap -F <target>                      # fast mode (top 100 ports)

## Scan Techniques
    nmap -sS <target>                     # SYN/stealth scan (default, requires root)
    nmap -sT <target>                     # TCP connect scan (no root needed)
    nmap -sU <target>                     # UDP scan
    nmap -sA <target>                     # ACK scan (firewall mapping)

## Service & Version Detection
    nmap -sV <target>                     # service/version detection
    nmap -sV --version-intensity 9 <target>  # maximum intensity
    nmap -sC <target>                     # run default NSE scripts
    nmap -A <target>                      # OS + version + scripts + traceroute

## OS Detection
    nmap -O <target>                      # OS detection (requires root)
    nmap -O --osscan-guess <target>       # aggressive OS guess

## NSE Scripts
    nmap --script=default <target>        # default scripts
    nmap --script=vuln <target>           # vulnerability checks
    nmap --script=safe <target>           # safe scripts only
    nmap --script=http-enum <target>      # web path enumeration
    nmap --script=smb-vuln-ms17-010 <target>  # EternalBlue check

## Output
    nmap -oN out.txt <target>             # normal output
    nmap -oX out.xml <target>             # XML
    nmap -oG out.gnmap <target>           # grepable
    nmap -oA out <target>                 # all three formats

## Timing & Evasion
    nmap -T0 <target>                     # paranoid (IDS evasion)
    nmap -T3 <target>                     # normal (default)
    nmap -T4 <target>                     # aggressive
    nmap -T5 <target>                     # insane
    nmap -f <target>                      # fragment packets
    nmap -D RND:10 <target>               # decoy scan

## Common One-liners
    nmap -sV -sC -O -p- -T4 <target>                          # full recon
    nmap -sn 192.168.1.0/24 -oG - | grep "Status: Up"         # live hosts
    nmap -p 80,443 --open 192.168.1.0/24                      # web hosts only
