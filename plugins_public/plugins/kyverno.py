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

import pathlib
import shutil
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional

import yaml
from jinja2 import Template
from pydantic.v1 import Field

from c2p.common.err import C2PError
from c2p.common.logging import getLogger
from c2p.common.utils import get_datetime, get_dict_safely
from c2p.framework.models import Policy, PVPResult, RawResult
from c2p.framework.models.pvp_result import (
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
    'skip': ResultEnum.Error,
    'error': ResultEnum.Error,
}


class PluginConfigKyverno(PluginConfig):
    policy_template_dir: str = Field(..., title='Path to Policy template directory')
    deliverable_policy_dir: str = Field(..., title='Path to deliverable (generated) policy directory')


class PluginKyverno(PluginSpec):

    def __init__(self, config: Optional[PluginConfigKyverno] = None) -> None:
        super().__init__()
        self.config = config

    def generate_pvp_result(self, raw_result: RawResult) -> PVPResult:
        pvp_result: PVPResult = PVPResult()
        observations: List[ObservationByCheck] = []

        polrs = list(
            filter(
                lambda x: x['apiVersion'] == 'wgpolicyk8s.io/v1alpha2' and x['kind'] == 'PolicyReport', raw_result.data
            )
        )
        cpolrs = list(
            filter(
                lambda x: x['apiVersion'] == 'wgpolicyk8s.io/v1alpha2' and x['kind'] == 'ClusterPolicyReport',
                raw_result.data,
            )
        )

        results = []
        for polr in polrs:
            for result in polr['results']:
                results.append(result)
        for cpolr in cpolrs:
            for result in cpolr['results']:
                results.append(result)

        policy_names = list(map(lambda x: x['policy'], results))  # policy_name is used as check_id
        policy_names = set(policy_names)

        for policy_name in policy_names:
            observation = ObservationByCheck(check_id=policy_name, methods=['AUTOMATED'], collected=get_datetime())

            results_per_policy = filter(lambda x: x['policy'] == policy_name, results)
            subjects = []
            for rpp in results_per_policy:
                result = rpp['result']
                result = status_dictionary[result] if result in status_dictionary else ResultEnum.Error
                timestamp = get_dict_safely(rpp, ['timestamp', 'seconds'], get_datetime().second)
                evaluated_on = datetime.fromtimestamp(timestamp, tz=timezone.utc)
                message = rpp['message']

                def to_subject(resource):
                    kind = get_dict_safely(resource, 'kind')
                    api_version = get_dict_safely(resource, 'apiVersion', '')
                    name = get_dict_safely(resource, 'name')
                    namespace = get_dict_safely(resource, 'namespace', '(ClusterScope)')
                    uid = get_dict_safely(resource, 'uid')
                    return Subject(
                        title=f'{api_version}/{kind} {name} {namespace}',
                        type='resource',
                        result=result,
                        resource_id=uid,
                        evaluated_on=evaluated_on,
                        reason=message,
                    )

                subjects = subjects + list(map(to_subject, get_dict_safely(rpp, 'resources')))

            observation.subjects = subjects
            observations.append(observation)

        pvp_result.observations_by_check = observations
        return pvp_result

    def generate_pvp_policy(self, policy: Policy):
        rule_sets = policy.rule_sets
        parameters = policy.parameters
        policy_template_dir = self.config.policy_template_dir
        deliverable_policy_dir = self.config.deliverable_policy_dir
        if not pathlib.Path(deliverable_policy_dir).exists():
            logger.info(f"The deliverable policy directory '{deliverable_policy_dir}' is not found. Creating...")
            pathlib.Path(deliverable_policy_dir).mkdir(parents=True)
        else:
            if not pathlib.Path(deliverable_policy_dir).is_dir():
                raise C2PError(f"The deliverable policy directory '{deliverable_policy_dir}' is not directory.")
        for rule_set in rule_sets:
            each_policy_template_dir = pathlib.Path(f'{policy_template_dir}/{rule_set.rule_id}')
            each_deliverable_policy_dir = pathlib.Path(f'{deliverable_policy_dir}/{rule_set.rule_id}')
            shutil.copytree(each_policy_template_dir, each_deliverable_policy_dir, dirs_exist_ok=True)
            contents = each_deliverable_policy_dir.glob('**/*')
            for path in list(contents):
                tp_str = path.open('r').read()
                yamldocs = yaml.safe_load_all(path.open('r'))
                if not self.__is_policy_file(yamldocs):
                    tp = Template(source=tp_str)
                    kv = dict(map(lambda x: (x.id, x.value), parameters))
                    rendered = tp.render(kv)
                    path.write_text(rendered)

    def __is_policy_file(self, yamldocs: List[Dict[str, Any]]) -> bool:
        for yamldoc in yamldocs:
            kind = get_dict_safely(yamldoc, 'kind', None)
            api_version = get_dict_safely(yamldoc, 'apiVersion', None)
            if kind in ['ClusterPolicy', 'Policy'] and api_version == 'kyverno.io/v1':
                return True
        return False
