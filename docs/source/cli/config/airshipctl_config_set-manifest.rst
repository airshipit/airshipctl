.. _airshipctl_config_set-manifest:

airshipctl config set-manifest
------------------------------

Airshipctl command to create/modify manifests in airship config

Synopsis
~~~~~~~~


Creates or modifies a manifests in the airshipctl config file based on the MANIFEST_NAME argument passed.
The optional flags that can be passed to the command are repo name, url, branch name, tag name, commit hash,
target-path and metadata-path. Use --force flag to enable force checkout of the repo. And use --phase flag
to enable phase repository. For any new site deployment, or testing of any new function or composite, this
config file will not have any customization, respective changes need to be done in the manifest files only.


::

  airshipctl config set-manifest MANIFEST_NAME [flags]

Examples
~~~~~~~~

::


  Create a new manifest
  # airshipctl config set-manifest exampleManifest --repo exampleRepo --url https://github.com/site \
    --branch master --phase --target-path exampleTargetpath

  Change the phase repo for manifest
  # airshipctl config set-manifest e2e --repo exampleRepo --phase

  Change the target-path for manifest
  # airshipctl config set-manifest e2e --target-path /tmp/e2e


Options
~~~~~~~

::

      --branch string          the branch to be associated with repository in this manifest
      --commithash string      the commit hash to be associated with repository in this manifest
      --force                  if set, enable force checkout in repository with this manifest
  -h, --help                   help for set-manifest
      --metadata-path string   the metadata path to be set for this manifest
      --phase                  if set, enable this repository as phase repository to be used with this manifest
      --repo string            the name of the repository to be associated with this manifest
      --tag string             the tag to be associated with repository in this manifest
      --target-path string     the target path to be set for this manifest
      --url string             the repository url to be associated with this manifest

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output

SEE ALSO
~~~~~~~~

* :ref:`airshipctl config <airshipctl_config>` 	 - Airshipctl command to manage airshipctl config file

