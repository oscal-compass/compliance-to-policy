# C2P with Prowler

[Prowler](https://github.com/prowler-cloud/prowler) is an open-source security
scanner covering AWS, Azure, GCP, Kubernetes, M365, GitHub, Google Workspace and
others. It is used here as a Policy Validation Point (PVP).

Prowler differs from a policy engine such as Kyverno or OCM in one way that
shapes the plugin: **it ships its own checks**. There are no policy templates to
render, so `generate_pvp_policy` performs check *selection* rather than policy
synthesis — it decides which of Prowler's checks run, and with what thresholds.

## Component definition

Prowler is a component of type `Validation` whose title matches
`C2PConfig.pvp_name`. Its rule sets carry the `Check_Id` values naming Prowler
checks. Check ids may be written bare (`iam_root_mfa_enabled`) or namespaced
(`prowler.iam_root_mfa_enabled`) when several tools share a rule; results are
reported back in whichever form the document used.

`Parameter_Id` values are dotted paths into Prowler's provider-keyed
configuration, for example `aws.max_console_access_days`. This keeps a threshold
a check compares against beside the rule it serves, under the same review as the
control, instead of in scanner configuration held separately.

## Scope: which account, project or resource

OSCAL describes a *kind* of system; an AWS account or a GCP project is an
*instance* of one. A component definition therefore carries no target, and C2P's
`Policy` model has no field for one — it answers which checks to run, never
where to run them.

Scope is supplied through plugin config by whatever runs the scan, which also
keeps account identifiers and credentials out of documents shared with auditors.
The same generated `checks.json` is reused across every target; only the scope
differs.

```bash
python samples_public/prowler/compliance_to_policy.py -p gcp \
  -s '{"project-ids": ["prod-1", "prod-2"]}'
```

```
prowler gcp --project-ids prod-1 prod-2 --checks-file .../checks.json ...
```

Flags are validated against the provider actually selected:

| Provider | Scope flags |
| --- | --- |
| `aws` | `profile`, `role`, `organizations-role`, `region`, `excluded-regions`, `resource-arns`, `resource-tags` |
| `gcp` | `project-ids`, `excluded-project-ids`, `organization-id`, `credentials-file`, `impersonate-service-account` |
| `azure` | `subscription-ids`, `tenant-id`, `azure-resource-groups`, `azure-region` |
| `kubernetes` | `context`, `namespaces`, `cluster-name`, `kubeconfig-file` |
| `m365` | `tenant-id`, `region` |
| `github` | `organizations`, `repositories`, `repo-list-file`, `exclude-workflows` |
| `googleworkspace` | none — the provider defines no arguments of its own |

Note that AWS has no `--account-id`: the account under scan is whichever the
credentials resolve to. An unsupported flag is rejected at generation time
rather than producing a command that fails later, once the compliance content
has already been reviewed.

## Compliance to Policy

```bash
python samples_public/prowler/compliance_to_policy.py -o /tmp/prowler-policy
```

This writes three files:

| File | Purpose |
| --- | --- |
| `checks.json` | the check selection, for `prowler <provider> --checks-file` |
| `config.yaml` | tunable thresholds, for `prowler <provider> --config-file` |
| `rules.json` | which rule each check serves, read back when results arrive |

and prints the Prowler invocation:

```
prowler aws --checks-file /tmp/prowler-policy/checks.json \
  --config-file /tmp/prowler-policy/config.yaml \
  --output-formats json-ocsf --output-directory .
```

## Policy to Compliance

```bash
python samples_public/prowler/result_to_compliance.py -d /tmp/prowler-policy
```

Prowler's OCSF Detection Findings become one observation per check, with one
subject per evaluated resource — so a check that passes on 99 resources and
fails on 1 is not flattened into a bare failure.

Passing `-d` (the directory from the previous step) additionally rolls the check
results up into one **finding per rule**. Without it the plugin still emits
observations; it simply has no rule mapping to aggregate against.

### How results are interpreted

| Prowler status | Result | Why |
| --- | --- | --- |
| `PASS` | pass | |
| `FAIL` | failure | |
| `MANUAL` | error | requires human determination; not evidence of compliance |
| `MUTED` | error | a mutelist entry is a suppressed failure, with no owner, rationale or expiry recorded. Reporting it as a pass would let scanner configuration silently satisfy a control |

A rule is `satisfied` only when every check serving it reported and every
evaluated resource passed. The distinction between failed and unproven is kept:

| Outcome | `state` | `reason` |
| --- | --- | --- |
| a check failed on some resource | `not-satisfied` | `fail` |
| a check errored, was muted, or never reported | `not-satisfied` | `other` |
| all checks passed everywhere | `satisfied` | `pass` |

A check that did not run is not a pass. Treating silence as success is the
failure mode this guards against.

See [prowler.result.md](./prowler.result.md) for the assessment results this
produces.
