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
"""Generated policy, compared byte for byte against a committed reference.

Driven from the bundled sample assets rather than from fixtures invented here, because the
claim worth testing is not "the renderer substitutes variables". `expected-policies` is a
committed snapshot of this generator's output: the test holds generated policy byte for byte
stable, so a change to a template or to the renderer has to be made deliberately and read as
a diff. It is not independent of the generator and does not claim to be.

The second test is the one that is independent of a snapshot: one authored parameter has to
reach two different engines as the same CEL expression, which no reference file asserts.

"Looks right" is not a criterion for generated policy. A CEL expression is compared as a
string, by a cluster, and one space in a list literal is the difference between two engines
agreeing about a control and silently disagreeing.
"""

import pathlib

import pytest
import yaml

from c2p.framework.c2p import C2P
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal
from plugins_public.plugins.kyverno_cel import PluginConfigKyvernoCel, PluginKyvernoCel
from plugins_public.plugins.kyverno_dsl import PluginConfigKyvernoDsl, PluginKyvernoDsl
from plugins_public.plugins.vap import PluginConfigVap, PluginVap

SAMPLE = pathlib.Path(__file__).resolve().parents[3] / 'samples_public' / 'kubernetes'
COMPONENT_DEFINITION = SAMPLE / 'component-definition.json'

# The two CEL engines, which have to render one authored parameter to the same expression.
CEL_ENGINES = [
    ('kyverno-cel', PluginKyvernoCel, PluginConfigKyvernoCel, 6),
    ('vap-cel', PluginVap, PluginConfigVap, 4),
]

# Every engine the component definition declares. kyverno-dsl generates the same six checks as
# kyverno-cel from the same rules, in Kyverno's earlier API: the control outlives the engine.
ENGINES = CEL_ENGINES + [('kyverno-dsl', PluginKyvernoDsl, PluginConfigKyvernoDsl, 6)]


def generate(engine, plugin_cls, config_cls, out_dir):
    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = COMPONENT_DEFINITION.as_posix()
    c2p_config.pvp_name = engine
    policy = C2P(c2p_config).get_policy()
    plugin_cls(
        config_cls(
            policy_template_dir=str(SAMPLE / 'templates' / engine),
            deliverable_policy_dir=str(out_dir),
        )
    ).generate_pvp_policy(policy)
    return policy


@pytest.mark.parametrize('engine,plugin_cls,config_cls,expected_checks', ENGINES)
def test_generated_policy_matches_the_reference_byte_for_byte(
    engine, plugin_cls, config_cls, expected_checks, tmp_path
):
    policy = generate(engine, plugin_cls, config_cls, tmp_path)
    assert len(policy.rule_sets) == expected_checks

    reference_root = SAMPLE / 'expected-policies' / engine
    references = sorted(reference_root.glob('**/*.yaml'))
    assert len(references) == expected_checks

    for reference in references:
        generated = tmp_path / reference.relative_to(reference_root)
        assert generated.exists(), f'{generated} was not generated'
        assert generated.read_text() == reference.read_text(), f'{generated.name} differs from the reference'


def test_one_parameter_renders_identically_in_both_cel_engines(tmp_path):
    """The point of an engine-neutral control: the same authored parameter becomes the same CEL
    list literal in Kyverno's CEL policy and in a ValidatingAdmissionPolicy."""
    for engine, plugin_cls, config_cls, _ in CEL_ENGINES:
        generate(engine, plugin_cls, config_cls, tmp_path / engine)

    def required_expression(path):
        """Read the variable out of the document, not out of the text.

        Matched against the raw lines this went looking for a `["` and found
        `apiGroups: ["apps"]` the moment the expression was quoted differently -- two engines
        agreeing about a line neither test nor policy depends on.
        """
        documents = [d for d in yaml.safe_load_all(path.read_text()) if d != None]
        policy = next(d for d in documents if 'variables' in d.get('spec', {}))
        return next(v['expression'] for v in policy['spec']['variables'] if v['name'] == 'required')

    rule = 'workload-inventory-labelled'
    cel = required_expression(tmp_path / 'kyverno-cel' / rule / f'kyverno-cel-{rule}.yaml')
    vap = required_expression(tmp_path / 'vap-cel' / rule / f'vap-{rule}.yaml')
    assert cel == vap
