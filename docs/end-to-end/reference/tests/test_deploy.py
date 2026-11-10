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
"""What gets applied is what was enumerated, and therefore what was read.

The step before this one is a person reading `out/`. That only means anything if the set of
documents they read is the set that reaches the cluster -- so the thing worth pinning is not
that `deploy` applies policy, but that it applies *these* files and nothing else that happens
to be sitting next to them.
"""

import pathlib
import sys

import pytest

from c2p.common.err import C2PError
from platform_k8s import deploy

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1]))

from deploy import deletion_command  # noqa: E402  -- the CLI beside platform_k8s/deploy.py


class FakeClient:

    def __init__(self):
        self.applied = None
        self.directories = []

    def apply_files(self, paths):
        self.applied = [pathlib.Path(p) for p in paths]

    def apply(self, path, recursive=True):  # pragma: no cover - must not be reached
        self.directories.append(pathlib.Path(path))


def policy_tree(root: pathlib.Path) -> pathlib.Path:
    out = root / 'out'
    for engine, rule, name in [
        ('kyverno-cel', 'image-pinned-by-digest', 'kyverno-cel-image-pinned-by-digest.yaml'),
        ('vap-cel', 'image-pinned-by-digest', 'vap-image-pinned-by-digest.yaml'),
    ]:
        document = out / engine / rule
        document.mkdir(parents=True, exist_ok=True)
        (document / name).write_text('kind: ValidatingAdmissionPolicy\n')
    return out


def test_only_the_enumerated_documents_are_applied(tmp_path):
    out = policy_tree(tmp_path)
    client = FakeClient()

    files = deploy.deploy(client, str(out))

    assert client.directories == [], 'the directory was handed to kubectl, not the files'
    assert client.applied == files
    assert [p.name for p in client.applied] == [
        'kyverno-cel-image-pinned-by-digest.yaml',
        'vap-image-pinned-by-digest.yaml',
    ]


def test_the_results_a_collection_run_writes_beside_the_policy_are_not_applied(tmp_path):
    """`collect_and_assess.py -o out` writes assessment-results.json and posture.json into the
    same tree. `kubectl apply -R -f out` would apply them: kubectl reads `.json` and `.yml` as
    well as `.yaml`, and neither was read as policy by anyone."""
    out = policy_tree(tmp_path)
    (out / 'assessment-results.json').write_text('{"assessment-results": {}}\n')
    (out / 'posture.json').write_text('{"schemaVersion": "c2p-posture/v1"}\n')
    (out / 'kyverno-cel' / 'stray.yml').write_text('kind: ClusterPolicy\n')
    client = FakeClient()

    deploy.deploy(client, str(out))

    names = [p.name for p in client.applied]
    assert 'assessment-results.json' not in names
    assert 'posture.json' not in names
    assert 'stray.yml' not in names
    assert len(names) == 2


def test_an_empty_tree_is_refused_rather_than_applied(tmp_path):
    empty = tmp_path / 'out'
    empty.mkdir()
    client = FakeClient()

    with pytest.raises(C2PError, match='No policy documents'):
        deploy.deploy(client, str(empty))

    assert client.applied == None


def test_a_missing_tree_says_to_generate_first(tmp_path):
    client = FakeClient()

    with pytest.raises(C2PError, match='Generate the policy first'):
        deploy.deploy(client, str(tmp_path / 'nope'))


def test_the_printed_deletion_names_the_same_files_and_nothing_else(tmp_path):
    """The command `deploy.py` prints gets pasted, so it has to remove what this run applied.

    Not a label selector: the template contract puts `compliance.oscal/check-id` on every policy
    C2P generates, so selecting on it reaches every C2P policy on the cluster -- another
    component definition's, another team's, whatever GitOps installed -- none of which this run
    applied or read.
    """
    out = policy_tree(tmp_path)
    client = FakeClient()
    files = deploy.deploy(client, str(out))

    command = deletion_command('kind-storefront', files)

    assert '-l ' not in command, 'a label selector reaches policies this run did not apply'
    assert '--selector' not in command
    assert '-R' not in command
    assert command.count('-f ') == len(files)
    for path in files:
        assert f'-f {path}' in command
    assert '--context kind-storefront' in command
