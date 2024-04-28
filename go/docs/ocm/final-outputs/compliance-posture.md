## Catalog
Electronic Version of NIST SP 800-53 Rev 5.1.1 Controls and SP 800-53A Rev 5.1.1 Assessment Procedures
## Component: Managed Kubernetes
#### Result of control: cm-6

Rule ID: test_configuration_check
<details><summary>Details</summary>

  - Subject UUID: 6fade0d6-93fc-11ee-a029-62f79297f1b7
    - Title: Cluster Name: cluster1
    - Result: fail
    - Reason:
      ```
      - eventName: c2p.policy-high-scan.179e2849d01e8567
        lastTimestamp: &#34;2023-12-06T05:53:26Z&#34;
        message: NonCompliant; violation - couldn&#39;t find mapping resource with kind ScanSettingBinding,
          please check if you have CRD deployed
      - eventName: c2p.policy-high-scan.179e2848595f9ba9
        lastTimestamp: &#34;2023-12-06T05:53:20Z&#34;
        message: NonCompliant; violation - couldn&#39;t find mapping resource with kind ComplianceSuite,
          please check if you have CRD deployed
      - eventName: c2p.policy-high-scan.179e284a97812778
        lastTimestamp: &#34;2023-12-06T05:53:30Z&#34;
        message: NonCompliant; violation - couldn&#39;t find mapping resource with kind ComplianceCheckResult,
          please check if you have CRD deployed
      
      ```

  - Subject UUID: 6fade374-93fc-11ee-a029-62f79297f1b7
    - Title: Cluster Name: cluster2
    - Result: fail
    - Reason:
      ```
      - eventName: c2p.policy-high-scan.179e284863bfbfab
        lastTimestamp: &#34;2023-12-06T05:53:20Z&#34;
        message: NonCompliant; violation - couldn&#39;t find mapping resource with kind ScanSettingBinding,
          please check if you have CRD deployed
      - eventName: c2p.policy-high-scan.179e284a53812e10
        lastTimestamp: &#34;2023-12-06T05:53:28Z&#34;
        message: NonCompliant; violation - couldn&#39;t find mapping resource with kind ComplianceSuite,
          please check if you have CRD deployed
      - eventName: c2p.policy-high-scan.179e2849950d51e5
        lastTimestamp: &#34;2023-12-06T05:53:25Z&#34;
        message: NonCompliant; violation - couldn&#39;t find mapping resource with kind ComplianceCheckResult,
          please check if you have CRD deployed
      
      ```
</details>


Rule ID: install_kyverno
<details><summary>Details</summary>

  - Subject UUID: 6fade0d6-93fc-11ee-a029-62f79297f1b7
    - Title: Cluster Name: cluster1
    - Result: pass
    - Reason:
      ```
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284be703d42d
        lastTimestamp: &#34;2023-12-06T05:53:35Z&#34;
        message: Compliant; notification - clusterroles [kyverno] found as specified, therefore
          this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284c7ace2ebf
        lastTimestamp: &#34;2023-12-06T05:53:38Z&#34;
        message: Compliant; notification - clusterroles [kyverno:admin-generaterequest]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284c9f4c379e
        lastTimestamp: &#34;2023-12-06T05:53:38Z&#34;
        message: Compliant; notification - clusterroles [kyverno:admin-policies] found as
          specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e28504ffc7000
        lastTimestamp: &#34;2023-12-06T05:53:54Z&#34;
        message: Compliant; notification - clusterroles [kyverno:admin-policyreport] found
          as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284befa43976
        lastTimestamp: &#34;2023-12-06T05:53:35Z&#34;
        message: Compliant; notification - clusterroles [kyverno:admin-reports] found as
          specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284e6ff55461
        lastTimestamp: &#34;2023-12-06T05:53:46Z&#34;
        message: Compliant; notification - clusterroles [kyverno:admin-updaterequest] found
          as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e285349215bae
        lastTimestamp: &#34;2023-12-06T05:54:07Z&#34;
        message: Compliant; notification - clusterroles [kyverno:events] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284d380ed6df
        lastTimestamp: &#34;2023-12-06T05:53:41Z&#34;
        message: Compliant; notification - clusterroles [kyverno:generate] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2853548333b1
        lastTimestamp: &#34;2023-12-06T05:54:07Z&#34;
        message: Compliant; notification - clusterroles [kyverno:policies] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284f80c03d5d
        lastTimestamp: &#34;2023-12-06T05:53:51Z&#34;
        message: Compliant; notification - clusterroles [kyverno:userinfo] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284ba3c9f6ea
        lastTimestamp: &#34;2023-12-06T05:53:34Z&#34;
        message: Compliant; notification - clusterroles [kyverno:view] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e28535e612839
        lastTimestamp: &#34;2023-12-06T05:54:07Z&#34;
        message: Compliant; notification - clusterroles [kyverno:webhook] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284fcb71f4df
        lastTimestamp: &#34;2023-12-06T05:53:52Z&#34;
        message: Compliant; notification - clusterrolebindings [kyverno] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e285680fe6ac8
        lastTimestamp: &#34;2023-12-06T05:54:21Z&#34;
        message: Compliant; notification - configmaps [kyverno-metrics] in namespace kyverno
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e28524f386c75
        lastTimestamp: &#34;2023-12-06T05:54:03Z&#34;
        message: Compliant; notification - configmaps [kyverno] in namespace kyverno found
          as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284bf8f4b48b
        lastTimestamp: &#34;2023-12-06T05:53:35Z&#34;
        message: Compliant; notification - customresourcedefinitions [admissionreports.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284bba53764a
        lastTimestamp: &#34;2023-12-06T05:53:34Z&#34;
        message: Compliant; notification - customresourcedefinitions [backgroundscanreports.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284fd5dac2d0
        lastTimestamp: &#34;2023-12-06T05:53:52Z&#34;
        message: Compliant; notification - customresourcedefinitions [clusteradmissionreports.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284f91c7ac2d
        lastTimestamp: &#34;2023-12-06T05:53:51Z&#34;
        message: Compliant; notification - customresourcedefinitions [clusterbackgroundscanreports.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2853336f0121
        lastTimestamp: &#34;2023-12-06T05:54:07Z&#34;
        message: Compliant; notification - customresourcedefinitions [clusterpolicies.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2853ae517829
        lastTimestamp: &#34;2023-12-06T05:54:09Z&#34;
        message: Compliant; notification - customresourcedefinitions [clusterpolicyreports.wgpolicyk8s.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2853bb96886f
        lastTimestamp: &#34;2023-12-06T05:54:09Z&#34;
        message: Compliant; notification - customresourcedefinitions [generaterequests.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e28567b4ab101
        lastTimestamp: &#34;2023-12-06T05:54:21Z&#34;
        message: Compliant; notification - customresourcedefinitions [policies.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e285155b5a7dc
        lastTimestamp: &#34;2023-12-06T05:53:58Z&#34;
        message: Compliant; notification - customresourcedefinitions [policyreports.wgpolicyk8s.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e285258c5e5f7
        lastTimestamp: &#34;2023-12-06T05:54:03Z&#34;
        message: Compliant; notification - customresourcedefinitions [updaterequests.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2856c0192717
        lastTimestamp: &#34;2023-12-06T05:54:22Z&#34;
        message: Compliant; notification - deployments [kyverno] in namespace kyverno found
          as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2851d0ebe5d4
        lastTimestamp: &#34;2023-12-06T05:54:01Z&#34;
        message: Compliant; notification - namespaces [kyverno] found as specified, therefore
          this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2851d422c597
        lastTimestamp: &#34;2023-12-06T05:54:01Z&#34;
        message: Compliant; notification - roles [kyverno:leaderelection] in namespace kyverno
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e28526cb03284
        lastTimestamp: &#34;2023-12-06T05:54:03Z&#34;
        message: Compliant; notification - rolebindings [kyverno:leaderelection] in namespace
          kyverno found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e285684956927
        lastTimestamp: &#34;2023-12-06T05:54:21Z&#34;
        message: Compliant; notification - services [kyverno-svc-metrics] in namespace kyverno
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e285277e643d2
        lastTimestamp: &#34;2023-12-06T05:54:03Z&#34;
        message: Compliant; notification - services [kyverno-svc] in namespace kyverno found
          as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2855b441709c
        lastTimestamp: &#34;2023-12-06T05:54:17Z&#34;
        message: Compliant; notification - serviceaccounts [kyverno] in namespace kyverno
          found as specified, therefore this Object template is compliant
      
      ```

  - Subject UUID: 6fade374-93fc-11ee-a029-62f79297f1b7
    - Title: Cluster Name: cluster2
    - Result: pass
    - Reason:
      ```
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284c9a97c784
        lastTimestamp: &#34;2023-12-06T05:53:38Z&#34;
        message: Compliant; notification - clusterroles [kyverno] found as specified, therefore
          this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284b35375584
        lastTimestamp: &#34;2023-12-06T05:53:32Z&#34;
        message: Compliant; notification - clusterroles [kyverno:admin-generaterequest]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284ef4862c6a
        lastTimestamp: &#34;2023-12-06T05:53:48Z&#34;
        message: Compliant; notification - clusterroles [kyverno:admin-policies] found as
          specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284dfd646310
        lastTimestamp: &#34;2023-12-06T05:53:44Z&#34;
        message: Compliant; notification - clusterroles [kyverno:admin-policyreport] found
          as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284dc9cf5a21
        lastTimestamp: &#34;2023-12-06T05:53:43Z&#34;
        message: Compliant; notification - clusterroles [kyverno:admin-reports] found as
          specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284b3482ad78
        lastTimestamp: &#34;2023-12-06T05:53:32Z&#34;
        message: Compliant; notification - clusterroles [kyverno:admin-updaterequest] found
          as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284f07938a0b
        lastTimestamp: &#34;2023-12-06T05:53:49Z&#34;
        message: Compliant; notification - clusterroles [kyverno:events] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284f17ff00f6
        lastTimestamp: &#34;2023-12-06T05:53:49Z&#34;
        message: Compliant; notification - clusterroles [kyverno:generate] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284ca95ae428
        lastTimestamp: &#34;2023-12-06T05:53:38Z&#34;
        message: Compliant; notification - clusterroles [kyverno:policies] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284e30919d74
        lastTimestamp: &#34;2023-12-06T05:53:45Z&#34;
        message: Compliant; notification - clusterroles [kyverno:userinfo] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284e4527ba38
        lastTimestamp: &#34;2023-12-06T05:53:45Z&#34;
        message: Compliant; notification - clusterroles [kyverno:view] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284c5820f7b0
        lastTimestamp: &#34;2023-12-06T05:53:37Z&#34;
        message: Compliant; notification - clusterroles [kyverno:webhook] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284dccc3cae5
        lastTimestamp: &#34;2023-12-06T05:53:43Z&#34;
        message: Compliant; notification - clusterrolebindings [kyverno] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e285889d4069c
        lastTimestamp: &#34;2023-12-06T05:54:29Z&#34;
        message: Compliant; notification - configmaps [kyverno-metrics] in namespace kyverno
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2853e3c830c7
        lastTimestamp: &#34;2023-12-06T05:54:09Z&#34;
        message: Compliant; notification - configmaps [kyverno] in namespace kyverno found
          as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2851c8b54cb1
        lastTimestamp: &#34;2023-12-06T05:54:00Z&#34;
        message: Compliant; notification - customresourcedefinitions [admissionreports.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284dda99ee7e
        lastTimestamp: &#34;2023-12-06T05:53:44Z&#34;
        message: Compliant; notification - customresourcedefinitions [backgroundscanreports.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e284cbfda9c70
        lastTimestamp: &#34;2023-12-06T05:53:39Z&#34;
        message: Compliant; notification - customresourcedefinitions [clusteradmissionreports.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2854d3f7806a
        lastTimestamp: &#34;2023-12-06T05:54:13Z&#34;
        message: Compliant; notification - customresourcedefinitions [clusterbackgroundscanreports.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2856e1ae7593
        lastTimestamp: &#34;2023-12-06T05:54:22Z&#34;
        message: Compliant; notification - customresourcedefinitions [clusterpolicies.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2851eb3a6bea
        lastTimestamp: &#34;2023-12-06T05:54:01Z&#34;
        message: Compliant; notification - customresourcedefinitions [clusterpolicyreports.wgpolicyk8s.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2851282fb972
        lastTimestamp: &#34;2023-12-06T05:53:58Z&#34;
        message: Compliant; notification - customresourcedefinitions [generaterequests.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2851fa561a0e
        lastTimestamp: &#34;2023-12-06T05:54:01Z&#34;
        message: Compliant; notification - customresourcedefinitions [policies.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2854e6af9e7a
        lastTimestamp: &#34;2023-12-06T05:54:14Z&#34;
        message: Compliant; notification - customresourcedefinitions [policyreports.wgpolicyk8s.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e285334bd6234
        lastTimestamp: &#34;2023-12-06T05:54:07Z&#34;
        message: Compliant; notification - customresourcedefinitions [updaterequests.kyverno.io]
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e28588a2bd302
        lastTimestamp: &#34;2023-12-06T05:54:29Z&#34;
        message: Compliant; notification - deployments [kyverno] in namespace kyverno found
          as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2851d0e23da9
        lastTimestamp: &#34;2023-12-06T05:54:01Z&#34;
        message: Compliant; notification - namespaces [kyverno] found as specified, therefore
          this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2853398624f1
        lastTimestamp: &#34;2023-12-06T05:54:07Z&#34;
        message: Compliant; notification - roles [kyverno:leaderelection] in namespace kyverno
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2858b1db4b8e
        lastTimestamp: &#34;2023-12-06T05:54:30Z&#34;
        message: Compliant; notification - rolebindings [kyverno:leaderelection] in namespace
          kyverno found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2857bc2aa4bd
        lastTimestamp: &#34;2023-12-06T05:54:26Z&#34;
        message: Compliant; notification - services [kyverno-svc-metrics] in namespace kyverno
          found as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e285346ad8e3c
        lastTimestamp: &#34;2023-12-06T05:54:07Z&#34;
        message: Compliant; notification - services [kyverno-svc] in namespace kyverno found
          as specified, therefore this Object template is compliant
      - eventName: c2p.policy-install-kyverno-from-manifests.179e2857c388cf77
        lastTimestamp: &#34;2023-12-06T05:54:26Z&#34;
        message: Compliant; notification - serviceaccounts [kyverno] in namespace kyverno
          found as specified, therefore this Object template is compliant
      
      ```
</details>


Rule ID: test_required_label
<details><summary>Details</summary>

  - Subject UUID: 6fade0d6-93fc-11ee-a029-62f79297f1b7
    - Title: Cluster Name: cluster1
    - Result: fail
    - Reason:
      ```
      - eventName: c2p.policy-kyverno-require-labels.179e2851c11fe04c
        lastTimestamp: &#34;2023-12-06T05:54:00Z&#34;
        message: Compliant; notification - clusterpolicies [require-labels] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-kyverno-require-labels.179e2862688eaee7
        lastTimestamp: &#34;2023-12-06T05:55:12Z&#34;
        message: &#39;NonCompliant; violation - policyreports found: [cpol-require-labels] in
          namespace local-path-storage&#39;
      
      ```

  - Subject UUID: 6fade374-93fc-11ee-a029-62f79297f1b7
    - Title: Cluster Name: cluster2
    - Result: fail
    - Reason:
      ```
      - eventName: c2p.policy-kyverno-require-labels.179e2855f5ab92dd
        lastTimestamp: &#34;2023-12-06T05:54:18Z&#34;
        message: Compliant; notification - clusterpolicies [require-labels] found as specified,
          therefore this Object template is compliant
      - eventName: c2p.policy-kyverno-require-labels.179e2862e1802d28
        lastTimestamp: &#34;2023-12-06T05:55:14Z&#34;
        message: &#39;NonCompliant; violation - policyreports found: [cpol-require-labels] in
          namespace local-path-storage&#39;
      
      ```
</details>

---
#### Result of control: cm-2

Rule ID: test_proxy_check
<details><summary>Details</summary>

  - Subject UUID: 6fade0d6-93fc-11ee-a029-62f79297f1b7
    - Title: Cluster Name: cluster1
    - Result: fail
    - Reason:
      ```
      - eventName: c2p.policy-deployment.179e284f776397b3
        lastTimestamp: &#34;2023-12-06T05:53:50Z&#34;
        message: &#39;NonCompliant; violation - deployments not found: [nginx-deployment] in
          namespace cluster1 missing; [nginx-deployment] in namespace default missing; [nginx-deployment]
          in namespace kube-node-lease missing; [nginx-deployment] in namespace kube-public
          missing; [nginx-deployment] in namespace kyverno missing; [nginx-deployment] in
          namespace local-path-storage missing&#39;
      
      ```

  - Subject UUID: 6fade374-93fc-11ee-a029-62f79297f1b7
    - Title: Cluster Name: cluster2
    - Result: fail
    - Reason:
      ```
      - eventName: c2p.policy-deployment.179e2854bed6d22e
        lastTimestamp: &#34;2023-12-06T05:54:13Z&#34;
        message: &#39;NonCompliant; violation - deployments not found: [nginx-deployment] in
          namespace cluster2 missing; [nginx-deployment] in namespace default missing; [nginx-deployment]
          in namespace kube-node-lease missing; [nginx-deployment] in namespace kube-public
          missing; [nginx-deployment] in namespace kyverno missing; [nginx-deployment] in
          namespace local-path-storage missing&#39;
      
      ```
</details>

---
