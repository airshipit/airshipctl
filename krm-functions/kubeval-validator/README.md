# Validation

This is a KRM function which implementing a validation function against
[kubeval](https://github.com/instrumenta/kubeval).

## Function implementation

The function is implemented as an [image](image), and built using `make image`.

### Function configuration

A number of settings can be modified for `kubeval` in the struct `Spec`. See
the `Spec` struct definition in [main.go](image/main.go) for the documentation.

## Function invocation

The function invokes by running validate command via `airshipctl`:

    airshipctl plan validate <plan_name>
    airshipctl phase validate <phase_name>

This exists non-zero if kubeval detects an invalid Resource.
