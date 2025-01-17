# Spinnaker Operator for Kubernetes

We've announced the [Spinnaker Operator](https://blog.armory.io/spinnaker-operator/): a Kubernetes operator to deploy and manage Spinnaker with the tools you're used to. We're sharing configuration in this repository (code to come soon) to let the community evaluate it and provide feedback.

## Goals
The Spinnaker operator:
- should be able to install any version of Spinnaker with a published BOM
- should perform preflight checks to confidently upgrade Spinnaker

More concretely, the operator:
- is configured via a `configMap` or a `secret`
- can deploy in a single namespace or in multiple namespaces
- garbage collect configuration (secrets, deployments, ...)
- provides a validating admission webhook to validate the configuration before it is applied

We plan to support many validations such as provider (AWS, Kubernetes,...) validation, connectivity to CI. Please let us know what would make your life easier when installing Spinnaker! You can use GitHub issues for the time being.


## Limitations
*The operator is in alpha and its CRD may change quite a bit. It is actively being developed.*
- Spinnaker configuration in `secret` is not supported at the moment.

## Requirements
The validating admission controller [requires](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites):
- Kubernetes server v1.13+
- Admission controllers enabled (`-enable-admission-plugins`)
- `ValidatingAdmissionWebhook` enabled in the kube-apiserver (should be the default)

Note: If you can't use the validation webhook, pass the `--without-admission-controller` to the operator (like in `deploy/operator/basic/deployment.yaml`).

## Operator Installation

First we'll install the `SpinnakerService` CRD:

```bash
$ git clone https://github.com/armory/spinnaker-operator
$ cd spinnaker-operator
$ kubectl apply -f deploy/crds/spinnaker_v1alpha1_spinnakerservice_crd.yaml
```

There are two modes for the operator:
- basic mode to install Spinnaker in a single namespace without validating admission webhook
- cluster mode works across namespaces and requires a `ClusterRole` to perform validation

The main difference between the two modes is that basic only requires a `Role` (vs a `ClusterRole`) and has no validating webhook.

Once installed you should see a new deployment representing the operator. The operator watches for changes to the `SpinnakerService` objects. You can check on the status of the operator using `kubectl`.

### Basic install (no validating webhook)
To install the operator run:

```bash
$ kubectl apply -n <namespace> -f deploy/operator/basic
```

`namespace` is the namespace where you want the operator to live and deploy to.

### Cluster install
To install the operator:
1. Edit the namespace in `deploy/operator/cluster/role_binding.yml` to be the namespace where you want the operator to live.
2. Run: 

```bash
$ kubectl apply -n <namespace> -f deploy/operator/cluster
```


## Spinnaker Installation

Once you've installed the operator, you can install Spinnaker by making a configuration (`configMap`). Check out examples in `deploy/spinnaker/examples`. If you prefer to use `kustomize`, we've added some kustomization in `deploy/spinnaker/kustomize` (WIP)


### Example 1: Installing version 1.15.1

```bash
$ kubectl -n <namespace> apply -f deploy/spinnaker/examples/basic
```

This configuration does not contain any connected accounts, just a persistent storage. 

### Example 2: Using Kustomize (TODO)

Set your own values in `deploy/spinnaker/kustomize/kustomization.yaml`, then:
 

```bash
$ kustomize build deploy/spinnaker/kustomize | kubectl -n <namespace> apply -f -
```

Or if using `kubectl` version 1.14+:
```bash
$ kubectl -n <namespace> apply -f deploy/spinnaker/kustomize
```


### Managing Spinnaker

You can manage your Spinnaker installations with `kubectl`. 

#### Listing Spinnaker instances
```bash
$ kubectl get spinnakerservice --all-namespaces
NAMESPACE     NAME        VERSION
mynamespace   spinnaker   1.15.1
```

The short name `spinsvc` is also available.

#### Describing Spinnaker instances
```bash
$ kubectl -n mynamespace describe spinnakerservice spinnaker
```

#### Deleting Spinnaker instances
Delete:
```bash
$ kubectl -n mynamespace deleted spinnakerservice spinnaker
spinnakerservice.spinnaker.io "spinnaker" deleted
```


## Configuring Spinnaker

### `SpinnakerService` (TODO)

The `SpinnakerService` points to the `configMap` with the configuration (see below).

### `configMap`

The `configMap` holding Spinnaker configuration can contain the following keys:
- `config`: the deployment configuration in the same format as in Halyard. For instance, given the following `~/.hal/config`:
```yaml
currentDeployment: default
deploymentConfigurations:
- name: default
  version: 1.15.2
  providers:
...
```

The `config` key would contain:
```yaml
name: default
version: 1.15.2
providers:
...
```

- `profiles`: the content of the local profile files (`~/.hal/<deployment>/profiles/`) by service name, e.g.:
```yaml
profiles: |
  gate:
    default.apiPort: 8085
```

- `serviceSettings`: the content of the service settings file (`~/.hal/<deployment>/profiles/`) by service name, e.g.:
```yaml
serviceSettings: |
  gate:
    artifactId: xxxxx
```

- `files__<relative path to other file>`: Other supporting files with a path relative to the main deployment. The file 
path is encoded with `__` as a path separator. This includes other profile files such as a custom packer template in
`files__profiles__rosco__packer__aws-custom.json`.   

#### Putting it all together:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: spinconfig-v001
data:
  config: |
    name: default
    version: 1.15.1
    ...
  profiles: |
    gate:
      default.apiPort: 8085
  serviceSettings: |
    gate:
      artifactId: xxxxx
  files__profiles__rosco__packer__aws-custom.json: |
    {
      "variables": {
        "docker_source_image": "null",
        "docker_target_image": null,
      },
      ...
    }
```
## Architecture (TODO)
