import argparse
import os
import pathlib
import sys
import tempfile

from c2p.framework.c2p import C2P
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal

sys.path.append(os.path.join(os.path.dirname(__file__), '../..'))
from plugins_public.plugins.ocm import PluginConfigOCM, PluginOCM

TEST_DATA_DIR = 'plugins_public/tests/data/ocm'

parser = argparse.ArgumentParser()
parser.add_argument(
    '-c',
    '--component_definition',
    type=str,
    default=f'{TEST_DATA_DIR}/component-definition.json',
    help=f'Path to component-definition.json (default: {TEST_DATA_DIR}/component-definition.json',
    required=False,
)
parser.add_argument(
    '-o', '--out', type=str, help='Path to output directory (default: system temporary directory)', required=False
)
args = parser.parse_args()

tmpdirname = args.out if args.out != None else tempfile.mkdtemp()

# Setup c2p_config
c2p_config = C2PConfig()
c2p_config.compliance = ComplianceOscal()
c2p_config.compliance.component_definition = args.component_definition
c2p_config.pvp_name = 'OCM'

# Construct C2P
c2p = C2P(c2p_config)

# Transform OSCAL (Compliance) to Policy
policy_template_dir = f'{TEST_DATA_DIR}/policy-resources'
config = PluginConfigOCM(
    policy_template_dir=policy_template_dir,
    deliverable_policy_dir=tmpdirname,
    namespace='c2p',
    paremeters_configmap_name='c2p-parameters',
    cluster_selectors={'environment': 'dev'},
    policy_set_name='c2p test',
)
PluginOCM(config).generate_pvp_policy(c2p.get_policy())


def tree(path: pathlib.Path, texts: list[str] = [], depth=0) -> list[str]:
    prefix = ''
    if depth > 0:
        for _ in range(depth - 1):
            prefix = prefix + '   '
        prefix = prefix + '- '
    for item in path.iterdir():
        texts.append(f'{prefix}{item.name}')
        if item.is_dir():
            tree(item, texts, depth=depth + 1)
    return texts


print('')
print(f'tree {tmpdirname}')
for text in tree(pathlib.Path(tmpdirname)):
    print(text)
