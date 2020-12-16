Function: generatesecrets-example
=================================

This function defines a secrets variable catalogue profile that
can be consumed by the generate-secrets function to generate secrets.
Using this example we can build other catalogues to generate passphrases
and certificates.

In the `example` defined passphrases and certificates fields are defined.
Sprig library templater functions and other custom defined functions
will be called to generate the respective passphrases and certificates.

In passphrases catalogue the `generationType` field has to be specified, so that the
passphrase generation happens based on the function. Here is the list of valid
`generationType` functions supported as of now: `randAscii`, `randAlpha`,
`randAlphaNum`, `randNumeric`, `derivePassword`, `regexGen`. Along with the
`generationType` the corresponding fields for that function has to be specified.
Refer to the `example` for the required fields for specific `generationType`.
If no `generationType` or inavlid type is specified an appropriate
error will be thrown and execution fails.

For certificate generation, commonName(`cn`), `validity`, `keyEncoding` are
the valid fields that are to be specified. If `cn` and `validity` are not
specified they take "kubernetes" and "365" days as default values.

The `/replacements` kustomization contains a substitution rule that injects
the variables specified into the generate-secrets function template, which will be
used to generate the respective passphrases and certificates based on the variables.
