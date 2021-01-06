[![Build Status](https://travis-ci.org/kubernetes-sigs/aws-efs-csi-driver.svg?branch=master)](https://travis-ci.org/kubernetes-sigs/aws-efs-csi-driver)
[![Coverage Status](https://coveralls.io/repos/github/kubernetes-sigs/aws-efs-csi-driver/badge.svg?branch=master)](https://coveralls.io/github/kubernetes-sigs/aws-efs-csi-driver?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes-sigs/aws-efs-csi-driver)](https://goreportcard.com/report/github.com/kubernetes-sigs/aws-efs-csi-driver)

## Amazon EFS CSI Driver

The [Amazon Elastic File System](https://aws.amazon.com/efs/) Container Storage Interface (CSI) Driver implements the [CSI](https://github.com/container-storage-interface/spec/blob/master/spec.md) specification for container orchestrators to manage the lifecycle of Amazon EFS file systems.

### CSI Specification Compatibility Matrix
| AWS EFS CSI Driver \ CSI Spec Version  | v0.3.0| v1.1.0 | v1.2.0 |
|----------------------------------------|-------|--------|--------|
| master branch                          | no    | no     | yes    |
| v1.0.0                                 | no    | no     | yes    |
| v0.3.0                                 | no    | yes    | no     |
| v0.2.0                                 | no    | yes    | no     |
| v0.1.0                                 | yes   | no     | no     |

## Features
Currently only static provisioning is supported. This means an AWS EFS file system needs to be created manually on AWS first. After that it can be mounted inside a container as a volume using the driver.

The following CSI interfaces are implemented:
* Node Service: NodePublishVolume, NodeUnpublishVolume, NodeGetCapabilities, NodeGetInfo, NodeGetId
* Identity Service: GetPluginInfo, GetPluginCapabilities, Probe

### Encryption In Transit
One of the advantages of using EFS is that it provides [encryption in transit](https://aws.amazon.com/blogs/aws/new-encryption-of-data-in-transit-for-amazon-efs/) support using TLS. Using encryption in transit, data will be encrypted during its transition over the network to the EFS service. This provides an extra layer of defence-in-depth for applications that requires strict security compliance.

Encryption in transit is enabled by default in the master branch version of the driver. To disable it and mount volumes using plain NFSv4, set `volumeAttributes` field `encryptInTransit` to `"false"` in your persistent volume manifest. For an example manifest, see [Encryption in Transit Example](../examples/kubernetes/encryption_in_transit/specs/pv.yaml).

**Note** Kubernetes version 1.13+ is required if you are using this feature in Kubernetes.

## EFS CSI Driver on Kubernetes
The following sections are Kubernetes specific. If you are a Kubernetes user, use this for driver features, installation steps and examples.

### Kubernetes Version Compability Matrix
| AWS EFS CSI Driver \ Kubernetes Version| maturity | v1.11 | v1.12 | v1.13 | v1.14 | v1.15 | v1.16 | v1.17 |
|----------------------------------------|----------|-------|-------|-------|-------|-------|-------|-------|
| master branch                          | GA       | no    | no    | no    | yes   | yes   | yes   | yes   |
| v1.0.0                                 | GA       | no    | no    | no    | yes   | yes   | yes   | yes   |
| v0.3.0                                 | beta     | no    | no    | no    | yes   | yes   | yes   | yes   |
| v0.2.0                                 | beta     | no    | no    | no    | yes   | yes   | yes   | yes   |
| v0.1.0                                 | alpha    | yes   | yes   | yes   | no    | no    | no    | no    |

### Container Images
|EFS CSI Driver Version     | Image                               |
|---------------------------|-------------------------------------|
|master branch              |amazon/aws-efs-csi-driver:master     |
|v1.0.0                     |amazon/aws-efs-csi-driver:v1.0.0     |
|v0.3.0                     |amazon/aws-efs-csi-driver:v0.3.0     |
|v0.2.0                     |amazon/aws-efs-csi-driver:v0.2.0     |
|v0.1.0                     |amazon/aws-efs-csi-driver:v0.1.0     |

### Features
* Static provisioning - EFS file system needs to be created manually first, then it could be mounted inside container as a persistent volume (PV) using the driver.
* Mount Options - Mount options can be specified in the persistent volume (PV) to define how the volume should be mounted.
* Encryption of data in transit - EFS file systems are mounted with encryption in transit enabled by default in the master branch version of the driver.

**Notes**:
* Since EFS is an elastic file system it doesn't really enforce any file system capacity. The actual storage capacity value in persistent volume and persistent volume claim is not used when creating the file system. However, since the storage capacity is a required field by Kubernetes, you must specify the value and you can use any valid value for the capacity.

### Installation
Deploy the driver:

If you want to deploy the stable driver:
```sh
kubectl apply -k "github.com/kubernetes-sigs/aws-efs-csi-driver/deploy/kubernetes/overlays/stable/?ref=release-1.0"
```

If you want to deploy the development driver:
```sh
kubectl apply -k "github.com/kubernetes-sigs/aws-efs-csi-driver/deploy/kubernetes/overlays/dev/?ref=master"
```

Alternatively, you could also install the driver using helm:
```sh
helm repo add aws-efs-csi-driver https://kubernetes-sigs.github.io/aws-efs-csi-driver/
helm repo update
helm upgrade --install aws-efs-csi-driver aws-efs-csi-driver/aws-efs-csi-driver
```

### Examples
Before the example, you need to:
* Get yourself familiar with how to setup Kubernetes on AWS and how to [create EFS file system](https://docs.aws.amazon.com/efs/latest/ug/getting-started.html).
* When creating EFS file system, make sure it is accessible from Kuberenetes cluster. This can be achieved by creating the file system inside the same VPC as Kubernetes cluster or using VPC peering.
* Install EFS CSI driver following the [Installation](README.md#Installation) steps.

#### Example links
* [Static provisioning](../examples/kubernetes/static_provisioning/README.md)
* [Encryption in transit](../examples/kubernetes/encryption_in_transit/README.md)
* [Accessing the file system from multiple pods](../examples/kubernetes/multiple_pods/README.md)
* [Consume EFS in StatefulSets](../examples/kubernetes/statefulset/README.md)
* [Mount subpath](../examples/kubernetes/volume_path/README.md)
* [Use Access Points](../examples/kubernetes/access_points/README.md)

## Development
Please go through [CSI Spec](https://github.com/container-storage-interface/spec/blob/master/spec.md) and [Kubernetes CSI Developer Documentation](https://kubernetes-csi.github.io/docs) to get some basic understanding of CSI driver before you start.

### Requirements
* Golang 1.13.4+

### Dependency
Dependencies are managed through go module. To build the project, first turn on go mod using `export GO111MODULE=on`, to build the project run: `make`

### Testing
To execute all unit tests, run: `make test`

## License
This library is licensed under the Apache 2.0 License.
