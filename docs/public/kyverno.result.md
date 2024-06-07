

## Component: Managed Kubernetes


#### Result of control cm-2: 



Rule `allowed-base-images (Kyverno)`:
- Building images which specify a base as their origin is a good start to improving supply chain security, but over time organizations may want to build an allow list of specific base images which are allowed to be used when constructing containers. This policy ensures that a container's base, found in an OCI annotation, is in a cluster-wide allow list.

<details><summary>Details</summary>


  - Subject UUID: 9dd754f9-107f-4bef-a8e8-1f6b48e99c18
    - Title: v1/Pod kube-scheduler-kind-control-plane kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 9133c837-a326-4f97-9831-1b7aa07a06d6
    - Title: v1/Pod coredns-5d78c9869d-gc25q kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 6a95995d-6604-468c-a7a9-3fbe41659d86
    - Title: v1/Pod kindnet-pbb9l kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 0f8ace75-0260-4ec6-8c25-44d7465100dd
    - Title: v1/Pod etcd-kind-control-plane kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 16719080-499e-4c3e-9bc5-a6c23c91c289
    - Title: v1/Pod kube-apiserver-kind-control-plane kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 18249536-8126-46b3-8c71-704e4a0a8189
    - Title: v1/Pod kube-proxy-zbddb kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 331edbaa-0c13-4734-8048-c97a83326eed
    - Title: v1/Pod coredns-5d78c9869d-2rbnq kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 4e1947d2-24da-4544-83c6-4f7c9573ef9d
    - Title: v1/Pod kube-controller-manager-kind-control-plane kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: f831cd49-b9f3-4c37-922b-ec4b3fc83ac7
    - Title: apps/v1/DaemonSet kindnet kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: ad94a91f-944c-48bd-9d0d-a9917e9f36c6
    - Title: apps/v1/DaemonSet kube-proxy kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: fae04582-1086-4372-ba26-4d75b679e08e
    - Title: apps/v1/ReplicaSet coredns-5d78c9869d kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 4b029216-ab36-4aa7-b084-93dd38df5bc8
    - Title: apps/v1/Deployment coredns kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 02840a50-7e97-427b-8a84-556e8ba00502
    - Title: v1/Pod kyverno-admission-controller-7cd788c8dd-gdnhp kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 4323bb0e-42c8-4ff4-ad81-cfa63fbc7282
    - Title: v1/Pod kyverno-reports-controller-7f94855747-tmnhr kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: e52ffeb4-5607-40c0-8c67-ef343f4659f1
    - Title: v1/Pod kyverno-cleanup-admission-reports-28551310-cc4k7 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: d5341516-c240-4034-a3ae-47b3b0bb8efb
    - Title: v1/Pod kyverno-cleanup-cluster-admission-reports-28551310-m4ld4 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 952284cc-7796-4951-8af0-31fa92fef354
    - Title: v1/Pod kyverno-cleanup-controller-ddf458755-9bnlb kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 791320ee-d753-4a01-9cbb-46ee1290511b
    - Title: v1/Pod kyverno-background-controller-74599787cf-s6nm2 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 40ee8280-f09f-413f-a6e3-ce64f081c040
    - Title: apps/v1/Deployment kyverno-cleanup-controller kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: dca6df7a-db81-4cee-869e-7e44cc3b0f43
    - Title: batch/v1/Job kyverno-cleanup-cluster-admission-reports-28551310 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 66d7a8c5-cb68-474d-830d-c30f1da9928f
    - Title: apps/v1/ReplicaSet kyverno-admission-controller-7cd788c8dd kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 1d850dad-4bfe-4c73-95de-782ae6cc68d3
    - Title: batch/v1/Job kyverno-cleanup-admission-reports-28551310 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: f4020009-ed61-43c4-a602-8077bcad4f45
    - Title: apps/v1/Deployment kyverno-background-controller kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 85c061bd-9d9e-4d49-95ff-1c64b4d59ac5
    - Title: apps/v1/ReplicaSet kyverno-cleanup-controller-ddf458755 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 940821f2-80b2-4fcc-9b84-4bfe19d31701
    - Title: apps/v1/ReplicaSet kyverno-background-controller-74599787cf kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 101f7630-5f5e-42b8-821c-f9a5de4f302d
    - Title: apps/v1/Deployment kyverno-reports-controller kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 7cd83df2-24eb-4c30-bb2f-531a2c01cb10
    - Title: apps/v1/ReplicaSet kyverno-reports-controller-7f94855747 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 52c44c56-c8bb-483a-97cb-3ef1099122ee
    - Title: apps/v1/Deployment kyverno-admission-controller kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 6016287c-8669-4557-9a49-1740ac712b5f
    - Title: batch/v1/CronJob kyverno-cleanup-cluster-admission-reports kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 4d0ca9de-f788-4078-8965-8799f0cd8dca
    - Title: batch/v1/CronJob kyverno-cleanup-admission-reports kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: e3459a30-68dd-437d-a9d3-85094b81e599
    - Title: v1/Pod local-path-provisioner-6bc4bddd6b-vlmww local-path-storage
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 440394c6-0252-4522-adab-de16854e4363
    - Title: apps/v1/ReplicaSet local-path-provisioner-6bc4bddd6b local-path-storage
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: ab966e76-3149-4826-a4cf-62f2fb402faa
    - Title: apps/v1/Deployment local-path-provisioner local-path-storage
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```

</details>


---

#### Result of control cm-2.1: 



Rule `disallow-capabilities (Kyverno)`:
- Adding capabilities beyond those listed in the policy must be disallowed.

<details><summary>Details</summary>


  - Subject UUID: 2a364519-28c5-4826-87ad-76f827642ee7
    - Title: v1/Pod kube-scheduler-kind-control-plane kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: c7b43d71-f295-4f57-94a6-b10abc182317
    - Title: v1/Pod coredns-5d78c9869d-gc25q kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: a7006a2b-4f28-4252-b43c-2399b2e57141
    - Title: v1/Pod kindnet-pbb9l kube-system
    - Result: failure :x:
    - Reason:
      ```
      Any capabilities added beyond the allowed list (AUDIT_WRITE, CHOWN, DAC_OVERRIDE, FOWNER, FSETID, KILL, MKNOD, NET_BIND_SERVICE, SETFCAP, SETGID, SETPCAP, SETUID, SYS_CHROOT) are disallowed.
      ```


  - Subject UUID: 9dfefd31-4aba-41e2-9cd9-3efe0879d6bc
    - Title: v1/Pod etcd-kind-control-plane kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: a0d5f6c4-a18a-4a2b-9e43-42cafa584899
    - Title: v1/Pod kube-apiserver-kind-control-plane kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: bad997e0-fbb9-4f1f-bf86-7a4149f065fb
    - Title: v1/Pod kube-proxy-zbddb kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: ac0acce0-cc5c-4fe7-9e04-9d3e4b120ea3
    - Title: v1/Pod coredns-5d78c9869d-2rbnq kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 41198c0b-4e36-48c3-9525-6e13307571ae
    - Title: v1/Pod kube-controller-manager-kind-control-plane kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 66769f8f-f401-428e-be5e-291a717c157a
    - Title: apps/v1/DaemonSet kindnet kube-system
    - Result: failure :x:
    - Reason:
      ```
      Any capabilities added beyond the allowed list (AUDIT_WRITE, CHOWN, DAC_OVERRIDE, FOWNER, FSETID, KILL, MKNOD, NET_BIND_SERVICE, SETFCAP, SETGID, SETPCAP, SETUID, SYS_CHROOT) are disallowed.
      ```


  - Subject UUID: 32b5f436-404c-4ce7-b0eb-c7aeb4227b2d
    - Title: apps/v1/DaemonSet kube-proxy kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 71903de3-edd3-41c2-bcb1-548deac9cd66
    - Title: apps/v1/ReplicaSet coredns-5d78c9869d kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: d7b2e553-8484-4832-92c3-469b0ad87fdc
    - Title: apps/v1/Deployment coredns kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 6db9c88c-05e9-49f6-9b54-e5b313ae5af5
    - Title: v1/Pod kyverno-admission-controller-7cd788c8dd-gdnhp kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: e616181c-d6a8-48ff-adb9-596b5c0a9a39
    - Title: v1/Pod kyverno-reports-controller-7f94855747-tmnhr kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 80b6f761-5f94-411d-a0ad-e744204bff81
    - Title: v1/Pod kyverno-cleanup-admission-reports-28551310-cc4k7 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: e562bcec-aae8-4bf2-899a-676acc03f69d
    - Title: v1/Pod kyverno-cleanup-cluster-admission-reports-28551310-m4ld4 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 247cafd6-ee9f-425f-82ff-6f4a28cb588d
    - Title: v1/Pod kyverno-cleanup-controller-ddf458755-9bnlb kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: d119d142-4ece-4acb-b190-70178a1f83a9
    - Title: v1/Pod kyverno-background-controller-74599787cf-s6nm2 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: addb8d0e-d015-4246-be86-9c46a02215f0
    - Title: apps/v1/Deployment kyverno-cleanup-controller kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 5c586f26-b61b-42f5-9d50-a6ca5cb9ca5c
    - Title: batch/v1/Job kyverno-cleanup-cluster-admission-reports-28551310 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 90b60a63-c2fb-4c8c-ab63-5265b6181952
    - Title: apps/v1/ReplicaSet kyverno-admission-controller-7cd788c8dd kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 8dffcd26-1ba7-4cfd-b05b-e90f13fb12f3
    - Title: batch/v1/Job kyverno-cleanup-admission-reports-28551310 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 1310f07c-2441-4d5d-80dd-02f20d60ca3a
    - Title: apps/v1/Deployment kyverno-background-controller kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 33125904-5f8d-4738-876c-826f66762d2e
    - Title: apps/v1/ReplicaSet kyverno-cleanup-controller-ddf458755 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: f2edb624-4092-479d-bc2f-af130a193133
    - Title: apps/v1/ReplicaSet kyverno-background-controller-74599787cf kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: f6684749-6df8-42f1-b22f-6201c6f8159f
    - Title: apps/v1/Deployment kyverno-reports-controller kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 943621bd-aa43-440c-94ca-30abf2b6aed9
    - Title: apps/v1/ReplicaSet kyverno-reports-controller-7f94855747 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 40170b0e-57ad-4f65-8cd4-f39bb4f7d681
    - Title: apps/v1/Deployment kyverno-admission-controller kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: e3b87bce-ab5f-4d2f-8ed1-486130e38558
    - Title: batch/v1/CronJob kyverno-cleanup-cluster-admission-reports kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-cronjob-adding-capabilities' passed.
      ```


  - Subject UUID: 9730a3f3-1e77-4037-954a-38b2317974fd
    - Title: batch/v1/CronJob kyverno-cleanup-admission-reports kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-cronjob-adding-capabilities' passed.
      ```


  - Subject UUID: 0b35bb44-2732-4367-8581-524704d65672
    - Title: v1/Pod local-path-provisioner-6bc4bddd6b-vlmww local-path-storage
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: ce2b855d-2830-444b-8410-93fb01ca0d70
    - Title: apps/v1/ReplicaSet local-path-provisioner-6bc4bddd6b local-path-storage
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: b31b2ccb-68a3-4d93-9842-d33ea3f74809
    - Title: apps/v1/Deployment local-path-provisioner local-path-storage
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```

</details>


---


