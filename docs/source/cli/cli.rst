..
      All Rights Reserved.

      Licensed under the Apache License, Version 2.0 (the "License"); you may
      not use this file except in compliance with the License. You may obtain
      a copy of the License at

          http://www.apache.org/licenses/LICENSE-2.0

      Unless required by applicable law or agreed to in writing, software
      distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
      WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
      License for the specific language governing permissions and limitations
      under the License.

.. _airshipctl-cli:

==============
AirshipCTL CLI
==============

The AirshipCTL CLI is used in conjunction with the binary created by running ``make build``.  This binary, by default,
is created in the ``airshipctl/bin/`` directory.


CLI Options
===========

**-h / \\-\\-help**

Prints help for a specific command or command group.

**\\-\\-debug** (Optional, default: false)

Enables verbose output of commands.

**\\-\\-airshipconf** (Optional, default: `$HOME/.airship/config`)

Path to file for airshipctl configuration.

**\\-\\-kubeconfig** (Optional, default: `$HOME/.airship/kubeconfig`)

Path to kubeconfig associated with airshipctl configuration.

.. _root-group:

Root Group
==========

Allows you to perform top level commands

::

  airshipctl <command>

Version
-------

Output the version of the airshipctl binary.

Usage:

::

    airshipctl version

Completion
----------

Generate autocompletion script for airshipctl for the specified shell (bash or zsh).

**shell** (Required)

Shell to generate autocompletion script for.  Supported values are `bash` and `zsh`

Usage:

::

    airshipctl completion <shell>

Examples
^^^^^^^^

This command can generate bash autocompletion. e.g.

::

    $ airshipctl completion bash

Which can be sourced as such:

::

    $ source <(airshipctl completion bash)

.. _bootstrap-group:

Bootstrap Group
===============

Used to bootstrap the ephemeral Kubernetes cluster.

ISOgen
-------

Generate bootstrap ISO image.

Usage:

::

    airshipctl bootstrap isogen

RemoteDirect
------------

Bootstrap ephemeral node.

Usage:

::

    airshipctl bootstrap remotedirect

.. _cluster-group:

Cluster Group
=============

Control Kubernetes cluster.

InitInfra
------------

Deploy initinfra components to cluster.

**cluster-type** (Optional, default:"ephemeral")

Select cluster type to deploy initial infrastructure to, currently only ephemeral is supported.

**\\-\\-dry-run** (Optional).

Don't deliver documents to the cluster, simulate the changes instead.

**\\-\\-prune** (Optional, default:false)

If set to true, command will delete all kubernetes resources that are not defined in airship documents and have
airshipit.org/deployed=initinfra label

Usage:

::

    airshipctl cluster initinfra <flags>

.. _config-group:

Config Group
============

Modify airshipctl config files

Get-Cluster
-----------

Display cluster information.

**name** (Optional, default: all defined clusters)

Displays a specific cluster if specified, or if left blank all defined clusters.

**\\-\\-cluster-type** (Required).

cluster-type for the cluster-entry in airshipctl config. Currently only ephemeral cluster types are supported.

Usage:

::

    airshipctl config get-cluster <name> --cluster-type=<cluster-type>

Examples
^^^^^^^^

List all the clusters airshipctl knows about:

::

    airshipctl config get-cluster

Display a specific cluster:

::

    airshipctl config get-cluster e2e --cluster-type=ephemeral

Get-Context
-----------

Displays context information

**name** (Optional, default: all defined contexts)

Displays a named context, if no name is provided display all defined contexts.

**\\-\\-current-context** (Optional, default:false)

Display the current context, supersedes the `name` argument.

Usage:

::

    airshipctl config get-context

Examples
^^^^^^^^

For all contexts:

::

    airshipctl config get-context

For the current context:

::

    airshipctl config get-context --current

For a named context:

::

    airshipctl config get-context e2e


Get-Credentials
---------------

Display a user's information.

**name** (Optional, default: all defined users)

Display a specific user's information.  If no name is specified, list all defined users.

Usage:

::

    airshipctl config get-credentials <NAME>

Examples
^^^^^^^^

List all the users airshipctl knows about:

::

    airshipctl config get-credentials

Display a specific user's information:

::

    airshipctl config get-credentials e2e

Init
----

Generate initial configuration files for airshipctl

Usage:

::

    airshipctl config init

Set-Cluster
-----------

Sets a cluster entry in the airshipctl config.

**name** (Required)

The name of the cluster to add to airshipctl config.

.. note::

    Specifying a name that already exists will merge new fields on top of existing values for those fields.

**\\-\\-certificate-authority** (Optional)

Path to certificate-authority file for the cluster entry in airshipctl config

**\\-\\-certificate-authority** (Required)

Cluster-type for the cluster entry in airshipctl config

**\\-\\-embed-certs** (Optional)

Embed-certs for the cluster entry in airshipctl config

**\\-\\-insecure-skip-tls-verify** (Optional, default:true)

Insecure-skip-tls-verify for the cluster entry in airshipctl config

**\\-\\-server** (Optional)

Server for the cluster entry in airshipctl config

Usage:

::

    airshipctl config set-cluster <name> <flags>

Examples
^^^^^^^^

Set only the server field on the e2e cluster entry without touching other values:

::

    airshipctl config set-cluster e2e --cluster-type=ephemeral --server=https://1.2.3.4

Embed certificate authority data for the e2e cluster entry:

::

    airshipctl config set-cluster e2e --cluster-type=target --certificate-authority-authority=~/.airship/e2e/kubernetes.ca.crt

Disable cert checking for the dev cluster entry:

::

    airshipctl config set-cluster e2e --cluster-type=target --insecure-skip-tls-verify=true

Configure client certificate:

::

    airshipctl config set-cluster e2e --cluster-type=target --embed-certs=true --client-certificate=".airship/cert_file"

Set-Context
-----------

Switch to a new context, or update context values in the airshipctl config

**name** (Required)

The name of the context to set.

**\\-\\-cluster-string**

Sets the cluster for the specified context in the airshipctl config.

**\\-\\-cluster-type**

Sets the cluster-type for the specified context in the airshipctl config.

**\\-\\-current**

Use current context from airshipctl config.

**\\-\\-manifest**

Sets the manifest for the specified context in the airshipctl config.

**\\-\\-namespace**

Sets the namespace for the specified context in the airshipctl config.

**\\-\\-user**

Sets the user for the specified context in the airshipctl config.

Usage:

::

    airshipctl config set-context <name> <flags>

Examples
^^^^^^^^

Create a completely new e2e context entry:

::

    airshipctl config set-context e2e --namespace=kube-system --manifest=manifest --user=auth-info --cluster-type=target

Update the current-context to e2e:

::

    airshipctl config set-context e2e

Update attributes of the current-context:

::

    airshipctl config set-context --current --manifest=manifest


Set-Credentials
---------------

Sets a user entry in the airshipctl config.

**name** (Required)

The user entry to update in airshipctl config.

.. note:: Specifying a name that already exists will merge new fields on top of existing values.

**\\-\\-client-certificate**

Path to client-certificate file for the user entry in airshipctl

**\\-\\-client-key**

Path to client-key file for the user entry in airshipctl

**\\-\\-embed-certs**

Embed client cert/key for the user entry in airshipctl

**\\-\\-password**

Password for the user entry in airshipctl

.. note:: Username and Password flags are mutually exclusive with Token flag

**\\-\\-token**

Token for the user entry in airshipctl

.. note:: Username and Password flags are mutually exclusive with Token flag

**\\-\\-username**

Username for the user entry in airshipctl

.. note:: Username and Password flags are mutually exclusive with Token flag

Usage:

::

    airshipctl config set-credentials <name> <flags>

Examples
^^^^^^^^

Set only the "client-key" field on the "cluster-admin" entry, without touching other values:

::

    airshipctl config set-credentials cluster-admin --username=~/.kube/admin.key

Set basic auth for the "cluster-admin" entry

::

    airshipctl config set-credentials cluster-admin --username=admin --password=uXFGweU9l35qcif

Embed client certificate data in the "cluster-admin" entry

::

    airshipctl config set-credentials cluster-admin --client-certificate=~/.kube/admin.crt --embed-certs=true

.. _document-group:

Document Group
==============

Manages deployment documents.

Pull
----

Pulls documents from remote git repository.

Usage:

::

    airshipctl document pull

Render
------

Render documents from model.

**-a / \\-\\-annotation**

Filter documents by Annotations.

**-g / \\-\\-apiversion**

Filter documents by API version.

**-f / \\-\\-filter**

Logical expression for document filtering.

**-k / \\-\\-kind**

Filter documents by Kinds.

**-l / \\-\\-label**

Filter documents by Labels.

Usage:

::

    airshipctl document render <flags>

.. _secret-group:

Secret Group
============

Manages secrets.

Generate
--------

Generates various secrets.

MasterPassphrase
^^^^^^^^^^^^^^^^

Generates a secure master passphrase.

Usage:

::

    airshipctl secret generate masterpassphrase
