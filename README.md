# compliance-to-policy template repository
Compliance-to-Policy (C2P) provides the framework to bridge the gap between compliance and policy administration.
This is the template repository used by C2P pipeline.

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
