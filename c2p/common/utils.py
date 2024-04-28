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

import json
import pathlib
import re
from datetime import datetime, timezone
from typing import Any, List, Union

from trestle.oscal.component import ComponentDefinition

from c2p.common import logging

logger = logging.getLogger('common:utils')


class Control:
    def __init__(self, control_id, impl_id, component_id):
        self.control_id = control_id
        self.impl_id = impl_id
        self.component_id = component_id


class ControlList:
    def __init__(self, items: List['Control']):
        self.items = items

    def get_control_ids(self) -> List[str]:
        def custom_sort(key):
            tokens = re.split('(\d+)', key)
            return list(map(lambda x: int(x) if x.isdigit() else x, tokens))

        control_ids = set(map(lambda x: x.control_id, self.items))
        return sorted(list(control_ids), key=custom_sort)


def get_control_list(path: str) -> ControlList:
    cdef = ComponentDefinition.oscal_read(pathlib.Path(path))
    controls: List[Control] = []
    for component in cdef.components:
        for control_impl in component.control_implementations:
            control_impl.uuid
            for impl_req in control_impl.implemented_requirements:
                control = Control(impl_req.control_id, control_impl.uuid, component.uuid)
                controls.append(control)

    return ControlList(controls)


def load_json_as_dict(path: Union[str, pathlib.Path]) -> Any:
    test_te_path: pathlib.Path
    if isinstance(path, str):
        test_te_path = pathlib.Path(path)
    elif isinstance(path, pathlib.Path):
        test_te_path = path
    else:
        return
    fh = test_te_path.open('r', encoding='utf8')
    return json.load(fh)


def get_datetime() -> datetime:
    return datetime.utcnow().replace(microsecond=0).replace(tzinfo=timezone.utc)


def get_dict_safely(d, key: Union[str, List[str]], default=None):
    if isinstance(key, str):
        if d is not None and isinstance(d, dict):
            return d[key] if key in d else default
        else:
            return default
    else:
        if len(key) > 0:
            k = key.pop(0)
            v = get_dict_safely(d, k, default)
            return get_dict_safely(v, key, default)
        else:
            return d


def remove_none(obj):
    if isinstance(obj, (list, tuple, set)):
        return type(obj)(remove_none(x) for x in obj if x is not None)
    elif isinstance(obj, dict):
        return type(obj)((remove_none(k), remove_none(v)) for k, v in obj.items() if k is not None and v is not None)
    else:
        return obj
