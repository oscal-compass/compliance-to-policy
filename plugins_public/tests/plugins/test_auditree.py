# -*- mode:python; coding:utf-8 -*-

# Copyright 2024 IBM Corporation
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
import shutil
import tempfile
from distutils.util import strtobool
from pathlib import Path

from c2p.framework.models import Parameter, Policy, PVPResult, RawResult
from plugins_public.plugins.auditree import PluginAuditree, PluginConfigAuditree
from plugins_public.tests.utils import load_yaml, load_yamls

TEST_DATA_DIR = os.getenv('TEST_DATA_DIR', 'plugins_public/tests/data/auditree')
TEST_DATA_DIR_PATH = Path(TEST_DATA_DIR)
OVERWRITE_EXPECTED_DATA = os.getenv('OVERWRITE_EXPECTED_DATA', 'false')


def test_auditree_pvp_result_to_compliance():
    check_results_path = TEST_DATA_DIR_PATH / 'check_results.json'
    with check_results_path.open('r') as f:
        check_results = json.load(f)
    raw_result = RawResult(
        data=check_results,
        additional_props={
            'locker_url': 'https://github.com/MY_ORG/MY_EVIDENCE_REPO',
        },
    )
    pvp_result = PluginAuditree().generate_pvp_result(raw_result)

    expected = TEST_DATA_DIR_PATH / 'pvp_result.json'
    if strtobool(OVERWRITE_EXPECTED_DATA):
        with expected.open('w') as f:
            f.write(pvp_result.json(indent=2, exclude_none=True))

    expected = PVPResult.parse_file(expected)

    assert len(expected.observations_by_check) == len(pvp_result.observations_by_check)

    check_ids = set([x.check_id for x in pvp_result.observations_by_check])
    check_ids_expected = set([x.check_id for x in expected.observations_by_check])
    assert check_ids_expected == check_ids

    # Check triplet of check_id, subject.title, subject.result
    def conv(x: PVPResult):
        return [
            [check.check_id, subject.title, subject.result]
            for check in x.observations_by_check
            for subject in check.subjects
        ]

    css = conv(pvp_result)
    css_expected = conv(expected)
    assert css_expected == css


def test_auditree_compliance_to_policy():
    with tempfile.TemporaryDirectory() as tmpdirname:
        generated_auditree_json = Path(tmpdirname) / 'auditree.json'
        config = PluginConfigAuditree(
            auditree_json_template=f'{TEST_DATA_DIR}/auditree.template.json', output=generated_auditree_json.as_posix()
        )
        parameters = [Parameter(id='org.gh.orgs', value='foo,bar')]
        policy = Policy(rule_sets=[], parameters=parameters)
        PluginAuditree(config).generate_pvp_policy(policy)

        expected = TEST_DATA_DIR_PATH / 'auditree.json'
        if strtobool(OVERWRITE_EXPECTED_DATA):
            shutil.copy(generated_auditree_json, expected)

        with generated_auditree_json.open('r') as f:
            actual = json.load(f)

        assert set(actual['org']['gh']['orgs']) == set(['foo', 'bar'])
