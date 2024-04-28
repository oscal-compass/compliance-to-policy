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
import shutil
import tempfile
from distutils.util import strtobool

from c2p.framework.models import Parameter, Policy, RawResult, RuleSet
from plugins_public.plugins.ocm import PluginConfigOCM, PluginOCM
from plugins_public.tests.utils import load_yaml

TEST_DATA_DIR = os.getenv('TEST_DATA_DIR', 'plugins_public/tests/data/ocm')
OVERWRITE_EXPECTED_DATA = os.getenv('OVERWRITE_EXPECTED_DATA', 'false')
KEEP_TEMP_FILE_AND_DIR = os.getenv('KEEP_TEMP_FILE_AND_DIR', 'false')


def test_ocm_pvp_result_to_compliance():
    pds = load_yaml(f'{TEST_DATA_DIR}/placementdecisions.cluster.open-cluster-management.io.yaml')
    policies = load_yaml(f'{TEST_DATA_DIR}/policies.policy.open-cluster-management.io.yaml')
    policy_sets = load_yaml(f'{TEST_DATA_DIR}/policysets.policy.open-cluster-management.io.yaml')
    raw = pds['items'] + policies['items'] + policy_sets['items']
    raw_result = RawResult(data=raw)
    pvp_result = PluginOCM().generate_pvp_result(raw_result)
    assert len(pvp_result.observations_by_check) == 3
    assert len(pvp_result.observations_by_check[0].subjects) == 2


def test_ocm_compliance_to_policy():
    tmpdir = tempfile.mkdtemp()
    policy_template_dir = pathlib.Path(f'{TEST_DATA_DIR}/policy-resources')
    deliverable_policy_dir = pathlib.Path(f'{tmpdir}/deliverable-policy')
    expected_deliverable_policy_dir = pathlib.Path(f'{TEST_DATA_DIR}/deliverable-policy')
    config = PluginConfigOCM(
        policy_template_dir=policy_template_dir.as_posix(),
        deliverable_policy_dir=deliverable_policy_dir.as_posix(),
        namespace='c2p',
        paremeters_configmap_name='c2p-parameters',
        cluster_selectors={'environment': 'dev'},
        policy_set_name='c2p test',
    )
    rule_sets = [
        RuleSet(rule_id='policy-deployment', check_id=''),
        RuleSet(rule_id='policy-disallowed-roles', check_id=''),
        RuleSet(rule_id='policy-high-scan', check_id=''),
    ]
    parameters = [Parameter(id='minimum_nginx_deployment_replicas', value='3')]
    policy = Policy(rule_sets=rule_sets, parameters=parameters)

    PluginOCM(config).generate_pvp_policy(policy)

    policy_dirs = filter(lambda x: x.is_dir(), deliverable_policy_dir.iterdir())
    assert set(['policy-disallowed-roles', 'policy-deployment', 'policy-high-scan']) == set(
        map(lambda x: x.name, policy_dirs)
    )
    assert set(['kustomization.yaml', 'parameters.yaml', 'policy-generator.yaml']) == set(
        map(lambda x: x.name, filter(lambda x1: x1.is_file(), deliverable_policy_dir.iterdir()))
    )

    expected = load_yaml(expected_deliverable_policy_dir / 'policy-generator.yaml')
    actual = load_yaml(deliverable_policy_dir / 'policy-generator.yaml')
    assert expected == actual

    expected = load_yaml(expected_deliverable_policy_dir / 'kustomization.yaml')
    actual = load_yaml(deliverable_policy_dir / 'kustomization.yaml')
    assert expected == actual

    expected = load_yaml(expected_deliverable_policy_dir / 'policy-generator.yaml')
    actual = load_yaml(deliverable_policy_dir / 'policy-generator.yaml')
    assert expected == actual

    # policy-disallowed-roles
    policy_dir = deliverable_policy_dir / 'policy-disallowed-roles'
    assert set(['policy-disallowed-roles-sample-role', 'kustomization.yaml', 'policy-generator.yaml']) == set(
        map(lambda x: x.name, policy_dir.iterdir())
    )
    assert set(['Role.noname.0.yaml']) == set(
        map(lambda x: x.name, (policy_dir / 'policy-disallowed-roles-sample-role').iterdir())
    )

    # policy-deployment
    policy_dir = deliverable_policy_dir / 'policy-deployment'
    assert set(['policy-nginx-deployment', 'kustomization.yaml', 'policy-generator.yaml']) == set(
        map(lambda x: x.name, policy_dir.iterdir())
    )
    assert set(['Deployment.nginx-deployment.0.yaml']) == set(
        map(lambda x: x.name, (policy_dir / 'policy-nginx-deployment').iterdir())
    )

    # policy-high-scan
    policy_dir = deliverable_policy_dir / 'policy-high-scan'
    assert set(
        [
            'compliance-high-scan',
            'compliance-suite-high',
            'compliance-suite-high-results',
            'kustomization.yaml',
            'policy-generator.yaml',
        ]
    ) == set(map(lambda x: x.name, policy_dir.iterdir()))
    assert set(['ScanSettingBinding.high.0.yaml']) == set(
        map(lambda x: x.name, (policy_dir / 'compliance-high-scan').iterdir())
    )
    assert set(['ComplianceSuite.high.0.yaml']) == set(
        map(lambda x: x.name, (policy_dir / 'compliance-suite-high').iterdir())
    )
    assert set(['ComplianceCheckResult.noname.0.yaml']) == set(
        map(lambda x: x.name, (policy_dir / 'compliance-suite-high-results').iterdir())
    )

    if strtobool(OVERWRITE_EXPECTED_DATA):
        shutil.rmtree(expected_deliverable_policy_dir)
        shutil.copytree(deliverable_policy_dir, expected_deliverable_policy_dir)

    if not strtobool(KEEP_TEMP_FILE_AND_DIR):
        shutil.rmtree(tmpdir)
