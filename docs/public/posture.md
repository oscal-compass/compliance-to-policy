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
state.

**An engine that could not decide has not found a violation.** An `error` from an engine, or a
`skip` from a matching PolicyException, is the absence of a verdict. Neither is counted in the
pass/fail tally; both reduce `assessedChecks`. A posture that counted them as failures would send
someone looking for a breach that is not there.

#### How a control's `state` is decided

The first row that matches wins. `fail`, `warning` and `pass` are verdicts; `error` and `skip`
are the absence of one.

| Condition | `state` | `reason` |
|---|---|---|
| The component definition declares no check for the control | `not-assessed` | `other` |
| It declares checks, but none of them reported | `not-assessed` | `other` |
| Checks reported, and every one of them reached no verdict | `not-assessed` | `other` |
| Any subject `fail`s | `not-satisfied` | `fail` |
| Any subject is a `warning` — non-compliant, reported by the engine as a warning rather than scored | `not-satisfied` | `other` |
| Subjects passed, and some subject reached no verdict | `not-assessed` | `other` |
| Every subject passed | `satisfied` | `pass` |

A check's own `result` follows the same idea one level down: the strongest verdict any of its
subjects reached, or `error` if none of them reached one, or `not-assessed` if no evidence
covered the check at all.

#### Known limitations

Two cases where the model does not yet say what it should. Both are tracked for a later change,
and until then a reader of a posture needs to know about them:

- **A check that errored and a check that stayed silent are not treated alike.** Two passing
  checks plus one that errored give `not-assessed`. Two passing checks plus one that reported
  nothing give `satisfied`, with `coverage` showing 2 of 3. Both are the same fact — one of the
  three checks did not answer — so one of the two states is wrong. The coverage figure is the
  reliable signal in both cases.
- **A rule with no checks is not in the coverage denominator.** `expectedChecks` counts checks
  declared by validation components. A rule that no validation component implements contributes
  nothing to either side of the fraction, so a control whose rules are entirely unimplemented
  reports `0 of 0` — with `state` `not-assessed`, which at least says something is wrong.

  The misleading case is a partial one. A control with one implemented rule that passes and
  another rule that no validation component implements reports `satisfied`, with `coverage`
  `1 of 1`: the unimplemented rule is in neither side of the fraction, so nothing in the state
  or the coverage says it is there. It survives in the JSON as a `rules[]` entry with an empty
  `checks` list, and is absent from the `md` and `html` views altogether. Until this changes,
  comparing a control's `rules[]` against the component definition is the only way to see it.

### `c2p tools posture-compare`

The two inputs are `-f json` postures, so produce them first — one from each collection:

```
c2p tools result2posture -ar /tmp/assessment-results.json     -cdef <cdef> -f json -o /tmp/posture.json
c2p tools result2posture -ar /tmp/assessment-results-vap.json -cdef <cdef> -f json -o /tmp/posture-vap-only.json
```

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

`-f json` writes this, and is the default. Serialization omits empty fields, so a field marked
`?` below is absent rather than null when it has no value.

As it stands a `not-assessed` check carries no `checkId`, since there was no reported check to
name. That is current behaviour rather than a guarantee — giving unassessed checks their declared
check id is a change under consideration.

```
schemaVersion  "c2p-posture/v1"
framework      str?
generatedAt    str?
source         { assessmentResults?, componentDefinition?, catalog? }   the inputs, as paths
               assessmentResults is one string: a repeated -ar is joined with ", "
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
