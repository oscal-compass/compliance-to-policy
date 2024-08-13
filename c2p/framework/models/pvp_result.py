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

from datetime import datetime
from enum import Enum
from typing import List, Optional

from pydantic.v1 import Field

from c2p.common.c2p_base_model import C2PBaseModel


class ResultEnum(str, Enum):
    Pass = 'pass'
    Failure = 'failure'
    Error = 'error'


class Property(C2PBaseModel):
    """
    An attribute, characteristic, or quality of the containing object expressed as a namespace qualified name/value pair. The value of a property is a simple scalar value, which may be expressed as a list of values.
    """

    name: str = Field(
        ...,
        description="A textual label that uniquely identifies a specific attribute, characteristic, or quality of the property's containing object.",
        title='Property Name',
    )
    value: str = Field(
        ...,
        description='Indicates the value of the attribute, characteristic, or quality.',
        title='Property Value',
    )


class Link(C2PBaseModel):
    """
    A reference to a local or remote resource
    """

    description: str = Field(
        ...,
        description='A human-readable description of this evidence.',
        title='Relevant Evidence Description',
    )
    href: str = Field(
        ...,
        description='A resolvable URL reference to relevant evidence.',
        title='Relevant Evidence Reference',
    )


class Subject(C2PBaseModel):
    """
    A human-oriented identifier reference to a resource. Use type to indicate whether the identified resource is a component, inventory item, location, user, or something else.
    """

    title: str = Field(title='Name of the object')
    type: str = Field(
        ...,
        title='Subject Universally Unique Identifier Reference Type',
    )
    resource_id: str = Field(..., title='Subject Universally Unique Identifier Reference')
    result: ResultEnum = Field(..., title='Assessment result')
    evaluated_on: Optional[datetime] = Field(
        None,
        title='Evaluated data/time',
        description='The date and time the subject was evaluated. If not given, observations_by_check.collected is used.',
    )
    reason: Optional[str] = Field(None, title='Reason')
    props: Optional[List[Property]] = Field(None)


class ObservationByCheck(C2PBaseModel):
    """
    Describes an individual observation based on each Check_Id defined in Component Definition.
    """

    title: Optional[str] = Field(
        None,
        description='The title for this observation for the check item. If not given, check id is used.',
        title='Observation Title',
    )
    description: Optional[str] = Field(
        None,
        description='A human-readable description of this assessment observation. If not given, check description is used.',
        title='Observation Description',
    )
    check_id: str = Field(..., description='Check_Id', title='Check_Id')
    methods: List[str] = Field(
        ...,
        description='Identifies how the observation was made.',
        title='Observation Method',
        example=['TEST-AUTOMATED'],
    )
    subjects: Optional[List[Subject]] = Field(None)
    collected: datetime = Field(
        ...,
        description='The date and time identifying when the finding information was collected.',
        title='Collected date/time',
    )
    relevant_evidences: Optional[List[Link]] = Field(None)
    props: Optional[List[Property]] = Field(None)


class PVPResult(C2PBaseModel):
    observations_by_check: Optional[List[ObservationByCheck]] = Field(None)
    links: Optional[List[Link]] = Field(None)


def set_defaults(pvp_result: PVPResult) -> PVPResult:
    for observation in pvp_result.observations_by_check:
        if observation.description == None:
            observation.title = observation.check_id
            observation.description = observation.check_id
        for subject in observation.subjects:
            if subject.evaluated_on == None:
                subject.evaluated_on = observation.collected
    return pvp_result
