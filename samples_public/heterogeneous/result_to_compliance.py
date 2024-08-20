import argparse
import subprocess
from typing import Dict, List

from trestle.oscal.assessment_results import AssessmentResults
from trestle.oscal.assessment_results import Model as AssessmentResultsRoot
from trestle.oscal.assessment_results import Result
from trestle.oscal.common import (
    ControlSelection,
    Metadata,
    ReviewedControls,
    SelectControlById,
)

from c2p.framework.oscal_utils import get_datetime, uuid

TEST_DATA_BASE_DIR = 'plugins_public/tests/data'
TEST_DATA_DIR = f'{TEST_DATA_BASE_DIR}/heterogeneous'

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
    '-r',
    '--result_directory',
    type=str,
    default=f'{TEST_DATA_BASE_DIR}',
    help=f'Path to  (default: {TEST_DATA_BASE_DIR}',
    required=False,
)

args = parser.parse_args()


def run(pvp, *additionals) -> AssessmentResultsRoot:
    command = [
        'python',
        f'samples_public/{pvp}/result_to_compliance.py',
        '-c',
        args.component_definition,
    ]
    command = command + list(additionals)
    ret = subprocess.run(command, capture_output=True, text=True)
    return AssessmentResultsRoot.parse_raw(ret.stdout)


ar_auditree = run('auditree', '-i', f'{args.result_directory}/auditree/check_results.json')
ar_kyverno = run(
    'kyverno',
    '-polr',
    f'{args.result_directory}/kyverno/policyreports.wgpolicyk8s.io.yaml',
    '-cpolr',
    f'{args.result_directory}/kyverno/clusterpolicyreports.wgpolicyk8s.io.yaml',
)
ar_ocm = run(
    'ocm',
    '-p',
    f'{args.result_directory}/ocm/policies.policy.open-cluster-management.io.yaml',
)
metadata = Metadata(
    title='System Assessment Results (using heterogeneous PVPs',
    last_modified=get_datetime(),
    oscal_version=ar_auditree.assessment_results.metadata.oscal_version,
    version='0.0.1',
)

include_controls_map: Dict[str, List[str]] = {}


def map_to_include_controls(ar: AssessmentResultsRoot) -> List[SelectControlById]:
    return [
        control
        for selection in ar.assessment_results.results[0].reviewed_controls.control_selections
        for control in selection.include_controls
    ]


include_controls = (
    map_to_include_controls(ar_auditree) + map_to_include_controls(ar_kyverno) + map_to_include_controls(ar_ocm)
)

for c in include_controls:
    if c.control_id in include_controls_map:
        include_controls_map[c.control_id] = include_controls_map[c.control_id] + c.statement_ids
    else:
        include_controls_map[c.control_id] = c.statement_ids
include_controls = [SelectControlById(control_id=x[0], statement_ids=x[1]) for x in include_controls_map.items()]
reviewed_controls = ReviewedControls(control_selections=[ControlSelection(include_controls=include_controls)])


def map_to_subjects(ar: AssessmentResultsRoot) -> List[SelectControlById]:
    return [
        control
        for selection in ar.assessment_results.results[0].reviewed_controls.control_selections
        for control in selection.include_controls
    ]


observations = (
    ar_auditree.assessment_results.results[0].observations
    + ar_kyverno.assessment_results.results[0].observations
    + ar_ocm.assessment_results.results[0].observations
)

result = Result(
    uuid=uuid(),
    title='System Assessment Results',
    description='System Assessment Results',
    start=get_datetime(),
    reviewed_controls=reviewed_controls,
    observations=observations,
)

ar = AssessmentResults(
    uuid=uuid(), metadata=metadata, import_ap=ar_auditree.assessment_results.import_ap, results=[result]
)
print(ar.oscal_serialize_json(pretty=True))
