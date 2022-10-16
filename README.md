# gateway admision controller

Originally based on the [k8s-at-home container template](https://github.com/k8s-at-home/template-container-image)
and the [example for Kubewebhook](https://github.com/slok/k8s-webhook-example/), this
[admision webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
changes the default gateway and, optionally, the DNS of processed pods. It does so by adding an
init container and a sidecar. The sidecar is used in case the IP of the gateway changes.

This is useful in order to send traffic to a VPN forwarder, traffic scanner, etc instead of using the
default cluster egress.

The [.github](.github) folder will get PRs from this template so you can apply the latest workflows.

## Prereqs

You need to create the following secrets (not needed within the k8s-at-home org - there we use org-wide secrets):
- GHCR_USERNAME            # Needed to upload container to the Github Container Registry
- GHCR_TOKEN               # Needed to upload container to the Github Container Registry

## How to build

1. Build and test local
    ```bash
    make
    ```
2. Build the container
    ```bash
    make docker-build
    ```

Check the [Makefile] for other build targets

## How to run

It is expected to be used from within a Helm chart but the binary might also
be run directly:

1. Run
    ```bash
    make run
    ```
2. Connect to <host IP>:8080

For more options you might run `make help`

