## Secret Generation Examples

Below examples show different function calls in templater that can be used for creation of secrets.

1. derivePassword function to generate password
```
password: {{ derivePassword 1 "long" (randAscii 10) "user" "example.com" }}
```

2. To generate randonm Ascii or AlphaNum or Numeric or Alpha keys
```
test1: {{ randAlphaNum <len> }}
test2: {{ randAlpha 5 }}
test3: {{ randNumeric 12 }}
test4: {{ randAscii 10 }}
```

3. Generate keys matching regex
```
regexkey: {{ regexGen "abc[x-z]+" 6 }}
```

4. To generate certificate authorities
```
{{- $ca := genCA <commonName> <validity> }}
{{- $ca := genCA "foo-ca" 365 }}
{{- $ca := genCAWithKey "foo-ca" 365 (genPrivateKey "rsa") }}
```

5. Certificate generation(selfsigned and signed)
```
{{- $cert := genSelfSignedCert <cn> <ip_list> <dns_list> <validity> }}
{{- $cert := genSelfSignedCertWithKey "foo.com" (list "10.0.0.1" "10.0.0.2") (list "bar.com" "bat.com") 365 (genPrivateKey "ecdsa") }}
{{- $cert := genSignedCert "foo.com" (list "10.0.0.1" "10.0.0.2") (list "bar.com" "bat.com") 365 $ca }}
{{- $cert := genSignedCert "foo.com" (list "10.0.0.1" "10.0.0.2") (list "bar.com" "bat.com") 365 $ca (genPrivateKey "ed25519")}}
```