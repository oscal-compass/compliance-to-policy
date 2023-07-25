## Catalog
Electronic Version of NIST SP 800-53 Rev 5 Controls and SP 800-53A Rev 5 Assessment Procedures
## Component: Managed Kubernetes

Compliance status: NonCompliant

Checked controls: [ac-6,cm-2,cm-6,]
#### Result of control: cm-6
**Compliance status: NonCompliant**

Rules:
- Rule ID: test_configuration_check
- Policy ID: policy-high-scan
- Status: fail
- Reason:
```
- clusterName: cluster1
  complianceState: NonCompliant
  messages:
  - eventName: c2p.policy-high-scan.176f1dcdc2b51b01
    lastTimestamp: "2023-07-05T23:52:34Z"
    message: NonCompliant; violation - couldn't find mapping resource with kind ScanSettingBinding,
      please check if you have CRD deployed
  - eventName: c2p.policy-high-scan.176f1ddc44adf035
    lastTimestamp: "2023-07-05T23:53:37Z"
    message: NonCompliant; violation - couldn't find mapping resource with kind ComplianceSuite,
      please check if you have CRD deployed
  - eventName: c2p.policy-high-scan.176f1ddc441457e5
    lastTimestamp: "2023-07-05T23:53:37Z"
    message: NonCompliant; violation - couldn't find mapping resource with kind ComplianceCheckResult,
      please check if you have CRD deployed
- clusterName: cluster2
  complianceState: NonCompliant
  messages:
  - eventName: c2p.policy-high-scan.176f1dc3684f9eb6
    lastTimestamp: "2023-07-05T23:51:50Z"
    message: NonCompliant; violation - couldn't find mapping resource with kind ScanSettingBinding,
      please check if you have CRD deployed
  - eventName: c2p.policy-high-scan.176f1dc426d20948
    lastTimestamp: "2023-07-05T23:51:53Z"
    message: NonCompliant; violation - couldn't find mapping resource with kind ComplianceSuite,
      please check if you have CRD deployed
  - eventName: c2p.policy-high-scan.176f1dc4e29e1221
    lastTimestamp: "2023-07-05T23:51:56Z"
    message: NonCompliant; violation - couldn't find mapping resource with kind ComplianceCheckResult,
      please check if you have CRD deployed

```
---
#### Result of control: cm-2
**Compliance status: NonCompliant**

Rules:
- Rule ID: test_proxy_check
- Policy ID: policy-deployment
- Status: fail
- Reason:
```
- clusterName: cluster1
  complianceState: NonCompliant
  messages:
  - eventName: c2p.policy-deployment.176f1ddc5591cb1c
    lastTimestamp: "2023-07-05T23:53:37Z"
    message: 'NonCompliant; violation - deployments not found: [nginx-deployment]
      in namespace cluster1 missing; [nginx-deployment] in namespace kube-node-lease
      missing; [nginx-deployment] in namespace kube-public missing; [nginx-deployment]
      in namespace local-path-storage missing'
- clusterName: cluster2
  complianceState: NonCompliant
  messages:
  - eventName: c2p.policy-deployment.176f1dc4e7de17cb
    lastTimestamp: "2023-07-05T23:51:56Z"
    message: 'NonCompliant; violation - deployments not found: [nginx-deployment]
      in namespace cluster2 missing; [nginx-deployment] in namespace default missing;
      [nginx-deployment] in namespace kube-node-lease missing; [nginx-deployment]
      in namespace kube-public missing; [nginx-deployment] in namespace local-path-storage
      missing'

```
---
#### Result of control: ac-6
**Compliance status: Compliant**

Rules:
- Rule ID: test_rbac_check
- Policy ID: policy-disallowed-roles
- Status: pass
- Reason:
```
- clusterName: cluster1
  complianceState: Compliant
  messages:
  - eventName: c2p.policy-disallowed-roles.176f1dcdc4c8d17e
    lastTimestamp: "2023-07-05T23:52:34Z"
    message: Compliant; notification - roles in namespace cluster1; in namespace default;
      in namespace kube-node-lease; in namespace kube-public; in namespace local-path-storage
      missing as expected, therefore this Object template is compliant
- clusterName: cluster2
  complianceState: Compliant
  messages:
  - eventName: c2p.policy-disallowed-roles.176f1dc36e36b7b2
    lastTimestamp: "2023-07-05T23:51:50Z"
    message: Compliant; notification - roles in namespace cluster2; in namespace default;
      in namespace kube-node-lease; in namespace kube-public; in namespace local-path-storage
      missing as expected, therefore this Object template is compliant

```
---
