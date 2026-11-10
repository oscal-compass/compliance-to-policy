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
"""The posture producer is built in code rather than from fixtures.

The cases that matter here -- a check nothing reported, one rule assessed by several PVPs, a
validation component with no Check_Description -- do not exist as committed artifacts, and
four near-identical Assessment Results on disk would hide what each test is actually about.
"""

import json
from typing import List, Optional

from trestle.oscal.assessment_results import AssessmentResults, Observation, Result
from trestle.oscal.common import Metadata, Property, SubjectReference
from trestle.oscal.component import (
    ComponentDefinition,
    ControlImplementation,
    DefinedComponent,
    ImplementedRequirement,
)

from c2p.framework import oscal_utils
from c2p.tools.posture import posture, render
from c2p.tools.posture.model import (
    RESULT_FAIL,
    RESULT_NOT_ASSESSED,
    RESULT_PASS,
    STATE_NOT_ASSESSED,
    STATE_NOT_SATISFIED,
    STATE_SATISFIED,
)

NOW = oscal_utils.get_datetime()


def metadata() -> Metadata:
    return Metadata(title='test', last_modified=NOW, version='0.0.1', oscal_version='1.1.2')


def rule_set_props(rule_id: str, check_id: str, check_description: Optional[str] = 'a check') -> List[Property]:
    """Props grouped by a shared remark, which is how c2p reads a rule set out of a component."""
    remarks = f'rule_set_{check_id}'
    props = [
        Property(name='Rule_Id', value=rule_id, remarks=remarks),
        Property(name='Rule_Description', value='a rule', remarks=remarks),
        Property(name='Check_Id', value=check_id, remarks=remarks),
    ]
    if check_description != None:
        props.append(Property(name='Check_Description', value=check_description, remarks=remarks))
    return props


def component_definition(engines: List[dict], control_id: str = 'ac-6', rule_id: str = 'a-rule') -> ComponentDefinition:
    """One service component implementing `control_id` via `rule_id`, plus one validation
    component per entry in `engines` -- the three-PVPs-per-rule shape."""
    service = DefinedComponent(
        uuid=oscal_utils.uuid(),
        type='service',
        title='target service',
        description='the system under assessment',
        control_implementations=[
            ControlImplementation(
                uuid=oscal_utils.uuid(),
                source='catalog.json',
                description='implementation',
                implemented_requirements=[
                    ImplementedRequirement(
                        uuid=oscal_utils.uuid(),
                        control_id=control_id,
                        description='requirement',
                        props=[Property(name='Rule_Id', value=rule_id)],
                    )
                ],
            )
        ],
    )
    validations = [
        DefinedComponent(
            uuid=oscal_utils.uuid(),
            type='validation',
            title=engine['title'],
            description='a policy validation point',
            props=rule_set_props(rule_id, engine['check_id'], engine.get('check_description', 'a check')),
        )
        for engine in engines
    ]
    return ComponentDefinition(uuid=oscal_utils.uuid(), metadata=metadata(), components=[service] + validations)


def assessment_results(observations: List[Observation]) -> AssessmentResults:
    result = Result(
        uuid=oscal_utils.uuid(),
        title='test',
        description='test',
        start=NOW,
        reviewed_controls={'control-selections': [{}]},
        observations=observations if observations != [] else None,
    )
    return AssessmentResults(
        uuid=oscal_utils.uuid(), metadata=metadata(), import_ap={'href': 'ap.json'}, results=[result]
    )


def observation(result: str, check_id: Optional[str] = None, rule_id: Optional[str] = None) -> Observation:
    props = []
    if rule_id != None:
        props.append(Property(name='assessment-rule-id', value=rule_id))
    if check_id != None:
        props.append(Property(name='assessment-check-id', value=check_id))
    return Observation(
        uuid=oscal_utils.uuid(),
        description='an observation',
        methods=['TEST-AUTOMATED'],
        collected=NOW,
        props=props,
        subjects=[
            SubjectReference(
                subject_uuid=oscal_utils.uuid(),
                type='resource',
                title='Deployment storefront/api',
                props=[
                    Property(name='result', value=result),
                    Property(name='reason', value='because'),
                    Property(name='resource-id', value='apps/v1/Deployment/storefront/api'),
                ],
            )
        ],
    )


def only_control(result):
    return result.components[0].controls[0]


def test_unassessed_check_is_reported_rather_than_omitted():
    """The whole point of the model: a control no engine could evaluate is neither satisfied
    nor not-satisfied, and the check that produced nothing still occupies a row."""
    cdef = component_definition([{'title': 'kyverno-cel', 'check_id': 'c1'}, {'title': 'vap-cel', 'check_id': 'c2'}])
    ar = assessment_results([observation(RESULT_PASS, check_id='c1')])

    control = only_control(posture.generate(ar, cdef))

    assert control.state == STATE_SATISFIED
    assert control.coverage.expected_checks == 2
    assert control.coverage.assessed_checks == 1
    results = {c.engine: c.result for c in control.rules[0].checks}
    assert results == {'kyverno-cel': RESULT_PASS, 'vap-cel': RESULT_NOT_ASSESSED}


def test_unassessed_check_omits_check_id_rather_than_nulling_it():
    cdef = component_definition([{'title': 'vap-cel', 'check_id': 'c2'}])
    ar = assessment_results([])

    serialized = json.loads(render.render_json(posture.generate(ar, cdef)))

    check = serialized['components'][0]['controls'][0]['rules'][0]['checks'][0]
    assert 'checkId' not in check
    assert check['result'] == RESULT_NOT_ASSESSED


def test_control_with_no_evidence_at_all_is_not_assessed():
    cdef = component_definition([{'title': 'vap-cel', 'check_id': 'c2'}])

    result = posture.generate(assessment_results([]), cdef)

    assert only_control(result).state == STATE_NOT_ASSESSED
    assert result.summary.not_assessed == 1
    assert result.summary.satisfied == 0


def test_failing_subject_makes_the_control_not_satisfied():
    cdef = component_definition([{'title': 'kyverno-cel', 'check_id': 'c1'}])
    ar = assessment_results([observation(RESULT_FAIL, check_id='c1')])

    control = only_control(posture.generate(ar, cdef))

    assert control.state == STATE_NOT_SATISFIED
    assert control.reason == 'fail'


def test_both_result_vocabularies_normalize():
    """This project emits `failure` where other producers emit `fail`; a posture must not render one of
    them as an unknown state."""
    cdef = component_definition([{'title': 'kyverno-cel', 'check_id': 'c1'}])

    for emitted in ['failure', 'fail']:
        ar = assessment_results([observation(emitted, check_id='c1')])
        control = only_control(posture.generate(ar, cdef))
        assert control.rules[0].checks[0].subjects[0].result == RESULT_FAIL, emitted
        assert control.state == STATE_NOT_SATISFIED, emitted


def test_observation_without_check_id_is_attributed_when_the_rule_has_one_check():
    """Python c2p stamps only `assessment-rule-id`. With a single check there is no ambiguity."""
    cdef = component_definition([{'title': 'kyverno-cel', 'check_id': 'c1'}])
    ar = assessment_results([observation(RESULT_PASS, rule_id='a-rule')])

    control = only_control(posture.generate(ar, cdef))

    assert control.coverage.assessed_checks == 1
    assert control.state == STATE_SATISFIED


def test_observation_without_check_id_is_not_guessed_across_several_checks():
    """Attributing a rule-keyed observation to one of three engines would invent coverage the
    evidence does not support, so every check stays unassessed instead."""
    cdef = component_definition(
        [
            {'title': 'kyverno-dsl', 'check_id': 'c1'},
            {'title': 'kyverno-cel', 'check_id': 'c2'},
            {'title': 'vap-cel', 'check_id': 'c3'},
        ]
    )
    ar = assessment_results([observation(RESULT_PASS, rule_id='a-rule')])

    control = only_control(posture.generate(ar, cdef))

    assert control.coverage.assessed_checks == 0
    assert control.state == STATE_NOT_ASSESSED


def test_missing_check_description_does_not_raise():
    """`c2p/tools/viewer/viewer.py` indexes Check_Description unguarded and raises KeyError on
    a component definition that omits it, which other producers' test data does."""
    cdef = component_definition([{'title': 'vap-cel', 'check_id': 'c2', 'check_description': None}])

    result = posture.generate(assessment_results([observation(RESULT_PASS, check_id='c2')]), cdef)

    assert only_control(result).state == STATE_SATISFIED


def test_markdown_renders_every_control_including_unassessed_ones():
    cdef = component_definition([{'title': 'vap-cel', 'check_id': 'c2'}])

    rendered = render.render_markdown(posture.generate(assessment_results([]), cdef))

    assert 'ac-6' in rendered
    assert STATE_NOT_ASSESSED in rendered


def test_html_view_carries_every_subject_and_its_reason():
    """A colour per control is not auditable: the engine's own message is the part a reader has
    to be able to disagree with."""
    cdef = component_definition([{'title': 'kyverno-cel', 'check_id': 'c1'}])
    ar = assessment_results([observation(RESULT_FAIL, check_id='c1')])

    rendered = render.render_html(posture.generate(ar, cdef))

    assert 'ac-6' in rendered
    assert 'Deployment storefront/api' in rendered
    assert 'because' in rendered


def test_several_assessment_results_combine_into_one_posture():
    """A c2p run assesses one PVP at a time, so several engines mean several documents."""
    cdef = component_definition([{'title': 'kyverno-cel', 'check_id': 'c1'}, {'title': 'vap-cel', 'check_id': 'c2'}])
    from_cel = assessment_results([observation(RESULT_PASS, check_id='c1')])
    from_vap = assessment_results([observation(RESULT_PASS, check_id='c2')])

    control = only_control(posture.generate([from_cel, from_vap], cdef))

    assert control.coverage.assessed_checks == 2
    assert control.state == STATE_SATISFIED
