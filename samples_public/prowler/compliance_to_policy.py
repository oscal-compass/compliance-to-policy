import argparse
import json
import os
import pathlib
import sys
import tempfile

from c2p.framework.c2p import C2P
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal

sys.path.append(os.path.join(os.path.dirname(__file__), '../..'))
from plugins_public.plugins.prowler import PluginConfigProwler, PluginProwler

TEST_DATA_DIR = 'plugins_public/tests/data/prowler'

parser = argparse.ArgumentParser()
parser.add_argument(
    '-c',
    '--component_definition',
    type=str,
    default=f'{TEST_DATA_DIR}/component-definition.json',
    help=f'Path to component-definition.json (default: {TEST_DATA_DIR}/component-definition.json)',
    required=False,
)
parser.add_argument('-p', '--provider', type=str, default='aws', help='Prowler provider', required=False)
parser.add_argument(
    '-o', '--out', type=str, help='Path to output directory (default: system temporary directory)', required=False
)
parser.add_argument(
    '-s',
    '--scope',
    type=str,
    help='JSON object of Prowler scope flags, e.g. \'{"project-ids": ["prod-1"]}\'. '
    'Scope is a runtime target, so it is supplied here rather than read from OSCAL.',
    required=False,
)
args = parser.parse_args()

tmpdirname = args.out if args.out != None else tempfile.mkdtemp()

# Setup c2p_config
c2p_config = C2PConfig()
c2p_config.compliance = ComplianceOscal()
c2p_config.compliance.component_definition = args.component_definition
c2p_config.pvp_name = 'Prowler'

# Construct C2P
c2p = C2P(c2p_config)

# Transform OSCAL (Compliance) to Policy. Prowler ships its own checks, so the
# generated policy is the check selection and the tunable config, not policy
# source.
config = PluginConfigProwler(
    provider=args.provider,
    deliverable_policy_dir=tmpdirname,
    scope=json.loads(args.scope) if args.scope else None,
)
generated = PluginProwler(config).generate_pvp_policy(c2p.get_policy())

print('')
print(f'Generated in {tmpdirname}:')
for name in sorted(p.name for p in pathlib.Path(tmpdirname).iterdir()):
    print(f'  {name}')
print('')
print('Run Prowler with:')
print(f"  {generated['command']}")
