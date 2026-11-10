# Verification

"Looks right" is not a criterion for generated policy. A CEL expression is compared as a string
by a cluster, so one space in a list literal is the difference between two engines agreeing
about a control and silently disagreeing. Three checks, in the order they are cheapest.

## 1. Coverage: every declared check has a template

The component definition is the authority on what should exist, not the template directory. Read
the checks each validation component declares and confirm one template each.

`policy_template.render` already fails on a missing template, so rendering is the check — but
read the error rather than assuming it means the component definition is wrong. It usually means
the engine genuinely cannot carry that check and the component definition should stop declaring
it. See [engines.md](engines.md).

## 2. Byte-for-byte against a committed reference

Render into a temporary directory and diff against a reference tree committed beside the
templates. This is the criterion a policy generator is held to, because the failure mode it
catches is invisible by inspection.

The bundled example is `plugins_public/tests/plugins/test_reference_parity.py`, driven from
`samples_public/kubernetes/`: it regenerates every engine's policy from the sample component
definition and the sample templates, and compares each document to
`samples_public/kubernetes/expected-policies/<engine>/`.

Add a new engine by adding its reference output and one row to that test's `ENGINES` list. The
check count in the row is deliberate: it asserts the number of rule sets the component
definition yields, so a check that silently disappears fails the test rather than passing with
fewer files.

**When the reference and the render disagree, decide which is wrong before editing either.** A
reference tree is only worth having if it is not casually overwritten.

## 3. The cross-engine parity of one parameter

If two engines express the same rule, assert that the authored parameter reaches both
identically. The sample test does this directly:

```python
rule = 'workload-inventory-labelled'
cel = required_expression(tmp_path / 'kyverno-cel' / rule / f'kyverno-cel-{rule}.yaml')
vap = required_expression(tmp_path / 'vap-cel' / rule / f'vap-{rule}.yaml')
assert cel == vap
```

This is the property that makes "one authored value, several engines" a claim rather than a
slogan. It is also the first thing to break when someone hand-formats a list instead of using
`tojson`.

## 4. Against a real API server

Schema validity is not expressiveness. A generated policy can be accepted by the API server and
still assert nothing.

- **Dry-run** every generated document before trusting it. This catches a CEL compile error,
  which no offline check will. Use **server-side apply**:

      kubectl apply --server-side --force-conflicts --dry-run=server -f <file>

  Plain `kubectl apply --dry-run=server` is not equivalent and will lie to you. On an object
  that already exists it computes a client-side patch first, and a failure there is reported as
  `error when applying patch` **before the request ever reaches the engine's admission
  webhook** — so a policy the engine would reject is reported as accepted. Measured: a template
  set that showed 6 of 6 accepted this way was 2 of 6 under server-side apply, the other four
  rejected by Kyverno's own webhook.
- **Confirm the negative.** Apply the policy and submit an object that should be rejected. A
  check that has never refused anything is indistinguishable from a check that always passes.
- For a reporting engine, wait for the background scan before collecting. There is no fixed
  interval worth sleeping; poll until every declared check has a result, or the measurement is
  of the wrong set. Collecting too early produces an empty posture that looks like a clean one.
