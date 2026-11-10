## The compliance posture

OSCAL Assessment Results say what each engine observed. A posture says what that means for each
**control**: its state, the checks behind it, and how much of the control was actually assessed.

Two commands produce it, and a documented JSON model underneath so that it is a document rather
than a rendering:

| Command | |
|---|---|
| `c2p tools result2posture` | Assessment Results → a posture, as `json`, `md` or `html` |
| `c2p tools posture-compare` | two postures, control by control |

#### Prerequisite

None beyond C2P. Both commands read committed files and contact no cluster.

### `c2p tools result2posture`

```
c2p tools result2posture \
    -ar /tmp/assessment-results.json \
    -cdef ./samples_public/kubernetes/component-definition.json \
    --catalog /tmp/nist_800_53_resolved/catalog.json \
    -f html -o /tmp/posture.html
```

| Flag | |
|---|---|
| `-ar`, `--assessment-results` | OSCAL Assessment Results. **Repeat the flag** to combine several — one C2P run assesses one PVP, so a cluster with several engines collected by separate pipelines produces one document per engine, and the posture is where they come back together |
| `-cdef`, `--component-definition` | OSCAL Component Definition. The posture is accountable for every check this declares, which is what lets an unassessed control be reported as such rather than omitted |
| `--catalog`, `-cat` | OSCAL Catalog, used to title the controls |
| `-n`, `--framework` | Compliance framework name recorded in the posture |
| `-f`, `--format` | `json` (the `c2p-posture/v1` document), `md` or `html` (views over it) |
| `--out`, `-o` | Output file. Omit to write to stdout |

**What it is accountable for.** The expected set of checks comes from the component definition,
not from the evidence. A check the component definition declares and no engine reported is a
`not-assessed` row rather than a missing one, and every control carries `coverage` alongside its
state. A control nothing could evaluate is `not-assessed` — neither compliant nor
non-compliant — because reporting it as either would invent a verdict the evidence does not
carry.

**An engine that could not decide has not found a violation.** An `error` from an engine, or a
`skip` from a matching PolicyException, is the absence of a verdict. Neither is counted in the
pass/fail tally; both reduce `assessedChecks`. A posture that counted them as failures would send
someone looking for a breach that is not there.

### `c2p tools posture-compare`

```
c2p tools posture-compare \
    -b /tmp/posture.json -p /tmp/posture-vap-only.json \
    --baseline-label 'every engine' --posture-label 'vap-cel only' \
    --condition 'when only ValidatingAdmissionPolicy is available'
```

| Flag | |
|---|---|
| `-b`, `--baseline` | the posture to compare against |
| `-p`, `--posture` | the posture being compared |
| `--baseline-label`, `--posture-label`, `--condition` | labels used in the rendered output |
| `-f`, `--format` | `text` or `json` |
| `-o`, `--out` | Output file. Omit to write to stdout |

A control counts as **dropped** when the baseline reported a state for it and the other posture
either does not mention it or could not assess it. Both are the same thing to a reader: the
control is no longer being answered. This is how the cost of a narrower set of engines is shown —
run a generation with one engine only, collect, and compare: the controls that engine cannot
express stop being reported, and this names them.

### The `c2p-posture/v1` document

`-f json` writes this. Serialization omits empty fields, so an unassessed check carries no
`checkId` rather than a null one.

```
schemaVersion  "c2p-posture/v1"
framework      str?
generatedAt    str?
source         { assessmentResults?, componentDefinition?, catalog? }   the inputs, as paths
components[]
  title        str
  controls[]
    controlId  str
    title      str?
    state      "satisfied" | "not-satisfied" | "not-assessed"
    reason     "pass" | "fail" | "other"      OSCAL ObjectiveStatus.reason, a classifier
    remarks    str?                           human-readable; the counts behind the state
    coverage   { expectedChecks, assessedChecks }
    evidence[] { description?, href }
    rules[]
      ruleId   str
      description str?
      checks[]
        checkId  str?        absent when nothing reported this check
        engine   str?        the validation component's title
        result   "pass" | "fail" | "warning" | "error" | "not-assessed"
        remarks  str?
        evidence[]
        subjects[]
          title       str
          resourceId  str?
          result      str
          reason      str?
          evaluatedOn str?
          props       { str: str }?
summary        { satisfied, notSatisfied, notAssessed }
```

`state` is a field rather than something a template computes. That is the point of the model: a
posture that only exists inside rendered markdown cannot be diffed, served or re-rendered, and
`md` and `html` are views over this document rather than the thing itself.

`reason` carries only the three values OSCAL's `ObjectiveStatus` permits. Anything a person
should read goes in `remarks`.

A worked example of the whole chain, with a committed `posture.json`, `posture.html` and a
`posture-compare` run against a real cluster, is the
[end-to-end walkthrough](/docs/end-to-end/README.md).
