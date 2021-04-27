# Additional Docker image root certificate authorities
If you require additional certificate authorities for your Docker image:
* Add ASCII PEM encoded .crt files to this directory
  * The files will be copied into your docker image at build time.

To update manually copy the .crt files to /usr/local/share/ca-certificates/ and run sudo update-ca-certificates.