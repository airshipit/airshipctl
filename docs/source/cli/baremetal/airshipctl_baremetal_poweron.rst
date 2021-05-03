.. _airshipctl_baremetal_poweron:

airshipctl baremetal poweron
----------------------------

Airshipctl command to power on host(s)

Synopsis
~~~~~~~~


Power on bare metal host(s). The command will target bare metal hosts from airship site inventory based on the
--name, --namespace and --labels flags provided. If no flags are provided, airshipctl will select all bare metal hosts in the site
inventory.


::

  airshipctl baremetal poweron [flags]

Examples
~~~~~~~~

::


  Perform poweron action against hosts with name rdm9r3s3 in all namespaces where the host is found
  # airshipctl baremetal poweron --name rdm9r3s3

  Perform poweron action against hosts with name rdm9r3s3 in metal3 namespace
  # airshipctl baremetal poweron --name rdm9r3s3 --namespace metal3

  Perform poweron action against all hosts defined in inventory
  # airshipctl baremetal poweron --all

  Perform poweron action against hosts with a label 'foo=bar'
  # airshipctl baremetal poweron --labels "foo=bar"


Options
~~~~~~~

::

      --all                specify this to target all hosts in the site inventory
  -h, --help               help for poweron
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

