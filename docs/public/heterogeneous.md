## Work on heterogeneous PVPs

Usecase of security checks against system (Github and Managed Kubernetes clusters) by multiple PVPs (Auditree, Kyverno, and OCM Policy).

![heterogeneous](https://github.com/oscal-compass/compliance-to-policy/assets/113283236/bb64f81a-986c-41fa-83c6-4e7e9165af76)

#### Steps
1. (Optional) Create OSCAL Component Defintion including multiple PVPs as validation components
    - [component-definition.csv](/plugins_public/tests/data/heterogeneous/component-definition.csv)
1. Generate PVP policies from the OSCAL Component Definition
    ```
    python samples_public/heterogeneous/compliance_to_policy.py \
      -c ./plugins_public/tests/data/heterogeneous/component-definition.json \
      -o ./policies
    ```
    1. Policies for each PVP are generated
        ```
        $ tree -L 2 policies
        policies
        ├── auditree
        │   └── auditree.json
        ├── kyverno
        │   ├── allowed-base-images
        │   └── disallow-capabilities
        └── ocm
            ├── kustomization.yaml
            ├── parameters.yaml
            ├── policy-deployment
            ├── policy-disallowed-roles
            ├── policy-generator.yaml
            └── policy-high-scan
        ```
1. (Optional) Collect policy validation results from system
    - Example all PVP results are located in [/plugins_public/tests/data](/plugins_public/tests/data).
1. Generate OSCAL Assessment Results from PVP results 
    ```
    python samples_public/heterogeneous/result_to_compliance.py \
      -c ./plugins_public/tests/data/heterogeneous/component-definition.json \
      -r ./plugins_public/tests/data > assessment-results.json
    ```
1. OSCAL Assessment Results is not human readable format. You can see the merged report in markdown by a quick viewer.
    ```
    c2p tools viewer \
      -cdef ./plugins_public/tests/data/heterogeneous/component-definition.json \
      -ar assessment-results.json
    ```
    e.g. [result.md](/docs/public/heterogeneous.result.md)