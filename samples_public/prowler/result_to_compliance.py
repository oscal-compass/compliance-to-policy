import argparse
import json
import os
import pathlib
import sys

from c2p.framework.c2p import C2P
from c2p.framework.models import RawResult
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal

sys.path.append(os.path.join(os.path.dirname(__file__), '../..'))
from plugins_public.plugins.prowler import PluginConfigProwler, PluginProwler

TEST_DATA_DIR = 'plugins_public/tests/data/prowler'

parser = argparse.ArgumentParser()
parser.add_argument(
    '-r',
    '--result',
    type=str,
    default=f'{TEST_DATA_DIR}/scan.ocsf.json',
    help='Path to the Prowler OCSF output (json-ocsf)',
    required=False,
)
parser.add_argument(
    '-c',
    '--component_definition',
    type=str,
    default=f'{TEST_DATA_DIR}/component-definition.json',
    help=f'Path to component-definition.json (default: {TEST_DATA_DIR}/component-definition.json)',
    required=False,
)
parser.add_argument(
    '-d',
    '--deliverable_policy_dir',
    type=str,
    help='Directory written by compliance_to_policy.py. Supplying it lets the '
    'plugin roll check results up into rule-level findings.',
    required=False,
)
args = parser.parse_args()

# Setup c2p_config
c2p_config = C2PConfig()
c2p_config.compliance = ComplianceOscal()
c2p_config.compliance.component_definition = args.component_definition
c2p_config.pvp_name = 'Prowler'
c2p_config.result_title = 'Prowler Assessment Results'
c2p_config.result_description = 'OSCAL Assessment Results from Prowler'

# Construct C2P
c2p = C2P(c2p_config)

# Create pvp_result from raw result via plugin
pvp_raw_result = RawResult(data=json.loads(pathlib.Path(args.result).read_text()))
plugin = (
    PluginProwler(PluginConfigProwler(provider='aws', deliverable_policy_dir=args.deliverable_policy_dir))
    if args.deliverable_policy_dir
    else PluginProwler()
)
pvp_result = plugin.generate_pvp_result(pvp_raw_result)

# Transform pvp_result to OSCAL Assessment Result
c2p.set_pvp_result(pvp_result)
oscal_assessment_results = c2p.result_to_oscal()

print(oscal_assessment_results.oscal_serialize_json(pretty=True))
