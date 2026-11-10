# The template contract

What a template is given, what it must be called, and where its output lands. Everything here
is enforced by `plugins_public/plugins/policy_template.py`; read that file if anything below
looks out of date.

## Naming and layout

```
<template_dir>/<Check_Id>.yaml.tmpl        ->  <deliverable_dir>/<Rule_Id>/<Check_Id>.yaml
```

One template per **check**, named after the check id exactly. One directory per engine, and the
plugin is pointed at the directory for the engine it generates.

The output path carries the **rule as well as the check**, so a rule implemented by several
engines does not collapse into one file and the generated tree can be applied per rule.

A check with no template raises `C2PError`. Do not catch it: see "A missing template is an
error" in [SKILL.md](SKILL.md).

## The render context

Jinja2, with this and nothing else:

| In the template | Source |
|---|---|
| `{{ engine }}` | the engine name the plugin was constructed for |
| `{{ rule.id }}` | `Rule_Id` |
| `{{ rule.description }}` | `Rule_Description` |
| `{{ check.id }}` | `Check_Id` |
| `{{ check.description }}` | `Check_Description` |
| `{{ params.<id> }}` | the resolved parameter value |

The context is structured rather than a flat namespace of parameters because a generated policy
has to carry its own provenance — which rule and check produced it, by which engine — and a
flat namespace cannot name any of that. Put those three in labels on every policy you generate;
the collector reads results back by check id.

**`params` holds resolved values**, not defaults: numbered parameter columns
(`$Parameter_Id_2`, …) and `set-parameters` are both read before rendering. Several parameters
per rule is normal — declare them by repeating the four parameter columns under a common
suffix.

**There are no control ids in the context.** `RuleSet` carries the rule and the check, not the
controls they map to. If a generated policy needs a control id, see the last section of
[SKILL.md](SKILL.md) before copying one in.

## Filters

| Filter | Does | Use for |
|---|---|---|
| `quote` | a double-quoted scalar | a value that has to survive as a YAML string |
| `tojson` | compact JSON, no space after separators | a list that has to be a CEL list literal, or a string that has to be a CEL string literal |
| `oneline` | collapses every line break to a space | a value going into a comment rather than a field |

`quote` and `tojson` are `json.dumps`, so their output is always ASCII, always balanced, and
never contains a line break. That is the property the rules below rest on.

The compact separators are not cosmetic. The same parameter has to render identically in every
engine's template, and a cluster compares a CEL expression as a string.

## Substituting a value from the component definition

`rule.description`, `check.description` and every `params.<id>` come from the component
definition, which in this flow an agent wrote. They are **data, not fragments of the document
you are generating**. A description may carry a carriage return, a parameter may carry a quote,
and trestle will not stop either: OSCAL validation refuses `LF` in a prop value but accepts
`CR`, `U+2028` and `U+0085`, as well as `"` and `'`.

Three rules, and they are not style preferences — each one is a way a value has escaped the
field it was written into:

**1. A filter's output is the whole token. Never add quotes around it.**

```jinja
expression: {{ (params.required_workload_labels.split(' ') | tojson) | quote }}
→           "[\"app.kubernetes.io/name\",\"app.kubernetes.io/part-of\"]"
```

Not `'{{ ... | tojson }}'`. A value containing `'` closes the scalar you opened, and the
generated document stops being the document you wrote — or stops being YAML at all, with no
error at generation time.

**2. A value inside a larger expression is concatenated in, never interpolated in.**

```jinja
expression: {{ ('resource.List("v1", "configmaps", ' ~ (params.ledger_namespace | tojson) ~ ')') | quote }}
messageExpression: >-
  'registry not in the ledger in namespace ' + {{ params.ledger_namespace | tojson }} + '. Offending: ' + ...
```

`tojson` of a string is a CEL string literal as well as a JSON one, so the value arrives as
one inert operand. Written the other way — `"{{ params.ledger_namespace }}"` inside a CEL
string — a value ending in `")` rewrites the expression the cluster compiles, and the result
is still valid YAML holding a valid CEL expression, so nothing downstream notices.

**3. A value in a comment goes through `oneline`.**

```jinja
# {{ check.description | oneline }}
```

A `CR` in a description otherwise ends the comment and starts a document of its own. The same
applies to a value inside a folded block scalar (`>-`), where a line break breaks the block.

A value used as a **mapping key** needs `quote` too — `{{ k | quote }}: "?*"`, not
`{{ k }}: "?*"`.

## What is checked after rendering

`policy_template.render` parses what it rendered before writing it, and raises `C2PError`
unless every document is a mapping whose `kind` is one the engine's plugin declares it
generates and whose `metadata.name` is the check id. A template that emits a document the
plugin never listed, or a second document under another name, fails generation.

This is a backstop, not a substitute for the rules above. It catches a value that broke out
into a new document; it cannot catch one that rewrote a CEL expression in place, because the
result is a well-formed document of the right kind under the right name. Only rule 2 catches
that.

Templates are also rendered in Jinja2's `SandboxedEnvironment` with `StrictUndefined`. A
template is executable, and in this flow it is written by an agent from untrusted input, so it
does not get the Python object graph. `StrictUndefined` is the other half: a mistyped
`params.foo` is a hard error rather than an empty string, because a check that silently
asserts nothing reports every workload compliant.

## Emitting literal braces

Kyverno's own JMESPath variables are written `{{ ... }}`, which Jinja2 would consume. Bind them
once at the top of the template and use the bindings:

```jinja
{%- set o = '{{' -%}
{%- set c = '}}' -%}
...
      - key: "{{ o }} split(element.image, '/') | [0] {{ c }}"
```

Both trim markers are needed, or the bindings leave blank lines at the top of the generated
document — and the generated document is compared byte for byte.

## A worked skeleton

```jinja
# Generated by C2P ({{ engine }}) — rule {{ rule.id }} / check {{ check.id }}
# {{ check.description | oneline }}
apiVersion: <engine's group>/<version>
kind: <engine's kind>
metadata:
  name: {{ check.id }}
  labels:
    compliance.oscal/rule-id: {{ rule.id }}
    compliance.oscal/check-id: {{ check.id }}
spec:
  ...
```

The header comment is worth keeping. A generated policy found loose on a cluster should say what
produced it and from which rule.
