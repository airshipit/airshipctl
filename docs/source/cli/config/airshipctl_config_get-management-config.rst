.. _airshipctl_config_get-management-config:

airshipctl config get-management-config
---------------------------------------

Airshipctl command to view management config(s) defined in the airshipctl config

Synopsis
~~~~~~~~


Displays a specific management config information, or all defined management configs if no name is provided.
The information relates to reboot-delays and retry in seconds along with management-type that has to be used.


::

  airshipctl config get-management-config MGMT_CONFIG_NAME [flags]

Examples
~~~~~~~~

::


  View all management configurations
  # airshipctl config get-management-configs

  View a specific management configuration named "default"
  # airshipctl config get-management-config default


Options
~~~~~~~

::

  -h, --help   help for get-management-config

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output

SEE ALSO
~~~~~~~~

* :ref:`airshipctl config <airshipctl_config>` 	 - Airshipctl command to manage airshipctl config file

