---
description: Secure Shell — remote login, tunneling, and key management
tags: [ssh, remote, tunnel, pentest]
---

# ssh

## Basic Connection
    ssh user@host                         # connect
    ssh -p 2222 user@host                 # custom port
    ssh -i ~/.ssh/id_rsa user@host        # specify private key
    ssh -v user@host                      # verbose (debug)

## Key Management
    ssh-keygen -t ed25519 -C "comment"    # generate ED25519 key (preferred)
    ssh-keygen -t rsa -b 4096             # generate RSA 4096 key
    ssh-copy-id user@host                 # install public key on server
    ssh-add ~/.ssh/id_rsa                 # add key to agent
    ssh-add -l                            # list keys in agent

## Tunneling
    ssh -L 8080:localhost:80 user@host    # local forward: localhost:8080 → host:80
    ssh -R 9090:localhost:3000 user@host  # remote forward: host:9090 → local:3000
    ssh -D 1080 user@host                 # dynamic SOCKS5 proxy on port 1080
    ssh -N -L 5432:db.internal:5432 bastion  # DB tunnel through bastion (no shell)

## File Transfer
    scp file.txt user@host:/tmp/          # copy file to remote
    scp user@host:/etc/passwd .           # copy file from remote
    scp -r ./dir user@host:/tmp/          # copy directory recursively
    rsync -avz ./local/ user@host:/remote/   # sync with progress

## SSH Config (~/.ssh/config)
    Host bastion
        HostName 1.2.3.4
        User deploy
        IdentityFile ~/.ssh/deploy_key
        Port 22

    Host internal
        HostName 10.0.0.5
        ProxyJump bastion

    # Then just: ssh internal

## ProxyJump / Bastion
    ssh -J bastion user@internal           # jump through bastion
    ssh -J user@bastion user@target        # one-hop pivot

## Useful Tricks
    ssh user@host 'id; whoami; hostname'   # run command remotely
    ssh user@host "cat /etc/shadow"        # read remote file
    ssh -t user@host 'sudo su -'           # force pseudo-TTY (for sudo)
    ssh user@host tail -f /var/log/syslog  # stream remote log

## Pentest Shortcuts
    ssh -L 80:127.0.0.1:80 user@target    # expose remote service locally
    ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null user@host  # no host check
