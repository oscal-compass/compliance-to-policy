import argparse
import os
import sys
from pathlib import Path

import yaml

from c2p.framework.c2p import C2P
from c2p.framework.models import RawResult
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal
from c2p.framework.models.raw_result import RawResult

sys.path.append(os.path.join(os.path.dirname(__file__), '../..'))
from plugins_public.plugins.auditree import PluginAuditree

TEST_DATA_DIR = 'plugins_public/tests/data/auditree'

parser = argparse.ArgumentParser()
parser.add_argument(
    '-i',
    '--input',
    type=str,
    default=f'{TEST_DATA_DIR}/check_results.json',
    help=f'Path to check_results.json (default: {TEST_DATA_DIR}/check_results.json)',
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
args = parser.parse_args()

# Setup c2p_config
c2p_config = C2PConfig()
c2p_config.compliance = ComplianceOscal()
c2p_config.compliance.component_definition = args.component_definition
c2p_config.pvp_name = 'Auditree'
c2p_config.result_title = 'Auditree Assessment Results'
c2p_config.result_description = 'OSCAL Assessment Results from Auditree'

# Construct C2P
c2p = C2P(c2p_config)

# Create pvp_result from raw result via plugin
check_results = yaml.safe_load(Path(args.input).open('r'))
pvp_raw_result = RawResult(
    data=check_results,
    additional_props={
        'locker_url': 'https://github.com/MY_ORG/MY_EVIDENCE_REPO',
    },
)
pvp_result = PluginAuditree().generate_pvp_result(pvp_raw_result)

# Transform pvp_result to OSCAL Assessment Result
c2p.set_pvp_result(pvp_result)
oscal_assessment_results = c2p.result_to_oscal()

print(oscal_assessment_results.oscal_serialize_json(pretty=True))
