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
"""The c2p-posture/v1 model.

Control status is a *field* here rather than something a template computes, which is the
point of this model: a posture that only exists inside rendered markdown cannot be diffed,
served or re-rendered.

Serialization goes through C2PBaseModel.serialize_json_bytes, which is
`dict(by_alias=True, exclude_none=True)` -- so an unassessed check omits `checkId` rather
than carrying a null discriminator that every consumer would have to special-case.
"""

from typing import Dict, List, Optional

from pydantic.v1 import Field

from c2p.common.c2p_base_model import C2PBaseModel

SCHEMA_VERSION = 'c2p-posture/v1'

# Control state. OSCAL's ObjectiveStatus allows satisfied | not-satisfied; NOT_ASSESSED is a
# third state this model adds, because "no engine could evaluate this" is neither compliant
# nor non-compliant and must not be rendered as either.
STATE_SATISFIED = 'satisfied'
STATE_NOT_SATISFIED = 'not-satisfied'
STATE_NOT_ASSESSED = 'not-assessed'

# OSCAL ObjectiveStatus.reason is a classifier, not free text: trestle permits exactly these
# three values. Human-readable text belongs in `remarks`.
REASON_PASS = 'pass'
REASON_FAIL = 'fail'
REASON_OTHER = 'other'

# Check result. PASS/FAIL/WARNING/ERROR mirror the engine-neutral evidence vocabulary;
# NOT_ASSESSED marks a check the component definition expects but no evidence covered.
RESULT_PASS = 'pass'
RESULT_FAIL = 'fail'
RESULT_WARNING = 'warning'
RESULT_ERROR = 'error'
RESULT_NOT_ASSESSED = 'not-assessed'


class PostureBaseModel(C2PBaseModel):

    class Config(C2PBaseModel.Config):
        allow_population_by_field_name = True


class Link(PostureBaseModel):
    description: Optional[str]
    href: str


class Subject(PostureBaseModel):
    title: str
    resource_id: Optional[str] = Field(None, alias='resourceId')
    result: str
    reason: Optional[str]
    evaluated_on: Optional[str] = Field(None, alias='evaluatedOn')
    props: Optional[Dict[str, str]]


class Check(PostureBaseModel):
    check_id: Optional[str] = Field(None, alias='checkId')
    engine: Optional[str]
    result: str
    remarks: Optional[str]
    evidence: List[Link] = []
    subjects: List[Subject] = []


class Rule(PostureBaseModel):
    rule_id: str = Field(..., alias='ruleId')
    description: Optional[str]
    checks: List[Check] = []


class Coverage(PostureBaseModel):
    expected_checks: int = Field(..., alias='expectedChecks')
    assessed_checks: int = Field(..., alias='assessedChecks')


class Control(PostureBaseModel):
    control_id: str = Field(..., alias='controlId')
    title: Optional[str]
    state: str
    reason: str
    remarks: Optional[str]
    coverage: Coverage
    evidence: List[Link] = []
    rules: List[Rule] = []


class Component(PostureBaseModel):
    title: str
    controls: List[Control] = []


class Source(PostureBaseModel):
    assessment_results: Optional[str] = Field(None, alias='assessmentResults')
    component_definition: Optional[str] = Field(None, alias='componentDefinition')
    catalog: Optional[str]


class Summary(PostureBaseModel):
    satisfied: int = 0
    not_satisfied: int = Field(0, alias='notSatisfied')
    not_assessed: int = Field(0, alias='notAssessed')


class Posture(PostureBaseModel):
    schema_version: str = Field(SCHEMA_VERSION, alias='schemaVersion')
    framework: Optional[str]
    generated_at: Optional[str] = Field(None, alias='generatedAt')
    source: Optional[Source]
    components: List[Component] = []
    summary: Summary = Summary()
