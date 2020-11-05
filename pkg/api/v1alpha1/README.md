# Generating *zz_generated.deepcopy.go* in api/v1alpha1

This directory contains the data types needed by *airshipctl phase run* command.

When you add a new data structure in this directory you will need to generate the file *zz_generated.deepcopy.go*.
To generate this file you will need the tool *controller-gen" executable.

If you don't have *controller-gen* in your machine, clone the following repository and compile it.

```bash
git clone https://github.com/kubernetes-sigs/controller-tools.git
cd controller-tools/cmd/controller-gen
go build -o controller-gen
```

Now you can generate the *zz_generated.deepcopy.go* using *controller-gen* as follow:

```bash
/path/to/controller-gen object paths=/path/to/airshipctl/pkg/api/v1alpha1/
```

At this point you should have a newly generated *zz_generated.deepcopy.go*.
Just check if your data structure has been added to this file and you are good to go.

>TODO: Add this task in the Makefile
