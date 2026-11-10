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

import json
import os

from c2p.framework.models import Parameter, Policy, RawResult, RuleSet
from c2p.framework.models.pvp_result import ResultEnum
from plugins_public.plugins.vap import PluginConfigVap, PluginVap
from plugins_public.tests.utils import load_yaml

TEST_DATA_DIR = os.getenv('TEST_DATA_DIR', 'plugins_public/tests/data/vap-cel')


def outcomes():
    return json.load(open(f'{TEST_DATA_DIR}/dryrun-outcomes.json'))


def test_generates_a_validating_admission_policy(tmp_path):
    config = PluginConfigVap(policy_template_dir=f'{TEST_DATA_DIR}/templates', deliverable_policy_dir=str(tmp_path))
    policy = Policy(
        rule_sets=[RuleSet(rule_id='labels-present', check_id='vap-labels-present')],
        parameters=[Parameter(id='required_labels', value='a b')],
    )

    PluginVap(config).generate_pvp_policy(policy)

    document = load_yaml(tmp_path / 'labels-present' / 'vap-labels-present.yaml')
    assert document['kind'] == 'ValidatingAdmissionPolicy'
    assert document['metadata']['name'] == 'vap-labels-present'


def test_the_same_parameter_renders_identically_for_both_cel_engines(tmp_path):
    """One authored parameter, byte-identical CEL in Kyverno CEL and in VAP. If these drift,
    the claim that a control outlives the engine stops being demonstrable."""
    from plugins_public.plugins.kyverno_cel import (
        PluginConfigKyvernoCel,
        PluginKyvernoCel,
    )

    parameters = [Parameter(id='required_labels', value='a b'), Parameter(id='owner', value='x')]
    vap_config = PluginConfigVap(
        policy_template_dir=f'{TEST_DATA_DIR}/templates', deliverable_policy_dir=str(tmp_path / 'vap')
    )
    PluginVap(vap_config).generate_pvp_policy(
        Policy(rule_sets=[RuleSet(rule_id='labels-present', check_id='vap-labels-present')], parameters=parameters)
    )
    cel_config = PluginConfigKyvernoCel(
        policy_template_dir='plugins_public/tests/data/kyverno-cel/templates',
        deliverable_policy_dir=str(tmp_path / 'cel'),
    )
    PluginKyvernoCel(cel_config).generate_pvp_policy(
        Policy(
            rule_sets=[RuleSet(rule_id='labels-present', check_id='kyverno-cel-labels-present')], parameters=parameters
        )
    )

    def required_expression(path):
        return next(line for line in path.read_text().splitlines() if 'expression' in line and '[' in line).strip()

    vap_line = required_expression(tmp_path / 'vap' / 'labels-present' / 'vap-labels-present.yaml')
    cel_line = required_expression(tmp_path / 'cel' / 'labels-present' / 'kyverno-cel-labels-present.yaml')
    assert vap_line == cel_line


def test_probe_outcomes_become_observations_per_check():
    pvp_result = PluginVap().generate_pvp_result(RawResult(data=outcomes()))

    by_check = {o.check_id: o for o in pvp_result.observations_by_check}
    assert sorted(by_check) == ['vap-digest-pinned', 'vap-labels-present']
    assert len(by_check['vap-labels-present'].subjects) == 2


def test_an_unattributable_probe_is_an_error_never_a_pass():
    """One dry run reveals only the first rejection. An outcome the probe could not attribute
    is unknown, and recording it as pass would be a false pass in the posture."""
    pvp_result = PluginVap().generate_pvp_result(RawResult(data=outcomes()))

    labels = next(o for o in pvp_result.observations_by_check if o.check_id == 'vap-labels-present')
    assert sorted(s.result for s in labels.subjects) == [ResultEnum.Error, ResultEnum.Failure]


def test_subject_props_carry_the_engine_detail():
    pvp_result = PluginVap().generate_pvp_result(RawResult(data=outcomes()))

    labels = next(o for o in pvp_result.observations_by_check if o.check_id == 'vap-labels-present')
    props = {p.name: p.value for p in labels.subjects[0].props}
    assert props['engine.object'] == 'vap-labels-present'


def test_an_evidence_document_is_accepted_as_well_as_a_bare_list():
    """Collection does not always happen in the same process: the same results replayed from a
    c2p-evidence/v1 document on disk have to produce the same observations."""
    document = {
        'schemaVersion': 'c2p-evidence/v1',
        'engine': 'vap-cel',
        'observations': [{'checkId': o['checkId'], 'subjects': [o]} for o in outcomes()],
    }

    from_document = PluginVap().generate_pvp_result(RawResult(data=document))
    from_list = PluginVap().generate_pvp_result(RawResult(data=outcomes()))

    assert [o.check_id for o in from_document.observations_by_check] == [
        o.check_id for o in from_list.observations_by_check
    ]
    assert [len(o.subjects) for o in from_document.observations_by_check] == [
        len(o.subjects) for o in from_list.observations_by_check
    ]
