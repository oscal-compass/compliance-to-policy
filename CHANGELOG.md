# CHANGELOG

## v1.0.0 (2026-11-10)

### Breaking

* feat!: major version release 1.0.0 (#62)

An empty commit, so that the release workflow cuts 1.0.0.

python-semantic-release derives the version from the commits since v0.5.0.
Those are `feat` and `fix`, which compute 0.6.0 -- running the workflow as it
stands would release that. The `!` in this subject is what raises the major
instead, and since there is no code change to attach it to, it goes on a
commit of its own.

That also means the subject has to survive the merge. Squashing this pull
request replaces it with the squash commit&#39;s own message, which is how #61
landed on main as `Merge pull request #61` and left nothing for the major
bump to key off. Either squash with this subject, or use a merge commit --
then this commit stays in the history as a parent and its subject is read
from there.

1.0.0 is #58, #59 and #60: policy generation for three Kubernetes engines
from one component definition, several engines aggregated into one OSCAL
Assessment Results, a control-level compliance posture, the agent skill that
writes the policy templates, and the documentation and end-to-end walkthrough
for all of it.

Upgrading from 0.5.0, note how rule parameters resolve -- see &#34;Rule
parameters&#34; in the README.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`91acafb`](https://github.com/oscal-compass/compliance-to-policy/commit/91acafbf1057df757a5597886b63a8ecd1cdb391))

### Chore

* chore: add a directive for C2P Go new repository

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`56a48a1`](https://github.com/oscal-compass/compliance-to-policy/commit/56a48a1727d2718f538d3619a8c7a18ac445aea6))

### Ci

* ci: sign the release commit and tag (#63)

Branch protection requires signed commits, and the release commit was
unsigned, so semantic-release could not push it.

The action already knows how to do this: given an SSH key pair it writes it
to ~/.ssh, loads it into an ssh-agent, and sets `gpg.format=ssh`,
`commit.gpgsign` and `tag.gpgsign` globally. That works here because
semantic-release creates its commit and tag by shelling out to git --
`repo.git.commit()` and `repo.git.tag(&#34;-a&#34;, ...)` -- rather than writing
objects through GitPython, so git&#39;s own signing configuration applies. The
container has openssh: `python:3.12-bookworm` comes from
`buildpack-deps:bookworm-scm`, which installs `openssh-client`.

`git_committer_name` and `git_committer_email` are not optional here. The
action refuses to configure signing without the email, and GitHub marks a
signature Verified only when the signing key belongs to the committer&#39;s
account and the committer email is verified on it. Setting both also makes
semantic-release use them as the commit author, through the
`GIT_COMMIT_AUTHOR` environment variable the action exports.

`Check signatures` is there because a green release is not evidence that any
of this worked: without it, a misconfigured key produces unsigned objects and
every later step still passes. It reads the objects that were pushed, then
asks GitHub whether it verified them, and fails on either.

While here, the release commit gets a subject. `commit_message` was the
python-semantic-release default repeated verbatim, which is a bare version
number -- `0.5.0`, `0.4.0` -- that no commit parser recognises, so every
release commit has been landing under &#34;Unknown&#34; in the changelog.
`chore(release):` parses, and `chore` is in neither `minor_tags` nor
`patch_tags`, so the release commit still cannot bump the next version.

Verified against the action&#39;s `action.sh` and `gitproject.py` at v9.8.7, and
by reproducing the same git configuration locally: `git tag -a` does sign
when `tag.gpgsign` is set, and the check fails as it should on an unsigned
commit and on an unsigned tag.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`241466c`](https://github.com/oscal-compass/compliance-to-policy/commit/241466c8ad1c1630f4b107e1efa3074b2f12b3f9))

* ci: move the artifact actions to v4

actions/upload-artifact@v3 and actions/download-artifact@v3 were shut down
on 2024-11-30, so the release workflow fails at the first upload step. v4 is
a drop-in for how this workflow uses them: one upload of dist/, one download
by the same name.

Independent of anything else in flight, and worth merging first so a release
can be cut while the rest is reviewed.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`362df8f`](https://github.com/oscal-compass/compliance-to-policy/commit/362df8feade8a9b72407438d054940db256f850b))

### Documentation

* docs: the partial case is the one that misleads

&#34;Known limitations&#34; had the all-or-nothing case: a control whose rules are
none of them implemented reports `0 of 0`. Nobody is misled by that -- the
state is `not-assessed`, which says something is wrong.

The case that misleads is partial. One implemented rule that passes and one
rule no validation component implements reports `satisfied` with coverage
`1 of 1`: the unimplemented rule is in neither side of the fraction, so
neither the state nor the coverage says it exists. Measured. The rule does
survive in the JSON, as a `rules[]` entry with an empty `checks` list -- it
is the `md` and `html` views it disappears from -- so comparing `rules[]`
against the component definition is how to find it until this changes.

Also, the parameter table said a duplicate id resolves to the later value
because rendering folds the list into a mapping. That is true of every plugin
that builds such a mapping, which is the three Kubernetes plugins and also
`kyverno` and `ocm`; `auditree` iterates the list and sees both. Named now.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`8f8676c`](https://github.com/oscal-compass/compliance-to-policy/commit/8f8676cd188ae01d03a04a1c1eb4893aa7b4503c))

* docs: how state is decided, what the plugins read back, and the rest of the parameter change

Three things the review asked for, and the smaller points with them.

**The parameter behaviour change was under-reported.** It named the lookup
order and nothing else. Three more things differ from before #58, and the
first can change what an existing component definition generates without
anyone touching it: parameters used to be read from the last non-validation
component only -- the loop reassigned its accumulator each time round -- and
are now read from every one of them. A parameter declared twice resolves to
the later value, silently, because rendering folds the list into one mapping.
And a parameter with no value anywhere renders as an empty string rather than
failing: `StrictUndefined` does not catch it, since the variable is defined,
so a check can be left asserting nothing.

**How a control&#39;s state is decided was not written down**, which is the thing
a reader of `c2p-posture/v1` most needs. `posture.md` now has the priority
table, matched row by row against `_roll_up`, and a &#34;Known limitations&#34;
section for the two cases #58 left as follow-ups: a check that errored and a
check that stayed silent do not produce the same state, and a rule no
validation component implements is in neither side of the coverage fraction.
Both are stated as limitations rather than left for a reader to discover.

**The result half of each plugin was missing.** `generate_pvp_result` takes
PolicyReports for the two Kyverno engines -- as a list or as the whole
`kubectl get -o json` envelope -- with the result mapping spelled out, and
for `vap` either a list of probe outcomes or a `c2p-evidence/v1` document,
which until now appeared only in the code and a test. Converting needs no
plugin configuration, which the page said nothing about. The applying step
now links the reference implementation #59 added.

Two claims were stronger than the code: &#34;neither is taken on trust&#34; now says
which two gaps remain open, and the paragraph on declaring no check now says
it holds only while some engine declares one.

Smaller points: `-f` defaults to `json`; a repeated `-ar` is joined into one
`source.assessmentResults` string; both posture-compare examples say where
their two inputs come from; and `checkId` on an unassessed check is described
as current behaviour rather than as a guarantee, since changing it is on the
follow-up list.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`6f9a8e9`](https://github.com/oscal-compass/compliance-to-policy/commit/6f9a8e952bfe804ef4228e64196446a1e5c670b3))

* docs: a link to the demo, under the steps it walks

The thumbnail is out. It made the first image in the component&#39;s README a
video still, and the reason for putting it there -- that the conference is
on -- expires with the conference. The recording is also not C2P&#39;s: it is an
OSCAL Compass end-to-end demo in which C2P is one component, so a thumbnail
for it belongs with the project&#39;s own outreach rather than here.

What is left is a link, placed under the numbered steps. Those describe the
chain in the abstract; the walkthrough is that same chain run against a
cluster, with the output of every command, so it reads better there than in
the list of plugin pages, where it was the one entry that is not a plugin.
`docs/end-to-end/README.md` gets the link back too, as a line rather than an
image, since that page is the demo this records.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`3840583`](https://github.com/oscal-compass/compliance-to-policy/commit/384058392b690f5a237a673c3b280d63c41608c3))

* docs: the posture commands and the document they produce

`c2p-posture/v1` was named in the README and nowhere specified, so a reader
could not write anything that consumes it. `docs/public/posture.md` gives it
a page: both commands with every flag, and the document&#39;s shape field by
field, with the values `state`, `reason` and `result` are allowed to take.

Two things worth saying once rather than in each caller. The expected set of
checks comes from the component definition rather than from the evidence,
which is what makes an unassessed control a row rather than an omission. And
an `error` from an engine, or a `skip` from a PolicyException, is the absence
of a verdict: neither is counted as a failure, both reduce `assessedChecks`.

`state` is a field and not something a template computes, because `md` and
`html` are views over this document rather than the thing itself.

Linked from the README&#39;s posture step, from the plugin list, and from the
Kubernetes page where the coverage gap between engines first comes up.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`54b87dd`](https://github.com/oscal-compass/compliance-to-policy/commit/54b87dd6e195eb090d3877ca59808bc1221dad8a))

* docs: two links that have never resolved

`CONTRIBUTING.md` pointed at `/CODE_OF_CONDUCT.md`, which this repository
does not have and never has. The code of conduct it refers to is the
organisation&#39;s, so the link now goes to `oscal-compass/community`.

The README&#39;s screenshot of the Kyverno assessment results, at
`/docs/public/images/assessment-results-md.kyverno.jpg`, was never committed
either -- `images/` holds the `ocm` and `auditree` ones only. Replaced with a
link to `docs/public/kyverno.result.md`, which is the same output as a
committed artefact rather than a picture of it, and is how
`docs/public/kyverno.md` already shows it.

Both predate this branch; they are here because this is the branch that
touches the documentation.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`fba44d8`](https://github.com/oscal-compass/compliance-to-policy/commit/fba44d8d5a661d8a12dd3931e632a91946f20350))

* docs: the demo video on the front page, without an event&#39;s section

The video belongs where someone arriving at the repository will see it.
#59 moved it to the walkthrough on the grounds that the root README should
not carry a section for one event, which is right about the section: there
is no `## KubeCon NA 2026 demo` heading here, no event in the prose, and
nothing that goes stale when the conference does. What is left is two lines
and a thumbnail, above the architecture diagram, saying what the chain does
and pointing at the walkthrough for reproducing it.

Removed from `docs/end-to-end/README.md` rather than left in both places.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`6ddde80`](https://github.com/oscal-compass/compliance-to-policy/commit/6ddde801cfde67fa3966bdb7f463ebfc9c018305))

* docs: the new engines, the posture commands, and what changed for parameters

Nothing this PR adds was in the README or under docs/public: the three
Kubernetes plugins, the two posture commands, aggregating several PVPs into
one Assessment Results, the assessment-check-id prop, the c2p-posture/v1
format, and a change to how rule parameters resolve.

docs/public/kubernetes.md covers the three engines together rather than one
page each. They share one component definition and one sample script, so
three pages would repeat the same walkthrough three times and bury the thing
worth saying -- that the same control reaches three engines, and that where
an engine cannot express a check the right answer is to declare none and let
the posture report the gap.

The posture commands are in the Quick demo beside `tools viewer`, at the same
weight: a command and what it is for. The difference from the viewer is worth
one sentence -- json is the c2p-posture/v1 model and md and html are views
over it, so a posture can be diffed and served, where the viewer renders and
forgets.

Parameter resolution is called out as a behaviour change. Earlier releases
read `Parameter_Value_Alternatives` and nothing else; `set-parameters` now
wins, `Parameter_Value_Default` comes next, and numbered columns are read.
An existing component definition carrying either of the first two will
generate different policy than it did, which a reader upgrading needs to be
told rather than left to discover.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`80f4c12`](https://github.com/oscal-compass/compliance-to-policy/commit/80f4c126cb994c63eafba4b64653b460c0963c4a))

* docs: one line in the root README, not an event&#39;s section

The KubeCon section came back in this branch, embedded video and all, after
#58 took it out. The root README is the component&#39;s entrance and should not
carry a section for one event, so the walkthrough is a line in &#34;Usage of C2P
Plugins&#34; beside the other engines&#39; pages.

The video moves to the top of `docs/end-to-end/README.md`, which is where
someone who followed that line has already decided they want the demo, with
the sentence that says it does not match the walkthrough command for command.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`c020d8e`](https://github.com/oscal-compass/compliance-to-policy/commit/c020d8eb067d28e8ce1f3ff96c762e461699d217))

* docs: the walkthrough&#39;s step references point at steps that exist

Three mentions of &#34;step 8d&#34;, which there is no such thing as. Step 8 has
three parts: 8a generates and is read, 8b applies, 8c collects. All three
meant 8c, and one of them said so in the same sentence -- &#34;run step 8d
again, because 8c applies nothing&#34; -- which is the collecting step twice
under two names.

A fourth was wrong the other way: &#34;there is no capture for step 8b, because
reading something produces no output&#34;. 8b is the applying, and it has one,
`B2-apply`. What produces no output is the reading in 8a, which is where
capture/README already puts the gap.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`8b28ce7`](https://github.com/oscal-compass/compliance-to-policy/commit/8b28ce7ad245883b5297c8a75c7bc74f50d343d5))

* docs: point at the walkthrough from the README and the Kubernetes sample

The references the core change removed, added back now that what they point
at is in the tree. The README gets its KubeCon section: the demo, and the
walkthrough as the way to reproduce it. The Kubernetes sample&#39;s README is end
to end again -- the chain diagram names where applying and collecting happen,
the commands that deploy and collect are back, and the paragraph on what the
VAP probe does to a cluster points at the reference implementation that lists
every operation.

Documentation only. `samples_public/kubernetes/generate.py` and
`plugins_public/plugins/vap.py` also carried references to the walkthrough
before the split, and they are deliberately left alone here: this branch adds
the walkthrough and does not touch library or sample code. Their docstrings
describe the applying half as a program C2P does not ship, which is true
whether or not an example of one is in the tree. If a pointer from the code
is wanted, it belongs in a change to that code.

Restored by hand rather than by reverting, because the sample README moved on
after those references came out -- it stopped describing `expected-policies`
as another implementation&#39;s output, which is kept.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`00f470f`](https://github.com/oscal-compass/compliance-to-policy/commit/00f470f280d11a7f5b99a3c031af8c8882a0590d))

* docs: re-render the artefacts that c2p-core&#39;s fixes changed

The walkthrough&#39;s committed artefacts were produced before the renderer and
the posture roll-up were fixed, so re-running the chain against them would
not reproduce what is committed here.

Six generated policies change shape, not meaning. The templates now put a
value from the component definition through a filter and never wrap it in
quotes of their own, which moves a CEL list literal from a single-quoted
YAML scalar into a double-quoted one and quotes label keys. Every regenerated
document parses to the same structure it did.

The postures change numbers, and the new ones are right. si-7.15 read
`20 of 34 subject(s) failing` and now reads `13 of 34 subject(s) failing, 7
subject(s) could not be evaluated`: the seven were a registry that could not
return a manifest during the recording, and the roll-up was reporting them as
violations of the control. No control changes state -- each of the two had
real failures beside the errors -- which is why this was only visible in the
remarks.

The posture tables in the B3 and B4 captures are `result2posture` over the
committed Assessment Results, so they are re-rendered too, and B1&#39;s last line
drops the sentence generate.py no longer prints. capture/README says which
files that applies to. Everything that came from the cluster is untouched.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`df0fb0d`](https://github.com/oscal-compass/compliance-to-policy/commit/df0fb0d26b9dd1e8d5b49727f72cde5ebb57f36f))

* docs: the end-to-end walkthrough, and a reference implementation for the cluster half

One NIST SP 800-53 control, authored once, becomes policy for three
Kubernetes engines, runs against a live cluster, and comes back as OSCAL
assessment results and a compliance posture. `docs/end-to-end/` walks that
in ten steps with the real inputs, the captured output of every command and
the artefacts each step produced.

The half of the round trip that talks to a cluster is here rather than in
the library, under `docs/end-to-end/reference/`, with its tests beside it.
It holds a kubeconfig, it writes to a cluster, and it copies engine messages
into a document an auditor may read. Those are decisions about one
particular cluster, so they are shown as a worked example rather than
shipped as a feature. C2P&#39;s job stops at generated policy and resumes at
collected results.

Between those two is a person. Generating and applying are separate
commands, and the step between them is reading what was generated -- the
templates are written by an agent, and a check that asserts nothing applies
cleanly, enforces nothing and reports every workload compliant.

Collecting from ValidatingAdmissionPolicy is the awkward part: it has no
report API and evaluates nothing retroactively, so a workload already
running has to be put back through admission. The probe adds a reporting
binding per check, scoped to a scratch namespace it owns, and submits a copy
of the workload there; it never removes, modifies or weakens a binding it
did not create, and it deletes only what it labelled. Measured on Kubernetes
v1.34.0 with four policies and one object that breaks all four: one
submission recovers four verdicts of four, against one of four for the
enforcing bindings.

A reported pass is the absence of a warning, which a misconfigured probe
also produces, so every check is first shown a canary written to fail it. A
check that does not report its own canary is recorded as `error` rather than
`pass`.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`9e3ea6f`](https://github.com/oscal-compass/compliance-to-policy/commit/9e3ea6f13657eb1004209a131567cabadf18be57))

* docs: do not refer to the end-to-end walkthrough from here

The walkthrough and its reference implementation are a separate change,
under docs/end-to-end/. Referring to them from here leaves eleven dangling
references if this merges first: two broken Markdown links, six commands in
the Kubernetes sample that name programs this branch does not contain, two
docstrings, and a line generate.py prints at the end of every run.

So the Kubernetes sample is a generation sample and now says so. Applying
the policy and collecting what the engines decide is named as the other half
of the chain and as a program C2P does not ship -- which is true of the
library with or without an example of one -- rather than by pointing at a
path. The README&#39;s KubeCon section goes with the walkthrough, since the
&#34;run it yourself&#34; link it ends with is the walkthrough.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`adf5c83`](https://github.com/oscal-compass/compliance-to-policy/commit/adf5c83a5c4b755405e76d081311abdfcb6281e7))

* docs: define C2P++ in the README, and redraw the architecture

C2P++ is C2P with agentic policy generation: the step where an agent writes
the policy templates the plugins render. The README now says that above the
diagram, and the diagram says the same. Nothing else is under the name --
the engines, the aggregation and the posture are ordinary C2P.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`da11629`](https://github.com/oscal-compass/compliance-to-policy/commit/da11629e8a254e42803c638aa2e7cb8432bc142a))

* docs: add initial security insights file

Signed-off-by: Jennifer Power &lt;barnabei.jennifer@gmail.com&gt; ([`3a97126`](https://github.com/oscal-compass/compliance-to-policy/commit/3a971268dd8c5509ff98fe9097deedc46af49f95))

### Feature

* feat(skills): a skill that writes the policy templates c2p renders

c2p turns a component definition into policy by filling one template per
check. Writing those templates is the step the component definition cannot
decide: it needs knowledge of the engine, not of c2p -- which apiVersion a
cluster will accept, that the CEL available to ValidatingAdmissionPolicy has
no flatten(), that a Kyverno verifyImages rule is a mutating webhook.

The skill does that, and states the rules that keep it honest:

- a check declared with no template is an error, not a silent omission;
- if an engine cannot express a check, do not write the check, so the
  posture reports a gap rather than a policy that always passes;
- one authored parameter must render identically across engines, because a
  cluster compares a CEL expression as a string;
- verify by applying to a cluster, not by reading -- plain
  &#39;kubectl apply --dry-run=server&#39; will lie to you.

template-contract.md is the interface between whatever writes a template and
the renderer: the render context is engine, rule, check and params, and
nothing else. It is worth reading even by someone who writes templates by
hand.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`aad551e`](https://github.com/oscal-compass/compliance-to-policy/commit/aad551e02d26932046874d59ce999ec9f42eb89f))

* feat(plugins): Kyverno DSL, Kyverno CEL and ValidatingAdmissionPolicy

Three engines for the same control, and a shared renderer for the two that
need one.

`kyverno_dsl` is the existing `kyverno.io/v1 ClusterPolicy`, deprecated in
Kyverno 1.19 and removed in 1.20. `kyverno_cel` is its replacement,
`policies.kyverno.io/v1`. Generating both from one component definition is
the argument for authoring the control rather than the policy: the control
outlives the engine.

`vap` is upstream ValidatingAdmissionPolicy, which needs nothing installed
and has no report API. Its generate half is ordinary; its result half
converts what a probe observed, which is why the probing itself is not in
this package and not in the library -- it needs a cluster, and holding
cluster credentials is a decision for whoever runs one.

`policy_template` renders one document per check from a template named after
the check, with a render context of engine, rule, check and params and
nothing else. Two properties it enforces, both because the alternative fails
silently:

- a check the component definition declares with no template is an error.
  Generating a policy set that is quietly missing a control is worse than
  not generating one.
- rule and check ids are validated as Kubernetes names before they become a
  path segment. They come from a component definition, which in the
  authoring flow these plugins exist for is written by an agent, and
  rendering writes unconditionally -- an id of `../../x` would otherwise
  land outside the deliverable directory, over whatever is there. Every id
  in the shipped sample already conforms, so no generated byte changes.

`samples_public/kubernetes` carries the component definition, the sixteen
templates, and `generate.py`, which turns the first into deployable policy
for all three engines without contacting anything.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`208106f`](https://github.com/oscal-compass/compliance-to-policy/commit/208106f1103145289e7977d5529de1ec1c5bbda4))

* feat(tools): a compliance posture, and a comparison between two of them

Assessment Results answer what each check said. The question an estate asks
is what the state of a control is, and how much of it anything actually
looked at. `c2p tools result2posture` answers the second from the first plus
the component definition, as a c2p-posture/v1 document rendered to JSON,
HTML or markdown.

Coverage is the part worth having. The component definition says which
checks a control expects; the assessment results say which reported. A
control nothing was able to assess comes back `not-assessed` with the count
that says so, rather than being quietly absent -- an estate that cannot see
what its tooling could not look at will conclude it is fine.

`c2p tools posture-compare` diffs two of these documents and names the
controls whose state changed, including the ones that stopped being reported
at all. That is how the cost of dropping an engine becomes a number rather
than an impression.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`47f31c6`](https://github.com/oscal-compass/compliance-to-policy/commit/47f31c627164914592cde4587aca433f44097697))

* feat(framework): aggregate several PVPs into one Assessment Results

A control satisfied by more than one engine was previously assessed by one
of them: c2p held a single PVP result, so running Kyverno and
ValidatingAdmissionPolicy over the same component definition produced two
documents and no way to say what the system&#39;s posture was. c2p now takes
several and emits one Assessment Results, because a single document is a
statement about the system whereas one file per engine is a record of how
the assessment happened to be organised.

Three fixes that running it that way found:

- observations now carry an &#39;assessment-check-id&#39; prop. Without it, an
  observation keyed only by rule is unattributable once two engines
  implement the same rule, and anything counting checks double-counts.
- an empty statement-ids list is omitted rather than emitted empty.
- parameter values resolve as set-parameters, then Parameter_Value_Default,
  then Parameter_Value_Alternatives. The previous order reached for the
  alternatives first, so a component definition carrying both got the
  alternative where it had authored a default. Release note worthy: a
  component definition with both will generate different policy.

jinja2 moves into the runtime dependencies: a plugin that renders templates
needs it, and it was previously only present transitively.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`70e369e`](https://github.com/oscal-compass/compliance-to-policy/commit/70e369e978dd18ebc4aa890d0d87b3eacc06dafa))

### Fix

* fix(reference): the printed deletion removes what this run applied

The command `deploy.py` printed selected on `-l compliance.oscal/check-id`,
which is a label-key existence selector. The template contract puts that
label on every policy C2P generates, so the command removed every C2P policy
on the cluster -- another component definition&#39;s, another team&#39;s, whatever
GitOps installed -- not the sixteen this run applied. On a cluster without
Kyverno&#39;s CEL CRDs the kind list made it fail outright as well.

The code comment, `reference/README.md` and the walkthrough all said it
matched the set that was applied. It did not. A printed command gets pasted,
so narrowing it is the fix rather than rewording the claim: `deletion_command`
now names the same enumeration `deploy` applied, one `-f` each, which makes
removing symmetrical with applying.

`test_the_printed_deletion_names_the_same_files_and_nothing_else` pins it,
and fails against the selector.

`capture/README.md` claimed less editing than had happened -- the `context:`
lines in B2, B3 and B4 and B2&#39;s deletion lines were adjusted to match what
the programs print, not re-run. It now lists every line that was touched and
says which of them came from running the renderer rather than from an editor.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`526be16`](https://github.com/oscal-compass/compliance-to-policy/commit/526be162899937271bd1edc9be0dde27317ac492))

* fix(reference): name the cluster, and apply only what was read

Two things the review asked for before this merges.

**The context is required, and said out loud.** `--context` defaulted to
`None`, so both programs fell through to whatever `kubectl config
current-context` happened to be -- while `platform_k8s/client.py` claimed
&#34;the kubeconfig and context are always passed explicitly&#34;. These apply Deny
policy and create namespaces and bindings; inheriting the ambient context is
how a run lands on a cluster nobody meant to change. Both now require it and
print it before writing anything. Every command in the walkthrough and in
`reference/README.md` already passed it, so nothing documented changes.

**Apply names the documents.** `deploy.py` enumerated `**/*.yaml` and then
handed `kubectl apply -R -f` the directory, which applies `.json` and `.yml`
too -- and after a collection run that directory also holds
`assessment-results.json` and `posture.json`. The set that was read and the
set that reached the cluster were two different sets. `reference/README.md`
already described the correct behaviour (&#34;the objects you reviewed, and only
those&#34;); now the code does it, applying the enumerated files in one request
so a policy and its binding still land together.

The `kubectl delete` that `deploy.py` prints had the same shape, and is now
scoped by the `compliance.oscal/check-id` label every generated document
carries -- exactly the set it applied, and nothing a person put there.

`tests/test_deploy.py` pins both: that the directory is never handed to
kubectl, and that results written beside the policy are not applied. Two of
the four fail against the previous code.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`e9101bb`](https://github.com/oscal-compass/compliance-to-policy/commit/e9101bbf156bcfba8ce28c1ff2fc2040b788720e))

* fix(tests): the reference is a snapshot of this generator, not of another one

`expected-policies` was described as the output of an independent
implementation, kept so that two implementations could be compared. That
implementation is not in this repository, so nothing here could check the
claim, and the quoting changes in the preceding commit regenerated the files
from this generator -- which settles it. What the files are is a committed
snapshot, so a change to a template or to the renderer is read as a diff.
Described as that now, in the test and in the sample README.

The parity test beside it was asserting nothing. It picked the expression to
compare by matching a raw line containing `[&#34;`, and once the expression was
quoted differently the first such line was `apiGroups: [&#34;apps&#34;]` -- identical
in both engines, as it is in every policy, so the test passed while saying
nothing about the authored parameter. It now reads the `required` variable
out of the parsed document. Verified by making one engine render the list
with a space in it: the test fails.

`SAMPLE` was a path relative to the working directory, so the test only ran
from the repository root. Resolved from `__file__`, as the test beside it
already does.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`5467dbe`](https://github.com/oscal-compass/compliance-to-policy/commit/5467dbe1cc52b10437973569ee6f99154524ec7d))

* fix(tools): an engine that could not evaluate has not found a violation

`_roll_up` counted `error` subjects alongside `fail`, so a control with two
engines passing and one unable to report came out `not-satisfied` with
reason `fail`. In the shipped sample that is what a VAP probe that cannot
see a check does to ac-6 and sc-39: a broken probe is published as a
compliance violation, and an estate cannot tell which of the two it has.
`vap.py` already says an outcome it could not stand behind arrives as
`error, never as pass`; the posture then read it as worse than a fail.

A verdict is now fail, warning or pass. `error` is the absence of one, and
is reported as coverage -- the check stays in `expectedChecks` and drops
out of `assessedChecks`, which is the fact that is true. A control whose
remaining verdicts all pass is `not-assessed` rather than `satisfied`: two
engines passing and one unable to say is not a clean control either. A real
failure beside an error is still a failure.

`skip` maps to Error and that is now coherent: Kyverno defines it as
evaluation bypassed by a matching PolicyException, which is no verdict
rather than a violation. Before this, a PolicyException made the control
non-compliant.

Three comments were wrong about `warn` and are corrected. Kyverno&#39;s report
result logic defines it as a non-compliant result reported as a warning,
the documented example being `policies.kyverno.io/scored: &#34;false&#34;`. It is
not an object that was admitted while the policy objected, and the check
remarks no longer described a warned subject as &#34;passing&#34;.

The behaviour had no tests, which is why it survived three reviews. Four of
the five added here fail against the previous code.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`e0c8791`](https://github.com/oscal-compass/compliance-to-policy/commit/e0c8791c62bed3ce9d66666ec6ddf5f5809d7c13))

* fix(skills): say which cluster, and ask before writing to it

The workflow told the agent to dry-run a generated document and then apply
it and submit something that should be rejected. It named no cluster, so
the agent applies to whatever `kubectl config current-context` happens to
be, and it asked no one -- a Deny policy on the wrong cluster rejects real
workloads. It also contradicted the principle the walkthrough is built on:
what is generated is read by a person before anything touches a cluster.

Dry-run is now as far as the workflow goes unaided. Everything past it --
apply, the scratch namespace a probe needs, the objects it submits -- is a
write, and wants an explicitly named throwaway cluster and the operator&#39;s
agreement. An unfamiliar context or a namespace with workloads in it is a
reason to stop rather than to guess.

The workflow also told the agent to read the component definition for the
mechanism it names, without saying that a component definition is data. It
is the same input that reaches the renderer, and text in it that reads as
an instruction is the component definition trying to write the template.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`71c22ca`](https://github.com/oscal-compass/compliance-to-policy/commit/71c22cacfa3da5dd61cba5b2ef0cf34b3e0b9cad))

* fix(plugins): a template, and what is substituted into it, are untrusted

In C2P++ the templates are written by an agent, from a component definition
and whatever documents the agent was pointed at. Rendering treated both as
if the project had written them.

**The template is executable.** Jinja2&#39;s default Environment hands a
template the Python object graph: `cycler.__init__.__globals__.os.popen(...)`
reaches os and runs. Rendering happens on a developer&#39;s machine or in CI,
where there is a kubeconfig and there are secrets. Reading the generated
policy is no defence -- the reading happens after the render, and the skill
says to read the output rather than the template. Now rendered in
SandboxedEnvironment, which refuses the attribute walk, with StrictUndefined
so a mistyped `params.foo` is an error instead of an empty string that makes
a check assert nothing. jinja2 is pinned to &gt;=3.1.6: the sandbox escapes
CVE-2024-56326 and CVE-2025-27516 are fixed in 3.1.5 and 3.1.6.

**The values are not quoted.** Descriptions and parameter values went into
YAML and CEL as written. OSCAL validation does not help: trestle refuses LF
in a prop value but accepts CR, U+2028, U+0085 and both quote characters
(measured, trestle 3.12.4). A CR in a check description ended the header
comment and added a cluster-admin ClusterRoleBinding to the generated file;
a quote in a parameter either broke the YAML or rewrote the CEL expression
the cluster compiles, with no error either way.

So the contract now says how a value is substituted, and the shipped
templates follow it: a filter&#39;s output is the whole token and is never
wrapped in quotes of your own, a value inside a larger expression is
concatenated in as `| tojson` rather than interpolated into a literal, and
a value in a comment goes through the new `oneline` filter. The generated
CEL is unchanged in meaning -- every regenerated document parses to the same
structure, and the four that differ textually differ only in YAML quoting
style.

And `render` now parses what it rendered before writing it: every document
must be a mapping whose kind is one the plugin declares it generates and
whose name is the check id. That catches a value that broke out into a new
document. It cannot catch one that rewrote an expression in place, which is
why the quoting rules come first and this is the backstop.

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`74006be`](https://github.com/oscal-compass/compliance-to-policy/commit/74006be6671f5fd627248371c66984f393fda89a))

* fix: add disclaimer (#52)

* fix: add disclaimer

Signed-off-by: degenaro &lt;lou.degenaro@gmail.com&gt;

* fix: use common banner

Signed-off-by: degenaro &lt;lou.degenaro@gmail.com&gt;

* fix: mdformat

Signed-off-by: degenaro &lt;lou.degenaro@gmail.com&gt;

---------

Signed-off-by: degenaro &lt;lou.degenaro@gmail.com&gt; ([`6ec821d`](https://github.com/oscal-compass/compliance-to-policy/commit/6ec821d4c253baf85e5b4d171ee9f9fb7affc1e0))

* fix: prevent execution of untrusted PR code with elevated privileges (#47)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`bbbcf8b`](https://github.com/oscal-compass/compliance-to-policy/commit/bbbcf8b3c952a55f3678ee882664438272074a32))

### Test

* test: ignore version/timestamp fields in expected output comparison

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`1d1b2e0`](https://github.com/oscal-compass/compliance-to-policy/commit/1d1b2e02e87aff31a44268da1221b8e7327991ee))

### Unknown

* Merge pull request #61 from yana1205/release/1.0.0

feat!: major version release 1.0.0 ([`58f6523`](https://github.com/oscal-compass/compliance-to-policy/commit/58f6523308f68383018044ada22c12719b0739d0))

* Merge pull request #60 from yana1205/docs/update

docs: pages for the Kubernetes engines and the compliance posture ([`cd6311a`](https://github.com/oscal-compass/compliance-to-policy/commit/cd6311a208fe86830714dd8736d47a45be2d9e20))

* Merge pull request #59 from yana1205/feat/end-to-end-walkthrough

docs: an end-to-end walkthrough, with a reference implementation for the cluster half ([`035d7dd`](https://github.com/oscal-compass/compliance-to-policy/commit/035d7dd18217107b8f2e9eb71b252dc5772f87b1))

* Merge pull request #58 from yana1205/feat/c2p-core

feat: C2P++ agentic policy generation, and multi-engine assessment ([`cf437d9`](https://github.com/oscal-compass/compliance-to-policy/commit/cf437d90a6071d44139b71d2d99a73a76535ba63))

* Merge pull request #42 from jpower432/docs/add-si

docs: add initial security insights file ([`846a644`](https://github.com/oscal-compass/compliance-to-policy/commit/846a644e8099e3b4630d05602020663705f00d80))

* Merge pull request #1 from yana1205/docs/add-si

test: ignore version/timestamp fields in expected output comparison ([`388c5d5`](https://github.com/oscal-compass/compliance-to-policy/commit/388c5d562e945d335db426f65c923b08e5777f96))

## v0.5.0 (2024-11-13)

### Feature

* feat: add local-definitions and findings (#34)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`248e1ef`](https://github.com/oscal-compass/compliance-to-policy/commit/248e1effb48322e3aad0e9ca2479f0c6adebb5c1))

### Fix

* fix: trestle version in unit test

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`0d86b0c`](https://github.com/oscal-compass/compliance-to-policy/commit/0d86b0c60c9768c9ddc64799367a4402d23d3daa))

### Unknown

* 0.5.0

Automatically generated by python-semantic-release ([`dd45b04`](https://github.com/oscal-compass/compliance-to-policy/commit/dd45b047c65700abcabf3308d55e966f9814f68d))

* Merge pull request #35 from yana1205/feat34

feat: add local-definitions and findings (#34) ([`389dc8f`](https://github.com/oscal-compass/compliance-to-policy/commit/389dc8fc7a79e2a0e076e975a0556cb0503aba8b))

* Chore: update go modules (#31)

* chore: update Go modules

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* chore: upgrade go-git to v5

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

---------

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`2b0b410`](https://github.com/oscal-compass/compliance-to-policy/commit/2b0b410a1542f32433ebe447b4ffe468ac25b514))

* LF footer website guidelines (#30)

* LF website

https://github.com/cncf/foundation/blob/main/website-guidelines.md

Signed-off-by: Lou DeGenaro &lt;lou.degenaro@gmail.com&gt;

* Fix content in readme

Signed-off-by: Yuji Watanabe &lt;muew@jp.ibm.com&gt;

* maintainers &amp; cncf logo

Signed-off-by: Lou DeGenaro &lt;lou.degenaro@gmail.com&gt;

* license &amp; authors

Signed-off-by: Lou DeGenaro &lt;lou.degenaro@gmail.com&gt;

* fix contributors link

Signed-off-by: Lou DeGenaro &lt;lou.degenaro@gmail.com&gt;

---------

Signed-off-by: Lou DeGenaro &lt;lou.degenaro@gmail.com&gt;
Signed-off-by: Yuji Watanabe &lt;muew@jp.ibm.com&gt;
Co-authored-by: Yuji Watanabe &lt;muew@jp.ibm.com&gt; ([`3c9e94e`](https://github.com/oscal-compass/compliance-to-policy/commit/3c9e94e4f38d0125b14ccb3f5d8f19e96b917179))

## v0.4.0 (2024-08-29)

### Documentation

* docs: removes CODE_OF_CONDUCT.md (#24)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`ea1915f`](https://github.com/oscal-compass/compliance-to-policy/commit/ea1915f1aa93cf34973245b71136803bb212ebb9))

### Feature

* feat: enhance release management (enable GitHub release and publish to PyPI) (#29)

* fix: search key for rule should be Rule_Id (#28)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* feat: add release automation (#28)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* fix: manually install detect-secret to manage direct dependency packages out side Pypi (#28)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* fix: Logo needs to be referenced by full url (#28)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* fix: makefile for UT and trestle verion in UT (#28)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* fix: replace namespace from ibm.github.com to oscal-compass.github.io (#28)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

---------

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`1ebcfd8`](https://github.com/oscal-compass/compliance-to-policy/commit/1ebcfd8f5d7cc1e76497029280d269b1a5c84c92))

### Unknown

* 0.4.0

Automatically generated by python-semantic-release ([`cfb955e`](https://github.com/oscal-compass/compliance-to-policy/commit/cfb955e385d85f8d6f8aa8266b2b219ac1bee39f))

* Merge pull request #26 from rahearn/update-trestle

Update to compliance-trestle ~= 3.3 ([`b028bb1`](https://github.com/oscal-compass/compliance-to-policy/commit/b028bb16a251c7f0f65e8bdb740151577cc9f13e))

* Merge pull request #1 from yana1205/yana1205/update-trestle

Update sample code to align compliance-trestle ~= 3.3 ([`a775146`](https://github.com/oscal-compass/compliance-to-policy/commit/a775146ef39435ed3d3c843d2646cd455a887246))

* Update sample code to align compliance-trestle ~= 3.3

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`900b0ef`](https://github.com/oscal-compass/compliance-to-policy/commit/900b0ef8e06e7bb361967fa7868e71868eefb7ce))

* Update to compliance-trestle ~= 3.3 ([`2951f07`](https://github.com/oscal-compass/compliance-to-policy/commit/2951f074299ccbb0c769092d08ac6417505cb876))

* add a brief description, design, and goals of Compliance-to-Policy project (#19) (#21)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`53edf65`](https://github.com/oscal-compass/compliance-to-policy/commit/53edf65adcc9c90524036aa9881ba8134e17217d))

* Update README.md (#20)

* fix logo size (#19)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* correct the description for Go version (#19)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

---------

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`604b5a6`](https://github.com/oscal-compass/compliance-to-policy/commit/604b5a62d2e6ee953e1ebdb2954d1bf4ff1a08e2))

* Merge pull request #18 from yana1205/yana/log

use non-transparent background logo ([`bcd764f`](https://github.com/oscal-compass/compliance-to-policy/commit/bcd764fcc61a2b66313b6971ba5ca25176ed2e27))

* use non-transparent background logo

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`03c22b2`](https://github.com/oscal-compass/compliance-to-policy/commit/03c22b2fe9f16cf93b462fd693d7d663b6aba4b5))

* update README to add compliance-to-policy logo (#17)

* update README to add compliance-to-policy logo

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* update version

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

---------

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`eee9cd9`](https://github.com/oscal-compass/compliance-to-policy/commit/eee9cd9c3899abad42b281d853015429eec47027))

## v0.3.2 (2024-06-05)

### Unknown

* Merge pull request #16 from yana1205/yana/dev

update README.md and version ([`194e8ae`](https://github.com/oscal-compass/compliance-to-policy/commit/194e8aecdc853860623e31214e886e2f4a569d41))

* update README.md and version

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`c453cff`](https://github.com/oscal-compass/compliance-to-policy/commit/c453cff5fc599cbd6d7d7b0f9550333f7835c0aa))

* Merge pull request #15 from yana1205/yana/heterogeneous

add example usage of heterogeneous PVPs ([`7bdc7ac`](https://github.com/oscal-compass/compliance-to-policy/commit/7bdc7ac4357b69186a334f70b1b9e7e72dc53249))

* add example usage of heterogeneous PVPs

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`9721932`](https://github.com/oscal-compass/compliance-to-policy/commit/9721932d4bf5c264fda307671566d72bd5b0347b))

* Implement Auditree plugin (#14)

* implement auditree plugin

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* format

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* update documents, including auditree usage

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

---------

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`2b0081c`](https://github.com/oscal-compass/compliance-to-policy/commit/2b0081cdd7e5e0850043ec71a8ed5392d7b9dc0e))

* Merge pull request #13 from oscal-compass/release

update version to v0.3.1 ([`2744462`](https://github.com/oscal-compass/compliance-to-policy/commit/27444629cd479db4bebc6a17e6f9a76149d3f2ce))

## v0.3.1 (2024-05-02)

### Unknown

* update version to v0.3.1

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`e116e93`](https://github.com/oscal-compass/compliance-to-policy/commit/e116e93ac9eefd08960b2f60f49b38b08be08a43))

* fix to add signature

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`ed206ac`](https://github.com/oscal-compass/compliance-to-policy/commit/ed206ac11ba04cb398e0883eac7f5b5cd611fc36))

* Merge pull request #11 from oscal-compass/dev

switch to use static versioning ([`f7f19c2`](https://github.com/oscal-compass/compliance-to-policy/commit/f7f19c2ef548f17ab342381acfe2e2c623af3dc9))

* switch to use static versioning

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`7232edd`](https://github.com/oscal-compass/compliance-to-policy/commit/7232edd92bdf29e4d176ca43a9673010293a722e))

## v0.3.0 (2024-05-02)

### Unknown

* Move to Python project for easily to leverage Trestle (#10)

* move go project to go/

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* update makefile to build cp2cli as submodule

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* update docs for url changes

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* import c2p python project

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* update README.md for message of deprecating C2P in Go

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* add Code of Conduct, Contributor guide, and maintainers list

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* update go/README.md release guide

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

---------

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`d8eb84e`](https://github.com/oscal-compass/compliance-to-policy/commit/d8eb84e68c5760fca81ef9acece0d81ad76b9f0e))

## v0.2.0 (2023-12-07)

### Unknown

* Merge pull request #9 from yana1205/up-to-date-ocm-plugin

Up to date ocm plugin ([`36a42c8`](https://github.com/oscal-compass/compliance-to-policy/commit/36a42c86d1e245f4a0f5071e10f61fb94996a1f2))

* Add images for example use case

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`c3c2bb6`](https://github.com/oscal-compass/compliance-to-policy/commit/c3c2bb6f462a055438a9242dfb33364b8320aacd))

* Separate docs from C2P for OCM and C2P for Kyverno

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`84d3e4b`](https://github.com/oscal-compass/compliance-to-policy/commit/84d3e4bbb7c44fb158a00dd0b48708cf9cb6dcec))

* Move oscal2posture to subcommand &#39;tools&#39;

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`fa24c87`](https://github.com/oscal-compass/compliance-to-policy/commit/fa24c874b58fce4b2f4dab4568bbb83998501ffd))

* Add end-to-end use case in .md

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`01d215e`](https://github.com/oscal-compass/compliance-to-policy/commit/01d215e34a223fd833db59f5bd9ae6ae0d45f675))

* Add version command

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`34d1ba2`](https://github.com/oscal-compass/compliance-to-policy/commit/34d1ba23d8cb14a082f42a86266451f46c5ee2be))

* add --results option to specify path to the directory of Policy Results

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`4352401`](https://github.com/oscal-compass/compliance-to-policy/commit/4352401f0ed8edbff4757e63e89896d8b5c24ae1))

* consolidate c2p configfile parser in each pvp

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`87450ea`](https://github.com/oscal-compass/compliance-to-policy/commit/87450ea2c45134ee753996dca04987b296cc2db5))

* -o option in result2oscal points output file instead of directory

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`882baa5`](https://github.com/oscal-compass/compliance-to-policy/commit/882baa53abaa46b4fbd3504a63963eba1874f4e4))

* Use List typed object

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`f990465`](https://github.com/oscal-compass/compliance-to-policy/commit/f9904654ee4b487d24f960ca717e6df19e7a59c4))

* Rename reporter to result2oscal

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`4ca1c16`](https://github.com/oscal-compass/compliance-to-policy/commit/4ca1c164c15010c3a8358594d807ef7e3039659f))

* Conversion to PolicyReport is no longer required

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`6ca8ef3`](https://github.com/oscal-compass/compliance-to-policy/commit/6ca8ef30ef9f85c7710cd37e98369f4e18cd67f5))

* Reuse oscal2posture implemented for C2P for Kyverno

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`cdb6913`](https://github.com/oscal-compass/compliance-to-policy/commit/cdb6913e748ccce011cd2d6124eacae04b1477f6))

* Align the latest OSCAL Assessment Results and consolidate ocm test data in ocm dedicated directory

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`fa87f02`](https://github.com/oscal-compass/compliance-to-policy/commit/fa87f02f6ae9ea0562796b5c4a43d80ba65db7eb))

* Merge pull request #8 from yana1205/main

Update demo movie ([`67dd6b9`](https://github.com/oscal-compass/compliance-to-policy/commit/67dd6b9e155dc83459a56c024b41e94252b61433))

* Update demo movie

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`ebce623`](https://github.com/oscal-compass/compliance-to-policy/commit/ebce6233bde00912d011bf72609af024235fb6d4))

## v0.1.0 (2023-10-31)

### Unknown

* Fix goreleaser

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`a7103c2`](https://github.com/oscal-compass/compliance-to-policy/commit/a7103c2513bef713c99eeddb24e2e829902ece9c))

* Extend c2p for Kyverno as PVP/PEP (#7)

* Add cmd to create policy resources from kyverno policy collection

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* Add oscal2policy cmd for kyverno to c2pcli

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* Implement results2oscal for kyverno

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* Implement sample oscal2posture

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* Consolidate commands relating with kyverno plugin

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* Rename composer to oscal2policy and fix name and description of CLI args

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* Consolidate command for tools for kyverno into kyverno plugin

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* Rename former c2p ocm plugins

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* Add sample cronjob for upsync

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* Move assessment-results from c2p-config to command args

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* Add coommand usage of C2P for Kyverno to readme

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* Move former c2p description for OCM to doc/ocm.md

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

---------

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`f6a2831`](https://github.com/oscal-compass/compliance-to-policy/commit/f6a28310486d05ba20e43aef116a056f71a9f3f6))

* add sample usage movie (#6)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`c66ed08`](https://github.com/oscal-compass/compliance-to-policy/commit/c66ed081d88251d36bb3aa2ee76d0540b71bb199))

* Fix indent in README

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`50db9a6`](https://github.com/oscal-compass/compliance-to-policy/commit/50db9a6c58cf74c0dee5d3255aabd7bfbb2ae982))

* Redesign of C2P for pipeline use case (#3)

* update k8s library version

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* decomposer outputs should contain only successfully decomposed policies

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* refactor

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* decomposed resources include policy-generator manifest

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* reuse policy-generator manifest generated at decomposition

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* composer takes OSCAL

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* composer generates policySet

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* composer takes c2pcr

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* refactor

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* add templatized parameter filling

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* add default namespace selector

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* refactor

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* use kustomize type provided by sigs.k8s.io

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* add component-title annotation

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* add reporter

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* add placement_type and helpers

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* convert policy to policy report

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* compliance-report from policy reports

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* do not ommit the InformGatekeeper/KyvernoPolicies field

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* refactor

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* create c2pcli

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* enable gorelease

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* move policyResultDir parameter to c2pcr

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* generate reports as default

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* add metadata to compliance report

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* add scripts

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* fix

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* separate generated OCM manifests and policy generator sources

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* fix NPE when unable to get statuses

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* add scripts for setting up argocd

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* ocm status collector

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* fix

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* generate md file

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* fix

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* add report-utils

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* use OCM gitops

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* remove the dependency on hub namepsace from Reporter

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* allow to use PolicyGenerator with empty namespace in PolicyDefaults

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

* upsate collector script

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt;

---------

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`fba0dd8`](https://github.com/oscal-compass/compliance-to-policy/commit/fba0dd88012c13e1fe9f1aedef8a477c45ad1205))

* Initial implementation (#2)

Signed-off-by: Takumi Yanagawa &lt;yana@jp.ibm.com&gt; ([`e9d9f53`](https://github.com/oscal-compass/compliance-to-policy/commit/e9d9f53110d586483ad4182b0bcc4f17bdf34da2))

* Initial commit ([`bad869f`](https://github.com/oscal-compass/compliance-to-policy/commit/bad869f330b6bc0b9d87e57db7ff5c8d44f5c853))
