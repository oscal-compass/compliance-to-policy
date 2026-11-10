# Reference implementation: driving a cluster

**This directory is not part of the C2P library, and it is not installed with it.** It is a
worked example of the half of the round trip C2P deliberately does not ship: applying generated
policy to a cluster and reading what the engines decided.

That half is left out on purpose. It holds cluster credentials, it writes to a cluster, and it
copies engine messages into a document an auditor may later read. Those are decisions about one
particular cluster and one particular operator's tolerance for risk, and a library that made
them on your behalf would be making them for every user at once. C2P's job starts at an authored
control and stops at generated policy; it resumes at collected results and produces OSCAL.

What sits between those two points is this directory, and a person.

## The step that is not automated

```
component definition  ->  generated policy  ->  [ a person reads it ]  ->  cluster
```

The generated policy is a function of the component definition, and in the authoring flow this
example was built for, the component definition and the policy templates are written by an
agent. A template is where engine-specific knowledge lives, so it is the one artefact in the
chain that cannot be derived from the control — and the one an agent can get subtly wrong in a
way that still generates, still applies, and still reports green.

So generation and deployment are separate commands, and reviewing the generated tree in between
is not a formality:

```bash
export PYTHONPATH=.
# 1. C2P generates. No cluster is contacted, and none of this is in this directory.
python samples_public/kubernetes/generate.py -o out

# 2. read what was generated
git diff --no-index /dev/null out/vap-cel    # or just: ls -R out && $EDITOR out

# 3. apply exactly what you read
python docs/end-to-end/reference/deploy.py --context kind-storefront -f out

# 4. collect what the engines then decided
python docs/end-to-end/reference/collect_and_assess.py --context kind-storefront -o out
```

Worth looking for specifically, because each has been produced by accident at least once:

- **A check that asserts nothing.** An expression over a field that no object in scope has is
  vacuously true, and reports every workload as compliant. This is the most expensive way for
  the chain to be wrong, because the output is indistinguishable from success.
- **A binding that matches nothing.** A policy with no binding, or a binding whose
  `namespaceSelector` does not match, enforces nothing. The probe's canaries catch this for
  ValidatingAdmissionPolicy; nothing catches it for you on the enforcement side.
- **Deny where you expected a report.** The generated Kyverno and VAP bindings here deny, which
  is what a control that says "must" asks for. Applying them to a running namespace can stop
  deployments that were working a minute ago. Check `validationActions` and
  `failureAction`/`validationFailureAction` before you apply, not after.
- **A trust root you did not choose.** The sample's image-signature check names a signer
  identity and an OIDC issuer. Those are the names your cluster will trust. Replace them.

## What this code does to a cluster

| Operation | Where | Scope |
|---|---|---|
| `apply --server-side` of the generated tree | `deploy.py` | the objects you reviewed, and only those |
| read PolicyReports | `platform_k8s/collect.py` | cluster-wide read |
| create a scratch namespace and `Warn` bindings | `platform_k8s/collect.py` | named and labelled per run |
| `apply --server-side --dry-run=server` of a copy of each workload | `platform_k8s/collect.py` | into the scratch namespace only |
| `delete -l compliance.oscal/probe-run=<run>` | `platform_k8s/collect.py` | only what this run created |

Three things it does not do, each because the alternative was tried and rejected:

- **It never deletes or modifies a policy or binding it did not create.** An earlier version
  unbound the other policies to isolate one check at a time. The bindings are cluster-scoped, so
  every probe opened a window in which the cluster was not enforcing, and a killed process left
  that window open. If this code is killed now, what it leaves behind is an empty namespace and
  bindings that deny nothing.
- **It never uses a client-side `kubectl apply`.** When a client-side apply is rejected, kubectl
  prints the patch it was attempting, and that patch carries a copy of the whole object —
  environment variables included — which then lands in Assessment Results. Server-side apply
  reports only the engine's own message.
- **It never copies a raw error stream into a result.** An engine may quote your object back at
  you in its message. What reaches Assessment Results is one line, capped.

There is deliberately no `undeploy`. Removing policy from a cluster only ever reduces enforcement,
so it stays a deliberate `kubectl delete` rather than a convenience sitting next to `deploy.py`
— which prints that command, naming the documents it applied, rather than offering to run it.

## Probing an admission-only engine

Kyverno publishes its findings to the PolicyReport API, so collecting from it is a read.
ValidatingAdmissionPolicy has no report API and evaluates nothing retroactively, so there is
nothing to read: the only way to learn what it thinks of a workload that is already running is
to submit that workload through admission again.

`platform_k8s/collect.py` carries the full account of how that is done without disturbing
anything, and the measurements behind it. The short version: an additional `Warn` binding per
check, scoped to a scratch namespace, and a dry-run submission of a copy of the workload there.
Warnings do not short-circuit, so one submission answers every check — measured at 4 of 4
verdicts, against 1 of 4 for the `Deny` bindings the cluster actually enforces with.

A `Warn` verdict of "pass" is the absence of a warning, which a probe wired up wrongly also
produces. So every check is first shown a **canary** from `canaries/` — an object written to
fail that check. If the check does not report the canary, the probe says `error` for that check
rather than `pass`. There is no flag to skip this, and a check with no canary is `error` too.

## Requirements

`--context` is required by both programs, and each prints the context it is about to act on
before it writes anything. There is no fallback to `kubectl config current-context`: these
programs apply Deny policy and create objects, and inheriting the ambient context is how a run
lands on a cluster nobody meant to change.

`kubectl` on `PATH` (or `--kubectl`), and a context with permission to: read the workloads being
assessed, read PolicyReports, apply the generated policy, and create and delete namespaces and
`validatingadmissionpolicybindings`. The last two are the probe's, and they are the reason this
is not a read-only tool.
