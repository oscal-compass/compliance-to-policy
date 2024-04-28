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
import shutil
from datetime import datetime
from typing import Any, Dict, List, Optional, TypeVar

import yaml
from pydantic import BaseModel, Field

from c2p.common.err import C2PError
from c2p.common.logging import getLogger
from c2p.common.utils import get_datetime, get_dict_safely, remove_none
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
    'Compliant': ResultEnum.Pass,
    'NonCompliant': ResultEnum.Failure,
}

ANNOTATION_COMPONENT_TITLE = "compliance-to-policy.component-title"


class Manifest(BaseModel):
    remediationAction: Optional[str] = None
    severity: Optional[str] = None
    complianceType: Optional[str] = None
    metadataComplianceType: Optional[str] = None
    evaluationInterval: Optional[Dict[str, Any]] = None
    namespaceSelector: Optional[Dict[str, Any]] = None
    pruneObjectBehavior: Optional[str] = None
    patches: Optional[Dict[str, Any]] = None
    path: Optional[str] = None
    extraDependencies: Optional[List[Dict[str, str]]] = None
    ignorePending: Optional[bool] = False


class PolicyConfig(BaseModel):
    name: str
    manifests: Optional[List[Manifest]] = []
    standards: Optional[List[str]] = []
    controls: Optional[List[str]] = []
    categories: Optional[List[str]] = []
    consolidateManifests: Optional[bool] = True
    orderManifests: Optional[bool] = False
    informGatekeeperPolicies: Optional[bool] = False
    informKyvernoPolicies: Optional[bool] = False
    remediationAction: Optional[str] = 'inform'
    severity: Optional[str] = 'high'
    complianceType: Optional[str] = 'mustnothave'


class PluginConfigOCM(PluginConfig):
    policy_template_dir: str = Field(..., title='Path to Policy template directory')
    deliverable_policy_dir: str = Field(..., title='Path to deliverable (generated) policy directory')
    namespace: str = Field(..., title='Namespace in OCM Hub to which policies are delivered')
    paremeters_configmap_name: str = Field('c2p-parameters', title='Name of configmap for parameters')
    cluster_selectors: Dict[str, str] = Field(
        ..., title='Pair of cluster label name and value to which policies are distributed to matched clusters'
    )
    policy_set_name: str = 'test'
    exclude_namespaces: List[str] = Field(
        [
            'kube-system',
            'open-cluster-management',
            'open-cluster-management-agent',
            'open-cluster-management-agent-addon',
        ],
        title='Namespaces that policy is not applied',
    )
    include_namespaces: List[str] = Field(['*'], title='Namespaces that policy must be applied')


class PluginOCM(PluginSpec):

    def __init__(self, config: Optional[PluginConfigOCM] = None) -> None:
        super().__init__()
        self.config = config

    def generate_pvp_result(self, raw_result: RawResult) -> PVPResult:
        pvp_result: PVPResult = PVPResult()
        observations: List[ObservationByCheck] = []

        policies = list(
            filter(
                lambda x: x['apiVersion'] == 'policy.open-cluster-management.io/v1' and x['kind'] == 'Policy',
                raw_result.data,
            )
        )

        # Root policy resource on Hub
        root_policies = list(
            filter(
                lambda x: get_dict_safely(x, ['metadata', 'labels', 'policy.open-cluster-management.io/cluster-name'])
                == None,
                policies,
            )
        )

        # Policy resources of each cluster to which the root policies are delivered
        each_policies = list(
            filter(
                lambda x: get_dict_safely(x, ['metadata', 'labels', 'policy.open-cluster-management.io/cluster-name'])
                != None,
                policies,
            )
        )

        # policy_name is used as check_id
        policy_namespace_names = list(map(lambda x: (x['metadata']['name'], x['metadata']['namespace']), root_policies))

        for policy_name, root_namespace in policy_namespace_names:
            observation = ObservationByCheck(check_id=policy_name, methods=['AUTOMATED'], collected=get_datetime())

            results_per_policy = filter(
                lambda x: x['metadata']['name'] == f'{root_namespace}.{policy_name}', each_policies
            )
            subjects = []
            for rpp in results_per_policy:
                name = get_dict_safely(rpp, ['metadata', 'name'])
                cluster_name = get_dict_safely(rpp, ['metadata', 'namespace'])
                result = get_dict_safely(rpp, ['status', 'compliant'])
                result = status_dictionary[result] if result in status_dictionary else ResultEnum.Error
                details = get_dict_safely(rpp, ['status', 'details'])
                if isinstance(details, list) and len(details) > 0:
                    for detail in details:
                        history = detail['history']
                        if isinstance(history, list) and len(history) > 0:
                            latest_history = history[0]
                            event_name = latest_history['eventName']
                            last_timestamp = latest_history['lastTimestamp']
                            message = latest_history['message']
                else:
                    logger.warn(f'"details" are not found for name "{name}" for "{cluster_name}"')

                evaluated_on = (
                    datetime.fromisoformat(last_timestamp.replace('Z', '+00:00'))
                    if last_timestamp != None
                    else get_datetime()
                )

                event_name = event_name if event_name != None else ''
                message = message if message != None else ''
                reason = f'[{event_name}] {message}' if event_name != '' and message != '' else None
                subject = Subject(
                    title=f'Cluster "{cluster_name}"',
                    type='cluster',
                    result=result,
                    resource_id=cluster_name,
                    evaluated_on=evaluated_on,
                    reason=reason,
                )
                subjects.append(subject)

            observation.subjects = subjects
            observations.append(observation)

        pvp_result.observations_by_check = observations
        return pvp_result

    def generate_pvp_policy(self, policy: Policy):
        rule_sets = policy.rule_sets
        parameters = policy.parameters
        policy_template_dir = pathlib.Path(self.config.policy_template_dir)
        deliverable_policy_dir = pathlib.Path(self.config.deliverable_policy_dir)
        if not deliverable_policy_dir.exists():
            logger.info(
                f"The deliverable policy directory '{deliverable_policy_dir.as_posix()}' is not found. Creating..."
            )
            deliverable_policy_dir.mkdir(parents=True)
        else:
            if not deliverable_policy_dir.is_dir():
                raise C2PError(
                    f"The deliverable policy directory '{deliverable_policy_dir.as_posix()}' is not directory."
                )
        policy_config_map: Dict[str, PolicyConfig] = {}
        for rule_set in rule_sets:
            policy_id = rule_set.rule_id
            each_policy_template_dir = policy_template_dir / policy_id
            each_deliverable_policy_dir = deliverable_policy_dir / policy_id
            shutil.copytree(each_policy_template_dir, each_deliverable_policy_dir, dirs_exist_ok=True)
            standards = []
            controls = []
            categories = []
            policy_generator_path = each_deliverable_policy_dir / 'policy-generator.yaml'
            policy_generator = yaml.safe_load(policy_generator_path.open('r'))
            policy_generator['policyDefaults']['namespace'] = self.config.namespace
            policy_generator['policyDefaults']['standards'] = standards
            policy_generator['policyDefaults']['controls'] = controls
            policy_generator['policyDefaults']['categories'] = categories
            policy_generator['policyDefaults']['placement'] = {'clusterSelectors': self.config.cluster_selectors}
            yaml.safe_dump(policy_generator, policy_generator_path.open('w'))

            pg_policy = policy_generator['policies'][0]
            if policy_id in policy_config_map:
                policy_config = policy_config_map[policy_id]
            else:
                policy_config = PolicyConfig.parse_obj(pg_policy)
                for idx, m in enumerate(policy_config.manifests):
                    policy_config.manifests[idx].path = m.path.replace('./', f'./{policy_id}/')
            policy_config.standards = self.__merge_uniquely(policy_config.standards, standards)
            policy_config.categories = self.__merge_uniquely(policy_config.controls, categories)
            policy_config.controls = self.__merge_uniquely(policy_config.categories, controls)
            policy_config_map[policy_id] = policy_config

        policy_set_name_sanitized = self.config.policy_set_name.lower().replace(' ', '-')  # DNS Compliant value
        policy_set = {
            'name': policy_set_name_sanitized,
            'policies': list(policy_config_map.keys()),
        }

        policy_set_generator = {
            'apiVersion': 'policy.open-cluster-management.io/v1',
            'kind': 'PolicyGenerator',
            'metadata': {'name': 'policy-set'},
            'placementBindingDefaults': {'name': 'policy-set'},
            'policyDefaults': {
                'placement': {'labelSelector': self.config.cluster_selectors},
                'consolidateManifests': False,
                'orderManifests': False,
                'informGatekeeperPolicies': False,
                'informKyvernoPolicies': False,
                'namespaceSelector': {
                    'exclude': self.config.exclude_namespaces,
                    'include': self.config.include_namespaces,
                },
                'namespace': self.config.namespace,
            },
            'policySetDefaults': {'placement': {'labelSelector': self.config.cluster_selectors}},
            'policies': list(map(lambda x: x.dict(), policy_config_map.values())),
            'policySets': [policy_set],
        }

        parameters_configmap = {
            'apiVersion': 'v1',
            'kind': 'ConfigMap',
            'metadata': {'name': self.config.paremeters_configmap_name, 'namespace': self.config.namespace},
            'data': dict(map(lambda x: (x.id, x.value), parameters)),
        }

        kustomize_patch = {
            'target': {'kind': 'PolicySet', 'name': policy_set_name_sanitized},
            'patch': json.dumps(
                [
                    {
                        'op': 'replace',
                        'path': f'/metadata/annotations/{ANNOTATION_COMPONENT_TITLE}',
                        'value': self.config.policy_set_name,
                    }
                ]
            ),
        }

        kustomize = {
            'generators': ['./policy-generator.yaml'],
            'patches': [kustomize_patch],
            'resources': ['./parameters.yaml'],
        }

        yaml.safe_dump(remove_none(policy_set_generator), (deliverable_policy_dir / 'policy-generator.yaml').open('w'))
        yaml.safe_dump(parameters_configmap, (deliverable_policy_dir / 'parameters.yaml').open('w'))
        yaml.safe_dump(kustomize, (deliverable_policy_dir / 'kustomization.yaml').open('w'))

    T = TypeVar('T')

    def __merge_uniquely(self, targets: Optional[List[T]], value: List[T]) -> List[T]:
        x = set(targets) if targets != None else set()
        for _ in value:
            x.add(value)
        return list(set(x))
