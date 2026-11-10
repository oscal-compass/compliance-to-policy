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
"""Several engines, one Assessment Results.

A cluster usually runs more than one policy engine, and each is assessed separately. One
document per engine makes the artifact a record of how the assessment was organised rather
than a statement about the system.
"""

import copy
import json
import pathlib

from c2p.framework.c2p import C2P
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal
from c2p.framework.models.pvp_result import (
    ObservationByCheck,
    PVPResult,
    ResultEnum,
    Subject,
)
from c2p.framework.oscal_utils import get_datetime, uuid

COMPONENT_DEFINITION_TEST_DATA = pathlib.Path('tests/data/framework/c2p/component-definition.json')


def a_result(check_id):
    return PVPResult(
        observations_by_check=[
            ObservationByCheck(
                check_id=check_id,
                methods=['TEST-AUTOMATED'],
                collected=get_datetime(),
                subjects=[
                    Subject(
                        title='Cluster "cluster1"',
                        type='resource',
                        resource_id='cluster1',
                        result=ResultEnum.Pass,
                        evaluated_on=get_datetime(),
                        reason='fine',
                    )
                ],
            )
        ]
    )


def two_engine_component_definition(tmp_path):
    """The fixture's single validation component, plus a second one owning another check."""
    root = json.load(COMPONENT_DEFINITION_TEST_DATA.open('r'))
    components = root['component-definition']['components']
    validation = next(c for c in components if c['type'].lower() == 'validation')

    second = copy.deepcopy(validation)
    second['uuid'] = uuid()
    second['title'] = 'Other'
    for prop in second['props']:
        if prop['name'] == 'Check_Id':
            prop['value'] = 'other-check'
    components.append(second)

    path = tmp_path / 'component-definition.json'
    path.open('w').write(json.dumps(root))
    return path


def configured(cdef_path):
    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = cdef_path.as_posix()
    c2p_config.pvp_name = 'OCM'
    c2p_config.result_title = 'TEST'
    c2p_config.result_description = 'TEST'
    return C2P(c2p_config)


def test_two_engines_produce_one_assessment_results(tmp_path):
    c2p = configured(two_engine_component_definition(tmp_path))
    c2p.add_pvp_result(a_result('policy-high-scan'))
    c2p.add_pvp_result(a_result('other-check'))

    assessment_results = c2p.result_to_oscal()

    assert len(assessment_results.results) == 1
    observations = assessment_results.results[0].observations
    assert len(observations) == 2
    check_ids = {
        prop.value for observation in observations for prop in observation.props if prop.name == 'assessment-check-id'
    }
    assert check_ids == {'policy-high-scan', 'other-check'}


def test_aggregating_matches_checks_from_every_validation_component(tmp_path):
    """With a single engine the configured PVP is the only one consulted, so a second engine's
    observation would be dropped as an unknown check."""
    c2p = configured(two_engine_component_definition(tmp_path))
    c2p.add_pvp_result(a_result('other-check'))

    observations = c2p.result_to_oscal().results[0].observations

    # one engine only: still filtered to the configured PVP, so nothing matches 'other-check'
    assert observations == []


def test_the_single_engine_path_is_unchanged(tmp_path):
    c2p = configured(two_engine_component_definition(tmp_path))
    c2p.set_pvp_result(a_result('policy-high-scan'))

    observations = c2p.result_to_oscal().results[0].observations

    assert len(observations) == 1
