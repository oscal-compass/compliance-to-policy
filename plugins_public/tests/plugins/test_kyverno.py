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

import os
import pathlib

from c2p.framework.models import Parameter, Policy, RawResult, RuleSet
from plugins_public.plugins.kyverno import PluginConfigKyverno, PluginKyverno
from plugins_public.tests.utils import load_yaml, load_yamls

TEST_DATA_DIR = os.getenv('TEST_DATA_DIR', 'plugins_public/tests/data/kyverno')


def test_kyverno_pvp_result_to_compliance():
    cpolr = load_yaml(f'{TEST_DATA_DIR}/clusterpolicyreports.wgpolicyk8s.io.yaml')
    polr = load_yaml(f'{TEST_DATA_DIR}/policyreports.wgpolicyk8s.io.yaml')
    raw = cpolr['items'] + polr['items']
    raw_result = RawResult(data=raw)
    pvp_result = PluginKyverno().generate_pvp_result(raw_result)
    assert 2 == len(pvp_result.observations_by_check)
    assert 33 == len(pvp_result.observations_by_check[0].subjects)
    assert 33 == len(pvp_result.observations_by_check[1].subjects)


def test_kyverno_compliance_to_policy():
    policy_template_dir = f'{TEST_DATA_DIR}/policy-resources'
    deliverable_policy_dir = f'{TEST_DATA_DIR}/deliverable-policy'
    config = PluginConfigKyverno(policy_template_dir=policy_template_dir, deliverable_policy_dir=deliverable_policy_dir)
    rule_sets = [
        RuleSet(rule_id='allowed-base-images', check_id=''),
        RuleSet(rule_id='disallow-capabilities', check_id=''),
    ]
    parameters = [Parameter(id='allowed_baseimages', value='gcr.io/distroless/static:root')]
    policy = Policy(rule_sets=rule_sets, parameters=parameters)

    PluginKyverno(config).generate_pvp_policy(policy)

    policy_template_dir = pathlib.Path(policy_template_dir)
    deliverable_policy_dir = pathlib.Path(deliverable_policy_dir)
    policy_dirs = filter(lambda x: x.is_dir(), deliverable_policy_dir.iterdir())
    assert set(['disallow-capabilities', 'allowed-base-images']) == set(map(lambda x: x.name, policy_dirs))

    # disallow-capabilities
    policy_dir = deliverable_policy_dir / 'disallow-capabilities'
    assert set(['disallow-capabilities.yaml']) == set(map(lambda x: x.name, policy_dir.iterdir()))

    policy = load_yaml(policy_dir / 'disallow-capabilities.yaml')
    expected = load_yaml(policy_template_dir / 'disallow-capabilities' / 'disallow-capabilities.yaml')
    assert expected == policy

    # allowed-base-images
    policy_dir = deliverable_policy_dir / 'allowed-base-images'
    assert set(['allowed-base-images.yaml', '02-setup-cm.yaml']) == set(map(lambda x: x.name, policy_dir.iterdir()))

    setup_yaml_path = policy_dir / '02-setup-cm.yaml'
    setup_yamls = load_yamls(setup_yaml_path)
    configmap = next(filter(lambda x: x['kind'] == 'ConfigMap', setup_yamls), None)
    assert 'gcr.io/distroless/static:root' == configmap['data']['allowedbaseimages']

    policy_path = policy_dir / 'allowed-base-images.yaml'
    policy = load_yaml(policy_path)
    expected = load_yaml(policy_template_dir / 'allowed-base-images' / 'allowed-base-images.yaml')
    assert expected == policy
