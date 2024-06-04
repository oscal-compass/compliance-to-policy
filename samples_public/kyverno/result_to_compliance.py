import argparse
import os
import pathlib
import sys

import yaml

from c2p.framework.c2p import C2P
from c2p.framework.models import RawResult
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal
from c2p.framework.models.raw_result import RawResult

sys.path.append(os.path.join(os.path.dirname(__file__), '../..'))
from plugins_public.plugins.kyverno import PluginKyverno

TEST_DATA_DIR = 'plugins_public/tests/data/kyverno'

parser = argparse.ArgumentParser()
parser.add_argument(
    '-polr',
    '--policy-report',
    type=str,
    default=f'{TEST_DATA_DIR}/policyreports.wgpolicyk8s.io.yaml',
    help='Path to policy report',
    required=False,
)
parser.add_argument(
    '-cpolr',
    '--cluster-policy-report',
    type=str,
    default=f'{TEST_DATA_DIR}/clusterpolicyreports.wgpolicyk8s.io.yaml',
    help='Path to cluster policy report',
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
c2p_config.pvp_name = 'Kyverno'
c2p_config.result_title = 'Kyverno Assessment Results'
c2p_config.result_description = 'OSCAL Assessment Results from Kyverno'

# Construct C2P
c2p = C2P(c2p_config)

# Create pvp_result from raw result via plugin
cpolr = yaml.safe_load(pathlib.Path(args.cluster_policy_report).open('r'))
polr = yaml.safe_load(pathlib.Path(args.policy_report).open('r'))
pvp_raw_result = RawResult(data=cpolr['items'] + polr['items'])
pvp_result = PluginKyverno().generate_pvp_result(pvp_raw_result)

# Transform pvp_result to OSCAL Assessment Result
c2p.set_pvp_result(pvp_result)
oscal_assessment_results = c2p.result_to_oscal()

print(oscal_assessment_results.oscal_serialize_json(pretty=True))
