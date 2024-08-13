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

from pydantic.v1 import Field

from c2p.common.c2p_base_model import C2PBaseModel


class RuleSet(C2PBaseModel):
    rule_id: str = Field(
        ...,
        title='A unique identifier of a policy (desired state)',
    )
    rule_description: Optional[str] = Field(title='Rule description')
    check_id: str = Field(
        ...,
        title='A unique identifier used to reference the result of the policy (desired state)',
    )
    check_description: Optional[str]
    raw: Optional[Dict[str, str]]


class Parameter(C2PBaseModel):
    id: str = Field(
        ...,
        title='A unique identifier of a parameter that can be used while PVP Policy generation',
    )
    description: Optional[str]
    value: str = Field(
        ...,
        title='The value of the parameter',
    )


class Policy(C2PBaseModel):
    rule_sets: List[RuleSet] = Field(None)
    parameters: List[Parameter] = Field(None)
