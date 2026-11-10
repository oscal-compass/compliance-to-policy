## Plugins for Kubernetes policy engines

Three engines for the same control: `kyverno-dsl`, `kyverno-cel` and `vap-cel`. One OSCAL
component definition, one template per check per engine, and C2P renders each engine's native
policy from it. Generating contacts no cluster.

They are separate plugins because they are not interchangeable — what each one can express, and
how its results come back, differ. That is covered under
[The engines are not interchangeable](#the-engines-are-not-interchangeable) below.

| Plugin | Generates | Config |
|---|---|---|
| `plugins_public.plugins.kyverno_dsl` | `kyverno.io/v1 ClusterPolicy` | `PluginConfigKyvernoDsl` |
| `plugins_public.plugins.kyverno_cel` | `policies.kyverno.io/v1 ValidatingPolicy`, `ImageValidatingPolicy` | `PluginConfigKyvernoCel` |
| `plugins_public.plugins.vap` | `admissionregistration.k8s.io/v1 ValidatingAdmissionPolicy` and its `ValidatingAdmissionPolicyBinding` | `PluginConfigVap` |

#### Prerequisite

Generating needs nothing: no `kubectl`, no credentials, no network. What the *generated* policy
needs is a cluster that will accept it — Kyverno 1.19 for both Kyverno policy families (the CEL
types need 1.19 or later and the legacy DSL is gone from 1.20), and Kubernetes 1.30 for
ValidatingAdmissionPolicy, which is built in and needs nothing installed.

#### Example usage of C2P

The bundled sample is [`samples_public/kubernetes/`](/samples_public/kubernetes/): six rules over
nine NIST SP 800-53 controls, each rule implemented by up to three engines.

1. Generate policy for every engine (C2P Compliance to Policy)
    ```
    export PYTHONPATH=.
    python samples_public/kubernetes/generate.py -o /tmp/deliverable-policy
    ```
    E.g.
    ```
    $ python samples_public/kubernetes/generate.py -o /tmp/deliverable-policy
    kyverno-dsl: generated 6 policy document(s) into /tmp/deliverable-policy/kyverno-dsl
    kyverno-cel: generated 6 policy document(s) into /tmp/deliverable-policy/kyverno-cel
    vap-cel: generated 4 policy document(s) into /tmp/deliverable-policy/vap-cel

    16 policy document(s) under /tmp/deliverable-policy. Read them before applying them: a check that asserts nothing applies cleanly and reports every workload compliant.
    ```
    `vap-cel` generates four, not six: two of the rules need I/O a CEL expression in the API
    server cannot do. The component definition declares no `vap-cel` check for them, so the
    posture reports the gap rather than a policy that always passes.

    One directory per engine, then one per rule, so a rule implemented by several engines does
    not collapse into one file:
    ```
    /tmp/deliverable-policy/kyverno-cel/image-pinned-by-digest/kyverno-cel-image-pinned-by-digest.yaml
    /tmp/deliverable-policy/vap-cel/image-pinned-by-digest/vap-image-pinned-by-digest.yaml
    ```

1. Read the generated policy, then apply it

    This step is deliberately not automated and not in the C2P library. The templates were
    written by an agent, and a check that asserts nothing applies cleanly, enforces nothing and
    reports every workload compliant — which is only catchable by reading it.

    Applying needs a program that talks to a cluster, which carries the credentials and the
    decision to write. C2P does not ship one.

1. Collect the engines' results and produce Assessment Results (C2P Result to Compliance)

    Kyverno DSL and Kyverno CEL both report through the PolicyReport API, so collecting is a
    read and both see violations that already existed. ValidatingAdmissionPolicy has no report
    API and evaluates nothing retroactively, so its verdicts have to be probed for by putting a
    workload back through admission as a server-side dry run; the `vap` plugin converts what
    such a probe observed.

    One Assessment Results document for all three engines — see
    [Aggregating several PVPs](/README.md#generate-compliance-assessment-results-from-pvp-native-results)
    in the README.

1. Render a compliance posture

    ```
    c2p tools result2posture -ar /tmp/assessment-results.json -cdef samples_public/kubernetes/component-definition.json -f html -o /tmp/posture.html
    ```

#### Using a plugin as a library

Each plugin takes a template directory and an output directory, and nothing else:

```python
from plugins_public.plugins.kyverno_cel import PluginConfigKyvernoCel, PluginKyvernoCel

plugin = PluginKyvernoCel(
    PluginConfigKyvernoCel(
        policy_template_dir='samples_public/kubernetes/templates/kyverno-cel',
        deliverable_policy_dir='/tmp/deliverable-policy/kyverno-cel',
    )
)
plugin.generate_pvp_policy(policy)
```

`policy` comes from `C2P(config).get_policy()`, as for any plugin — see
[Usage of C2P as a library](/README.md#usage-of-c2p-as-a-library).

The template for a check is `<policy_template_dir>/<Check_Id>.yaml.tmpl`. A check the component
definition declares with no template is an error, not a silent omission: generation raises
`C2PError` rather than producing a policy set that quietly misses a control. The template
contract — the render context, the filters, and how a value from the component definition is
substituted — is [`skills/policy-generation/template-contract.md`](/skills/policy-generation/template-contract.md).

Templates are rendered in a Jinja2 sandbox with `StrictUndefined`, and what they rendered is
parsed before it is written: every document must be a kind the plugin declares it generates,
named after the check. A template is executable and its inputs come from a component definition,
so neither is taken on trust.

#### The engines are not interchangeable

**Kyverno DSL** (`kyverno.io/v1 ClusterPolicy`) is deprecated in Kyverno 1.19 and removed in
1.20. **Kyverno CEL** (`policies.kyverno.io/v1`) is its replacement. Generating both from one
component definition is the argument for authoring the control rather than the policy: the
control outlives the engine.

**ValidatingAdmissionPolicy** needs nothing installed, but it decides at admission time only. It
runs no background scan over objects that already exist and publishes no report. And it cannot
express a check that needs I/O — verifying a signature against a registry and a transparency
log, or consulting an in-cluster ledger — because CEL in the API server cannot fetch from a
registry or read another object.

Where an engine cannot express a check, the right answer is to declare no check for it. The
control then appears in the posture as *not assessed by that engine* rather than as a pass, and
`c2p tools posture-compare` names which controls stop being reported when only that engine is
available. A control with no assessment capability behind it is neither compliant nor
non-compliant, and a posture that quietly omits it is how an estate convinces itself it is fine.

More on each engine's dialect, and the failures that recur when generating for it, in
[`skills/policy-generation/engines.md`](/skills/policy-generation/engines.md).
