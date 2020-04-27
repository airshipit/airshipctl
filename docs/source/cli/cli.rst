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

The AirshipCTL CLI is used in conjunction with the binary created by running
``make build``.  This binary, by default, is created in the ``airshipctl/bin/``
directory.


CLI Options
===========

**-h / \\-\\-help**

Prints help for a specific command or command group.

**\\-\\-debug** (default: false)

Enables verbose output of commands.

**\\-\\-airshipconf** (default: `$HOME/.airship/config`)

Path to file for airshipctl configuration.

**\\-\\-kubeconfig** (default: `$HOME/.airship/kubeconfig`)

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

Generate completion script for airshipctl for the specified shell (bash or zsh).

**shell** (Required)

Shell to generate completion script for.  Supported values are `bash` and `zsh`

Usage:

::

    airshipctl completion <shell>

Examples
^^^^^^^^

Save shell completion to a file

::

    $ airshipctl completion bash > $HOME/.airship_completions

Apply completions to the current shell

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

This command provides capabilities for interacting with a Kubernetes cluster,
such as getting status and deploying initial infrastructure.

InitInfra
------------

Deploy initinfra components to cluster.

**cluster-type** (default:"ephemeral")

Select cluster type to deploy initial infrastructure to, currently only ephemeral is supported.

**\\-\\-dry-run**

Don't deliver documents to the cluster, simulate the changes instead.

**\\-\\-prune** (default: false)

If set to true, command will delete all kubernetes resources that are not defined in airship documents and have
airshipit.org/deployed=initinfra label

Usage:

::

    airshipctl cluster initinfra <flags>

.. _config-group:

Config Group
============

Manage the airshipctl config file

Get-Cluster
-----------

Get cluster information from the airshipctl config.

**name** (Optional, default: all defined clusters)

Display a specific cluster or all defined clusters if no name is provided.

**\\-\\-cluster-type** (Required if **name** is provided).

The type of the desired cluster. Valid values are from [ephemeral|target].

Usage:

::

    airshipctl config get-cluster <name> --cluster-type=<cluster-type>

Examples
^^^^^^^^

List all the clusters:

::

    airshipctl config get-cluster

Display a specific cluster:

::

    airshipctl config get-cluster e2e --cluster-type=ephemeral

Get-Context
-----------

Display information about contexts such as associated manifests, users, and clusters.

**name** (Optional, default: all defined contexts)

Displays a named context, if no name is provided display all defined contexts.

**\\-\\-current-context** (default: false)

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

    airshipctl config get-context exampleContext


Get-Credentials
---------------

Get user credentials from the airshipctl config.

**name** (Optional, default: all defined users)

Display a specific user's credentials, or all defined user credentials if no name is provided.

Usage:

::

    airshipctl config get-credentials <name>

Examples
^^^^^^^^

List all user credentials:

::

    airshipctl config get-credentials

Display a specific user's credentials:

::

    airshipctl config get-credentials exampleUser

Init
----

Generate an airshipctl config file and its associated kubeConfig file.
These files will be written to the $HOME/.airship directory, and will contain
default configurations.

.. note:: This will overwrite any existing config files in $HOME/.airship

Usage:

::

    airshipctl config init

Set-Cluster
-----------

Create or modify a cluster in the airshipctl config files.

Since a cluster can be either "ephemeral" or "target", you must specify
cluster-type when managing clusters.

**name** (Required)

The name of the cluster to add or modify in the airshipctl config file.

**\\-\\-certificate-authority**

Path to a certificate authority file

**\\-\\-certificate-authority** (Required)

The type of the cluster to add or modify

**\\-\\-embed-certs** (default: false)

If set, embed the client certificate/key into the cluster

**\\-\\-insecure-skip-tls-verify** (default: true)

If set, disable certificate checking

**\\-\\-server**

Server to use for the cluster

Usage:

::

    airshipctl config set-cluster <name> <flags>

Examples
^^^^^^^^

Set the server field on the ephemeral exampleCluster:

::

    airshipctl config set-cluster exampleCluster \
      --cluster-type=ephemeral \
      --server=https://1.2.3.4

Embed certificate authority data for the target exampleCluster:

::

    airshipctl config set-cluster exampleCluster \
      --cluster-type=target \
      --client-certificate-authority=$HOME/.airship/ca/kubernetes.ca.crt \
      --embed-certs

Disable certificate checking for the target exampleCluster:

::

    airshipctl config set-cluster exampleCluster
      --cluster-type=target \
      --insecure-skip-tls-verify

Configure client certificate for the target exampleCluster:

::

    airshipctl config set-cluster exampleCluster \
      --cluster-type=target \
      --embed-certs \
      --client-certificate=$HOME/.airship/cert_file

Set-Context
-----------

Create or modify a context in the airshipctl config files.

**name** (Required)

The name of the context to add or modify in the airshipctl config file.

**\\-\\-cluster**

Set the cluster for the specified context.

**\\-\\-cluster-type**

Set the cluster-type for the specified context.

**\\-\\-current**

Update the current context.

**\\-\\-manifest**

Set the manifest for the specified context.

**\\-\\-namespace**

Set the namespace for the specified context.

**\\-\\-user**

Set the user for the specified context.

Usage:

::

    airshipctl config set-context <name> <flags>

Examples
^^^^^^^^

Create a new context named "exampleContext":

::

    airshipctl config set-context exampleContext \
      --namespace=kube-system \
      --manifest=exampleManifest \
      --user=exampleUser
      --cluster-type=target

Update the manifest of the current-context:

::

   airshipctl config set-context \
     --current \
     --manifest=exampleManifest


Set-Credentials
---------------

Create or modify a user credential in the airshipctl config file.

.. note:: Specifying more than one authentication method is an error.

**name** (Required)

The user entry to update in airshipctl config.

**\\-\\-client-certificate**

Path to a certificate file.

**\\-\\-client-key**

Path to a key file.

**\\-\\-embed-certs**

If set, embed the client certificate/key into the credential.

**\\-\\-password**

Password for the credential

.. note:: Username and Password flags are mutually exclusive with Token flag

**\\-\\-token**

Token to use for the credential

.. note:: Username and Password flags are mutually exclusive with Token flag

**\\-\\-username**

Username for the credential

.. note:: Username and Password flags are mutually exclusive with Token flag

Usage:

::

    airshipctl config set-credentials <name> <flags>

Examples
^^^^^^^^

Create a new user credential with basic auth:

::

    airshipctl config set-credentials exampleUser \
      --username=exampleUser \
      --password=examplePassword

Change the client-key of a user named admin

::

    airshipctl config set-credentials admin \
      --client-key=$HOME/.kube/admin.key

Change the username and password of the admin user

::

    airshipctl config set-credentials admin \
      --username=admin \
      --password=uXFGweU9l35qcif

Embed client certificate data of the admin user

::

    airshipctl config set-credentials admin \
      --client-certificate=$HOME/.kube/admin.crt \
      --embed-certs

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
