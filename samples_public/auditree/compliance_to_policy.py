import argparse
import os
import sys
import tempfile
from pathlib import Path

from c2p.framework.c2p import C2P
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal

sys.path.append(os.path.join(os.path.dirname(__file__), '../..'))
from plugins_public.plugins.auditree import PluginAuditree, PluginConfigAuditree

TEST_DATA_DIR = 'plugins_public/tests/data/auditree'

parser = argparse.ArgumentParser()
parser.add_argument(
    '-i',
    '--input',
    type=str,
    default=f'{TEST_DATA_DIR}/auditree.template.json',
    help=f'Path to auditree.json template (default: {TEST_DATA_DIR}/auditree.template.json)',
    required=False,
)
parser.add_argument(
    '-c',
    '--component_definition',
    type=str,
    default=f'{TEST_DATA_DIR}/component-definition.json',
    help=f'Path to component-definition.json (default: {TEST_DATA_DIR}/component-definition.json',
    required=False,
)
parser.add_argument(
    '-o',
    '--out',
    type=str,
    help='Path to generated auditree.json (default: system temporary directory)',
    required=False,
)
args = parser.parse_args()

with Path(args.out).open('w') if args.out != None else tempfile.NamedTemporaryFile() as output:
    # Setup c2p_config
    c2p_config = C2PConfig()
    c2p_config.compliance = ComplianceOscal()
    c2p_config.compliance.component_definition = args.component_definition
    c2p_config.pvp_name = 'Auditree'
    c2p_config.result_title = 'Auditree Assessment Results'
    c2p_config.result_description = 'OSCAL Assessment Results from Auditree'

    # Construct C2P
    c2p = C2P(c2p_config)

    # Transform OSCAL (Compliance) to Policy
    config = PluginConfigAuditree(auditree_json_template=args.input, output=output.name)
    PluginAuditree(config).generate_pvp_policy(c2p.get_policy())
