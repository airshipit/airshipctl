.. _airshipctl_document_pull:

airshipctl document pull
------------------------

Airshipctl command to pull manifests from remote git repositories

Synopsis
~~~~~~~~


The remote manifests repositories as well as the target path where
the repositories will be cloned are defined in the airship config file.

By default the airship config file is initialized with the
repository "https://opendev.org/airship/treasuremap" as a source of
manifests and with the manifests target path "$HOME/.airship/default".


::

  airshipctl document pull [flags]

Examples
~~~~~~~~

::


  Pull manifests from remote repos
  # airshipctl document pull
  For the below sample airship config file, it will pull from remote repository where URL mentioned
  to the target location /home/airship with manifests->treasuremap->repositories->airshipctl->checkout
  options branch, commitHash & tag mentioned in manifest section.
  In the URL section, instead of a remote repository location we can also mention already checkout directory,
  in this case we need not use document pull otherwise, any temporary changes will be overwritten.
  >>>>>>Sample Config File<<<<<<<<<
  cat ~/.airship/config
  apiVersion: airshipit.org/v1alpha1
  contexts:
    ephemeral-cluster:
      managementConfiguration: treasuremap_config
      manifest: treasuremap
    target-cluster:
      managementConfiguration: treasuremap_config
      manifest: treasuremap
  currentContext: ephemeral-cluster
  kind: Config
  managementConfiguration:
    treasuremap_config:
      insecure: true
      systemActionRetries: 30
      systemRebootDelay: 30
      type: redfish
  manifests:
    treasuremap:
      inventoryRepositoryName: primary
      metadataPath: manifests/site/eric-test-site/metadata.yaml
      phaseRepositoryName: primary
      repositories:
        airshipctl:
          checkout:
            branch: ""
            commitHash: f4cb1c44e0283c38a8bc1be5b8d71020b5d30dfb
            force: false
            localBranch: false
            tag: ""
          url: https://opendev.org/airship/airshipctl.git
        primary:
          checkout:
            branch: ""
            commitHash: 5556edbd386191de6c1ba90757d640c1c63c6339
            force: false
            localBranch: false
            tag: ""
          url: https://opendev.org/airship/treasuremap.git
      targetPath: /home/airship
  permissions:
    DirectoryPermission: 488
    FilePermission: 416
  >>>>>>>>Sample output of document pull for above configuration<<<<<<<<<
  pkg/document/pull/pull.go:36: Reading current context manifest information from /home/airship/.airship/config
  (currentContext:)
  pkg/document/pull/pull.go:51: Downloading airshipctl repository airshipctl from https://opendev.org/airship/
  airshipctl.git into /home/airship (url: & targetPath:)
  pkg/document/repo/repo.go:141: Attempting to download the repository airshipctl
  pkg/document/repo/repo.go:126: Attempting to clone the repository airshipctl from https://opendev.org/airship/
  airshipctl.git
  pkg/document/repo/repo.go:120: Attempting to open repository airshipctl
  pkg/document/repo/repo.go:110: Attempting to checkout the repository airshipctl from commit hash #####
  pkg/document/pull/pull.go:51: Downloading primary repository treasuremap from https://opendev.org/airship/
  treasuremap.git into /home/airship  (repository name taken from url path last content)
  pkg/document/repo/repo.go:141: Attempting to download the repository treasuremap
  pkg/document/repo/repo.go:126: Attempting to clone the repository treasuremap from /home/airship/treasuremap
  pkg/document/repo/repo.go:120: Attempting to open repository treasuremap
  pkg/document/repo/repo.go:110: Attempting to checkout the repository treasuremap from commit hash #####


Options
~~~~~~~

::

  -h, --help          help for pull
  -n, --no-checkout   no checkout is performed after the clone is complete.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output

SEE ALSO
~~~~~~~~

* :ref:`airshipctl document <airshipctl_document>` 	 - Airshipctl command to manage site manifest documents

