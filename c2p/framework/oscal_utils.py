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

from datetime import datetime, timezone
from enum import Enum
from typing import Any, Dict, List, Union
from uuid import uuid4

from pydantic.v1 import BaseModel
from trestle.common.common_types import TypeWithProps
from trestle.common.list_utils import as_list
from trestle.oscal.common import (
    ControlSelection,
    Property,
    ReviewedControls,
    SelectControlById,
)
from trestle.oscal.component import ComponentDefinition

from c2p.common.oscal import is_component_type_validation


def uuid() -> str:
    """Return uuid."""
    return str(uuid4())


def group_props_by_remarks(item: TypeWithProps) -> List[Dict[str, str]]:
    """Group props by remarks and return as dict of [remark, prop]."""
    grouped = {}
    for prop in as_list(item.props):
        remarks = prop.remarks
        if not remarks in grouped:
            grouped[remarks] = {}
        grouped[remarks][prop.name] = prop.value
    return list(map(lambda x: x[1], grouped.items()))


def reviewed_controls(component_definition: ComponentDefinition) -> ReviewedControls:
    """Return reviewed controls."""
    control_selections = []
    for component in component_definition.components:
        if is_component_type_validation(component.type):
            continue
        for control_impl in component.control_implementations:
            selectControls = []
            for impl_req in control_impl.implemented_requirements:
                statement_ids = []
                for stmt in impl_req.statements if impl_req.statements != None else []:
                    statement_ids.append(stmt.statement_id)
                selectControl = SelectControlById(control_id=impl_req.control_id, statement_ids=statement_ids)
                selectControls.append(selectControl)
            control_selections.append(ControlSelection(include_controls=selectControls))
    rval = ReviewedControls(control_selections=control_selections)
    return rval


def add_prop(props: List[Property], name: str, data: Union[str, Dict, BaseModel], keys: List[str]) -> None:
    try:
        if isinstance(data, str):
            value = data
        else:
            if isinstance(data, BaseModel):
                data = data.dict()
            value = get_value(data, keys)
        if value == None:
            return None
        prop = Property(name=normalize(name), value=whitespace(value))
        props.append(prop)
        return prop
    except KeyError:
        return None


def get_value(data: Dict, keys: List[str]) -> Any:
    """Descend yaml layers to get value for order list of keys."""
    try:
        value = data
        for key in keys:
            value = value[key]
        if isinstance(value, Enum):
            value = value.value
        if isinstance(value, datetime):
            value = value.isoformat()
    except KeyError:
        raise KeyError
    return value


def whitespace(text: str) -> str:
    """Replace line ends with blanks."""
    return str(text).replace('\n', ' ')


def normalize(text: str) -> str:
    """Replace slashes with underscores."""
    return text.replace('/', '_')


def get_datetime() -> datetime:
    return datetime.utcnow().replace(microsecond=0).replace(tzinfo=timezone.utc)


def get_datetime_str() -> str:
    return get_datetime().isoformat()
