# -*- mode:python; coding:utf-8 -*-

# Copyright 2026 IBM Corporation
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
"""Where a rule id and a check id become a filesystem path.

A component definition is authored input, and in the flow these plugins exist for it is authored
by an agent. Both ids are used unescaped as a path segment -- the output directory and the
template to read -- so an id is as trusted as the path it builds. Generation writes without
asking, so an id that escapes the output directory overwrites whatever is there.
"""

import json
import pathlib

import pytest

from c2p.common.err import C2PError
from c2p.framework.models.policy import Policy, RuleSet
from plugins_public.plugins import policy_template

SAMPLE_CD = pathlib.Path(__file__).resolve().parents[3] / 'samples_public' / 'kubernetes' / 'component-definition.json'


def _policy(rule_id: str, check_id: str) -> Policy:
    policy = Policy()
    policy.rule_sets = [RuleSet(rule_id=rule_id, rule_description='d', check_id=check_id, check_description='d')]
    policy.parameters = []
    return policy


KINDS = ['ValidatingAdmissionPolicy']

# The smallest document that is what a vap-cel template claims to emit: rendering is now
# checked against that claim, so a fixture emitting `name: ...` alone would be refused.
TEMPLATE = (
    'apiVersion: admissionregistration.k8s.io/v1\n'
    'kind: ValidatingAdmissionPolicy\n'
    'metadata:\n'
    '  name: {{ check.id }}\n'
)


def _templates(directory: pathlib.Path, check_id: str, body: str = TEMPLATE) -> str:
    (directory / f'{check_id}{policy_template.TEMPLATE_SUFFIX}').write_text(body)
    return str(directory)


def test_a_valid_identifier_renders_where_it_says(tmp_path):
    template_dir = tmp_path / 'tmpl'
    template_dir.mkdir()
    _templates(template_dir, 'check-a')
    out = tmp_path / 'out'

    written = policy_template.render(_policy('rule-a', 'check-a'), str(template_dir), str(out), 'vap-cel', KINDS)

    assert written == [out / 'rule-a' / 'check-a.yaml']
    assert written[0].read_text() == TEMPLATE.replace('{{ check.id }}', 'check-a')


def test_a_rule_id_that_climbs_out_of_the_output_directory_is_refused(tmp_path):
    template_dir = tmp_path / 'tmpl'
    template_dir.mkdir()
    _templates(template_dir, 'check-a')
    out = tmp_path / 'out'
    out.mkdir()
    bystander = tmp_path / 'chk.yaml'
    bystander.write_text('do not overwrite me\n')

    with pytest.raises(C2PError, match='not a valid Kubernetes name'):
        policy_template.render(_policy('../', 'chk'), str(template_dir), str(out), 'vap-cel', KINDS)

    assert bystander.read_text() == 'do not overwrite me\n'


def test_a_check_id_that_climbs_out_of_the_template_directory_is_refused(tmp_path):
    """The read side, which is the worse half: the escaped file is rendered as a template and its
    contents are written into the generated policy."""
    template_dir = tmp_path / 'tmpl'
    template_dir.mkdir()
    (tmp_path / f'secret{policy_template.TEMPLATE_SUFFIX}').write_text('a-token\n')

    with pytest.raises(C2PError, match='not a valid Kubernetes name'):
        policy_template.render(
            _policy('rule-a', '../secret'), str(template_dir), str(tmp_path / 'out'), 'vap-cel', KINDS
        )


@pytest.mark.parametrize(
    'identifier',
    ['Check-A', 'check_a', 'check.a', '-check', 'check-', '', 'a' * 254, '../x', 'x/y'],
)
def test_an_identifier_kubernetes_would_not_accept_is_refused(identifier):
    """A refused id could not have been applied to a cluster either: every generated object is
    named after it. Failing at generation is the earlier and clearer of the two."""
    with pytest.raises(C2PError):
        policy_template.validate_identifier('Check_Id', identifier)


def test_the_shipped_sample_is_unaffected():
    """The validation was added after the sample, so this is the regression that matters: no
    generated path and no expected output changes."""
    document = json.loads(SAMPLE_CD.read_text())
    seen = []

    def walk(node):
        if isinstance(node, dict):
            if node.get('name') in ['Rule_Id', 'Check_Id']:
                seen.append((node['name'], node.get('value')))
            for value in node.values():
                walk(value)
        elif isinstance(node, list):
            for value in node:
                walk(value)

    walk(document)
    assert len(seen) >= 20
    for kind, value in seen:
        assert policy_template.validate_identifier(kind, value) == value
