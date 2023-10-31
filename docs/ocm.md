## C2P for OCM

### Prerequisites
1. Install [Policy Generator Plugin](https://github.com/open-cluster-management-io/policy-generator-plugin#as-a-kustomize-plugin)

### C2P Decomposer
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
### C2P Composer
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

### Setup pipeline
1. Create two repositories (one is configuration repository that's used for pipeline from OSCAL to Policy and another is evidence repository that's used for pipeline from OCM statuses to Compliance result)
    - For example, c2p-for-ocm-pipeline01-config and c2p-for-ocm-pipeline01-evidence
1. Create Github Personal Access Token having following permissions
    - Repository permission of `Contents`, `Pull Requests`, and `Workflows` with read-and-write against both the configuration repository and the evidence repository.
1. Fork C2P repository (yana1205/compliance-to-policy.git) and checkout `template`
1. Set required parameters for github action to initialize your configuration and evidence repo
    1. Go to Settings tab
    1. Go to `Actions` under `Secrets and variables`
    1. Create `New repository secret`
        - Name: PAT
        - Secret: Created Github Personal Access Token  
    1. Go to `Variables` tab to create `New repository variable`
    1. Create `CONFIGURATION_REPOSITORY` variable
        - Name: CONFIGURATION_REPOSITORY
        - Value: `<configuration repository org>/<configuration repository name> (e.g. yana1205/c2p-for-ocm-pipeline01-config)`
    1. Create `EVIDENCE_REPOSITORY` variable
        - Name: EVIDENCE_REPOSITORY
        - Value: `<evidence repository org>/<evidence repository name> (e.g. yana1205/c2p-for-ocm-pipeline01-evidence)`
1. Run Action `Initialize repositories` with branch `template`
1. Go to the configuration repository and create `New repository secret`
    - Name: PAT
    - Secret: Created Github Personal Access Token
1. Go to the evidence repository and create `New repository secret`
    - Name: PAT
    - Secret: Created Github Personal Access Token

### Run oscal-to-pocliy
1. Go to the configuration repository
1. Go to `Actions` tab
1. Run `OSCAL to Policy`
    1. This action generates manifests from OSCAL and then generate a PR of changes for a directory `ocm-policy-manifests` containing the generated manifests.
1. Merge the PR

### Integrate with GitOps
1. Sync `ocm-policy-manifests` directory with your OCM Hub by OCM GitOps (OCM Channel and Subscription addon)

### Deploy collector to your OCM Hub
1. Apply RBAC for collector
    ```
    kubectl apply -f https://raw.githubusercontent.com/yana1205/compliance-to-policy/redesign.0622/scripts/collect/rbac.yaml
    ```
1. Create Secret for Github access
    ```
    kubectl -n c2p create secret generic --save-config collect-ocm-status-secret --from-literal=user=<github user> --from-literal=token=<github PAT> --from-literal=org=<evidence org name> --from-literal=repo=<evidence repo name>
    ```
    e.g.
    ```
    kubectl -n c2p create secret generic --save-config collect-ocm-status-secret --from-literal=user=yana1205 --from-literal=token=github_pat_xxx --from-literal=org=yana1205 --from-literal=repo=c2p-for-ocm-pipeline01-evidence
    ```
1. Deploy collector cronjob
    ```
    kubectl apply -f https://raw.githubusercontent.com/IBM/compliance-to-policy/main/scripts/collect/cronjob.yaml
    ```

### Cleanup
```
kubectl delete -f https://raw.githubusercontent.com/IBM/compliance-to-policy/main/scripts/collect/cronjob.yaml
kubectl -n c2p delete secret collect-ocm-status-secret 
kubectl delete -f https://raw.githubusercontent.com/IBM/compliance-to-policy/main/scripts/collect/rbac.yaml
```

## C2P as controller (deprecated)
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