## Plugin for Kyverno

#### Prerequisite
- Install KinD and Kyverno 1.10

#### Example usage of C2P

1. Generate Kyverno Policy (C2P Compliance to Policy)
    ```
    python samples_public/kyverno/compliance_to_policy.py -o /tmp/deliverable-policy
    ```
    E.g.
    ```
    $ python samples_public/kyverno/compliance_to_policy.py -o /tmp/deliverable-policy

    tree /tmp/deliverable-policy
    disallow-capabilities
    - disallow-capabilities.yaml
    allowed-base-images
    - 02-setup-cm.yaml
    - allowed-base-images.yaml
    ```
1. Deploy the generated policies
    ```
    kubectl apply -R -f /tmp/deliverable-policy
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
1. Check policy results
    ```
    $ kubectl get policyreport,clusterpolicyreport -A
    NAMESPACE            NAME                                                     PASS   FAIL   WARN   ERROR   SKIP   AGE
    kube-system          policyreport.wgpolicyk8s.io/cpol-allowed-base-images     0      12     0      0       0      19s
    kube-system          policyreport.wgpolicyk8s.io/cpol-disallow-capabilities   9      2      0      0       0      19s
    kyverno              policyreport.wgpolicyk8s.io/cpol-allowed-base-images     0      18     0      0       0      9s
    kyverno              policyreport.wgpolicyk8s.io/cpol-disallow-capabilities   18     0      0      0       0      9s
    local-path-storage   policyreport.wgpolicyk8s.io/cpol-allowed-base-images     0      3      0      0       0      16s
    local-path-storage   policyreport.wgpolicyk8s.io/cpol-disallow-capabilities   3      0      0      0       0      16s
    ```
1. Collect policy/cluster policy reports as PVP Raw results
    ```
    kubectl get policyreport -A -o yaml > /tmp/policyreports.wgpolicyk8s.io.yaml
    kubectl get clusterpolicyreport -o yaml > /tmp/clusterpolicyreports.wgpolicyk8s.io.yaml
    ```
1. Generate Assessment Result (C2P Result to Compliance)
    ```
    python samples_public/kyverno/result_to_compliance.py \
     -polr /tmp/policyreports.wgpolicyk8s.io.yaml \
     -cpolr /tmp/clusterpolicyreports.wgpolicyk8s.io.yaml \
     > /tmp/assessment_results.json
    ```
1. OSCAL Assessment Results is not human readable format. You can see the merged report in markdown by a quick viewer.
    ```
    c2p tools viewer -ar /tmp/assessment_results.json -cdef ./plugins_public/tests/data/kyverno/component-definition.json -o /tmp/assessment_results.md
    ```
    ![assessment-results-md.kyverno.jpg](/docs/public/images/assessment-results-md.kyverno.jpg)