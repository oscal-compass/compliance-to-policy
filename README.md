# compliance-to-policy
Compliance-to-Policy (C2P) provides the framework to bridge the gap between compliance and policy administration.

## Prerequisites
1. Install [Policy Generator Plugin](https://github.com/open-cluster-management-io/policy-generator-plugin#as-a-kustomize-plugin)

## C2P Decomposer
Decompose OCM poicy collection to kubernetes resources composing each OCM policy (we call it policy resource).

1. Clone [Policy Collection](https://github.com/open-cluster-management-io/policy-collection)
    ```
    git clone --depth 1 https://github.com/open-cluster-management-io/policy-collection.git /tmp/policy-collection
    ```
1. Run C2P Decomposer
    ```
    go run ./cmd/decompose/decompose.go --policy-collection-dir=/tmp/policy-collection --out=/tmp/c2p-output
    ```
1. Decomposed policy resources are ouput in `/tmp/c2p-output/decomposed/resources`
    ```
    $ tree -L 1 /tmp/c2p-output/decomposed
    /tmp/c2p-output/decomposed
    ├── _sources
    └── resources
    ```
    Individual decomposed resource contains k8s manifests and configuration files (policy-generator.yaml and kustomization.yaml) for PolicyGenerator. 
    ```
    $ tree -L 3 /tmp/c2p-output/decomposed/resources
    /tmp/c2p-output/decomposed/resources
    ├── add-chrony
    │   ├── add-chrony-worker
    │   │   └── MachineConfig.50-worker-chrony.0.yaml
    │   ├── kustomization.yaml
    │   └── policy-generator.yaml
    ├── add-tvk-license
    │   ├── add-tvk-license
    │   │   └── License.triliovault-license.0.yaml
    │   ├── kustomization.yaml
    ```
## C2P Composer
Compose OCM Policy from policy resources from compliance information (for example, [compliance.yaml](cmd/compose/compliance.yaml))

1. Run C2P Composer
    ```
    go run cmd/compose-by-c2pcr/main.go --c2pcr ./cmd/compose-by-c2pcr/c2pcr.yaml --out /tmp/c2p-output
    ```
1. Composed OCM policies are output in `/tmp/c2p-output`
    ```
    $ tree /tmp/c2p-output                                                                             
    /tmp/c2p-output
    ├── add-chrony
    │   ├── add-chrony-worker
    │   │   └── MachineConfig.50-worker-chrony.0.yaml
    │   ├── kustomization.yaml
    │   └── policy-generator.yaml
    ├── install-odf-lvm-operator
    │   ├── kustomization.yaml
    │   ├── odf-lvmcluster
    │   │   └── LVMCluster.odf-lvmcluster.0.yaml
    │   ├── policy-generator.yaml
    │   └── policy-odf-lvm-operator
    │       ├── Namespace.openshift-storage.0.yaml
    │       ├── OperatorGroup.openshift-storage-operatorgroup.0.yaml
    │       └── Subscription.lvm-operator.0.yaml
    ├── kustomization.yaml
    ├── policy-generator.yaml
    └── policy-sets.yaml
    ```

## C2P Controller
1. Build image
    ```
    make docker-build docker-push IMG=<controller image>
    ```
1. Create KinD cluster
    ```
    kind create cluster
    ```
1. Install (if you use OCM, install-ocm-related-crds may fail since the required CRDs are already there.)
    ```
    make install
    make install-ocm-related-crds
    ```
1. Deploy
    ```
    make deploy IMG=<controller image>
    ```
1. Create CR
    ```
    kubectl apply -f ./config/samples/compliance-to-policy_v1alpha1_compliancedeployment.yaml -n compliance-to-policy-system 
    ```
1. Check if Policy, PlacmenetBinding/Rule are created
    ```
    kubectl get policies,placementbindings,placementrules -n compliance-high
    ```
1. Cleanup
    ```
    make undeploy
    make uninstall
    ```
