# Captured output

Three files per command: `.cmd` (what ran, with every variable expanded), `.out` (stdout and
stderr, merged, verbatim) and `.rc` (the exit code).

The order is the order the review gate imposes. Generation and deployment are separate commands
because reading the generated policy in between is the step that is not automated — see
[`../reference/README.md`](../reference/README.md).

| | |
|---|---|
| `B0-policy-templates.*` | the `policy-generation` skill writing one template, and its report. `.generated.yaml.tmpl` is what that run produced — a second pass, not one of the sixteen the walkthrough uses |
| `B1-generate.*` | generate policy for three engines and stop, contacting no cluster |
| | *a person reads `out/` — no output, so no capture* |
| `B2-apply.*` | the reference implementation applies the reviewed tree, unmodified |
| `B3-collect.*` | collect from all three engines, assess, posture |
| `B4-collect-vap-only.*` | the same with `-e vap-cel`, which is how the coverage gap is shown |
| `B5-posture-compare.*` | the two postures, compared |

Steps before `B0` have no captures. They are prompts, and their output is the OSCAL artefacts
committed under `../oscal/` and `../authoring/`, with frames from the runs under `../images/`.

`B0` was run in a scratch workspace holding the component definition, with the skill installed
under `.claude/skills/`; the paths in its output are that workspace's. The frame in step 7 of the
walkthrough is a **different run of the same prompt**, taken from a terminal rather than from
`claude -p`, so the two do not agree line for line. Step 7 is the one step here that does not
reproduce exactly, and the walkthrough says so where it shows them.

Captured against the demo cluster: kind, Kubernetes v1.34.0, Kyverno v1.19.2.

Some of these have been brought up to date since that run, so that what is committed here agrees
with what the committed artefacts and the current code produce rather than with the code as it
stood in November. In full:

| | |
|---|---|
| `B1` last line | lost the sentence `generate.py` used to print pointing here |
| `B2` first line, `B3` and `B4` first line | gained the `context:` line both programs now print before they write |
| `B2` last lines | the `kubectl delete` the program prints, which now names the documents it applied |
| `B3` and `B4` posture tables | `result2posture` over the committed `assessment-results.json`, re-rendered when the roll-up stopped counting a check that could not evaluate as a failure -- the numbers are the same evidence, read correctly |

The posture tables were re-rendered by running the renderer. The rest were edited to match what
the programs now print, not re-run -- the cluster they were recorded against is gone. Every line
that came from that cluster is as it was.

`B0`'s `.generated.yaml.tmpl` predates the template contract's rules on substituting a value
from the component definition, so its header comment lacks the `oneline` filter the contract now
requires. It is left as the run produced it; the sixteen templates the walkthrough uses are
under `samples_public/kubernetes/templates/` and follow the contract.

`B1`, `B2` and `B3` are three programs on purpose. Generating is C2P over committed files and needs
no cluster; applying and collecting need one and are a reference implementation. Keeping them apart
is what puts the person reading, between `B1` and `B2`, somewhere other than inside a flag.
