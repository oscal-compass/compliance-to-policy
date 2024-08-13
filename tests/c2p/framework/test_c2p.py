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
import pathlib
from typing import Any, Dict, Tuple

from pydantic.v1 import BaseModel
from trestle.oscal.assessment_results import AssessmentResults

from c2p.framework.c2p import C2P
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal
from c2p.framework.models.pvp_result import PVPResult
from tests.c2p import write_test_result_to_file

COMPONENT_DEFINITION_TEST_DATA = pathlib.Path('tests/data/framework/c2p/component-definition.json')
PVP_RESULT_TEST_DATA = pathlib.Path('tests/data/framework/c2p/pvp-result.json')
EXPECTED_ASSESSMENT_RESULTS_DATA = pathlib.Path('tests/data/framework/c2p/assessment-results.json')
OUTPUT_PATH = EXPECTED_ASSESSMENT_RESULTS_DATA


def extract_dicts(d: Dict[str, Any], excludes=[]) -> Tuple[Dict[str, Any], Dict[str, Any]]:
    primitives = []
    nested = []
    for key, value in d.items():
        if key in excludes:
            continue
        if isinstance(value, list) or isinstance(value, dict):
            nested.append((key, value))
        else:
            primitives.append((key, value))

    return dict(primitives), dict(nested)


def assert_pydantic_object(actual: BaseModel, expect: BaseModel, exludes=[]):
    actual, _ = extract_dicts(actual.dict(), excludes=exludes)
    expect, _ = extract_dicts(expect.dict(), excludes=exludes)
    assert actual == expect


def test_result_to_oscal():
    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = COMPONENT_DEFINITION_TEST_DATA.as_posix()
    c2p_config.pvp_name = 'OCM'
    c2p_config.result_title = 'TEST Assessment Results'
    c2p_config.result_description = 'OSCAL Assessment Results from TEST'

    pvp_result = json.load(PVP_RESULT_TEST_DATA.open('r'))
    c2p_config.pvp_result = PVPResult.parse_obj(pvp_result)

    c2p = C2P(c2p_config)
    assessment_results = c2p.result_to_oscal()
    expect = AssessmentResults.parse_file(EXPECTED_ASSESSMENT_RESULTS_DATA)

    assert_pydantic_object(assessment_results.metadata, expect.metadata, exludes=['last_modified'])
    assert_pydantic_object(assessment_results.import_ap, expect.import_ap)
    actual_result = assessment_results.results[0]
    expect_result = expect.results[0]
    assert_pydantic_object(actual_result, expect_result, exludes=['uuid', 'start'])
    actual_reviewed_controls = actual_result.reviewed_controls
    expect_reviewed_controls = expect_result.reviewed_controls
    assert expect_reviewed_controls == actual_reviewed_controls

    actual_observations = actual_result.observations
    expect_observations = expect_result.observations

    assert len(actual_observations) == len(expect_observations)
    for expect_o in expect_observations:
        actual_o = next(filter(lambda x: x.title == expect_o.title, actual_observations), None)
        assert actual_o != None
        assert expect_o.props == actual_o.props
        for expect_s in expect_o.subjects:
            actual_s = next(filter(lambda x: x.title == expect_s.title, actual_o.subjects), None)
            assert actual_s != None
            assert expect_s.type == actual_s.type
            assert list(filter(lambda x: x.name != 'evaluated-on', expect_s.props)) == list(
                filter(lambda x: x.name != 'evaluated-on', actual_s.props)
            )
    write_test_result_to_file(OUTPUT_PATH, assessment_results.json(exclude_none=True, indent=2))
