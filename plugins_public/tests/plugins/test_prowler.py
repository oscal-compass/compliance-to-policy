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

import pytest
import yaml

from c2p.common.err import C2PError
from c2p.framework.models import Policy, RawResult
from c2p.framework.models.policy import Parameter, RuleSet
from c2p.framework.models.pvp_result import ResultEnum
from plugins_public.plugins.prowler import PluginConfigProwler, PluginProwler

DATA = pathlib.Path(__file__).parent.parent / 'data' / 'prowler'


def _policy() -> Policy:
    # A rule verified by two checks arrives as two rule sets sharing a rule_id.
    return Policy(
        rule_sets=[
            RuleSet(
                rule_id='aws-root-mfa',
                rule_description='Root account carries MFA',
                check_id='prowler.iam_root_mfa_enabled',
            ),
            RuleSet(rule_id='aws-root-mfa', check_id='prowler.iam_root_hardware_mfa_enabled'),
            RuleSet(rule_id='aws-console-mfa', check_id='iam_user_mfa_enabled_console_access'),
        ]
    )


def _raw() -> RawResult:
    return RawResult(data=json.loads((DATA / 'scan.ocsf.json').read_text()))


def test_generate_pvp_policy_writes_check_selection(tmp_path):
    plugin = PluginProwler(PluginConfigProwler(provider='aws', deliverable_policy_dir=str(tmp_path / 'out')))
    result = plugin.generate_pvp_policy(_policy())

    # The namespace prefix is stripped: Prowler knows only the bare check id.
    assert result['checks'] == [
        'iam_root_hardware_mfa_enabled',
        'iam_root_mfa_enabled',
        'iam_user_mfa_enabled_console_access',
    ]
    # A rule verified by several checks keeps all of them, recorded as the
    # component definition wrote them so results can be reported back in kind.
    assert result['rules']['aws-root-mfa'] == [
        'prowler.iam_root_mfa_enabled',
        'prowler.iam_root_hardware_mfa_enabled',
    ]
    # Prowler reads json_file[provider]; a bare array leaves its check set None.
    written = json.loads(pathlib.Path(result['checks_file']).read_text())
    assert written == {'aws': result['checks']}


def test_generate_pvp_policy_rejects_unknown_provider(tmp_path):
    plugin = PluginProwler(PluginConfigProwler(provider='not-a-cloud', deliverable_policy_dir=str(tmp_path / 'out')))
    with pytest.raises(C2PError):
        plugin.generate_pvp_policy(_policy())


def test_generate_pvp_result_groups_by_check():
    observations = PluginProwler().generate_pvp_result(_raw()).observations_by_check
    by_check = {o.check_id: o for o in observations}

    # Four distinct checks in the fixture, one observation each.
    assert len(observations) == 4
    # Two findings for the same check collapse into one observation, two subjects.
    console = by_check['iam_user_mfa_enabled_console_access']
    assert len(console.subjects) == 2
    assert {s.result for s in console.subjects} == {ResultEnum.Pass, ResultEnum.Failure}


def test_check_ids_are_reported_as_the_document_wrote_them(tmp_path):
    """Observations must match the component definition's dialect, or C2P drops them.

    Standalone, the plugin reports Prowler's bare ids. After a policy run whose
    rule sets namespaced them, the same findings come back namespaced.
    """
    bare = PluginProwler().generate_pvp_result(_raw()).observations_by_check
    assert all(not o.check_id.startswith('prowler.') for o in bare)

    plugin = PluginProwler(PluginConfigProwler(provider='aws', deliverable_policy_dir=str(tmp_path / 'out')))
    plugin.generate_pvp_policy(_policy())  # rule sets use 'prowler.'-prefixed ids
    namespaced = {o.check_id for o in plugin.generate_pvp_result(_raw()).observations_by_check}
    assert 'prowler.iam_root_mfa_enabled' in namespaced
    # A check absent from the rule sets keeps the id Prowler reported.
    assert 'cloudtrail_multi_region_enabled' in namespaced


def test_muted_is_not_reported_as_pass():
    """A mutelist entry is a suppressed failure, not a satisfied control."""
    observations = PluginProwler().generate_pvp_result(_raw()).observations_by_check
    muted = next(o for o in observations if o.check_id == 'iam_administrator_access_with_mfa')
    assert muted.subjects[0].result == ResultEnum.Error
    assert any(p.name == 'resources-muted' and p.value == '1' for p in muted.props)


def test_manual_is_reported_as_error():
    observations = PluginProwler().generate_pvp_result(_raw()).observations_by_check
    manual = next(o for o in observations if o.check_id == 'cloudtrail_multi_region_enabled')
    assert manual.subjects[0].result == ResultEnum.Error


def test_resources_evaluated_is_recorded():
    """'0 failed' is meaningless without knowing how many were examined."""
    observations = PluginProwler().generate_pvp_result(_raw()).observations_by_check
    console = next(o for o in observations if o.check_id == 'iam_user_mfa_enabled_console_access')
    assert any(p.name == 'resources-evaluated' and p.value == '2' for p in console.props)


def test_missing_event_code_raises():
    """A silent miss would read downstream as 'check not run', not 'adapter broken'."""
    with pytest.raises(C2PError):
        PluginProwler().generate_pvp_result(RawResult(data=[{'status_code': 'PASS'}]))


def _val(x):
    return x.value if hasattr(x, 'value') else x


def _findings_by_rule(tmp_path):
    plugin = PluginProwler(PluginConfigProwler(provider='aws', deliverable_policy_dir=str(tmp_path / 'out')))
    plugin.generate_pvp_policy(_policy())  # records the rule -> checks mapping
    result = plugin.generate_pvp_result(_raw())
    return {f.target.target_id: f for f in result.findings}


def test_findings_roll_checks_up_to_rules(tmp_path):
    """Observations say what a check saw; findings say whether the rule holds."""
    by_rule = _findings_by_rule(tmp_path)
    assert set(by_rule) == {'aws-root-mfa', 'aws-console-mfa'}


def test_one_failing_resource_fails_the_rule(tmp_path):
    """The console check passes on one user and fails on another."""
    status = _findings_by_rule(tmp_path)['aws-console-mfa'].target.status
    assert _val(status.state) == 'not-satisfied'
    assert _val(status.reason) == 'fail'


def test_check_that_did_not_report_leaves_rule_unproven(tmp_path):
    """A silent check is not a pass. 'other' distinguishes unproven from failed."""
    status = _findings_by_rule(tmp_path)['aws-root-mfa'].target.status
    assert _val(status.state) == 'not-satisfied'
    assert _val(status.reason) == 'other'
    assert 'did not report' in status.remarks


def test_findings_are_omitted_without_a_rule_mapping():
    """Result collection can run standalone; it degrades to observations only."""
    assert PluginProwler().generate_pvp_result(_raw()).findings is None


def _generate(tmp_path, parameters):
    plugin = PluginProwler(PluginConfigProwler(provider='aws', deliverable_policy_dir=str(tmp_path / 'out')))
    policy = _policy()
    policy.parameters = parameters
    return plugin.generate_pvp_policy(policy)


def test_parameters_nest_into_the_provider_config(tmp_path):
    result = _generate(tmp_path, [Parameter(id='aws.max_ebs_snapshots', value='50')])
    assert result['config'] == {'aws': {'max_ebs_snapshots': 50}}
    written = yaml.safe_load(pathlib.Path(result['config_file']).read_text())
    assert written == {'aws': {'max_ebs_snapshots': 50}}
    assert f"--config-file {result['config_file']}" in result['command']


def test_parameters_for_other_plugins_are_ignored(tmp_path):
    """C2P hands every plugin every parameter; only ours belong in our config."""
    result = _generate(
        tmp_path,
        [
            Parameter(id='aws.max_ebs_snapshots', value='50'),
            Parameter(id='kubernetes.audit_log_maxage', value='30'),
            Parameter(id='org.gh.orgs', value='nasa,esa'),
        ],
    )
    assert result['config'] == {'aws': {'max_ebs_snapshots': 50}}


def test_parameter_types_follow_yaml_scalar_rules(tmp_path):
    """A version like "1.28" is a string; unquoted 1.28 is a float."""
    result = _generate(
        tmp_path,
        [
            Parameter(id='aws.threshold', value='0.3'),
            Parameter(id='aws.version', value='"1.28"'),
            Parameter(id='aws.enabled', value='true'),
            Parameter(id='aws.severities', value='[High, MEDIUM]'),
        ],
    )
    assert result['config']['aws'] == {
        'threshold': 0.3,
        'version': '1.28',
        'enabled': True,
        'severities': ['High', 'MEDIUM'],
    }


def test_no_parameters_means_no_config_file(tmp_path):
    result = _generate(tmp_path, [])
    assert result['config_file'] is None
    assert '--config-file' not in result['command']
    assert not (tmp_path / 'out' / 'config.yaml').exists()


def _scoped(tmp_path, provider, scope):
    plugin = PluginProwler(
        PluginConfigProwler(provider=provider, deliverable_policy_dir=str(tmp_path / 'out'), scope=scope)
    )
    return plugin.generate_pvp_policy(_policy())


def test_scope_narrows_the_generated_command(tmp_path):
    result = _scoped(tmp_path, 'gcp', {'project-ids': ['prod-1', 'prod-2']})
    assert '--project-ids prod-1 prod-2' in result['command']


def test_scope_accepts_underscores(tmp_path):
    result = _scoped(tmp_path, 'kubernetes', {'namespaces': 'kube-system'})
    assert '--namespaces kube-system' in result['command']


def test_aws_account_id_is_rejected(tmp_path):
    """Prowler has no --account-id; the account follows from the credentials.

    Silently emitting it would produce a command that fails at scan time, long
    after the compliance content was reviewed.
    """
    with pytest.raises(C2PError) as err:
        _scoped(tmp_path, 'aws', {'account-id': '123456789012'})
    assert 'account-id' in str(err.value)
    assert 'profile' in str(err.value)


def test_scope_flag_from_another_provider_is_rejected(tmp_path):
    with pytest.raises(C2PError):
        _scoped(tmp_path, 'aws', {'project-ids': ['prod-1']})


def test_provider_without_scope_flags_says_so(tmp_path):
    with pytest.raises(C2PError) as err:
        _scoped(tmp_path, 'googleworkspace', {'organization-id': 'x'})
    assert 'takes no scope flags' in str(err.value)


def test_no_scope_leaves_the_command_unchanged(tmp_path):
    result = _scoped(tmp_path, 'aws', None)
    assert result['command'].startswith('prowler aws --checks-file')
