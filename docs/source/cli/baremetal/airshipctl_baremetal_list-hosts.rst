.. _airshipctl_baremetal_list-hosts:

airshipctl baremetal list-hosts
-------------------------------

Airshipctl command to list bare metal host(s)

Synopsis
~~~~~~~~


List bare metal host(s).

::

  airshipctl baremetal list-hosts [flags]

Examples
~~~~~~~~

::


  	Retrieve list of baremetal hosts, default output option is 'table'
  	# airshipctl baremetal list-hosts
  	# airshipctl baremetal list-hosts --namespace default
  	# airshipctl baremetal list-hosts --namespace default --output table
  	# airshipctl baremetal list-hosts --output yaml


Options
~~~~~~~

::

  -h, --help               help for list-hosts
  -l, --labels string      label(s) to filter desired bare metal host from site manifest documents
  -n, --namespace string   airshipctl phase that contains the desired bare metal host from site manifest document(s)
  -o, --output string      output formats. Supported options are 'table' and 'yaml' (default "table")
      --timeout duration   timeout on bare metal action (default 10m0s)

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output

SEE ALSO
~~~~~~~~

* :ref:`airshipctl baremetal <airshipctl_baremetal>` 	 - Airshipctl command to manage bare metal host(s)

