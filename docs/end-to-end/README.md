# KubeCon NA 2026 demo: from a NIST control to a cluster, and back

NIST SP 800-53 rev 5 → OSCAL → three Kubernetes policy engines → evidence back into OSCAL → a
control-by-control compliance posture. On a real cluster, and every link in the chain is an OSCAL
Compass component.

[![Watch the demo on YouTube](https://img.youtube.com/vi/pk3EsgA4EBQ/maxresdefault.jpg)](https://www.youtube.com/watch?v=pk3EsgA4EBQ)

The video is an introduction and moves at a presentation's pace, so it does not look command for
command like this page — it shows the shape of the chain, and this is where it can be reproduced.

**What it shows.** Two components doing an entire compliance lifecycle between them. The
[`agentic-agile-authoring`](https://github.com/oscal-compass/agentic-agile-authoring) skills author
the OSCAL: a person writes two prose documents and gives five prompts, and the skills produce the
profile, the organisation-defined parameter values, the resolved catalog and the component
definition, issuing every trestle operation through the
[`compliance-trestle-mcp`](https://github.com/oscal-compass/compliance-trestle-mcp) server.
`compliance-to-policy` does everything after that: policy for three Kubernetes engines, evidence
collection once that policy is on a cluster, aggregation into one OSCAL Assessment Results, and a
posture carrying control status, per-check results and reasons. Putting the policy on the cluster
sits between those two, and it is not automated here: a person reads what was generated first. The
code that then drives the cluster is a worked example under [`reference/`](reference/), outside the
`c2p` library on purpose.

The result is nine NIST SP 800-53 controls, each with a status and the evidence behind it, traceable
back to the control a person selected. No code was written for this demo.

## Read it instead

**[WALKTHROUGH.md](WALKTHROUGH.md)** — one chain of ten steps, with the real inputs, the real
stdout and the real output file at each one. Captured on 2026-11-10 against Kubernetes 1.34.0 and
Kyverno 1.19.2.

Steps 1 to 6 are the skills authoring the OSCAL and step 7 is a skill writing the policy templates
— which is what C2P++ names, and all it names. Step 8 is `compliance-to-policy` taking the
component definition and those templates and returning assessment results and a posture from a
live cluster. It is four
commands rather than one, because generating policy and trusting it are separate acts and a person
belongs between them. Steps 9 and 10 re-run part of that against a single engine, which produces a
posture with a coverage gap in it — an estate needs to see what its tooling could not look at.

Steps 3 to 7 are prompts rather than commands. They go to an agent running the
[`agentic-agile-authoring`](https://github.com/oscal-compass/agentic-agile-authoring) skills, which
issue every OSCAL operation through the
[`compliance-trestle-mcp`](https://github.com/oscal-compass/compliance-trestle-mcp) server; step 7
goes to the [`policy-generation`](../../skills/policy-generation/) skill in this repository, which
writes files rather than issuing OSCAL operations. All five prompts are committed verbatim, with
frames from the recorded runs.

Steps 8 to 10 are this repository:
[`samples_public/kubernetes/generate.py`](../../samples_public/kubernetes/generate.py),
[`reference/deploy.py`](reference/deploy.py),
[`reference/collect_and_assess.py`](reference/collect_and_assess.py) and `c2p tools
posture-compare`. Step 8 takes step 6's component definition as its `-c`. The per-engine templates
it renders came from [`skills/policy-generation/`](../../skills/policy-generation/), also here.

Generating the policy needs no cluster — that is step 8a, plain C2P over committed files, and you
can read everything it produces. Applying and collecting do. Every other step is a function of
files committed in this directory, step 10 included.

## What is in this directory

| | |
|---|---|
| [`prompts/`](prompts/) | the five prompts from steps 3 to 7, verbatim |
| [`authoring/`](authoring/) | the two prose documents a human writes, and step 6's `cd.csv` |
| [`oscal/`](oscal/) | the profile, the resolved catalog and the component definition |
| [`out/`](out/) | step 8 — 16 policy documents, the Assessment Results, the posture |
| [`out-vap/`](out-vap/) | step 9 — the same with one engine: an Assessment Results and the posture showing the coverage gap |
| [`reference/`](reference/) | the code that applies policy and collects evidence — a worked example, not part of the library |
| [`capture/`](capture/) | `.cmd`, `.out` and `.rc` for steps 8, 9 and 10 |
| [`images/`](images/) | the frames embedded in steps 1 to 6 |

## The three engines

An estate does not run one policy engine, so the component definition declares three, and the same
authored control produces policy for all of them.

| engine | apiVersion | kinds | checks |
|---|---|---|---|
| `kyverno-dsl` | `kyverno.io/v1` | `ClusterPolicy` | 6 |
| `kyverno-cel` | `policies.kyverno.io/v1` | `ValidatingPolicy`, `ImageValidatingPolicy` | 6 |
| `vap-cel` | `admissionregistration.k8s.io/v1` | `ValidatingAdmissionPolicy` + `Binding` | 4 |

`kyverno.io/v1 ClusterPolicy` is deprecated in Kyverno 1.19 and gone in 1.20;
`policies.kyverno.io/v1` is its replacement. Generating both from one component definition is the
case for authoring the control rather than the policy: the control outlives the engine.

`vap-cel` declares four checks rather than six because two rules carry no ValidatingAdmissionPolicy
check — an authored decision recorded in those rules' descriptions, since a check that always passed
would make a control green with no evidence behind it. Step 6 of the walkthrough quotes the
reasoning from the artefact.
