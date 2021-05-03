.. _airshipctl_baremetal_poweroff:

airshipctl baremetal poweroff
-----------------------------

Airshipctl command to shutdown bare metal host(s)

Synopsis
~~~~~~~~


Power off bare metal host(s). The command will target bare metal hosts from airship site inventory based on the
--name, --namespace and --labels flags provided. If no flags are provided, airshipctl will select all bare metal hosts in the site
inventory.


::

  airshipctl baremetal poweroff [flags]

Examples
~~~~~~~~

::


  Perform poweroff action against hosts with name rdm9r3s3 in all namespaces where the host is found
  # airshipctl baremetal poweroff --name rdm9r3s3

  Perform poweroff action against hosts with name rdm9r3s3 in metal3 namespace
  # airshipctl baremetal poweroff --name rdm9r3s3 --namespace metal3

  Perform poweroff action against all hosts defined in inventory
  # airshipctl baremetal poweroff --all

  Perform poweroff action against hosts with a label 'foo=bar'
  # airshipctl baremetal poweroff --labels "foo=bar"


Options
~~~~~~~

::

      --all                specify this to target all hosts in the site inventory
  -h, --help               help for poweroff
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

