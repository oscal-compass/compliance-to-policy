## Plugin for OCM

#### Prerequisite
- Install KinD and setup Open Cluster Management Hub cluster and managed clusters ([Setup OCM](#setup-ocm))

#### Example usage of C2P

1. Generate OCM Policy (C2P Compliance to Policy)
    ```
    python samples_public/ocm/compliance_to_policy.py -o /tmp/deliverable-policy
    ```
    E.g.
    ```
    $ python samples_public/ocm/compliance_to_policy.py 

    tree /tmp/deliverable-policy
    parameters.yaml
    policy-high-scan
    - compliance-high-scan
       - ScanSettingBinding.high.0.yaml
    - policy-generator.yaml
    - kustomization.yaml
    - compliance-suite-high
       - ComplianceSuite.high.0.yaml
    - compliance-suite-high-results
       - ComplianceCheckResult.noname.0.yaml
    policy-deployment
    - policy-generator.yaml
    - kustomization.yaml
    - policy-nginx-deployment
       - Deployment.nginx-deployment.0.yaml
    policy-disallowed-roles
    - policy-disallowed-roles-sample-role
       - Role.noname.0.yaml
    - policy-generator.yaml
    - kustomization.yaml
    policy-generator.yaml
    ```
1. Deploy the generated policies
    ```
    kustomize build --enable-alpha-plugins /tmp/deliverable-policy | kubectl apply -f -
    ```
    E.g.
    ```
    $ kubectl apply -R -f /tmp/deliverable-policy
    namespace/platform created
    configmap/baseimages created
    Warning: Validation failure actions enforce/audit are deprecated, use Enforce/Audit instead.
    clusterpolicy.kyverno.io/allowed-base-images created
    clusterpolicy.kyverno.io/disallow-capabilities created
    ```
1. Check policy statuses at hub cluster
    ```
    $ kubectl get policy -A
    NAMESPACE   NAME                          REMEDIATION ACTION   COMPLIANCE STATE   AGE
    c2p         policy-deployment             inform               NonCompliant       55s
    c2p         policy-disallowed-roles       inform               Compliant          55s
    c2p         policy-high-scan              inform               NonCompliant       55s
    cluster1    c2p.policy-deployment         inform               NonCompliant       54s
    cluster1    c2p.policy-disallowed-roles   inform               Compliant          54s
    cluster1    c2p.policy-high-scan          inform               NonCompliant       54s
    cluster2    c2p.policy-deployment         inform               NonCompliant       54s
    cluster2    c2p.policy-disallowed-roles   inform               Compliant          54s
    cluster2    c2p.policy-high-scan          inform               NonCompliant       54s
    ```
1. Collect policies as PVP Raw results
    ```
    kubectl get policy -A -o yaml > /tmp/policies.policy.open-cluster-management.io.yaml
    kubectl get placementdecisions -A -o yaml > /tmp/placementdecisions.cluster.open-cluster-management.io.yaml
    kubectl get policysets -A -o yaml > /tmp/policysets.policy.open-cluster-management.io.yaml
    ```
1. Generate Assessment Result (C2P Result to Compliance)
    ```
    python samples_public/ocm/result_to_compliance.py \
     -p /tmp/policies.policy.open-cluster-management.io.yaml \
     > /tmp/assessment_results.json
    ```
1. OSCAL Assessment Results is not human readable format. You can see the merged report in markdown by a quick viewer.
    ```
    c2p tools viewer -ar /tmp/assessment_results.json -cdef ./plugins_public/tests/data/ocm/component-definition.json -o /tmp/assessment_results.md
    ```
    ![assessment-results-md.ocm.jpg](/docs/public/images/assessment-results-md.ocm.jpg)

## Setup OCM
1. Prerequisite
    - kind
        ```
        $ kind version
        kind v0.19.0 go1.20.4 darwin/arm64
        ```
    - clusteradm
        ```
        $ clusteradm version
        client          version :v0.8.1-0-g3aea9c5
        server release  version :v1.26.0
        default bundle  version :0.13.1
        ```
1. Create 3 KinD clusters (hub, cluster1 and 2)
    ```
    kind create cluster --name hub --image kindest/node:v1.26.0 --wait 5m
    kind create cluster --name cluster1 --image kindest/node:v1.26.0 --wait 5m
    kind create cluster --name cluster2 --image kindest/node:v1.26.0 --wait 5m
    ```
1. Install OCM Hub
    ```
    kubectl config use-context kind-hub
    clusteradm init --wait
    ```
1. Join clusters to Hub
    ```
    kubectl config use-context kind-hub
    token=`clusteradm get token | head -n 1 | clusteradm get token | head -n 1 | cut -f 2 -d "="`
    server=`kubectl config view --minify -o=jsonpath='{.clusters[0].cluster.server}'`
    kubectl config use-context kind-cluster1
    clusteradm join --hub-token $token --hub-apiserver $server --cluster-name cluster1 --force-internal-endpoint-lookup --wait
    kubectl config use-context kind-cluster2
    clusteradm join --hub-token $token --hub-apiserver $server --cluster-name cluster2 --force-internal-endpoint-lookup --wait
    kubectl config use-context kind-hub
    clusteradm accept --clusters cluster2
    ```
1. Enable governance-policy-framework
    ```
    kubectl config use-context kind-hub
    clusteradm install hub-addon --names governance-policy-framework
    kubectl -n open-cluster-management wait deployment --all --for=condition=Available --timeout 3m
    ```
1. Deploy synchronization components to manages clusters
    ```
    kubectl config use-context kind-hub
    clusteradm addon enable --names governance-policy-framework --clusters cluster1,cluster2
    for c in cluster1 cluster2
    do
        kubectl -n $c wait managedclusteraddon --all --for=condition=Available --timeout 3m
    done
    ```
1. Deploy configuration policy controller to the managed cluster(s)
    ```
    kubectl config use-context kind-hub
    clusteradm addon enable --names config-policy-controller --clusters cluster1,cluster2
    for c in cluster1 cluster2
    do
        kubectl -n $c wait managedclusteraddon --all --for=condition=Available --timeout 3m
    done
    ```
1. Labeling "environment=dev" to managed clusters
    ```
    kubectl config use-context kind-hub
    for c in cluster1 cluster2
    do
        kubectl label managedcluster $c environment=dev
    done
    ```
1. Create managedclusterset
    ```
    kubectl config use-context kind-hub
    kubectl apply -f - << EOL
    apiVersion: cluster.open-cluster-management.io/v1beta2
    kind: ManagedClusterSet
    metadata:
      name: myclusterset
    spec:
      clusterSelector:
        labelSelector:
          matchExpressions:
          - key: environment
            operator: In
            values:
            - dev
        selectorType: LabelSelector
    EOL
    ```
1. Create "c2p" namespace and bind managed clusters to "c2p" namespace
    ```
    kubectl config use-context kind-hub
    kubectl create ns c2p
    clusteradm clusterset bind myclusterset --namespace c2p
    ```
1. The final cluster configuration
    ```
    $ clusteradm get clustersets
    <ManagedClusterSet> 
    └── <default> 
    │   ├── <BoundNamespace> 
    │   ├── <Status> 2 ManagedClusters selected
    │   ├── <Clusters> [cluster1 cluster2]
    └── <global> 
    │   ├── <BoundNamespace> 
    │   ├── <Status> 2 ManagedClusters selected
    │   ├── <Clusters> [cluster1 cluster2]
    └── <myclusterset> 
        └── <BoundNamespace> c2p
        └── <Status> 2 ManagedClusters selected
        └── <Clusters> [cluster1 cluster2]
    ```
