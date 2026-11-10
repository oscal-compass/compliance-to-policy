# Kubernetes policy generation

One control definition to deployable policy for three engines:

```
component-definition.json
  -> policy per engine        plugins_public/plugins/{kyverno_dsl,kyverno_cel,vap}.py
  -> read by a person         not automated, and not optional
```

Applying that policy to a cluster and collecting what the engines then decide is the other half
of the chain, and it is not in this sample. C2P ships the conversion for it — the plugins'
`generate_pvp_result`, `c2p.result_to_oscal` and `c2p tools posture` — but not the program that
talks to a cluster, which carries the credentials and the decision to write.

`plugins_public` is not an installed package, so the repository root has to be on the path:

```bash
export PYTHONPATH=.

# generate, contacting no cluster
python samples_public/kubernetes/generate.py -o out
```

The generating and the applying are separate commands on purpose. The templates were written by
an agent, and a check that asserts nothing applies cleanly, enforces nothing and reports every
workload compliant — which is only catchable by reading it.

## What is here

| Path | |
|---|---|
| `component-definition.json` | six rules over nine NIST SP 800-53 controls, each rule implemented by up to three engines: `kyverno-dsl`, `kyverno-cel`, `vap-cel` |
| `templates/<engine>/<Check_Id>.yaml.tmpl` | one template per check; the render context is `engine`, `rule`, `check`, `params` |
| `expected-policies/` | a committed snapshot of this generator's output, so a change to a template or the renderer has to be read as a diff |

`plugins_public/tests/plugins/test_reference_parity.py` regenerates from the component
definition and the templates here and compares the result to `expected-policies/` byte for
byte. It also checks the thing no snapshot can: that one authored parameter reaches both CEL
engines as the same expression. That is the criterion for a policy generator: a CEL expression
is compared as a string by a cluster, so one space in a list literal is the difference between
two engines agreeing about a control and silently disagreeing.

## The engines are not interchangeable, and that is the point

**Kyverno DSL** (`kyverno.io/v1 ClusterPolicy`) and **Kyverno CEL** (`policies.kyverno.io/v1`)
both run a background scan and report through the same PolicyReport API, so collecting is a read
and both see violations that already existed — where most real findings live. The DSL is
deprecated in Kyverno 1.19 and removed in 1.20, and generating both from one component
definition is the argument: the control outlives the engine.

**ValidatingAdmissionPolicy** needs nothing installed, but it has no report API and evaluates
nothing retroactively. Its results have to be probed for by putting the workload back through
admission as a server-side dry run. And two of the six rules have no VAP check at all, because
they need I/O a CEL expression cannot do: verifying a signature against a registry and a
transparency log, and consulting an in-cluster registry ledger.

So generate with `-e vap-cel` alone and two of the six rules produce no policy at all, which is
how those controls stop being reported in the posture that follows — not green, not red,
absent. `tools posture-compare` names which ones. A control with no assessment capability
behind it is neither compliant nor non-compliant, and a posture that quietly omits it is how an
estate convinces itself it is fine.

## Requirements

`generate.py` contacts no cluster at all and is plain C2P: no `kubectl`, no credentials, no
network. Generating is a function of files, which is what makes reading the result before any
of it is applied possible at all.

What the generated policy needs is a cluster that will accept it. Kyverno 1.19 for both policy
families — the CEL types need 1.19 or later and the legacy DSL is gone from 1.20; VAP is built
in from Kubernetes 1.30.
