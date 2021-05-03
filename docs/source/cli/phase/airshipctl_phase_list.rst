.. _airshipctl_phase_list:

airshipctl phase list
---------------------

Airshipctl command to list phases

Synopsis
~~~~~~~~


List phases defined in site manifests by plan. Phases within a plan are
executed sequentially. Multiple phase plans are executed in parallel.


::

  airshipctl phase list PHASE_NAME [flags]

Examples
~~~~~~~~

::


  List phases of phasePlan
  # airshipctl phase list --plan phasePlan

  To output the contents in table format (default operation)
  # airshipctl phase list --plan phasePlan -o table

  To output the contents in yaml format
  # airshipctl phase list --plan phasePlan -o yaml

  List all phases
  # airshipctl phase list

  List phases with clustername
  # airshipctl phase list --cluster-name clustername


Options
~~~~~~~

::

  -c, --cluster-name string   filter documents by cluster name
  -h, --help                  help for list
  -o, --output string         output format. Supported formats are 'table' and 'yaml' (default "table")
      --plan string           plan name of a plan

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output

SEE ALSO
~~~~~~~~

* :ref:`airshipctl phase <airshipctl_phase>` 	 - Airshipctl command to manage phases

