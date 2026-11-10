# The three Kubernetes engines

| Engine id | apiVersion | Kind | Evidence |
|---|---|---|---|
| `kyverno-dsl` | `kyverno.io/v1` | `ClusterPolicy` | background scan → PolicyReport |
| `kyverno-cel` | `policies.kyverno.io/v1` | `ValidatingPolicy`, `ImageValidatingPolicy` | background scan → PolicyReport |
| `vap-cel` | `admissionregistration.k8s.io/v1` | `ValidatingAdmissionPolicy` + `Binding` | nothing; must be probed |

An estate does not have one engine, which is why a component definition can declare several and
why the same rule gets a template in each.

## Reporting engines and admission-only engines

Both Kyverno families run a **background scan** and write to the `wgpolicyk8s.io` PolicyReport
API. Collection is therefore a read, and they see a violation that **already existed** — which
is where most real findings live.

`ValidatingAdmissionPolicy` needs nothing installed, but it has no report API and evaluates
nothing retroactively. There is no result to read: the object has to be re-submitted with a
server-side dry run to find out what the engine thinks of it. Write templates knowing that an
object admitted before the policy existed will never be judged by it.

A consequence for authoring: a `vap-cel` template that depends on seeing the whole cluster is
not merely hard, it is meaningless. Scope VAP checks to what is decidable about the object in
the request.

## `kyverno-dsl` — kyverno.io/v1 ClusterPolicy

Deprecated in Kyverno 1.19 and removed in 1.20. Generate it anyway when the estate runs it; that
an engine is retiring is the argument for authoring the control rather than the policy.

Shape: `spec.validationFailureAction: Enforce`, `spec.background: true`, `rules[].match.any[].resources`
with `kinds` and `namespaces`, then either

- `validate.pattern` for a structural assertion (`"?*"` means "present and non-empty"), or
- `validate.foreach[].list` with `deny.conditions.all[]` (`key` / `operator` / `value`) when the
  assertion is per-container or per-element.

To read cluster state, use `context[].apiCall` with `urlPath` and `jmesPath`.

**Measured trap:** make the `apiCall` a **LIST** of the objects you want, not a GET of one named
object. A GET that 404s makes the rule *error*, and a posture folds error into fail — so a
missing ConfigMap shows up as a red control that looks like a violation. A list of zero items is
an honest empty allow-list.

For image signatures, `rules[].verifyImages` takes a keyless cosign attestor (subject, issuer,
Rekor) or a key read from a Secret or ConfigMap.

**Measured trap, and the one most likely to be repeated:** a `foreach` variable such as
`element.image` or `element.name` may appear **only inside the `foreach` block**. Referencing it
in the rule's top-level `validate.message` is rejected by Kyverno's own admission webhook:

    admission webhook "validate-policy.kyverno.svc" denied the request:
    variable 'element.image' present outside of foreach at path /validate/message

The message that names the offending element has to be the `deny` message inside the `foreach`.
A top-level message must be static, naming the parameter rather than the element.

## `kyverno-cel` — policies.kyverno.io/v1

Kyverno's current policy types, written in CEL. Needs Kyverno 1.19 or later.

Shape: `spec.validationActions: [Deny, Audit]`, `spec.evaluation.background.enabled: true`,
`spec.matchConstraints` with a `namespaceSelector` and `resourceRules`, then `spec.variables`
(name + expression) and `spec.validations` (expression + `messageExpression`).

Put the parameter in a variable and assert over the variable. It keeps the assertion readable
and it makes the parameter's rendering visible in one place when you diff against another
engine.

`ImageValidatingPolicy` carries the `cosign` blocks. It has no `secretRef` for a key, so an
approved key is supplied through `cosign.key.expression`.

**Measured on 1.19.2:** `resource.List` needs its result wrapped in `dyn`. Mutating policies run
before validating ones, so a mutation can change what a validation sees — do not write a
validation that assumes the object as submitted.

**The transparency log field is named differently in each Kyverno API, and the API server
rejects the wrong one outright** (`strict decoding error: unknown field
"spec.attestors[0].cosign.rekor"`):

| API | attestor path | field |
|---|---|---|
| `policies.kyverno.io/v1 ImageValidatingPolicy` | `spec.attestors[].cosign` | **`ctlog`** |
| `kyverno.io/v1 ClusterPolicy` | `verifyImages[].attestors[].entries[].keyless` | **`rekor`** |

Same log, same URL, different key. Do not carry one over from the other by analogy.

## `vap-cel` — admissionregistration.k8s.io/v1

In-tree from Kubernetes 1.30, nothing to install. That raises the obvious question of why anyone
would run a policy engine, and the honest answer is in what it cannot do.

Shape: the `ValidatingAdmissionPolicy` **and** its `ValidatingAdmissionPolicyBinding` in the
same document, separated by `---`. They must travel together: a policy without its binding
enforces nothing, and in a posture that is indistinguishable from a policy that passed. Put
`matchConstraints.resourceRules` on the policy and `matchResources.namespaceSelector` on the
binding. `failurePolicy: Fail`. Use `auditAnnotations` to stamp the check id, so a probe result
can be attributed.

### What it cannot express, and why

Recognising this is most of the skill. Both of these are real, not a matter of effort:

- **Verifying a signature.** It needs a registry fetch and a transparency-log round trip. CEL in
  the API server has no network I/O and no cosign primitive.
- **Testing membership against a list held in the cluster.** It has no resource library, so it
  cannot read another object. `paramRef` looks like the way around it and is not: a `paramRef`
  that selects several objects is evaluated once per object and must pass for **all** of them,
  which is conjunctive, while an allow-list is existential. Inlining the list into CEL satisfies
  the control text while destroying the property the external list exists for — that the list
  can change without redeploying policy.

**Measured:** VAP's CEL has no `flatten`.

When you hit one of these, declare no check for that engine. The control will come back
`not-assessed` with its coverage showing which engines were expected — which is the true
statement.
