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

from typing import Dict, List

from trestle.common import const
from trestle.common.common_types import TypeWithParamId, TypeWithParts, TypeWithProps
from trestle.common.list_utils import as_filtered_list, as_list, none_if_empty


def is_component_type_validation(component_type: str) -> bool:
    return component_type.lower() == 'validation'


def get_rule_sets(item: TypeWithProps) -> List[Dict[str, str]]:
    """Get all rules found in this items props."""
    # rules is dict containing rule_id and description
    rules_dict = {}
    for prop in as_list(item.props):
        remarks = prop.remarks
        if not remarks in rules_dict:
            rules_dict[remarks] = {}
        rules_dict[remarks][prop.name] = prop.value
    return list(map(lambda x: x[1], rules_dict.items()))
