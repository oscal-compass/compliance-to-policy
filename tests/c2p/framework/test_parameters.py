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

import copy
import json
import pathlib

from c2p.framework.c2p import C2P
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal
from c2p.framework.oscal_utils import uuid

COMPONENT_DEFINITION_TEST_DATA = pathlib.Path('tests/data/framework/c2p/component-definition.json')


def parameter_props(parameter_id: str):
    remarks = f'rule_set_{parameter_id}'
    return [
        {'name': 'Parameter_Id', 'value': parameter_id, 'remarks': remarks},
        {'name': 'Parameter_Description', 'value': f'description of {parameter_id}', 'remarks': remarks},
        {'name': 'Parameter_Value_Alternatives', 'value': 'a-value', 'remarks': remarks},
    ]


def test_parameters_from_every_non_validation_component(tmp_path):
    """A component definition with more than one non-validation component used to keep only the
    last one's parameters, because _get_parameters assigned inside its loop over components."""
    root = json.load(COMPONENT_DEFINITION_TEST_DATA.open('r'))
    components = root['component-definition']['components']
    service = next(c for c in components if c['type'].lower() != 'validation')
    service['props'] = service.get('props', []) + parameter_props('param_a')

    second = copy.deepcopy(service)
    second['uuid'] = uuid()
    second['title'] = 'A second service'
    second['props'] = [
        p
        for p in second['props']
        if p['name'] not in ['Parameter_Id', 'Parameter_Description', 'Parameter_Value_Alternatives']
    ]
    second['props'] = second['props'] + parameter_props('param_b')
    components.append(second)

    cdef_path = tmp_path / 'component-definition.json'
    cdef_path.open('w').write(json.dumps(root))

    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = cdef_path.as_posix()
    c2p_config.pvp_name = 'OCM'

    parameters = C2P(c2p_config).get_parameters()

    assert sorted(p.id for p in parameters) == ['param_a', 'param_b']


def test_numbered_parameters_are_all_read(tmp_path):
    """A rule carries several parameters by numbering the columns -- Parameter_Id,
    Parameter_Id_2, ... -- and a reader that looks only at the unsuffixed name finds one."""
    root = json.load(COMPONENT_DEFINITION_TEST_DATA.open('r'))
    components = root['component-definition']['components']
    service = next(c for c in components if c['type'].lower() != 'validation')
    remarks = 'rule_set_numbered'
    service['props'] = service.get('props', []) + [
        {'name': 'Parameter_Id', 'value': 'first', 'remarks': remarks},
        {'name': 'Parameter_Value_Alternatives', 'value': 'one', 'remarks': remarks},
        {'name': 'Parameter_Id_2', 'value': 'second', 'remarks': remarks},
        {'name': 'Parameter_Value_Alternatives_2', 'value': 'two', 'remarks': remarks},
    ]

    cdef_path = tmp_path / 'component-definition.json'
    cdef_path.open('w').write(json.dumps(root))
    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = cdef_path.as_posix()
    c2p_config.pvp_name = 'OCM'

    parameters = {p.id: p.value for p in C2P(c2p_config).get_parameters()}

    assert parameters == {'first': 'one', 'second': 'two'}


def test_parameter_values_come_from_set_parameters(tmp_path):
    """trestle's csv-to-oscal-cd puts the resolved value in `set-parameters` and leaves only
    the id and description in the props, so props alone yield parameters with no value."""
    root = json.load(COMPONENT_DEFINITION_TEST_DATA.open('r'))
    components = root['component-definition']['components']
    service = next(c for c in components if c['type'].lower() != 'validation')
    service['props'] = service.get('props', []) + [
        {'name': 'Parameter_Id', 'value': 'required_labels', 'remarks': 'rule_set_resolved'},
        {'name': 'Parameter_Description', 'value': 'labels', 'remarks': 'rule_set_resolved'},
    ]
    service['control-implementations'][0]['set-parameters'] = [{'param-id': 'required_labels', 'values': ['a b c']}]

    cdef_path = tmp_path / 'component-definition.json'
    cdef_path.open('w').write(json.dumps(root))
    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = cdef_path.as_posix()
    c2p_config.pvp_name = 'OCM'

    parameters = {p.id: p.value for p in C2P(c2p_config).get_parameters()}

    assert parameters['required_labels'] == 'a b c'


def test_observations_name_the_check_as_well_as_the_rule():
    """One rule is routinely implemented by several PVPs. Without the check id an Assessment
    Results cannot say which engine produced an observation, so anything reading it across
    engines -- a posture, for instance -- cannot attribute the result to a check at all."""
    import json as _json

    from trestle.oscal.assessment_results import AssessmentResults

    from c2p.framework.models.pvp_result import PVPResult

    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = COMPONENT_DEFINITION_TEST_DATA.as_posix()
    c2p_config.pvp_name = 'OCM'
    c2p_config.result_title = 'TEST'
    c2p_config.result_description = 'TEST'
    c2p_config.pvp_result = PVPResult.parse_obj(
        _json.load(pathlib.Path('tests/data/framework/c2p/pvp-result.json').open('r'))
    )

    assessment_results = C2P(c2p_config).result_to_oscal()

    observation = assessment_results.results[0].observations[0]
    names = {prop.name: prop.value for prop in observation.props}
    assert 'assessment-rule-id' in names
    assert names['assessment-check-id'] == 'policy-high-scan'
