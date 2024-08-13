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

from enum import Enum
from typing import Dict, List, Literal, Optional, Union

from pydantic.v1 import Field

from c2p.common.c2p_base_model import C2PBaseModel
from c2p.framework.models import PVPResult


class ComplianceType(str, Enum):
    OSCAL = 'oscal'


class ComplianceOscal(C2PBaseModel):
    type: Literal[ComplianceType.OSCAL] = Field(ComplianceType.OSCAL, title='Compliance Type')
    catalog: Optional[str]
    profile: Optional[str]
    component_definition: Optional[str]
    rule_id_column: Optional[str] = Field(
        'Rule_Id',
        title='Column name of Rule Id in component-definition',
    )
    rule_description_column: Optional[str] = Field(
        'Rule_Description',
        title='Column name of Rule Description in component-definition',
    )
    check_id_column: Optional[str] = Field(
        'Check_Id',
        title='Column name of Check Id in component-definition',
    )
    check_description_column: Optional[str] = Field(
        'Check_Description',
        title='Column name of Check Description in component-definition',
    )


class C2PConfig(C2PBaseModel):
    compliance: Optional[Union[ComplianceOscal]]
    pvp_result: Optional[PVPResult]
    pvp_name: Optional[str]
    result_title: Optional[str]
    result_description: Optional[str]
    result_labels: Optional[List[str]] = None
