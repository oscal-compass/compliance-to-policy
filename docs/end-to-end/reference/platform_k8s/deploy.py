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
"""Apply a generated policy tree.

This runs after a person has read the tree, which is the step `../README.md` is about and the
reason applying is a command of its own rather than something the generator does on its way
past. What is applied is exactly what was read: the same directory, unmodified.

There is no `undeploy`. Removing policy from a cluster only ever reduces enforcement, and
`kubectl delete -R -f <dir>` is a decision to make deliberately rather than a convenience to
put next to this.
"""

import pathlib
from typing import List

from c2p.common.err import C2PError
from c2p.common.logging import getLogger

from .client import KubeClient

logger = getLogger(__name__)


def policy_files(policy_dir: str) -> List[pathlib.Path]:
    root = pathlib.Path(policy_dir)
    if not root.is_dir():
        raise C2PError(f'Generated policy directory {policy_dir} does not exist. Generate the policy first.')
    return sorted(root.glob('**/*.yaml'))


def deploy(client: KubeClient, policy_dir: str) -> List[pathlib.Path]:
    """Apply every generated policy document under `policy_dir`.

    Applied as a directory rather than file by file so that a policy and the binding that
    activates it land together; a policy without its binding enforces nothing, and that is
    indistinguishable in a posture from a policy that passed.
    """
    files = policy_files(policy_dir)
    if files == []:
        raise C2PError(f'No policy documents under {policy_dir}')
    client.apply(pathlib.Path(policy_dir))
    logger.info(f'Applied {len(files)} policy document(s) from {policy_dir}')
    return files
