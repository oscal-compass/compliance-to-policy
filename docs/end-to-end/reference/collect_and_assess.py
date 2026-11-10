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
"""What the engines on a cluster decided, back into OSCAL and a posture.

    policy already applied to a cluster
      -> results gathered per engine      (this script, a reference implementation)
      -> OSCAL Assessment Results         (c2p, one document for every engine)
      -> compliance posture               (c2p tools result2posture)

The chain before this one is `samples_public/kubernetes/generate.py`, then a person reading what
it produced, then `kubectl apply`. **This script is a reference implementation and not part of
the C2P library.** It holds a kubeconfig and talks to a cluster, which are decisions about one
particular cluster; `../README.md` says what it does with that access, and why the step before it
is not automated.

It does not render a single template. Collection is a function of the component definition --
which checks each engine declares -- and of what the engines reported. That is why this half can
be pointed at a cluster whose policy was applied by a GitOps sync or a controller, by someone who
has never run the generator.

The engines differ in a way worth seeing. Both Kyverno families report through the wgpolicyk8s.io
PolicyReport API, so collecting is a read and it includes violations that already existed.
ValidatingAdmissionPolicy has no report API and evaluates nothing retroactively, so its verdicts
have to be probed for -- see `platform_k8s/collect.py`, which is mostly about doing that without
disturbing what the cluster is enforcing.
"""

import argparse
import pathlib
import sys

HERE = pathlib.Path(__file__).resolve().parent
REPO_ROOT = HERE.parents[2]
# Run from anywhere: the repository root for `c2p` and `plugins_public`, this directory for the
# reference platform package. `docs/end-to-end` has a hyphen in it, so this package cannot be
# addressed as a dotted path and has to be found this way.
sys.path[:0] = [str(REPO_ROOT), str(HERE)]

from trestle.oscal.catalog import Catalog  # noqa: E402
from trestle.oscal.component import ComponentDefinition  # noqa: E402

from c2p.framework.c2p import C2P  # noqa: E402
from c2p.framework.models import RawResult  # noqa: E402
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal  # noqa: E402
from c2p.tools.posture import posture, render  # noqa: E402
from c2p.tools.posture.model import Source  # noqa: E402
from platform_k8s import KubeClient, collect  # noqa: E402
from plugins_public.plugins.kyverno_cel import PluginConfigKyvernoCel, PluginKyvernoCel  # noqa: E402
from plugins_public.plugins.kyverno_dsl import PluginConfigKyvernoDsl, PluginKyvernoDsl  # noqa: E402
from plugins_public.plugins.vap import PluginConfigVap, PluginVap  # noqa: E402

SAMPLE = REPO_ROOT / 'samples_public' / 'kubernetes'

REPORTED = 'reported'
PROBED = 'probed'

ENGINES = {
    'kyverno-dsl': (PluginKyvernoDsl, PluginConfigKyvernoDsl, REPORTED),
    'kyverno-cel': (PluginKyvernoCel, PluginConfigKyvernoCel, REPORTED),
    'vap-cel': (PluginVap, PluginConfigVap, PROBED),
}


def parse_args():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('-c', '--component-definition', type=str, default=str(SAMPLE / 'component-definition.json'))
    parser.add_argument('-o', '--out', type=str, default='out', help='Assessment Results and posture go here')
    parser.add_argument('-e', '--engine', action='append', choices=list(ENGINES), help='Repeatable; default is all')
    parser.add_argument('--catalog', type=str, default=None, help='Titles the controls in the posture')
    parser.add_argument('--framework', type=str, default=None)
    parser.add_argument('-n', '--namespace', type=str, default='storefront')
    parser.add_argument(
        '--canaries',
        type=str,
        default=str(HERE / 'canaries'),
        help='One <check-id>.canary.yaml per probed check: an object that must fail that check',
    )
    parser.add_argument('--kubeconfig', type=str, default=None)
    parser.add_argument(
        '--context',
        type=str,
        required=True,
        help='kubectl context to act on. Required: this writes to a cluster, and inheriting '
        'current-context is how a run lands on the wrong one',
    )
    parser.add_argument('--kubectl', type=str, default='kubectl')
    return parser.parse_args()


def collect_one(engine, args, client):
    """Ask one engine what it saw, and convert that into this engine's PVP result."""
    plugin_cls, config_cls, mode = ENGINES[engine]

    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = args.component_definition
    c2p_config.pvp_name = engine
    policy = C2P(c2p_config).get_policy()
    check_ids = [rule_set.check_id for rule_set in policy.rule_sets]

    if mode == REPORTED:
        # Scoped to this engine's checks: every Kyverno family reports to the same PolicyReport
        # API, so an unfiltered collection makes each plugin report the other engine's checks too.
        raw = collect.collect_policy_reports(client, check_ids)
    else:
        raw = collect.probe_admission_policies(
            client,
            check_ids,
            namespace=args.namespace,
            canary_dir=args.canaries,
        )
    print(f'{engine}: collected ({mode}) for {len(check_ids)} check(s)')
    return plugin_cls(config_cls()).generate_pvp_result(RawResult(data=raw))


def main():
    args = parse_args()
    # Said before anything is written: the VAP probe creates a namespace and bindings, so this
    # is not a read-only run, and which cluster it acted on belongs in its output.
    print(f'context: {args.context}')
    if args.kubeconfig != None:
        print(f'kubeconfig: {args.kubeconfig}')
    out_root = pathlib.Path(args.out)
    out_root.mkdir(parents=True, exist_ok=True)
    client = KubeClient(kubeconfig=args.kubeconfig, context=args.context, kubectl=args.kubectl)
    engines = args.engine if args.engine != None else list(ENGINES)

    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = args.component_definition
    c2p_config.pvp_name = engines[0]
    c2p_config.result_title = 'Assessment Results'
    c2p_config.result_description = f'Assessed by: {", ".join(engines)}'
    c2p = C2P(c2p_config)

    for engine in engines:
        c2p.add_pvp_result(collect_one(engine, args, client))

    assessment_results = c2p.result_to_oscal()
    ar_path = out_root / 'assessment-results.json'
    ar_path.write_text(assessment_results.oscal_serialize_json(pretty=True))
    print(f'\nwrote {ar_path} ({len(assessment_results.results[0].observations)} observations)')

    component_definition = ComponentDefinition.oscal_read(pathlib.Path(args.component_definition))
    catalog = Catalog.oscal_read(pathlib.Path(args.catalog)) if args.catalog != None else None
    result = posture.generate(
        assessment_results,
        component_definition,
        catalog=catalog,
        framework=args.framework,
        source=Source(assessment_results=str(ar_path), component_definition=args.component_definition),
    )
    (out_root / 'posture.json').write_text(render.render_json(result))
    (out_root / 'posture.html').write_text(render.render_html(result))
    print(f'wrote {out_root / "posture.json"} and posture.html\n')
    print(render.render_markdown(result))


if __name__ == '__main__':
    main()
