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
    go run ./cmd/compose/compose.go --policy-resources-dir=/tmp/c2p-output/decomposed/resources --compliance-yaml=./cmd/compose/compliance.yaml --out=/tmp/c2p-output
    ```
1. Composed OCM policies are output in `/tmp/c2p-output/composed`
    ```
    $ tree -L 1 /tmp/c2p-output/composed  
    /tmp/c2p-output/composed
    ├── add-chrony.yaml
    └── install-odf-lvm-operator.yaml
    ```
1. If you want to see the intermidiate files to generate OCM Policy, please set `--temp-dir=<something to directory>` in the previous C2P Composer command.
    ```
    $ mkdir -p /tmp/c2p-temp
    $ go run ./cmd/compose/compose.go --policy-resources-dir=/tmp/c2p-output/decomposed/resources --compliance-yaml=./cmd/compose/compliance.yaml --out=/tmp/c2p-output --temp-dir=/tmp/c2p-temp
    $ tree -L 4 /tmp/c2p-temp
    /tmp/c2p-temp
    └── tmp-747478669
        └── CM-2 Baseline Configuration
            ├── add-chrony
            │   ├── kustomization.yaml
            │   ├── policy-generator.yaml
            │   └── resources
            └── install-odf-lvm-operator
                ├── kustomization.yaml
                ├── policy-generator.yaml
                └── resources
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
