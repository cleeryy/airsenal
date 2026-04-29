---
description: LDAP directory query and enumeration tool
tags: [ldap, active-directory, enumeration, pentest]
---

# ldapsearch

## Basic Syntax
    ldapsearch -H ldap://<host> -x -b "dc=example,dc=com" "(objectClass=*)"
    # -H: LDAP URI   -x: simple auth   -b: base DN   filter in quotes

## Anonymous Bind (Unauthenticated)
    ldapsearch -H ldap://dc.example.com -x -b "" -s base "(objectClass=*)"   # root DSE
    ldapsearch -H ldap://dc.example.com -x -b "dc=example,dc=com" "(objectClass=*)"

## Authenticated Bind
    ldapsearch -H ldap://dc.example.com \
               -x -D "cn=user,dc=example,dc=com" -W \
               -b "dc=example,dc=com" "(objectClass=*)"

    # Active Directory (NTLM / Kerberos)
    ldapsearch -H ldap://dc.example.com \
               -x -D "user@example.com" -W \
               -b "dc=example,dc=com" "(objectClass=user)"

## LDAPS (TLS)
    ldapsearch -H ldaps://dc.example.com -x -b "dc=example,dc=com" "(objectClass=*)"
    ldapsearch -H ldap://dc.example.com -Z -x -b "dc=example,dc=com" "(objectClass=*)"  # StartTLS

## Common Filters (Active Directory)
    # All users
    ldapsearch ... "(objectClass=user)"
    # All groups
    ldapsearch ... "(objectClass=group)"
    # Specific user
    ldapsearch ... "(sAMAccountName=jdoe)"
    # Domain admins
    ldapsearch ... "(memberOf=CN=Domain Admins,CN=Users,DC=example,DC=com)"
    # All computers
    ldapsearch ... "(objectClass=computer)"
    # Users with SPN (Kerberoastable)
    ldapsearch ... "(&(objectClass=user)(servicePrincipalName=*))"
    # Users with no pre-auth (ASREProastable)
    ldapsearch ... "(&(objectClass=user)(userAccountControl:1.2.840.113556.1.4.803:=4194304))"
    # Password never expires
    ldapsearch ... "(&(objectClass=user)(userAccountControl:1.2.840.113556.1.4.803:=65536))"

## Selecting Attributes
    ldapsearch ... "(objectClass=user)" sAMAccountName mail displayName
    ldapsearch ... "(objectClass=user)" "*"              # all attributes
    ldapsearch ... "(objectClass=user)" "+"              # operational attributes

## Output Formats
    ldapsearch ... -LLL           # minimal LDIF (suppress comments and version)
    ldapsearch ... -o ldif-wrap=no  # no line wrapping

## Useful Enumeration Sequence
    # 1. Discover naming context
    ldapsearch -H ldap://dc -x -b "" -s base namingContexts

    # 2. Enumerate all users
    ldapsearch -H ldap://dc -x -D "user@domain" -W \
               -b "dc=domain,dc=com" "(objectClass=user)" \
               sAMAccountName mail memberOf -LLL

    # 3. Find kerberoastable accounts
    ldapsearch -H ldap://dc -x -D "user@domain" -W \
               -b "dc=domain,dc=com" \
               "(&(objectClass=user)(servicePrincipalName=*))" \
               sAMAccountName servicePrincipalName -LLL
