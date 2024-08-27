

## Component: GitHub


#### Result of control ac-2: 



Rule `rule_github_org_member (Auditree)`:
- Check whether the GitHub org is not empty.

<details><summary>Details</summary>


  - Subject UUID: de01a6a4-4ebe-4191-b566-e1dc48e8c613
    - Title: Auditree Check: demo_examples.checks.test_github.GitHubOrgs.test_members_is_not_empty_0_oscal_compass
    - Result: failure :x:
    - Reason:
      ```
      {'oscal-compass': ['There are people in there, but less than 5!']}
      ```


  - Subject UUID: f933f9fa-fb6e-4a62-a708-2b4cf59009c2
    - Title: Auditree Check: demo_examples.checks.test_github.GitHubOrgs.test_members_is_not_empty_1_esa
    - Result: pass :white_check_mark:
    - Reason:
      ```
      {}
      ```

</details>


---

#### Result of control cm-2: 



Rule `rule_github_api_version (Auditree)`:
- Check whether there are any supported versions.

<details><summary>Details</summary>


  - Subject UUID: 841cc8b0-29a7-46ff-81fb-8f1279b1be7b
    - Title: Auditree Check: demo_examples.checks.test_github.GitHubAPIVersionsCheck.test_supported_versions
    - Result: failure :x:
    - Reason:
      ```
      {'Supported GitHub API Versions Warning': ['There is only one supported version. Get with the program: 2022-11-28']}
      ```

</details>


---


