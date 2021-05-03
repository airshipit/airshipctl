.. _airshipctl_phase_render:

airshipctl phase render
-----------------------

Airshipctl command to render phase documents from model

Synopsis
~~~~~~~~


Render documents for a phase.


::

  airshipctl phase render PHASE_NAME [flags]

Examples
~~~~~~~~

::


  Get all 'initinfra' phase documents containing labels "app=helm" and "service=tiller"
  # airshipctl phase render initinfra -l app=helm,service=tiller

  Get all phase documents containing labels "app=helm" and "service=tiller" and kind 'Deployment'
  # airshipctl phase render initinfra -l app=helm,service=tiller -k Deployment

  Get all documents from config bundle
  # airshipctl phase render --source config

  Get all documents executor rendered documents for a phase
  # airshipctl phase render initinfra --source executor


Options
~~~~~~~

::

  -a, --annotation string   filter documents by Annotations
  -g, --apiversion string   filter documents by API version
  -d, --decrypt             ensure that decryption of encrypted documents has finished successfully
  -h, --help                help for render
  -k, --kind string         filter documents by Kind
  -l, --label string        filter documents by Labels
  -s, --source string       phase: phase entrypoint will be rendered by kustomize, if entrypoint is not specified error will be returned
                            executor: rendering will be performed by executor if the phase
                            config: this will render bundle containing phase and executor documents (default "phase")

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output

SEE ALSO
~~~~~~~~

* :ref:`airshipctl phase <airshipctl_phase>` 	 - Airshipctl command to manage phases

