.. _airshipctl_config_use-context:

airshipctl config use-context
-----------------------------

Airshipctl command to switch to a different context

Synopsis
~~~~~~~~


Switch to a different context defined in the airshipctl config file.
This command doesn't change the context for the kubeconfig file.


::

  airshipctl config use-context CONTEXT_NAME [flags]

Examples
~~~~~~~~

::


  Switch to a context named "exampleContext" in airshipctl config file
  # airshipctl config use-context exampleContext


Options
~~~~~~~

::

  -h, --help   help for use-context

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output

SEE ALSO
~~~~~~~~

* :ref:`airshipctl config <airshipctl_config>` 	 - Airshipctl command to manage airshipctl config file

