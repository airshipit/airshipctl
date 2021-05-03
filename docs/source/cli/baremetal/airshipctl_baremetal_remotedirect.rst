.. _airshipctl_baremetal_remotedirect:

airshipctl baremetal remotedirect
---------------------------------

Airshipctl command to bootstrap the ephemeral host

Synopsis
~~~~~~~~


Bootstrap bare metal host. It targets bare metal host from airship inventory based
on the --iso-url, --name, --namespace, --label and --timeout flags provided.


::

  airshipctl baremetal remotedirect [flags]

Examples
~~~~~~~~

::


  Perform bootstrap action against hosts with name rdm9r3s3 in all namespaces where the host is found
  # airshipctl baremetal remotedirect --name rdm9r3s3

  Perform bootstrap action against hosts with name rdm9r3s3 in metal3 namespace
  # airshipctl baremetal remotedirect --name rdm9r3s3 --namespace metal3

  Perform bootstrap action against hosts with a label 'foo=bar'
  # airshipctl baremetal remotedirect --labels "foo=bar"


Options
~~~~~~~

::

  -h, --help               help for remotedirect
      --iso-url string     specify iso url for host to boot from
  -l, --labels string      label(s) to filter desired bare metal host from site manifest documents
      --name string        name to filter desired bare metal host from site manifest document
  -n, --namespace string   airshipctl phase that contains the desired bare metal host from site manifest document(s)
      --timeout duration   timeout on bare metal action (default 10m0s)

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output

SEE ALSO
~~~~~~~~

* :ref:`airshipctl baremetal <airshipctl_baremetal>` 	 - Airshipctl command to manage bare metal host(s)

