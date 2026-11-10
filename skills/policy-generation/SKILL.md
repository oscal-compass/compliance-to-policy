---
name: policy-generation
description: Generate the policy templates compliance-to-policy renders into engine-native policy. Given an OSCAL component definition and the name of a policy engine (PVP), produce one template per check that engine declares. Use when user wants to generate policy from OSCAL for a named engine, add an engine to an existing component definition, write a Kyverno ClusterPolicy / Kyverno CEL / ValidatingAdmissionPolicy template for a check, or decide whether an engine can express a check at all.
argument-hint: Optional engine or check id (e.g., "vap-cel", "image-signature-verified")
license: Complete terms in LICENSE.txt
---

# Policy Generation

## Purpose

Generate the policy templates c2p renders. **Given a component definition and the name of a
policy engine, produce one template per check that engine declares.**

A component definition says *which* control a rule satisfies and *which* check verifies that
rule. It does not say what the check does on a cluster. That is the template's job, and this
skill generates them.

The templates are the engine expertise in this pipeline, so this is not a mechanical
transformation and the skill does not pretend it is. Two things make it work: the component
definition has to carry enough specificity (see **Inputs**), and the result has to be verified
on a cluster rather than by reading it (see [verification.md](verification.md)). Expect a first
pass that a cluster corrects.

## Inputs

| Required | |
|---|---|
| **The component definition** | which engines are declared as validation components, which checks each declares, the rule and check descriptions, and the resolved parameter values |
| **The engine (PVP) to generate for** | named by the user, matching a validation component's title |

| Optional, and worth asking for | |
|---|---|
| **What the target system is** | beyond the component descriptions: the platform, its version, what is installed, what the namespace layout is |
| **Engine version** | `kyverno.io/v1 ClusterPolicy` is gone from Kyverno 1.20; `policies.kyverno.io/v1` needs 1.19+; `ValidatingAdmissionPolicy` needs Kubernetes 1.30+ |
| **An existing template for the same rule in another engine** | so one authored parameter renders identically in both |

### How specific the descriptions have to be

This skill can only generate what the component definition already decided. A rule description
of "images must be signed" yields a guess; the generation works when the descriptions name the
mechanism and the subject. From a measured example that does work:

> **rule** — *Every container image is pulled from a registry that has an entry in the
> cluster-held approved-registry ledger (one ConfigMap per registry labelled
> `supply-chain.example.com/ledger=approved-registry`). The allow-list is cluster state and
> changes without redeploying policy.*
>
> **check** — *ClusterPolicy `context.apiCall` lists ConfigMaps labelled
> `supply-chain.example.com/ledger=approved-registry` in `approved_registry_ledger_namespace`
> and denies any image whose registry host is absent from the ledger.*

The rule says what must be true and why; the check says, for that engine, which mechanism
establishes it. That is the granularity to require. If the component definition is thinner than
this, the honest move is to improve it rather than to invent the missing decisions here — those
are compliance decisions, and they belong to whoever authored the component definition.

A validation component's own description is worth reading too. In the same example, `vap-cel`
describes itself as *"Admission time only. No external data and no background scan"* — which is
the constraint that decides whether a check is expressible at all.

## The division of labour

This is the whole point, so be explicit about it before writing a line of YAML.

| Carried by the component definition | Carried by the template |
|---|---|
| which control the rule satisfies | the policy's `apiVersion` and `kind` |
| which check verifies the rule | which resources it matches |
| the parameter the organisation chose, and its value | the expression or pattern that evaluates the rule |
| the rule and check descriptions | the message a violation produces |
| which engines are expected to implement the rule | any companion object the engine needs |

Neither side can be derived from the other. The OSCAL document holds compliance decisions; the
template holds a statement about Kubernetes. Keeping that line clean is what lets the same
authored parameter reach three engines, and what lets a control outlive the engine that
currently enforces it.

**Do not restate in a template anything the render context already carries.** The render context
is small on purpose — see [template-contract.md](template-contract.md).

## Main tasks

1. Read the component definition and list the checks the named engine declares
2. For each check, decide whether that engine can express it. If it cannot, say so and
   **generate no template** — the component definition should stop declaring it
3. Generate one template per remaining check
4. Render, and read the output rather than the template
5. Dry-run against a cluster, confirm a negative, and fix what the cluster rejects

## Sub-skills

- [template-contract.md](template-contract.md) — **START HERE** — file naming, the render
  context, the filters, output paths, and emitting literal braces
- [engines.md](engines.md) — the three Kubernetes engines, the shape each one wants, and what
  each one cannot do
- [verification.md](verification.md) — byte-for-byte parity, coverage, and dry-run

## Workflow

1. Confirm the component definition and the engine to generate for. Ask for the optional inputs
   above; the engine version in particular changes what is available
2. List the checks that engine declares. Every one needs a template, or the render fails
3. Read each check description for the mechanism it names, and the rule description for what
   must be true. Where they disagree, ask — do not pick one
4. Generate `<template_dir>/<Check_Id>.yaml.tmpl` per check
5. Render with the engine's plugin and read the output, not the template
6. If another engine already implements the same rule, diff the rendered expressions: an
   authored parameter must reach both identically
7. Dry-run every generated document against an API server, then apply it and submit something
   it should reject
8. Fix what the cluster rejects, and record the fix — engine dialect failures recur. Several are
   already collected in [engines.md](engines.md)
9. Commit the rendered output as a reference, and add it to the parity test

## Key learnings

### A missing template is an error, not a silent omission

`policy_template.render` raises on a check that has no template. This is deliberate: a
component definition declaring a check the engine never received is a control that silently
stops being assessed, which is worse than a failure — a red control is a finding someone can
work, an absent one looks like nothing happened.

### If an engine cannot express a check, do not write the check

A template that always passes is the most expensive thing in this pipeline. It makes a control
green with no evidence behind it. Declaring no check for that engine is honest and visible: the
control comes back `not-assessed`, with coverage showing which engines were expected and which
reported.

Two of the reference set's six rules have no ValidatingAdmissionPolicy check for exactly this
reason, and that absence is a finding about the engine rather than a gap in the work.

### One authored parameter must render identically across engines

If two engines express the same rule, the parameter has to arrive the same way in both. A CEL
expression is compared as a string by a cluster, so one space in a list literal means two
engines silently disagree about a control. The `tojson` filter emits compact separators for
this reason; do not hand-format a list.

### Do not duplicate the component definition inside the template

A measured example from the bundled reference set: `compliance.oscal/control-ids` is a literal
in all sixteen templates, because the render context carries no control ids — `RuleSet` has the
rule and the check, not the controls they map to. It is correct today and will go stale the
first time a rule's control mapping changes, and the parity test cannot catch it, because the
template is where the expected value comes from.

So when the context does not carry something, either check it against the component definition
outside the template, or carry it in the context. Do not quietly copy it and hope.
