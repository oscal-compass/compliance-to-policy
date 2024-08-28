# <img alt="Logo" width="50px" src="https://raw.githubusercontent.com/oscal-compass/compliance-to-policy/main/assets/compliance-to-policy-800x800.PNG" style="vertical-align: middle;" /> Compliance-to-Policy (also known as `C2P`)

Compliance-to-Policy (C2P) is designed to bridge Compliance as Code such as Open Security Controls Assessment Language (OSCAL) and Policy as Code used by Policy Validation Point (PVP). It generates policies in native format of PVP from OSCAL Component Definitions and produces OSCAL Assessment Results from the native assessment results of PVP. C2P can be used both as a command-line tool and a Python library, making it easy and flexible to integrate into your Continuous Compliance pipelines, such as GitHub Actions, Tekton Pipelines, or Agile Authoring Pipelines. It supports multiple PVP engines, including [Kyverno](https://kyverno.io/), [Open Cluster Management Policy Framework](https://open-cluster-management.io/), and the open-source [Auditree](https://auditree.github.io/), through dedicated plugins for each. Custom plugins can be implemented with a small amount of Python code.

![C2P Overview](/assets/architecture.png)

1. Compliance-to-Policy (C2P) is running in GitOps Pipeline, Kubernetes controller, or Python/Go environment
1. C2P receives Compliance as Code, for example OSCAL Component Definition that represents mapping between controls and policies (policy names/ids)
1. C2P generates policies through plugin for each policy engine
    - The plugin is responsible for implementing a function that takes policy names/ids and returns policies
1. Policies are delivered to policy engines by GitOps sync, the subsequence pipeline task, Kubernetes controller, or a deployment automation program 
1. Results are collected from policy engines by a scheduled task or Kubernetes controller
1. C2P aggregates the results of policy engines by controls through plugin for each policy engine
    - The plugin is responsible for implementing a function that takes the results of the policy engine and returns verdicts (pass/fail/error), reason, and/or resource name for each respective policy by its names/IDs.
1. C2P produces Compliance Assessment Results, for example OSCAL Assessment Results that represents the assessment results of each control 

Demo:
- [Kyverno as PVP](docs/public/kyverno.md)
- [Heterogeneous PVPs (mixing Kyverno, OCM Policy, and Auditree)](docs/public/heterogeneous.md)

## Goals
Provide seamless integration with compliance frameworks and existing policy engines, and enable to use heterogeneous policy engines in compliance check operation
- Flexibility in choise of policy engines and compliance frameworks
    - Provide plugins to cover various policy engines including proprietary/open source policy validation/enforcement engines, or in-house policy validation/enforcement program
    - Cover various compliance frameworks not only OSCAL but also other GRC frameworks and Cloud Security Posture Management services
- Community-driven plugin extension
    - Provide an efficient plugin interface and development method

## Supported Compliance Frameworks
- [Open Security Controls Assessment Language (OSCAL)](https://pages.nist.gov/OSCAL/documentation/)
    - OSCAL standard provides a compliance framework and the corresponding set of key compliance artifacts expressed in machine processable formats enabling all compliance documents to be treated as code and therefore processed and managed in the same manner.

## Supported Policy Engines
- [Kyverno](https://kyverno.io/) (for Kubernetes resources)
    - Kyverno is a policy engine designed for Kubernetes, where policies are managed as Kubernetes resources. Kyverno policies can validate, mutate, generate, and clean up Kubernetes resources.
- [Open Cluster Management Policy Framework](https://open-cluster-management.io/) (for Kubernetes resources)
    - OCM is a multi-cluster management platform that provides governance of Kubernetes policies. [Its policy framework](https://open-cluster-management.io/concepts/policy/) allows for the validation and enforcement of policies across multiple clusters.
- [Auditree](https://auditree.github.io/) (for any target, especially well-suited for resources of PaaS/SaaS/IaaS available through REST API.)
    - Auditree is a GitOps based workflow automation that enables the collection and verification of evidence, building a long-term store of evidence in an git "evidence locker." Evidence is gathered by code scripts called "fetchers" and verified by "checks."

Roadmap:
- [OPA/Gatekeeper](https://github.com/open-policy-agent/gatekeeper) (for Kubernetes resources)
- [Ansible](https://www.ansible.com/) (for any target, especially for VMs and OnPremise hosts)

## C2P in Go language
The Go verion is available in the [go directory](/go/README.md). 

## Install 

#### From git repo
```
pip install git+https://github.com/oscal-compass/compliance-to-policy.git
``` 
You may be asked passphrase of SSH key to access to the git repo.

#### From source
1. Clone the repository
    ```
    git clone https://github.com/oscal-compass/compliance-to-policy.git
    ```
1. Go to `compliance-to-policy`
    ```
    cd compliance-to-policy
    ```
1. Install
    ```
    make install
    ```

## Quick demo

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

## Usage of C2P Plugins
- [Kyverno](docs/public/kyverno.md)
- [Open Cluster Management Governance Policy Framework](docs/public/ocm.md)
- [Auditree](docs/public/auditree.md)
- [Heterogeneous PVPs (mixing Kyverno, OCM Policy, and Auditree)](docs/public/heterogeneous.md)

## Usage of C2P as a library

#### Generate PVP Policies from Compliance
1. Create `C2PConfig` object to supply compliance requirements and some metadata (See also [kyverno/compliance_to_policy.py](/samples_public/kyverno/compliance_to_policy.py) for a real example)
    ```python
    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = 'plugins_public/tests/data/kyverno/component-definition.json'
    c2p_config.pvp_name = 'Kyverno'
    c2p_config.result_title = 'Kyverno Assessment Results'
    c2p_config.result_description = 'OSCAL Assessment Results from Kyverno'
    ```
1. Select a plugin for supported PVPs (`PluginKyverno`, `PluginOCM`) and create `PluginConfig` object to supply plugin specific properties
    ```python
    from plugins_public.plugins.kyverno import PluginConfigKyverno, PluginKyverno
    policy_template_dir = 'plugins_public/tests/data/kyverno/policy-resources'
    config = PluginConfigKyverno(policy_template_dir=policy_template_dir, deliverable_policy_dir='/tmp/deliverable-policies')
    ```
1. Create `C2P` and `Plugin`
    ```python
    c2p = C2P(c2p_config)
    plugin = PluginKyverno(config)
    ```
1. Get policy from `c2p` and generate PVP policy by `generate_pvp_policy()`
    ```python
    policy = c2p.get_policy()
    plugin.generate_pvp_policy(policy)
    ```
1. The deliverable policies are output in '/tmp/deliverable-policies'
    ```
    $ tree /tmp/deliverable-policy
    /tmp/deliverable-policy
    ├── allowed-base-images
    │   ├── 02-setup-cm.yaml
    │   └── allowed-base-images.yaml
    └── disallow-capabilities
        └── disallow-capabilities.yaml
    ```
#### Generate Compliance Assessment Results from PVP native results
1. Create `C2PConfig` object to supply compliance requirements and some metadata (See also [kyverno/compliance_to_policy.py](/samples_public/kyverno/result_to_compliance.py) for a real example)
    ```python
    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = 'plugins_public/tests/data/kyverno/component-definition.json'
    c2p_config.pvp_name = 'Kyverno'
    c2p_config.result_title = 'Kyverno Assessment Results'
    c2p_config.result_description = 'OSCAL Assessment Results from Kyverno'
    ```
1. Select a plugin for supported PVPs (`PluginKyverno`, `PluginOCM`) and create `PluginConfig` object to supply plugin specific properties
    ```python
    from plugins_public.plugins.kyverno import PluginConfigKyverno, PluginKyverno
    config = PluginConfigKyverno()
    ```
1. Create `C2P` and `Plugin`
    ```python
    c2p = C2P(c2p_config)
    plugin = PluginKyverno(config)
    ```
1. Load PVP native results
    ```python
    policy_report_file = 'plugins_public/tests/data/kyverno/policyreports.wgpolicyk8s.io.yaml'
    cluster_policy_report_file = 'plugins_public/tests/data/kyverno/clusterpolicyreports.wgpolicyk8s.io.yaml'
    policy_report = yaml.safe_load(pathlib.Path(policy_report_file).open('r'))
    cluster_policy = yaml.safe_load(pathlib.Path(cluster_policy_report_file).open('r'))
    pvp_raw_result = RawResult(data=policy_report['items'] + cluster_policy['items'])
    ```
1. Call `generate_pvp_result()` of the plugin to get a formatted PVP result
    ```python
    pvp_result = PluginKyverno().generate_pvp_result(pvp_raw_result)
    ```
1. Create `C2P` and call `result_to_oscal()` to obtain Compliance Assessment Results
    ```python
    c2p.set_pvp_result(pvp_result)
    oscal_assessment_results = c2p.result_to_oscal()
    print(oscal_assessment_results.oscal_serialize_json(pretty=True))
    ```
1. (Optional) you may reformat OSCAL Assessment Results in markdown style.
    ```
    c2p tools viewer -ar <OSCAL Assessment Results (json)> -cdef ./plugins_public/tests/data/ocm/component-definition.json -o /tmp/assessment_results.md
    ```

## How to support your own PVP in C2P

You can create a custom plugin by overriding `PluginSpec` and `PluginConfig`. 
`PluginSpec` has two interfaces `generate_pvp_policy` and `generate_pvp_result`. 
C2P framework will instantiate `PluginSpec` with `PluginConfig`.

#### PluginConfig
1. Extend PluginConfig with custom fields as the plugin needs 
    ```python
    from c2p.framework.plugin_spec import PluginSpec
    class YourPluginConfig(PluginConfig):
        custom_field: str = Field(..., title='Custom field for your plugin')
    ```
1. Extend PluginSpec and define __init__ with YourPluginConfig
    ```python
    class YourPlugin(PluginSpec):
        def __init__(self, config: Optional[YourPluginConfig] = None) -> None:
            super().__init__()
            self.config = config # work on config
    ```

#### PluginSpec.generate_pvp_policy
1. `generate_pvp_policy()` in `PluginSpec` accepts one argument `policy: c2p.framework.models.Policy`.
    The object has two fields (`rule_sets` and `parameters`). `rule_sets` and `parameters` are a list of Rule_Id, Check_Id, Parameter_Id, Parameter_Value, etc of the components handled by your PVP in OSCAL Component Definition.
1. Implement the logic to generate PVP policy from provided rule_sets and parameters. 
    ```python
        def generate_pvp_policy(self, policy: Policy):
            rule_sets: List[RuleSet] = policy.rule_sets
            parameters: List[Parameter] = policy.parameters
            # generate deliverable policy from rule_sets and parameters
    ```

#### PluginSpec.generate_pvp_result
1. `generate_pvp_result()` is expected to generate the summarized raw results of your PVP per unit in `PVPResult` format. This unit must be associated with a unique id called Check_Id. For example of [PluginKyverno](/plugins_public/plugins/kyverno.py), Policy Reports is the raw results and are summarized by policy name.
    ```python
    def generate_pvp_result(self, raw_result: RawResult) -> PVPResult:
        pvp_result: PVPResult = PVPResult()
        observations: List[ObservationByCheck] = []

        polrs = list(
            filter(
                lambda x: x['apiVersion'] == 'wgpolicyk8s.io/v1alpha2' and x['kind'] == 'PolicyReport', raw_result.data
            )
        )
        cpolrs = list(
            filter(
                lambda x: x['apiVersion'] == 'wgpolicyk8s.io/v1alpha2' and x['kind'] == 'ClusterPolicyReport',
                raw_result.data,
            )
        )

        results = []
        for polr in polrs:
            for result in polr['results']:
                results.append(result)
        for cpolr in cpolrs:
            for result in cpolr['results']:
                results.append(result)

        policy_names = list(map(lambda x: x['policy'], results))  # policy_name is used as check_id
        policy_names = set(policy_names)

        for policy_name in policy_names:
            observation = ObservationByCheck(check_id=policy_name, methods=['AUTOMATED'], collected=get_datetime())
    ```

1. The input argument `raw_result` has `data` field that is serialized raw results as dict. You can define your preferable format of the data. C2P Framework will pass PVP native results to plugin with this format.

#### Publish plugin
1. Put the plugin in plugin directory [/plugins_public/plugins](/plugins_public/plugins) or Python module path when you use C2P.

## Development

### Developing
1. Install Python
    ```
    $ python --version
    Python 3.10.12
    ```
1. Setup venv
    ```
    python -m venv .venv
    ```
1. Install dependant modules
    ```
    make install-dev
    ```
1. Enable detect-secret
    ```
    pre-commit install
    ```

### Test
```
$ make test       

plugins_public/tests/plugins/test_kyverno.py::test_kyverno_pvp_result_to_compliance PASSED                                                                                                   [ 25%]
plugins_public/tests/plugins/test_kyverno.py::test_kyverno_compliance_to_policy PASSED                                                                                                       [ 50%]
plugins_public/tests/plugins/test_ocm.py::test_ocm_pvp_result_to_compliance PASSED                                                                                                           [ 75%]
plugins_public/tests/plugins/test_ocm.py::test_ocm_compliance_to_policy 
------------------------------------------------------------------------------------------ live log call -------------------------------------------------------------------------------------------
2024-04-25 05:31:48 [    INFO] The deliverable policy directory '/var/folders/yx/1mv5rdh53xd93bphsc459ht00000gn/T/tmpxtvpcrpr/deliverable-policy' is not found. Creating... (ocm.py:191)
PASSED                                                                                                                                                                                       [100%]

======================================================================================== 4 passed in 0.31s =========================================================================================

tests/c2p/framework/test_c2p.py::test_result_to_oscal PASSED                                                                                                                                 [ 33%]
tests/c2p/test_cli.py::test_run PASSED                                                                                                                                                       [ 66%]
tests/c2p/test_cli.py::test_version PASSED                                                                                                                                                   [100%]

======================================================================================== 3 passed in 0.26s =========================================================================================
```

### Cleanup caches
```
make clean
```