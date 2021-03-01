# Cloud-init

This function generates yaml files (user-data and network-data) needed for building the bootstrap ISO image.
It assumes a `ResourceList` is passed to its input where `items` field is a phase executor document bundle
that contains necessary data. To get the data from the bundle the `functionConfig` field must contain
`IsoConfiguration` document that defines document selector parameters.
