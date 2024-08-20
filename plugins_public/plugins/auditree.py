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

import copy
import json
import re
from datetime import datetime, timezone
from pathlib import Path
from typing import List, Optional, Tuple, Union

from pydantic.v1 import Field
from pydantic.v1.utils import deep_update

from c2p.common.logging import getLogger
from c2p.common.utils import get_dict_safely
from c2p.framework.models import Policy, PVPResult, RawResult
from c2p.framework.models.pvp_result import (
    Link,
    ObservationByCheck,
    PVPResult,
    ResultEnum,
    Subject,
)
from c2p.framework.plugin_spec import PluginConfig, PluginSpec

logger = getLogger(__name__)

status_dictionary = {
    'pass': ResultEnum.Pass,
    'fail': ResultEnum.Failure,
    'warn': ResultEnum.Failure,
    'error': ResultEnum.Error,
}


class PluginConfigAuditree(PluginConfig):
    auditree_json_template: str = Field(..., title='Path to auditree.json template')
    output: str = Field('auditree.json', title='Path to the generated auditree.json (default: ./auditree.json)')


class PluginAuditree(PluginSpec):

    def __init__(self, config: Optional[PluginConfigAuditree] = None) -> None:
        super().__init__()
        self.config = config

    def generate_pvp_policy(self, policy: Policy):
        with Path(self.config.auditree_json_template).open('r') as f:
            auditree_json = json.load(f)
        parameters = policy.parameters
        for parameter in parameters:
            key = parameter.id
            value = get_dict_safely(auditree_json, key.split('.'))
            if value is not None:
                try:
                    if isinstance(value, list):
                        updated = parameter.value.split(',')
                    elif isinstance(value, str):
                        updated = parameter.value
                    elif isinstance(value, int):
                        updated = int(parameter.value)
                    elif isinstance(value, float):
                        updated = float(parameter.value)
                    else:
                        raise Exception(f'Unsupported parameter value format (parameter_id: {key})')
                except Exception as e:
                    raise Exception(f'Invalid parameter value format (parameter_id: {key})') from e
                auditree_json = update_dict(auditree_json, key.split('.'), updated)
        json.dump(auditree_json, Path(self.config.output).open('w'), indent=2)

    def generate_pvp_result(self, raw_result: RawResult) -> PVPResult:
        locker_url = get_dict_safely(raw_result.additional_props, 'locker_url', 'files:///tmp/compliance')
        pvp_result: PVPResult = PVPResult()
        observations: List[ObservationByCheck] = []
        for check_class_name, check_class_result in raw_result.data.items():
            for check_method_name, check_method_result in get_dict_safely(check_class_result, 'checks', []).items():
                check_id = f'{check_class_name}.{check_method_name}'
                timestamp = get_dict_safely(check_method_result, 'timestamp')
                dt = datetime.fromtimestamp(timestamp, tz=timezone.utc)

                observation = ObservationByCheck(
                    check_id=check_id,
                    methods=['AUTOMATED'],
                    collected=dt,
                )

                evidences = get_dict_safely(check_class_result, 'evidence', [])
                relevant_evidences = []
                for evidence in evidences:
                    href = f'{locker_url}/{get_dict_safely(evidence, "path", "")}'
                    description = get_dict_safely(evidence, 'description', '')
                    relevant_evidences.append(Link(description=description, href=href))

                status = get_dict_safely(check_method_result, 'status', 'not_found')
                if status is None:
                    reason = f'Status not found for this check {check_id}.'
                    status = ResultEnum.Error

                def generate_reason(status) -> str:
                    successes = get_dict_safely(check_method_result, 'successes', {})
                    warnings = get_dict_safely(check_method_result, 'warnings', {})
                    failures = get_dict_safely(check_method_result, 'failures', {})
                    exception = get_dict_safely(check_method_result, 'exception', {})
                    res = {}
                    if status == 'pass':
                        res = successes
                    elif status == 'warn':
                        res = warnings
                    elif status == 'fail':
                        res = failures
                    elif status == 'error':
                        res = exception
                    else:
                        res = successes
                        res.update(warnings)
                        res.update(failures)
                        res.update({'exception': exception} if exception != '' else {})
                    return f'{res}'

                subject = Subject(
                    title=f'Auditree Check: {check_id}',
                    type='inventory-item',
                    result=status_dictionary[status] if status in status_dictionary else ResultEnum.Error,
                    resource_id=check_id,
                    evaluated_on=dt,
                    reason=generate_reason(status),
                )
                observation.subjects = [subject]
                observations.append(observation)

        # merge observations whose check id is generated by parametrized expansion (see parameterized.expand())
        merged_observations: List[ObservationByCheck] = []
        merged_list: List[Tuple[str, ObservationByCheck]] = []

        for observation in observations:
            *classname, parametrized_method = observation.check_id.split('.')
            res = re.search('(.*)_([0-9]+)_(.+)$', parametrized_method)
            if res is not None:
                method = res.group(1)
                normalized_check_id = '.'.join(classname + [method])
                merged_list.append((normalized_check_id, observation))
            else:
                merged_observations.append(observation)

        for normalized_check_id in set([x[0] for x in merged_list]):
            group = [x[1] for x in merged_list if x[0] == normalized_check_id]
            merged_subjects = [subject for x in group for subject in x.subjects]
            observation = ObservationByCheck(
                check_id=normalized_check_id,
                methods=['AUTOMATED'],
                collected=group[0].collected,
                subjects=merged_subjects,
            )
            merged_observations.append(observation)

        pvp_result.observations_by_check = merged_observations

        return pvp_result


def update_dict(d, key: Union[str, List[str]], value):
    if isinstance(key, str):
        data = copy.deepcopy(d)
        data[key] = value
        return data
    else:
        update = {key.pop(): value}
        for _key in reversed(key):
            update = {_key: update}
        return deep_update(d, update)
