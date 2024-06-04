

## Component: GitHub


#### Result of control cm-2: 


---

#### Result of control ac-2: 



Rule `demo_examples.checks.test_github.GitHubOrgs.test_members_is_not_empty (Auditree)`:
- Check whether the GitHub org is not empty.

<details><summary>Details</summary>


  - Subject UUID: ae5c2bac-47be-4734-b847-beaad450a76e
    - Title: Auditree Check: demo_examples.checks.test_github.GitHubOrgs.test_members_is_not_empty_0_nasa
    - Result: pass :white_check_mark:
    - Reason:
      ```
      {}
      ```


  - Subject UUID: adc99d7c-b9fd-4d31-961f-c076635f2d53
    - Title: Auditree Check: demo_examples.checks.test_github.GitHubOrgs.test_members_is_not_empty_1_esa
    - Result: pass :white_check_mark:
    - Reason:
      ```
      {}
      ```

</details>


---


## Component: Managed Kubernetes


#### Result of control cm-2: 



Rule `allowed-base-images (Kyverno)`:
- Building images which specify a base as their origin is a good start to improving supply chain security, but over time organizations may want to build an allow list of specific base images which are allowed to be used when constructing containers. This policy ensures that a container's base, found in an OCI annotation, is in a cluster-wide allow list.

<details><summary>Details</summary>


  - Subject UUID: 70057a02-062f-4fb7-9dff-6407e633e4a1
    - Title: v1/Pod kube-scheduler-kind-control-plane kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 03b044ca-6739-41cb-9c9e-038db3e48b9f
    - Title: v1/Pod coredns-5d78c9869d-gc25q kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: b393c6ea-dce4-496c-bc65-52ae123564a5
    - Title: v1/Pod kindnet-pbb9l kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 3d51ff83-c708-484c-8bdb-858ca48d14d3
    - Title: v1/Pod etcd-kind-control-plane kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 499afc36-5aa5-4d3b-8cac-b7a58043e14a
    - Title: v1/Pod kube-apiserver-kind-control-plane kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: f770e021-b3a7-4699-bca7-58cf6c70bb14
    - Title: v1/Pod kube-proxy-zbddb kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 321d5475-59c1-4b77-98fc-624f23b0deba
    - Title: v1/Pod coredns-5d78c9869d-2rbnq kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 037a870a-6256-4b0f-8d2b-70e3bb877c6c
    - Title: v1/Pod kube-controller-manager-kind-control-plane kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 79f831d8-7b72-47a1-8f44-9fe9835caea9
    - Title: apps/v1/DaemonSet kindnet kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 906d7d99-2eae-461e-9def-5557e4c488ca
    - Title: apps/v1/DaemonSet kube-proxy kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: ab72cc56-7c1f-42aa-9d96-3cafd843b464
    - Title: apps/v1/ReplicaSet coredns-5d78c9869d kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 6672c835-890d-4c68-a271-321ec432d26f
    - Title: apps/v1/Deployment coredns kube-system
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: cbc28d59-713f-4b25-9dc3-d6cedd0eb8cf
    - Title: v1/Pod kyverno-admission-controller-7cd788c8dd-gdnhp kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: be053523-a394-4fef-956f-c4fdbe5b841e
    - Title: v1/Pod kyverno-reports-controller-7f94855747-tmnhr kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: a753ba1d-df9e-4c4a-b0a9-6e435da74b9d
    - Title: v1/Pod kyverno-cleanup-admission-reports-28551310-cc4k7 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: d696fc41-1f67-4dd7-a1b4-1df455a2c607
    - Title: v1/Pod kyverno-cleanup-cluster-admission-reports-28551310-m4ld4 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 2ca0e393-5de6-49b4-bcb0-c0a98f89bace
    - Title: v1/Pod kyverno-cleanup-controller-ddf458755-9bnlb kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: e02c3cb0-69b9-4947-b4cb-5b9447194bf0
    - Title: v1/Pod kyverno-background-controller-74599787cf-s6nm2 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 3eac92cd-05b3-4340-9d2d-ac06509b6aba
    - Title: apps/v1/Deployment kyverno-cleanup-controller kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 32965aa3-0813-4d49-8b52-d6623764c16e
    - Title: batch/v1/Job kyverno-cleanup-cluster-admission-reports-28551310 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: b009839e-1f23-44b1-bae1-61e2f1177613
    - Title: apps/v1/ReplicaSet kyverno-admission-controller-7cd788c8dd kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 05e47efc-7798-4087-a007-4a0f1b0ee925
    - Title: batch/v1/Job kyverno-cleanup-admission-reports-28551310 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 019c7fa5-e4ac-43a2-93d4-f31be5e894d2
    - Title: apps/v1/Deployment kyverno-background-controller kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: a1853c39-c418-4f76-aa2b-2d19f0705cf4
    - Title: apps/v1/ReplicaSet kyverno-cleanup-controller-ddf458755 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 7dee76f1-87bd-4125-af7f-b16b3940ee07
    - Title: apps/v1/ReplicaSet kyverno-background-controller-74599787cf kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 5dc41d4d-f497-4f9c-8e0a-4514a990e30b
    - Title: apps/v1/Deployment kyverno-reports-controller kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 583c7bb1-3b9a-4dc2-a820-e17c525fcfe9
    - Title: apps/v1/ReplicaSet kyverno-reports-controller-7f94855747 kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 09b9733c-fd96-43ba-a32f-fa0374341aa0
    - Title: apps/v1/Deployment kyverno-admission-controller kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: e91e2d61-5d80-4d78-81dd-6295518904a3
    - Title: batch/v1/CronJob kyverno-cleanup-cluster-admission-reports kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: ff4e74e7-952d-4bd2-a9fa-77357c87869a
    - Title: batch/v1/CronJob kyverno-cleanup-admission-reports kyverno
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 0441a56f-4a31-43c7-89b5-b40a8072ccf8
    - Title: v1/Pod local-path-provisioner-6bc4bddd6b-vlmww local-path-storage
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 876d0a33-5d16-48f2-a73b-02d458e3e53c
    - Title: apps/v1/ReplicaSet local-path-provisioner-6bc4bddd6b local-path-storage
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```


  - Subject UUID: 637ce887-14c0-4773-af38-bf5dd77b7ac1
    - Title: apps/v1/Deployment local-path-provisioner local-path-storage
    - Result: failure :x:
    - Reason:
      ```
      validation failure: This container image's base is not in the approved list or is not specified. Only pre-approved base images may be used. Please contact the platform team for assistance.
      ```

</details>



Rule `policy-deployment (OCM)`:
- Ensure NGINX is deployed and running with given minimum instances

<details><summary>Details</summary>


  - Subject UUID: 8d550c54-dcd4-4ba8-9e0a-a6b2d2158120
    - Title: Cluster "cluster1"
    - Result: failure :x:
    - Reason:
      ```
      [c2p.policy-deployment.176f1ddc5591cb1c] NonCompliant; violation - deployments not found: [nginx-deployment] in namespace cluster1 missing; [nginx-deployment] in namespace kube-node-lease missing; [nginx-deployment] in namespace kube-public missing; [nginx-deployment] in namespace local-path-storage missing
      ```


  - Subject UUID: 9632eb0a-0a37-4aa5-8f51-e738acc95dab
    - Title: Cluster "cluster2"
    - Result: failure :x:
    - Reason:
      ```
      [c2p.policy-deployment.176f1dc4e7de17cb] NonCompliant; violation - deployments not found: [nginx-deployment] in namespace cluster2 missing; [nginx-deployment] in namespace default missing; [nginx-deployment] in namespace kube-node-lease missing; [nginx-deployment] in namespace kube-public missing; [nginx-deployment] in namespace local-path-storage missing
      ```

</details>


---

#### Result of control cm-2.1: 



Rule `disallow-capabilities (Kyverno)`:
- Adding capabilities beyond those listed in the policy must be disallowed.

<details><summary>Details</summary>


  - Subject UUID: e904838d-6ae5-4e6d-a6af-f23251af41a9
    - Title: v1/Pod kube-scheduler-kind-control-plane kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 9efd1da2-49a2-4c8a-876c-d7ba3903d131
    - Title: v1/Pod coredns-5d78c9869d-gc25q kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 12fcdc21-0bef-4e9a-a6fc-5f0d610dca7c
    - Title: v1/Pod kindnet-pbb9l kube-system
    - Result: failure :x:
    - Reason:
      ```
      Any capabilities added beyond the allowed list (AUDIT_WRITE, CHOWN, DAC_OVERRIDE, FOWNER, FSETID, KILL, MKNOD, NET_BIND_SERVICE, SETFCAP, SETGID, SETPCAP, SETUID, SYS_CHROOT) are disallowed.
      ```


  - Subject UUID: 46f484c3-310c-400d-b827-fd93f71ee2a9
    - Title: v1/Pod etcd-kind-control-plane kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: e51ea171-1445-4ed5-a87b-f2a471aba1d7
    - Title: v1/Pod kube-apiserver-kind-control-plane kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 43561035-6e0d-42c1-9d73-db80d0ef91b8
    - Title: v1/Pod kube-proxy-zbddb kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 98d01ce3-bbb7-424a-90b9-0500444d6410
    - Title: v1/Pod coredns-5d78c9869d-2rbnq kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: b0fe2b9e-0dab-4895-a3f4-999fcf161ac0
    - Title: v1/Pod kube-controller-manager-kind-control-plane kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 09678adf-4299-4c5c-8cd9-dd69c85b4891
    - Title: apps/v1/DaemonSet kindnet kube-system
    - Result: failure :x:
    - Reason:
      ```
      Any capabilities added beyond the allowed list (AUDIT_WRITE, CHOWN, DAC_OVERRIDE, FOWNER, FSETID, KILL, MKNOD, NET_BIND_SERVICE, SETFCAP, SETGID, SETPCAP, SETUID, SYS_CHROOT) are disallowed.
      ```


  - Subject UUID: e81dac74-cb49-4c70-8d77-6418fcbbe670
    - Title: apps/v1/DaemonSet kube-proxy kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 21c7ca64-e77e-494d-8f12-0f4038abe410
    - Title: apps/v1/ReplicaSet coredns-5d78c9869d kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 85e58cf6-ebe9-4edc-bf35-177864c7b1cd
    - Title: apps/v1/Deployment coredns kube-system
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 44942bba-b51d-42a2-9449-df7dd48eacf2
    - Title: v1/Pod kyverno-admission-controller-7cd788c8dd-gdnhp kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 43b02afc-fbbd-474e-82dd-2f63035f6a43
    - Title: v1/Pod kyverno-reports-controller-7f94855747-tmnhr kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 10a62320-3a50-46fa-8228-0b2dbc1c8a85
    - Title: v1/Pod kyverno-cleanup-admission-reports-28551310-cc4k7 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: efc7b716-8c48-4f64-a921-7ce7f18a87bb
    - Title: v1/Pod kyverno-cleanup-cluster-admission-reports-28551310-m4ld4 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: cb34f9c9-cb3d-40e4-95f6-4d08b91fc41f
    - Title: v1/Pod kyverno-cleanup-controller-ddf458755-9bnlb kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 4ce8797f-7de4-46f4-ba8a-f07dd6e5c825
    - Title: v1/Pod kyverno-background-controller-74599787cf-s6nm2 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: e12d0bf6-d8f8-40e3-bf34-8769be9d2242
    - Title: apps/v1/Deployment kyverno-cleanup-controller kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: f85edce1-7872-4e08-8c15-4b31acc57752
    - Title: batch/v1/Job kyverno-cleanup-cluster-admission-reports-28551310 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 16d35cd5-68a8-4e6b-a7eb-ca1dd5d53484
    - Title: apps/v1/ReplicaSet kyverno-admission-controller-7cd788c8dd kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 970a79f6-7816-4ae3-a2e7-461fcefcd59c
    - Title: batch/v1/Job kyverno-cleanup-admission-reports-28551310 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: cc36dd97-e864-4fde-a6ea-9f597a10816e
    - Title: apps/v1/Deployment kyverno-background-controller kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 096a66b2-b4f0-4699-8c7a-73a533a39fc9
    - Title: apps/v1/ReplicaSet kyverno-cleanup-controller-ddf458755 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 09265b08-841a-40e4-8905-f5c3a69852d1
    - Title: apps/v1/ReplicaSet kyverno-background-controller-74599787cf kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: a92bb864-4709-435b-99a3-af182b9d99ee
    - Title: apps/v1/Deployment kyverno-reports-controller kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 470a4b28-bfdf-4b42-beb4-fdfa633d21cf
    - Title: apps/v1/ReplicaSet kyverno-reports-controller-7f94855747 kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: 4fea8c6a-1493-4522-8093-b560a5d6521f
    - Title: apps/v1/Deployment kyverno-admission-controller kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: aac1da24-1ef4-4b9f-a6c3-78736dc74d6b
    - Title: batch/v1/CronJob kyverno-cleanup-cluster-admission-reports kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-cronjob-adding-capabilities' passed.
      ```


  - Subject UUID: 71c4745a-55b5-4616-a976-dd17212f720d
    - Title: batch/v1/CronJob kyverno-cleanup-admission-reports kyverno
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-cronjob-adding-capabilities' passed.
      ```


  - Subject UUID: ebf2ea59-6634-4b02-b582-c08a3247d4bd
    - Title: v1/Pod local-path-provisioner-6bc4bddd6b-vlmww local-path-storage
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'adding-capabilities' passed.
      ```


  - Subject UUID: 65eafe6f-9a3a-4e90-9cb9-73f7448d056c
    - Title: apps/v1/ReplicaSet local-path-provisioner-6bc4bddd6b local-path-storage
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```


  - Subject UUID: fa06ba30-837c-44da-985c-f0409d1ae14f
    - Title: apps/v1/Deployment local-path-provisioner local-path-storage
    - Result: pass :white_check_mark:
    - Reason:
      ```
      validation rule 'autogen-adding-capabilities' passed.
      ```

</details>


---

#### Result of control ac-1: 



Rule `policy-disallowed-roles (OCM)`:
- Ensure roles are set to only allowed values

<details><summary>Details</summary>


  - Subject UUID: 0dc6be0d-a543-4be8-b44e-93807b239f97
    - Title: Cluster "cluster1"
    - Result: pass :white_check_mark:
    - Reason:
      ```
      [c2p.policy-disallowed-roles.176f1dcdc4c8d17e] Compliant; notification - roles in namespace cluster1; in namespace default; in namespace kube-node-lease; in namespace kube-public; in namespace local-path-storage missing as expected, therefore this Object template is compliant
      ```


  - Subject UUID: abdd951e-9d9b-44e6-9709-a783d5c3ad32
    - Title: Cluster "cluster2"
    - Result: pass :white_check_mark:
    - Reason:
      ```
      [c2p.policy-disallowed-roles.176f1dc36e36b7b2] Compliant; notification - roles in namespace cluster2; in namespace default; in namespace kube-node-lease; in namespace kube-public; in namespace local-path-storage missing as expected, therefore this Object template is compliant
      ```

</details>


---

#### Result of control cm-6: 



Rule `policy-high-scan (OCM)`:
- Ensure scan is enabled with high level

<details><summary>Details</summary>


  - Subject UUID: 2558054b-c8f5-477d-91e1-aab3b2d58c04
    - Title: Cluster "cluster1"
    - Result: failure :x:
    - Reason:
      ```
      [c2p.policy-high-scan.176f1ddc441457e5] NonCompliant; violation - couldn't find mapping resource with kind ComplianceCheckResult, please check if you have CRD deployed
      ```


  - Subject UUID: e50230da-d99d-4a4c-84fd-89e79b733297
    - Title: Cluster "cluster2"
    - Result: failure :x:
    - Reason:
      ```
      [c2p.policy-high-scan.176f1dc4e29e1221] NonCompliant; violation - couldn't find mapping resource with kind ComplianceCheckResult, please check if you have CRD deployed
      ```

</details>


---


