<!--
 Copyright 2023 Dimitri Koshkin. All rights reserved.
 SPDX-License-Identifier: Apache-2.0
 -->

# kubernetes-upgrader

A set of Kubernetes controllers to automate Kubernetes clusters upgrade using [Cluster API's ClusterClass](https://cluster-api.sigs.k8s.io/tasks/experimental-features/cluster-class/).

## Description

This project is a set of Kubernetes controllers that build Kubernetes machine images using upstream [image-builder](https://github.com/kubernetes-sigs/image-builder),
and then use those images to upgrade Kubernetes clusters that are using [Cluster API's ClusterClass](https://cluster-api.sigs.k8s.io/tasks/experimental-features/cluster-class/).

This should work with any Infrastructure provider supported by image-builder,
but it was only tested with [Docker's CAPD](https://github.com/kubernetes-sigs/cluster-api/tree/main/test/infrastructure/docker) and [vSphere's CAPV](https://github.com/kubernetes-sigs/cluster-api-provider-vsphere).

The API and controllers design take inspiration from [Cluster API's](https://cluster-api.sigs.k8s.io/) including:

-   There are template types that are used to create the actual types.
    This is similar to CAPI's bootstrap and infrastructure templates.
-   Some of the types comply to API contracts and share information through status fields.
    This is similar to CAPI's infrastructure provider contracts and allows external controllers implementations.

### API Types

![Architecture](https://lucid.app/publicSegments/view/7e0cdb4a-8908-48dc-87a1-5933652564df/image.png "Architecture")

-   `DebianRepositorySource` satisfies the `MachineImageSyncer` contract, fetching all available versions from a Kubernetes Debian repository and setting `status.versions`.
-   `MachineImageTemplate` is a template for `MachineImage` used by `MachineImageSyncer`.
-   `MachineImageSyncer` creates a new `MachineImage` objects with the latest version from `sourceRef` object. `DebianRepositorySource` is one implementation of the contract, but any type can be used as long as it sets the `status.versions` field.
-   `MachineImage` runs a Job to build a Kubernetes machine image. The examples use [image-builder](https://github.com/kubernetes-sigs/image-builder),
    but any tool can be used as long as it generates a Kubernetes machine image and labels the Job with `kubernetesupgraded.dimitrikoshkin.com/image-id`.
    The controller sets `status.ready` once the Job succeeds and copies the label value to `spec.id`.
-   `Plan` finds the latest `MachineImage` with `spec.version` that is in `spec.versionRange` and matches an optional `machineImageSelector`. It then sets `status.machineImageDetails` with `version` and image `id`.
-   `ClusterClassClusterUpgrader` satisfies the Plan contract. It will update `spec.topology.version` and (an optional) field from `topologyVariable` of the CAPI Cluster, using the values from the Plan's `status.machineImageDetails`. .

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

This will deploy the controllers and a sample to build a vSphere OVA,
read the [upstream docs](https://image-builder.sigs.k8s.io/capi/providers/vsphere) on how to configure the vSphere provider.

1.  Deploy the CRDs and the controllers:

    ```sh
    kubectl apply -f https://github.com/dkoshkin/kubernetes-upgrader/releases/latest/download/components.yaml
    ```

1.  Create a Secret with vSphere credentials:

    ```sh
    cat << 'EOF' > vsphere.json
    {
      "vcenter_server":"$VSPHERE_SERVER",
      "insecure_connection": "true",
      "username":"$VSPHERE_USERNAME",
      "password":"$VSPHERE_PASSWORD",
      "vsphere_datacenter": "$VSPHERE_DATACENTER",
      "cluster": "$VSPHERE_CLUSTER",
      "datastore":"$VSPHERE_DATASTORE",
      "folder": "$VSPHERE_TEMPLATE_FOLDER",
      "network": "$VSPHERE_NETWORK",
      "convert_to_template": "true"
    }
    EOF
    kubectl create secret generic image-builder-vsphere-vars --from-file=vsphere.json
    ```

1.  Deploy the samples:

    ```sh
    kubectl apply -f https://github.com/dkoshkin/kubernetes-upgrader/releases/latest/download/example-vsphere-with-job-template.yaml
    ```

1.  The controller will create a Job to build the image, after some time you should see the image in the vSphere UI.
    Check the status of `MachineImage` to see if the image was successfully built:

    ```sh
    kubectl get MachineImage -o yaml
    ```

    You should see `status.ready` set to `true` and `spec.id` set to a newly created OVA template.

## For Developers

You’ll need a Kubernetes cluster to run against.
Follow [CAPI's Quickstart documentation](https://cluster-api.sigs.k8s.io/user/quick-start.html) to create a cluster using [KIND](https://sigs.k8s.io/kind) and the Docker provider.
Use Kubernetes version `v1.26.3` if you are planning on using the sample config.

### Creating workload cluster

1.  Create a KIND bootstrap cluster:

    ```sh
    kind create cluster --config - <<EOF
    kind: Cluster
    apiVersion: kind.x-k8s.io/v1alpha4
    networking:
      ipFamily: dual
    nodes:
    - role: control-plane
      extraMounts:
      - hostPath: /var/run/docker.sock
        containerPath: /var/run/docker.sock
    EOF
    ```

1.  Deploy CAPI providers:

    ```sh
    export CLUSTER_TOPOLOGY=true
    clusterctl init --infrastructure docker --wait-providers
    ```

1.  Create a workload cluster:

    ```sh
    # The list of service CIDR, default ["10.128.0.0/12"]
    export SERVICE_CIDR=["10.96.0.0/12"]
    # The list of pod CIDR, default ["192.168.0.0/16"]
    export POD_CIDR=["192.168.0.0/16"]
    # The service domain, default "cluster.local"
    export SERVICE_DOMAIN="k8s.test"
    # Create the cluster
    clusterctl generate cluster capi-quickstart --flavor development \
    --kubernetes-version v1.26.3 \
    --control-plane-machine-count=1 \
    --worker-machine-count=1 | kubectl apply -f -
    ```

### Running on the cluster

1.  Generated the components manifests and build the image:

    ```sh
    make release-snapshot
    make docker-push IMG=ghcr.io/dkoshkin/kubernetes-upgrader:$(gojq -r '.version' dist/metadata.json)
    ```

1.  If using a local KIND cluster:

    ```sh
    kind load docker-image ghcr.io/dkoshkin/kubernetes-upgrader:$(gojq -r '.version' dist/metadata.json)
    ```

1.  Deploy the controller to the cluster with the image specified by `IMG`:

    ```sh
    make deploy IMG=ghcr.io/dkoshkin/kubernetes-upgrader:$(gojq -r '.version' dist/metadata.json)
    ```

1.  Deploy the samples:

    ```sh
    kubectl apply -f config/samples/example-docker-static.yaml
    ```

### Undeploy controller

UnDeploy the controller from the cluster:

```sh
make undeploy
```

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
