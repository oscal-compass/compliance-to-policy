# Component definition — authoring instructions

How this system's rules are decomposed, and what the component definition has to say about
them. The decomposition is **fixed**: this estate's policy templates were generated against
it, so rule ids and check ids must match exactly.

## Components

Two service components:

| component | what it covers |
|---|---|
| `container-supply-chain` | provenance of images admitted to the cluster |
| `three-tier-app` | the workloads themselves |

Three validation components, which are the policy engines:

| component | engine |
|---|---|
| `kyverno-dsl` | `kyverno.io/v1` ClusterPolicy |
| `kyverno-cel` | `policies.kyverno.io/v1` |
| `vap-cel` | upstream ValidatingAdmissionPolicy — admission time only, no external data, no background scan |

## Rules and their control mappings

| rule | controls | engines |
|---|---|---|
| `image-signature-verified` | si-7.15 si-7 | kyverno-dsl, kyverno-cel |
| `image-registry-approved` | cm-7.5 cm-8.3 | kyverno-dsl, kyverno-cel |
| `image-pinned-by-digest` | cm-2 si-7 | kyverno-dsl, kyverno-cel, vap-cel |
| `container-runs-unprivileged` | ac-6 sc-39 | kyverno-dsl, kyverno-cel, vap-cel |
| `workload-inventory-labelled` | cm-8 | kyverno-dsl, kyverno-cel, vap-cel |
| `db-credentials-from-secret` | ia-5 | kyverno-dsl, kyverno-cel, vap-cel |

## Check ids

`<engine>-<rule-id>`, except `vap-cel`, which uses `vap-<rule-id>`.

Sixteen checks in total: six, six and four.

## Why the first two rules carry no vap-cel check

This is deliberate, not an omission, and the rules' descriptions have to say so rather than
leaving a weak check in place of the explanation.

- **`image-signature-verified`** — an upstream ValidatingAdmissionPolicy evaluates CEL inside
  the API server. It can neither fetch the signature from the registry nor make the
  transparency-log round trip that cosign verification requires.
- **`image-registry-approved`** — an upstream ValidatingAdmissionPolicy has no resource
  library, so its CEL cannot test membership against a list held in cluster ConfigMaps.

## What the descriptions must carry

A description is the contract a policy template is generated from. Write rule and check
descriptions that name the **mechanism and the subject**, not just the intent.

Carry the parameters the rules need, including:

- the approved signer identity and its OIDC issuer
- the approved-registry ledger namespace, label and data key
- the required label keys
- the credential env keys

## Working order

Author the CSV, render the markdown preview and run the conversion in one pass. Do not hold
the conversion waiting for sign-off on the CSV — if the conversion rejects a row, fix the row
and run it again.
