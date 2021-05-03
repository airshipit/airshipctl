.. _airshipctl_config_get-manifest:

airshipctl config get-manifest
------------------------------

Airshipctl command to get a specific or all manifest(s) information from the airshipctl config

Synopsis
~~~~~~~~


Displays a specific manifest information, or all defined manifests if no name is provided. The information
includes the repository details related to site manifest along with the local targetPath for them.


::

  airshipctl config get-manifest MANIFEST_NAME [flags]

Examples
~~~~~~~~

::


  List all the manifests
  # airshipctl config get-manifests

  Display a specific manifest
  # airshipctl config get-manifest e2e


Options
~~~~~~~

::

  -h, --help   help for get-manifest

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output

SEE ALSO
~~~~~~~~

* :ref:`airshipctl config <airshipctl_config>` 	 - Airshipctl command to manage airshipctl config file

