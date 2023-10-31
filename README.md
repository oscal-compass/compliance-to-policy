# compliance-to-policy
Compliance-to-Policy (C2P) provides the framework to bridge Compliance administration and Policy administration by [OSCAL](https://pages.nist.gov/OSCAL/). OSCAL (Open Security Controls Assessment Language) is a standardized framework developed by NIST for expressing and automating the assessment and management of security controls in machine-readable format (xml, json, yaml)

## Continuous Compliance by C2P 

https://github.com/IBM/compliance-to-policy/assets/113283236/da3518d0-53de-4bd6-8703-04ce94e9dfba

## Usage of C2P commands

### C2P for Kyverno
Prepare Kyverno Policy Resources
- You can use [policy-resources for test](/pkg/testdata/kyverno/policy-resources)
- You can load Kyverno Policy Resource from Kyverno Policies (https://github.com/kyverno/policies)
    1. Run `kyverno tools load-policy-resources` command
        ```
        $ go run cmd/c2pcli/main.go kyverno tools load-policy-resources --src https://github.com/kyverno/policies --dest /tmp/policies
        ```
        ```
        $ tree /tmp/policies
        /tmp/policies
        ├── add-apparmor-annotations
        │   └── add-apparmor-annotations.yaml
        ├── add-capabilities
        │   └── add-capabilities.yaml
        ├── add-castai-removal-disabled
        │   └── add-castai-removal-disabled.yaml
        ├── add-certificates-volume
        │   └── add-certificates-volume.yaml
        ├── add-default-resources
        ...
        ```
    - You can check result.json about what resources are downloaded.
        ```
        $ cat /tmp/policies/result.json

        ```
    - There are some policies that depend on context. Please add the context resources manually. result.json contains list of the policies that have context field
        ```
        $ jq -r .summary.resourcesHavingContext /tmp/policies/result.json
        [
            "allowed-podpriorities",
            "allowed-base-images",
            "advanced-restrict-image-registries",
            ...
            "require-linkerd-server"
        ]
        ```
#### Convert OSCAL to Kyverno Policy
```
$ go run cmd/c2pcli/main.go kyverno oscal2policy -c ./pkg/testdata/kyverno/c2p-config.yaml -o /tmp/kyverno-policies
2023-10-31T07:23:56.291+0900    INFO    kyverno/c2pcr   kyverno/configparser.go:53      Component-definition is loaded from ./pkg/testdata/kyverno/component-definition.json

$ tree /tmp/kyverno-policies 
/tmp/kyverno-policies
└── allowed-base-images
    ├── 02-setup-cm.yaml
    └── allowed-base-images.yaml
```

#### Convert Policy Report to OSCAL Assessment Results
```
$ go run cmd/c2pcli/main.go kyverno result2oscal -c ./pkg/testdata/kyverno/c2p-config.yaml -o /tmp/assessment-results

$ tree /tmp/assessment-results 
/tmp/assessment-results
└── assessment-results.json
```

Reformat in human friendly format (markdown file) since OSCAL is not machine friendly format.
```
$ go run cmd/c2pcli/main.go kyverno oscal2posture -c ./pkg/testdata/kyverno/c2p-config.yaml --assessment-results /tmp/assessment-results/assessment-results.json -o /tmp/compliance-report.md
```

```
$ head -n 15 /tmp/compliance-report.md
## Catalog

## Component: Kubernetes
#### Result of control: cm-8.3_smt.a

Rule ID: allowed-base-images
<details><summary>Details</summary>

  - Subject UUID: 0b1adf1c-f6e2-46af-889e-39255e669655
    - Title: ApiVersion: v1, Kind: Pod, Namespace: argocd, Name: argocd-application-controller-0
    - Result: fail
    - Reason:
      ```
      validation failure: This container image&#39;s base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```
```

### C2P for Open Cluster Management (OCM)
OCM has Policy Governance Framework, where the policy is OCM Policy and the PVP audit result is status of deployed OCM Policy.

#### Convert OSCAL Component Definition to OCM Policy
TBD
#### Convert OCM Policy Status to OSCAL Assessment Results
TBD

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

---
## Utilities
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
