<!--
 Copyright 2023 Dimitri Koshkin. All rights reserved.
 SPDX-License-Identifier: Apache-2.0
 -->

# kubernetes-upgrader

A set of Kubernetes controllers to automate Kubernetes clusters upgrade using [Cluster API's ClusterClass](https://cluster-api.sigs.k8s.io/tasks/experimental-features/cluster-class/).

## Description

This project is a set of Kubernetes controllers that build Kubernetes machine images using upstream [image-builder](https://github.com/kubernetes-sigs/image-builder),
and then use those images to upgrade Kubernetes clusters that are using [Cluster API's ClusterClass](https://cluster-api.sigs.k8s.io/tasks/experimental-features/cluster-class/).

This should work with any Infrastructure provider supported by image-builder, but it was only tested with [vSphere's CAPV](https://github.com/kubernetes-sigs/cluster-api-provider-vsphere).

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
    kubectl apply -f https://github.com/dkoshkin/kubernetes-upgrader/releases/latest/download/sample-with-job-template.yaml
    ```

1.  The controller will create a Job to build the image, after some time you should see the image in the vSphere UI.
    Check the status of `MachineImage` to see if the image was successfully built:

    ```sh
    kubectl get MachineImage -o yaml
    ```

    You should see `status.ready` set to `true` and `spec.imageID` set to a newly created OVA template.

## For Developers

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

The controller deploys a webhook which requires a TLS certificate. The easiest way to get a certificate is to use [cert-manager](https://cert-manager.io/docs/installation/kubernetes/).
You can deploy cert-manager using the following command:

```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.4/cert-manager.yaml
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
    kubectl apply -f config/samples/
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
