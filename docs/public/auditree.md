## Plugin for Auditree

### Example usage of C2P w/ mock data
1. Generate auditree.json (C2P Compliance to Policy)
    ```sh
    $ python ./samples_public/auditree/compliance_to_policy.py -h
    usage: compliance_to_policy.py [-h] [-i INPUT] [-c COMPONENT_DEFINITION] [-o OUT]

    options:
    -h, --help            show this help message and exit
    -i INPUT, --input INPUT
                            Path to auditree.json template (default: plugins_public/tests/data/auditree/auditree.template.json)
    -c COMPONENT_DEFINITION, --component_definition COMPONENT_DEFINITION
                            Path to component-definition.json (default: plugins_public/tests/data/auditree/component-
                            definition.json
    -o OUT, --out OUT     Path to generated auditree.json (default: system temporary directory)
    ```
    e.g.
    ```sh
    $ python ./samples_public/auditree/compliance_to_policy.py -o auditree.json
    $ cat auditree.json 
    {
        "locker": {
            "default_branch": "main",
            "repo_url": "https://github.com/MY_ORG/MY_EVIDENCE_REPO"
    },...
    ```
1. Generate Assessment Result (C2P Result to Compliance)
    ```sh
    $ python ./samples_public/auditree/result_to_compliance.py -h
    usage: result_to_compliance.py [-h] [-i INPUT] [-c COMPONENT_DEFINITION]

    options:
    -h, --help            show this help message and exit
    -i INPUT, --input INPUT
                            Path to check_results.json (default: plugins_public/tests/data/auditree/check_results.json)
    -c COMPONENT_DEFINITION, --component_definition COMPONENT_DEFINITION
                            Path to component-definition.json (default: plugins_public/tests/data/auditree/component-
                            definition.json
    ```
    e.g.
    ```sh
    $ python ./samples_public/auditree/result_to_compliance.py
    ...
        "results": [
        {
            "uuid": "853eeb24-6970-4f73-8fcc-fc274be669ec",
            "title": "Auditree Assessment Results",
            "description": "OSCAL Assessment Results from Auditree",
            "start": "2024-06-02T08:42:22+00:00",
            "reviewed-controls": {
            "control-selections": [
                {
                "include-controls": [
                    {
                    "control-id": "cm-2",
                    "statement-ids": []
                    },
                    {
                    "control-id": "ac-2",
                    "statement-ids": []
                    }
                ]
                }
            ]
            },
            "observations": [
            {
                "uuid": "3ea6d5dd-7a69-4f18-828c-a0e578594c63",
                "title": "demo_examples.checks.test_github.GitHubOrgs.test_members_is_not_empty",
                "description": "demo_examples.checks.test_github.GitHubOrgs.test_members_is_not_empty",
                "props": [
                {
                    "name": "assessment-rule-id",
                    "value": "demo_examples.checks.test_github.GitHubOrgs.test_members_is_not_empty"
                }
                ],
                "methods": [
                "AUTOMATED"
                ],
                "subjects": [
                {
                    "subject-uuid": "e3789a4f-f32a-4d59-b777-44df643631e6",
                    "type": "inventory-item",
                    "title": "Auditree Check: demo_examples.checks.test_github.GitHubOrgs.test_members_is_not_empty_0_nasa",
                    "props": [
                    {
                        "name": "resource-id",
                        "value": "demo_examples.checks.test_github.GitHubOrgs.test_members_is_not_empty_0_nasa"
                    },
                    {
                        "name": "result",
                        "value": "pass"
            ...
    ```
### Example usage of C2P w/ Auditree

Prerequisite:
- Install Python packages for Auditree:
    - https://pypi.org/project/auditree-framework/
    - https://pypi.org/project/auditree-arboretum/

1. Clone auditree-framework and go to `demo` directory (See also https://complianceascode.github.io/auditree-framework/quick-start.html)
    ```
    git clone https://github.com/ComplianceAsCode/auditree-framework.git
    cd auditree-framework/demo
    ```
1. Clone c2p
    ```
    git clone https://github.com/oscal-compass/compliance-to-policy.git
    ```
1. Generate auditree.json (C2P Compliance to Policy)
    1. Create OSCAL component-definition.json
        
        `sed 's/nasa/oscal-compass/g' ./compliance-to-policy/plugins_public/tests/data/auditree/component-definition.json > ./component-definition.json`
        
        1. (Optional) You can edit it in Spreadsheet [component-definition.csv](/plugins_public/tests/data/auditree/component-definition.csv) and then convert it to OSCAL JSON format through Trestle. To convert it, C2P also provides an utility (internally using Trestle)

            `c2p tools csv-to-oscal-cd --title "Sample Component Definition using Auditree as PVP" --csv ./compliance-to-policy/plugins_public/tests/data/auditree/component-definition.csv  --out <path to output directory>`

    1. Generate auditree.json

        `python ./compliance-to-policy/samples_public/auditree/compliance_to_policy.py -i ./auditree_demo.json -c ./component-definition.json -o auditree.json`

1. Run policy validation (Auditree fetchers and checks)
    ```
    compliance --fetch --evidence local -C auditree.json -v
    ```
    ```
    compliance --check demo.arboretum.accred,demo.custom.accred --evidence local -C auditree.json -v
    ```
    You'll see the path to the local evidence locker directory in the log.
    
    e.g.
    ```
    $ compliance --check demo.arboretum.accred,demo.custom.accred --evidence local -C auditree.json -v

    INFO: Using locker found in /var/folders/yx/1mv5rdh53xd93bphsc459ht00000gn/T/compliance...
    ...
    ```
1. Generate Assessment Result (C2P Result to Compliance)
    
    `python ./compliance-to-policy/samples_public/auditree/result_to_compliance.py -i <PATH/TO/EVIDENCE_LOCKER/check_results.json> -c ./component-definition.json` > assessment-results.json

    e.g.
    ```
    $ python ./compliance-to-policy/samples_public/auditree/result_to_compliance.py -i /var/folders/yx/1mv5rdh53xd93bphsc459ht00000gn/T/compliance/check_results.json -c ./component-definition.json > assessment_results.json
    ```
1. OSCAL Assessment Results is not human readable format. You can see the merged report in markdown by a quick viewer.
    ```
    c2p tools viewer -ar assessment_results.json -cdef ./component-definition.json
    ```
    ![assessment-results-md.auditree.jpg](/docs/public/images/assessment-results-md.auditree.jpg)