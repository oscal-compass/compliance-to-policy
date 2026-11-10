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
"""One component definition to deployable policy for three engines.

    OSCAL component definition  +  one template per check
      -> policy per engine                                   (this script)

This is C2P doing the job C2P has always done: a plugin per engine turns a control into that
engine's native policy. Nothing here is new except the engines and where the templates came
from — they were written by an agent skill rather than by hand, and C2P cannot tell.

**No cluster, no credentials, no network.** Generating is a function of files, which is what
makes the next step possible: a person reads the result before any of it is applied. Applying
it, and collecting what the engines then decide, is deliberately not in this script and not in
the C2P library: it belongs to the program that holds the credentials for the cluster.
"""

import argparse
import os
import pathlib

from c2p.framework.c2p import C2P
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal
from plugins_public.plugins.kyverno_cel import PluginConfigKyvernoCel, PluginKyvernoCel
from plugins_public.plugins.kyverno_dsl import PluginConfigKyvernoDsl, PluginKyvernoDsl
from plugins_public.plugins.vap import PluginConfigVap, PluginVap

HERE = os.path.dirname(os.path.abspath(__file__))

ENGINES = {
    'kyverno-dsl': (PluginKyvernoDsl, PluginConfigKyvernoDsl),
    'kyverno-cel': (PluginKyvernoCel, PluginConfigKyvernoCel),
    'vap-cel': (PluginVap, PluginConfigVap),
}


def parse_args():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('-c', '--component-definition', type=str, default=f'{HERE}/component-definition.json')
    parser.add_argument(
        '-t', '--template-root', type=str, default=f'{HERE}/templates', help='Holds one directory per engine'
    )
    parser.add_argument('-o', '--out', type=str, default='out', help='One directory per engine is written here')
    parser.add_argument('-e', '--engine', action='append', choices=list(ENGINES), help='Repeatable; default is all')
    return parser.parse_args()


def generate(engine: str, component_definition: str, template_root: str, out_root: pathlib.Path) -> int:
    plugin_cls, config_cls = ENGINES[engine]

    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = component_definition
    c2p_config.pvp_name = engine
    policy = C2P(c2p_config).get_policy()

    policy_dir = out_root / engine
    plugin = plugin_cls(
        config_cls(
            policy_template_dir=f'{template_root}/{engine}',
            deliverable_policy_dir=str(policy_dir),
        )
    )
    plugin.generate_pvp_policy(policy)
    print(f'{engine}: generated {len(policy.rule_sets)} policy document(s) into {policy_dir}')
    return len(policy.rule_sets)


def main():
    args = parse_args()
    out_root = pathlib.Path(args.out)
    out_root.mkdir(parents=True, exist_ok=True)
    engines = args.engine if args.engine != None else list(ENGINES)

    total = 0
    for engine in engines:
        total = total + generate(engine, args.component_definition, args.template_root, out_root)

    print(
        f'\n{total} policy document(s) under {out_root}. Read them before applying them: a check '
        f'that asserts nothing applies cleanly and reports every workload compliant.'
    )


if __name__ == '__main__':
    main()
