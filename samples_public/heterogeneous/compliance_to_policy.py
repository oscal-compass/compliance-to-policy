import argparse
import subprocess
from pathlib import Path

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
    '-o',
    '--out',
    type=str,
    help=f'Path to output directory',
    required=True,
)

args = parser.parse_args()
out = args.out


def run(pvp, *additionals):
    command = [
        'python',
        f'samples_public/{pvp}/compliance_to_policy.py',
        '-c',
        args.component_definition,
    ]
    command = command + list(additionals)
    subprocess.run(command, capture_output=True, text=True)


Path(out).mkdir(exist_ok=True)
(Path(out) / 'auditree').mkdir(exist_ok=True)

run('auditree', '-o', f'{out}/auditree/auditree.json')
run('kyverno', '-o', f'{out}/kyverno')
run('ocm', '-o', f'{out}/ocm')
