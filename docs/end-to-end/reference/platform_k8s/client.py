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
"""A thin, explicit wrapper over `kubectl`.

Shelling out rather than importing a Kubernetes client is a deliberate trade for a reference
implementation: the operations here are one flag each, and keeping them as `kubectl` argument
lists means a reader can run any of them by hand to check what this code claims.

Two rules hold throughout, and both are load-bearing rather than stylistic.

**Every apply is a server-side apply.** A client-side `kubectl apply` computes a patch, and
when that patch is rejected kubectl prints the patch body -- which carries a
`last-applied-configuration` annotation holding the whole object, environment variables
included -- into its error message. Anything that records such a message has copied the
object's secrets into its output. Measured on Kubernetes v1.34.0 with a Kyverno webhook doing
the rejecting: client-side 2353 bytes with the object echoed, server-side 328 bytes with only
the policy's own message. Client-side apply also *skips admission entirely* when the computed
patch is empty, reporting `unchanged` with exit 0, which any probe would read as a pass.

**Nothing here deletes a resource by name.** Deletion is by label selector only, and the only
label this module ever selects on is one it wrote itself during this run. There is no code
path by which an identifier taken from a component definition reaches a `kubectl delete`.

The kubeconfig and context are always passed explicitly. Ambient cluster selection is how a
collection run silently reports on the wrong cluster.
"""

import json
import pathlib
import subprocess
from typing import Any, Dict, List, Optional

from c2p.common.err import C2PError
from c2p.common.logging import getLogger

logger = getLogger(__name__)

# Written by this module onto everything it creates, and the only selector it deletes by.
PROBE_RUN_LABEL = 'compliance.oscal/probe-run'


class KubeClient:

    def __init__(
        self,
        kubeconfig: Optional[str] = None,
        context: Optional[str] = None,
        kubectl: str = 'kubectl',
        timeout: int = 120,
        field_manager: str = 'c2p-reference',
    ) -> None:
        self.kubeconfig = kubeconfig
        self.context = context
        self.kubectl = kubectl
        self.timeout = timeout
        self.field_manager = field_manager

    def _args(self, args: List[str]) -> List[str]:
        prefix = [self.kubectl]
        if self.kubeconfig != None:
            prefix = prefix + ['--kubeconfig', self.kubeconfig]
        if self.context != None:
            prefix = prefix + ['--context', self.context]
        return prefix + args

    def run(self, args: List[str], check: bool = True, stdin: Optional[str] = None) -> subprocess.CompletedProcess:
        completed = subprocess.run(
            self._args(args),
            input=stdin,
            capture_output=True,
            text=True,
            timeout=self.timeout,
        )
        if check and completed.returncode != 0:
            raise C2PError(f'kubectl {" ".join(args)} failed: {completed.stderr.strip()}')
        return completed

    def version(self) -> str:
        return self.run(['version', '-o', 'json']).stdout

    def get_json(self, resources: str, namespace: Optional[str] = None, all_namespaces: bool = False) -> Dict[str, Any]:
        """Read resources as JSON, distinguishing "this engine reports nothing" from a failure.

        A CRD that is not installed means the engine reports nothing, which the posture should
        show as a gap. Anything else -- an authentication failure, the wrong context, an
        unreachable API server -- is a broken run, and reporting it as an empty result set
        would turn it into a uniform not-assessed posture that looks like a finding.
        """
        args = ['get', resources, '-o', 'json']
        if all_namespaces:
            args = args + ['-A']
        elif namespace != None:
            args = args + ['-n', namespace]
        completed = self.run(args, check=False)
        if completed.returncode != 0:
            stderr = completed.stderr.strip()
            if 'the server doesn\'t have a resource type' in stderr or 'NotFound' in stderr:
                logger.warning(f'{resources} is not installed on this cluster: reporting nothing')
                return {'items': []}
            raise C2PError(f'kubectl get {resources} failed, so this run cannot be trusted: {stderr}')
        return json.loads(completed.stdout)

    def names(self, resources: str, namespace: Optional[str] = None) -> List[str]:
        args = ['get', resources, '-o', 'name']
        if namespace != None:
            args = args + ['-n', namespace]
        completed = self.run(args, check=False)
        if completed.returncode != 0:
            raise C2PError(f'kubectl get {resources} failed: {completed.stderr.strip()}')
        return [line for line in completed.stdout.splitlines() if line.strip() != '']

    def get_manifest(self, reference: str, namespace: Optional[str] = None) -> Optional[str]:
        args = ['get', reference, '-o', 'yaml']
        if namespace != None:
            args = args + ['-n', namespace]
        completed = self.run(args, check=False)
        return completed.stdout if completed.returncode == 0 else None

    def apply(self, path: pathlib.Path, recursive: bool = True) -> None:
        """Apply a directory of manifests that someone has read."""
        args = ['apply', '--server-side', '--force-conflicts', '--field-manager', self.field_manager]
        args = args + (['-R', '-f', str(path)] if recursive else ['-f', str(path)])
        self.run(args)

    def apply_manifest(self, manifest: str, namespace: Optional[str] = None) -> None:
        args = ['apply', '--server-side', '--field-manager', self.field_manager]
        if namespace != None:
            args = args + ['-n', namespace]
        self.run(args + ['-f', '-'], stdin=manifest)

    def create_namespace(self, name: str, labels: Dict[str, str]) -> None:
        """Create a namespace, or adopt it if a previous run left it behind.

        Server-side apply rather than `create` so that a leftover namespace from a run that
        was killed is reused instead of failing the next one.
        """
        rendered = '\n'.join([f'    {k}: {json.dumps(v)}' for k, v in sorted(labels.items())])
        self.apply_manifest(f'apiVersion: v1\nkind: Namespace\nmetadata:\n  name: {name}\n  labels:\n{rendered}\n')

    def delete_by_label(self, resources: str, label: str, value: str) -> None:
        """Delete only what this run labelled.

        The sole deletion path in this module. `resources` is a literal from this package and
        `value` is a run identifier this process generated, so no value derived from a
        component definition can reach it.
        """
        self.run(['delete', resources, '-l', f'{label}={value}', '--ignore-not-found'], check=False)

    def dry_run_apply(self, manifest: str, namespace: Optional[str] = None) -> subprocess.CompletedProcess:
        """Submit a manifest with a server-side dry run, and hand back the whole result.

        Both streams matter and they mean different things, so neither is collapsed here. An
        admission policy bound with `validationActions: [Warn]` reports a failure as an API
        warning on stderr *and lets the request through*, so a rejection on stderr with a
        non-zero exit is a different event from a warning on stderr with exit 0.
        """
        args = ['apply', '--server-side', '--force-conflicts', '--field-manager', self.field_manager]
        if namespace != None:
            args = args + ['-n', namespace]
        return self.run(args + ['--dry-run=server', '-f', '-'], check=False, stdin=manifest)
