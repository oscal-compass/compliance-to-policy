# compliance-to-policy
Compliance-to-Policy (C2P) provides the framework to bridge Compliance administration and Policy administration by [OSCAL](https://pages.nist.gov/OSCAL/). OSCAL (Open Security Controls Assessment Language) is a standardized framework developed by NIST for expressing and automating the assessment and management of security controls in machine-readable format (xml, json, yaml)

## Continuous Compliance by C2P 

https://github.com/IBM/compliance-to-policy/assets/113283236/4b0b5357-4025-46c8-8d88-1f4c00538795

## Usage of C2P commands

### C2P for Kyverno
Prepare Kyverno Policy Resources
- You can use [policy-resources for test](/pkg/testdata/kyverno/policy-resources)
- For bring your own policies, please see [Bring your own Kyverno Policy Resources](#bring-your-own-kyverno-policy-resources)

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

#### Reformat in human-friendly format (markdown file)
```
$ go run cmd/c2pcli/main.go kyverno tools oscal2posture -c ./pkg/testdata/kyverno/c2p-config.yaml --assessment-results /tmp/assessment-results/assessment-results.json -o /tmp/compliance-report.md
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

### Bring your own Kyverno Policy Resources
- You can download Kyverno Policies (https://github.com/kyverno/policies) as Policy Resources and modify them
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

## Build at local
```
goreleaser release --snapshot --clean
```

## Test
```
make test-pkg
```