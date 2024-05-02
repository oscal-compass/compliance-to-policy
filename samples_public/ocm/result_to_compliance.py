import argparse
import os
import pathlib
import sys

import yaml

from c2p.framework.c2p import C2P
from c2p.framework.models.c2p_config import C2PConfig, ComplianceOscal
from c2p.framework.models.raw_result import RawResult

sys.path.append(os.path.join(os.path.dirname(__file__), '../..'))
from plugins_public.plugins.ocm import PluginOCM

TEST_DATA_DIR = 'plugins_public/tests/data/ocm'

parser = argparse.ArgumentParser()
parser.add_argument(
    '-p',
    '--policy-result',
    type=str,
    default=f'{TEST_DATA_DIR}/policies.policy.open-cluster-management.io.yaml',
    help='Path to a yaml file in which policies.policy.open-cluster-management.io resources are dumped.',
    required=False,
)
args = parser.parse_args()

# Setup c2p_config
c2p_config = C2PConfig()
c2p_config.compliance = ComplianceOscal()
c2p_config.compliance.component_definition = 'plugins_public/tests/data/ocm/component-definition.json'
c2p_config.pvp_name = 'OCM'
c2p_config.result_title = 'OCM Assessment Results'
c2p_config.result_description = 'OSCAL Assessment Results from OCM'

# Create pvp_result from raw result via plugin
policies = yaml.safe_load(pathlib.Path(args.policy_result).open('r'))
pvp_raw_result = RawResult(data=policies['items'])
c2p_config.pvp_result = PluginOCM().generate_pvp_result(pvp_raw_result)

# Transform pvp_result to OSCAL Assessment Result
c2p = C2P(c2p_config)
oscal_assessment_results = c2p.result_to_oscal()

print(oscal_assessment_results.oscal_serialize_json(pretty=True))
