# -*- mode:python; coding:utf-8 -*-

# Copyright 2026 IBM Corporation
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

import os

import pytest

from c2p.common.err import C2PError
from c2p.framework.models import Parameter, Policy, RawResult, RuleSet
from c2p.framework.models.pvp_result import ResultEnum
from plugins_public.plugins.kyverno_cel import PluginConfigKyvernoCel, PluginKyvernoCel
from plugins_public.tests.utils import load_yaml

TEST_DATA_DIR = os.getenv('TEST_DATA_DIR', 'plugins_public/tests/data/kyverno-cel')


def a_policy():
    return Policy(
        rule_sets=[
            RuleSet(
                rule_id='labels-present',
                rule_description='workloads carry inventory labels',
                check_id='kyverno-cel-labels-present',
                check_description='every required label key is present and non-empty',
            )
        ],
        parameters=[
            Parameter(id='required_labels', value='app.kubernetes.io/name compliance.example.com/owner'),
            Parameter(id='owner', value='platform-team'),
        ],
    )


def test_generates_one_policy_per_check_carrying_its_provenance(tmp_path):
    config = PluginConfigKyvernoCel(
        policy_template_dir=f'{TEST_DATA_DIR}/templates', deliverable_policy_dir=str(tmp_path)
    )

    PluginKyvernoCel(config).generate_pvp_policy(a_policy())

    generated = tmp_path / 'labels-present' / 'kyverno-cel-labels-present.yaml'
    assert generated.exists()
    text = generated.read_text()
    # A generated policy has to say which rule, check and engine produced it, or an operator
    # holding only the cluster cannot get back to the control.
    assert 'rule labels-present / check kyverno-cel-labels-present' in text
    assert '(kyverno-cel)' in text
    assert 'every required label key is present and non-empty' in text

    document = load_yaml(generated)
    assert document['metadata']['name'] == 'kyverno-cel-labels-present'
    assert document['metadata']['labels']['compliance.oscal/rule-id'] == 'labels-present'


def test_a_blank_separated_parameter_renders_as_a_compact_cel_list(tmp_path):
    """The same parameter must render identically for every engine, and a CEL expression is
    compared as a string -- so no space after the separators."""
    config = PluginConfigKyvernoCel(
        policy_template_dir=f'{TEST_DATA_DIR}/templates', deliverable_policy_dir=str(tmp_path)
    )

    PluginKyvernoCel(config).generate_pvp_policy(a_policy())

    text = (tmp_path / 'labels-present' / 'kyverno-cel-labels-present.yaml').read_text()
    assert '\'["app.kubernetes.io/name","compliance.example.com/owner"]\'' in text
    assert '"platform-team"' in text


def test_a_check_with_no_template_is_an_error_not_a_silent_omission(tmp_path):
    config = PluginConfigKyvernoCel(
        policy_template_dir=f'{TEST_DATA_DIR}/templates', deliverable_policy_dir=str(tmp_path)
    )
    policy = a_policy()
    policy.rule_sets[0].check_id = 'kyverno-cel-does-not-exist'

    with pytest.raises(C2PError):
        PluginKyvernoCel(config).generate_pvp_policy(policy)


def test_subject_comes_from_the_report_scope_when_results_name_no_resource():
    """These reports name their subject once, at the report level. Reading only
    results[].resources yields a posture where every subject is an unnamed resource."""
    reports = load_yaml(f'{TEST_DATA_DIR}/policyreports.wgpolicyk8s.io.yaml')

    pvp_result = PluginKyvernoCel().generate_pvp_result(RawResult(data=reports['items']))

    by_check = {o.check_id: o for o in pvp_result.observations_by_check}
    assert sorted(by_check) == ['kyverno-cel-digest-pinned', 'kyverno-cel-labels-present']
    subject = by_check['kyverno-cel-digest-pinned'].subjects[0]
    assert subject.title == 'StatefulSet storefront/db'
    assert subject.resource_id == 'apps/v1/StatefulSet/storefront/db'
    assert subject.result == ResultEnum.Pass


def test_an_admitted_objection_is_a_warning_not_a_failure():
    """`warn` means the engine objected and the object got in anyway: the control was not
    enforced, which is a different claim from the control being violated."""
    reports = load_yaml(f'{TEST_DATA_DIR}/policyreports.wgpolicyk8s.io.yaml')

    pvp_result = PluginKyvernoCel().generate_pvp_result(RawResult(data=reports['items']))

    labels = next(o for o in pvp_result.observations_by_check if o.check_id == 'kyverno-cel-labels-present')
    assert [s.result for s in labels.subjects] == [ResultEnum.Failure, ResultEnum.Warning]


def test_non_report_objects_are_ignored():
    reports = load_yaml(f'{TEST_DATA_DIR}/policyreports.wgpolicyk8s.io.yaml')

    pvp_result = PluginKyvernoCel().generate_pvp_result(RawResult(data=reports['items']))

    subjects = [s for o in pvp_result.observations_by_check for s in o.subjects]
    assert len(subjects) == 3


def test_evaluation_time_comes_from_the_report_not_collection_time():
    reports = load_yaml(f'{TEST_DATA_DIR}/policyreports.wgpolicyk8s.io.yaml')

    pvp_result = PluginKyvernoCel().generate_pvp_result(RawResult(data=reports['items']))

    labels = next(o for o in pvp_result.observations_by_check if o.check_id == 'kyverno-cel-labels-present')
    assert labels.subjects[0].evaluated_on.year == 2025
