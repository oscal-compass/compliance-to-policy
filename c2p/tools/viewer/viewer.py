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

from typing import Dict, List, Optional

from jinja2 import Template
from pydantic.v1 import BaseModel
from trestle.oscal.assessment_results import AssessmentResults, Observation
from trestle.oscal.common import Property
from trestle.oscal.component import ComponentDefinition, DefinedComponent

from c2p.common.oscal import is_component_type_validation
from c2p.framework.oscal_utils import group_props_by_remarks
from c2p.tools.viewer import TEMPLATE


class SubjectResult(BaseModel):
    uuid: str
    title: str
    result: str
    reason: str


class RuleResult(BaseModel):
    id: str
    description: str
    subjects: List[SubjectResult] = []


class ControlResult(BaseModel):
    id: str
    rule_results: List[RuleResult] = []


class RenderedComponent(BaseModel):
    title: str
    control_results: List[ControlResult] = []


def find_observation(observations: List[Observation], check_id) -> Optional[Observation]:
    for observation in observations:
        for prop in observation.props:
            if prop.name == 'assessment-rule-id' and prop.value == check_id:
                return observation
    return None


def get_prop_value(props: List[Property], name):
    p = next(filter(lambda x: x.name == name, props), None)
    return p.value if p != None else None


def get_pass_fail_icon(result):
    if result == 'pass':
        return ':white_check_mark:'
    elif result == 'failure':
        return ':x:'
    else:
        return ':warning:'


def render(assessment_results: AssessmentResults, component_definition: ComponentDefinition) -> str:
    rule_sets_map: Dict[str, List[Dict[str, str]]] = {}
    for component in component_definition.components:
        if is_component_type_validation(component.type):
            rule_sets_map[component.title] = group_props_by_remarks(component)

    components: List[DefinedComponent] = list(
        filter(lambda x: not is_component_type_validation(x.type), component_definition.components)
    )

    def get_pvp_rule_pair(rule_id):
        for pvp, rule_sets in rule_sets_map.items():
            for rule in rule_sets:
                if rule['Rule_Id'] == rule_id:
                    return (pvp, rule)
        return None, None

    render_components = []
    for component in components:
        rendered_component = RenderedComponent(title=component.title)
        for control_imple in component.control_implementations:
            for imple_req in control_imple.implemented_requirements:
                control_id = imple_req.control_id
                control_result = ControlResult(id=control_id)
                for prop in filter(lambda x: x.name == 'Rule_Id', imple_req.props):
                    rule_id = prop.value
                    pvp, rule_set = get_pvp_rule_pair(rule_id)
                    if rule_set != None:
                        rule_result = RuleResult(id=f'{rule_id} ({pvp})', description=rule_set['Check_Description'])
                        o = find_observation(assessment_results.results[0].observations, rule_set['Rule_Id'])
                        if o != None:
                            for subject in o.subjects:
                                result = get_prop_value(subject.props, 'result')
                                result = f'{result} {get_pass_fail_icon(result)}'
                                reason = get_prop_value(subject.props, 'reason')
                                sr = SubjectResult(
                                    uuid=subject.subject_uuid, title=subject.title, result=result, reason=reason
                                )
                                rule_result.subjects.append(sr)
                            control_result.rule_results.append(rule_result)
                rendered_component.control_results.append(control_result)
        render_components.append(rendered_component)

    tp = Template(source=TEMPLATE)
    rendered = tp.render(components=render_components)
    return rendered
