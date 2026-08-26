# -*- mode:python; coding:utf-8 -*-

# Copyright 2026 John Earle
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import json
import pathlib
import uuid as uuid_module
from typing import Any, Dict, List, Optional

import yaml
from pydantic.v1 import Field
from trestle.oscal.common import Finding, FindingTarget, ObjectiveStatus

from c2p.common.err import C2PError
from c2p.common.logging import getLogger
from c2p.common.utils import get_datetime, get_dict_safely
from c2p.framework.models import Policy, PVPResult, RawResult
from c2p.framework.models.pvp_result import (
    ObservationByCheck,
    Property,
    PVPResult,
    ResultEnum,
    Subject,
)
from c2p.framework.plugin_spec import PluginConfig, PluginSpec

logger = getLogger(__name__)

# Prowler emits OCSF Detection Findings whose status_code is one of
# PASS / FAIL / MANUAL / MUTED.
#
# MUTED maps to Error, not Pass. A muted finding is a failure suppressed by
# Prowler's mutelist - a decision made outside the compliance artifacts, with
# no owner, rationale or expiry attached. Reporting it as a pass would let a
# scanner configuration silently satisfy a control.
status_dictionary = {
    'PASS': ResultEnum.Pass,
    'FAIL': ResultEnum.Failure,
    'MANUAL': ResultEnum.Error,
    'MUTED': ResultEnum.Error,
}

# A component definition may namespace its check ids (e.g.
# "prowler.iam_root_mfa_enabled") so several tools can coexist in one rule set,
# or use Prowler's bare ids directly. Both are supported: the prefix is stripped
# before the id is handed to Prowler, and restored on the way back out, so
# observations speak the same dialect as the document they will be matched
# against.
CHECK_ID_PREFIX = 'prowler.'

# Written by generate_pvp_policy, read back by generate_pvp_result. Policy
# generation and result collection are separate runs - often separate processes,
# with the scan in between - so the rule-to-check mapping is recorded on disk in
# the deliverable directory rather than held in memory.
CHECKS_FILENAME = 'checks.json'
RULES_FILENAME = 'rules.json'

# Prowler's tunable thresholds, consumed via `--config-file`. Its schema is
# keyed by provider at the top level (aws, azure, gcp, kubernetes, github, ...),
# so a Parameter_Id is the dotted path into that tree, e.g.
# `aws.max_ebs_snapshots`.
CONFIG_FILENAME = 'config.yaml'

# Flags that narrow what a scan covers, verified against
# prowler/providers/<provider>/lib/arguments/arguments.py.
#
# Scope is a runtime concern, not a compliance claim: a component definition
# describes a *kind* of system, while an account, project or bucket is an
# instance of it. So scope is supplied by the runner through plugin config and
# never read from OSCAL - which also keeps targets and credentials out of
# documents that get published to auditors.
#
# Note there is no AWS `--account-id`: the account under scan is whichever one
# the credentials resolve to, selected with --profile or --role.
SCOPE_FLAGS = {
    'aws': ['profile', 'role', 'organizations-role', 'region', 'excluded-regions', 'resource-arns', 'resource-tags'],
    'gcp': [
        'project-ids',
        'excluded-project-ids',
        'organization-id',
        'credentials-file',
        'impersonate-service-account',
    ],
    'azure': ['subscription-ids', 'tenant-id', 'azure-resource-groups', 'azure-region'],
    'kubernetes': ['context', 'namespaces', 'cluster-name', 'kubeconfig-file'],
    'm365': ['tenant-id', 'region'],
    'github': ['organizations', 'repositories', 'repo-list-file', 'exclude-workflows'],
    'googleworkspace': [],  # the provider defines no arguments of its own
}

SUPPORTED_PROVIDERS = [
    'aws',
    'azure',
    'gcp',
    'kubernetes',
    'm365',
    'github',
    'googleworkspace',
]


class PluginConfigProwler(PluginConfig):
    provider: str = Field(..., title='Prowler provider (aws, azure, gcp, kubernetes, m365, ...)')
    deliverable_policy_dir: str = Field(..., title='Path to deliverable (generated) policy directory')
    output_dir: Optional[str] = Field('.', title='Directory Prowler writes its OCSF output to')
    scope: Optional[Dict[str, Any]] = Field(
        None,
        title='Scope narrowing the scan, as Prowler flags without the leading dashes '
        "(e.g. {'project-ids': ['prod-1']}). Underscores are accepted in place of dashes.",
    )


class PluginProwler(PluginSpec):
    """Prowler as a Policy Validation Point.

    Prowler ships its own checks, so policy generation is check *selection*
    rather than policy synthesis: the rule set determines which of Prowler's
    checks run. Results arrive as OCSF Detection Findings and are grouped by
    check id into one observation each, with one subject per evaluated
    resource.
    """

    def __init__(self, config: Optional[PluginConfigProwler] = None) -> None:
        super().__init__()
        self.config = config

    # ------------------------------------------------------------------ helpers
    @staticmethod
    def _to_native_check_id(check_id: str) -> str:
        return check_id[len(CHECK_ID_PREFIX) :] if check_id.startswith(CHECK_ID_PREFIX) else check_id

    @staticmethod
    def _check_id_of(finding: Dict[str, Any]) -> str:
        check_id = get_dict_safely(finding, ['metadata', 'event_code'])
        if not check_id:
            raise C2PError(
                'Unable to determine the check id of a Prowler finding: '
                'metadata.event_code is missing. Prowler\'s OCSF output format may have changed.'
            )
        return check_id

    @staticmethod
    def _parse_value(value: Optional[str]) -> Any:
        """Resolve a parameter value as a YAML scalar.

        Prowler's config mixes types, and some values that look numeric are
        genuinely strings - a Kubernetes version like "1.28" must not become
        the float 1.28. Rather than guessing, the value is read as YAML, so the
        author controls the type the same way they would in the config file
        itself: `1.28` is a float, `"1.28"` is a string, `[a, b]` is a list.
        Anything YAML cannot parse is kept verbatim as a string.
        """
        if value is None:
            return None
        try:
            return yaml.safe_load(value)
        except yaml.YAMLError:
            return value

    def _build_config(self, parameters: Optional[List[Any]]) -> Dict[str, Any]:
        """Nest dotted Parameter_Ids into Prowler's provider-keyed config tree.

        C2P sources parameters from the non-validation component, so in a
        heterogeneous component definition every plugin is handed every
        parameter - including ones meant for Kyverno or OCM. Only parameters
        rooted at this run's provider are ours to act on; the rest are left
        alone rather than written into a config Prowler would reject.
        """
        config: Dict[str, Any] = {}
        for parameter in parameters if parameters else []:
            if not parameter.id:
                continue
            segments = parameter.id.split('.')
            if segments[0] != self.config.provider or len(segments) < 2:
                logger.debug(f"Parameter '{parameter.id}' is not a {self.config.provider} config key; skipping.")
                continue
            node = config
            for segment in segments[:-1]:
                node = node.setdefault(segment, {})
                if not isinstance(node, dict):
                    raise C2PError(
                        f"Parameter '{parameter.id}' conflicts with another parameter: "
                        f"'{segment}' is used both as a value and as a group."
                    )
            node[segments[-1]] = self._parse_value(parameter.value)
        return config

    def _scope_args(self) -> List[str]:
        """Render the configured scope as Prowler arguments."""
        if not self.config.scope:
            return []
        allowed = SCOPE_FLAGS[self.config.provider]
        args: List[str] = []
        for key, value in self.config.scope.items():
            flag = key.lstrip('-').replace('_', '-')
            if flag not in allowed:
                raise C2PError(
                    f"'{flag}' is not a scope flag of the Prowler {self.config.provider} provider. "
                    + (
                        f'Supported: {", ".join(allowed)}.'
                        if allowed
                        else f'The {self.config.provider} provider takes no scope flags.'
                    )
                )
            values = value if isinstance(value, list) else [value]
            args.append(f'--{flag} ' + ' '.join(str(v) for v in values))
        return args

    # ----------------------------------------------------------------- generate
    def generate_pvp_policy(self, policy: Policy) -> Any:
        """Write the check selection Prowler should run.

        Emits `checks.json` (consumable via `prowler <provider> --checks-file`)
        alongside a human-readable `rules.json` recording which rule each check
        serves, so a result can be traced back to the rule that required it.

        Any parameters rooted at this provider are written to `config.yaml`
        (`prowler <provider> --config-file`), so a threshold a check compares
        against is declared in the component definition and reviewed alongside
        the rule, rather than living in scanner configuration off to one side.
        """
        if self.config is None:
            raise C2PError('PluginConfigProwler is required to generate a Prowler policy.')
        if self.config.provider not in SUPPORTED_PROVIDERS:
            raise C2PError(
                f"Unsupported Prowler provider '{self.config.provider}'. "
                f'Supported providers: {", ".join(SUPPORTED_PROVIDERS)}'
            )

        scope_args = self._scope_args()

        deliverable_dir = pathlib.Path(self.config.deliverable_policy_dir)
        if not deliverable_dir.exists():
            logger.info(f"The deliverable policy directory '{deliverable_dir}' is not found. Creating...")
            deliverable_dir.mkdir(parents=True)

        checks: List[str] = []
        rules: Dict[str, List[str]] = {}
        for rule_set in policy.rule_sets if policy.rule_sets else []:
            checks.append(self._to_native_check_id(rule_set.check_id))
            # Recorded as written in the component definition, so results can be
            # reported back in the same form the document uses.
            # A rule may be verified by several checks; C2P delivers those as
            # separate rule sets sharing a rule_id.
            rules.setdefault(rule_set.rule_id, [])
            if rule_set.check_id not in rules[rule_set.rule_id]:
                rules[rule_set.rule_id].append(rule_set.check_id)

        checks = sorted(set(checks))
        if not checks:
            logger.warning('No checks were derived from the component definition.')

        # Prowler reads this as json_file[provider] (lib/check/check.py), so the
        # list must be keyed by provider. A bare array parses, then fails inside
        # Prowler's own exception handler and leaves the check set as None.
        checks_file = deliverable_dir / CHECKS_FILENAME
        checks_file.write_text(json.dumps({self.config.provider: checks}, indent=2))
        rules_file = deliverable_dir / RULES_FILENAME
        rules_file.write_text(json.dumps(rules, indent=2, sort_keys=True))

        # Thresholds a check compares against belong beside the rule they serve,
        # not in scanner configuration nobody reviews.
        prowler_config = self._build_config(policy.parameters)
        config_file = None
        if prowler_config:
            config_file = deliverable_dir / CONFIG_FILENAME
            config_file.write_text(yaml.safe_dump(prowler_config, default_flow_style=False, sort_keys=True))
            logger.info(f'Wrote {len(prowler_config[self.config.provider])} config parameter(s) to {config_file}')

        logger.info(f'Generated {len(checks)} check(s) for provider {self.config.provider} in {deliverable_dir}')
        return {
            'provider': self.config.provider,
            'checks': checks,
            'rules': rules,
            'checks_file': str(checks_file),
            'config_file': str(config_file) if config_file else None,
            'config': prowler_config,
            'scope': ' '.join(scope_args),
            'command': (
                f'prowler {self.config.provider} '
                + (' '.join(scope_args) + ' ' if scope_args else '')
                + f'--checks-file {checks_file} '
                + (f'--config-file {config_file} ' if config_file else '')
                + f'--output-formats json-ocsf --output-directory {self.config.output_dir}'
            ),
        }

    # -------------------------------------------------------------------- result
    def generate_pvp_result(self, raw_result: RawResult) -> PVPResult:
        """Convert Prowler's OCSF Detection Findings into observations.

        One observation per check; one subject per evaluated resource, so a
        check that passes on 99 resources and fails on 1 is not reported as a
        simple failure.
        """
        pvp_result: PVPResult = PVPResult()
        observations: List[ObservationByCheck] = []

        findings = raw_result.data if isinstance(raw_result.data, list) else [raw_result.data]

        # Prowler reports bare ids; map each back to the form the component
        # definition used, so C2P can match the observation to its rule set.
        reported_as = {self._to_native_check_id(c): c for c in self._all_check_ids()}

        findings_by_check: Dict[str, List[Dict[str, Any]]] = {}
        for finding in findings:
            native = self._check_id_of(finding)
            findings_by_check.setdefault(reported_as.get(native, native), []).append(finding)

        for check_id, check_findings in sorted(findings_by_check.items()):
            subjects: List[Subject] = []
            muted = 0
            for finding in check_findings:
                status_code = str(get_dict_safely(finding, 'status_code', '')).upper()
                if status_code == 'MUTED':
                    muted = muted + 1
                result = status_dictionary[status_code] if status_code in status_dictionary else ResultEnum.Error

                resources = get_dict_safely(finding, 'resources')
                resources = resources if isinstance(resources, list) and resources else [{}]
                for resource in resources:
                    name = get_dict_safely(resource, 'name')
                    uid = get_dict_safely(resource, 'uid')
                    resource_type = get_dict_safely(resource, 'type', 'resource')
                    region = get_dict_safely(resource, 'region', '')
                    subjects.append(
                        Subject(
                            title=f'{resource_type} {name if name else uid} {region}'.strip(),
                            type='resource',
                            resource_id=uid if uid else (name if name else 'unknown'),
                            result=result,
                            reason=get_dict_safely(finding, 'status_detail'),
                        )
                    )

            props = [
                Property(name='assessment-check-id', value=check_id),
                Property(name='resources-evaluated', value=str(len(subjects))),
            ]
            if muted > 0:
                # Surfaced explicitly: a muted result is suppressed, not passing.
                props.append(Property(name='resources-muted', value=str(muted)))

            observations.append(
                ObservationByCheck(
                    check_id=check_id,
                    title=check_id,
                    description=get_dict_safely(check_findings[0], ['finding_info', 'title'], check_id),
                    methods=['AUTOMATED'],
                    collected=get_datetime(),
                    subjects=subjects,
                    props=props,
                )
            )

        pvp_result.observations_by_check = observations
        findings = self._generate_findings(observations)
        if findings:
            pvp_result.findings = findings
        return pvp_result

    # ------------------------------------------------------------------ findings
    def _load_rule_mapping(self) -> Dict[str, List[str]]:
        """Recover the rule -> checks mapping recorded during policy generation."""
        if self.config is None:
            return {}
        rules_file = pathlib.Path(self.config.deliverable_policy_dir) / RULES_FILENAME
        if not rules_file.exists():
            logger.info(
                f"No '{RULES_FILENAME}' in '{self.config.deliverable_policy_dir}'; "
                'reporting observations without rule-level findings.'
            )
            return {}
        return json.loads(rules_file.read_text())

    def _all_check_ids(self) -> List[str]:
        """Every check id recorded during policy generation, as the document wrote it."""
        return [check for checks in self._load_rule_mapping().values() for check in checks]

    def _generate_findings(self, observations: List[ObservationByCheck]) -> List[Finding]:
        """Roll check observations up into one finding per rule.

        An observation says what a check saw; a finding says whether the rule
        holds. A rule is satisfied only when every check serving it ran and
        every evaluated resource passed - one failing resource fails the rule,
        and a check that errored or was muted leaves the rule unproven rather
        than passing. Without this the assessment results carry evidence but no
        judgement, and a reader has to re-derive the verdict by hand.
        """
        rule_mapping = self._load_rule_mapping()
        if not rule_mapping:
            return []

        results_by_check: Dict[str, List[ResultEnum]] = {
            o.check_id: [s.result for s in o.subjects] for o in observations
        }

        findings: List[Finding] = []
        for rule_id, checks in sorted(rule_mapping.items()):
            failed: List[str] = []
            unproven: List[str] = []
            for check in checks:
                if check not in results_by_check:
                    unproven.append(f'{check} (did not report)')
                    continue
                results = results_by_check[check]
                if any(r == ResultEnum.Failure for r in results):
                    failed.append(check)
                elif any(r != ResultEnum.Pass for r in results):
                    unproven.append(f'{check} (error or suppressed)')

            # `reason` is an OSCAL controlled vocabulary (pass / fail / other);
            # the prose belongs in `remarks`. 'other' is the honest value for a
            # rule that was neither proven nor disproven.
            if failed:
                state, reason = 'not-satisfied', 'fail'
                remarks = 'Failing checks: ' + ', '.join(sorted(failed))
            elif unproven:
                state, reason = 'not-satisfied', 'other'
                remarks = 'Not proven: ' + '; '.join(sorted(unproven))
            else:
                state, reason = 'satisfied', 'pass'
                remarks = 'All checks passed on every evaluated resource: ' + ', '.join(sorted(checks))

            findings.append(
                Finding(
                    uuid=str(uuid_module.uuid4()),
                    title=rule_id,
                    description=remarks,
                    target=FindingTarget(
                        type='objective-id',
                        target_id=rule_id,
                        status=ObjectiveStatus(state=state, reason=reason, remarks=remarks),
                    ),
                )
            )
        return findings
