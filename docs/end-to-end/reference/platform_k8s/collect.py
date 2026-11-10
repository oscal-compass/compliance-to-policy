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
"""Gather what the engines on a cluster saw, in the shape their plugin expects.

Two shapes, because two kinds of engine:

A **reporting** engine writes its findings to the API and runs a background scan, so collection
is a read and it can see a violation that already exists. Kyverno is one.

An **admission-only** engine has no report API and evaluates nothing retroactively, so there is
nothing to read: the object has to be re-submitted through admission to find out what the engine
thinks of it. ValidatingAdmissionPolicy is one. That probe is the awkward case and most of this
module is about doing it without touching anything that was already running.

The probe adds; it never removes
--------------------------------

A `validationActions: [Deny]` binding stops the request at the first policy that objects, so a
single submission reveals one verdict and leaves the rest unknown. The obvious way out -- unbind
the other policies, probe one check, put them back -- is what an earlier version of this file
did, and it is unacceptable: the bindings are cluster-scoped, so every probe opened a window in
which the cluster was not enforcing, and a process killed mid-probe left that window open.

A `validationActions: [Warn]` binding reports a failure as an API warning and *admits* the
request, so nothing short-circuits and every bound policy gets to speak. So this module creates
an additional Warn binding per check, scoped by `namespaceSelector` to a scratch namespace it
owns, and submits a copy of the object there as a server-side dry run. Measured on Kubernetes
v1.34.0 against four policies and one object that violates all four:

    Deny, one dry run                  1 of 4 verdicts recovered
    added Warn bindings, one dry run   4 of 4 verdicts recovered, exit 0, nothing deleted

The scratch namespace is not decoration. A *mutating* webhook runs before validating admission
policies, so when one rejects the submission the API server never evaluates the policies at all
-- measured: with a Kyverno image-verification policy denying, 0 of 4 verdicts came back. The
demo's Kyverno rules are matched on `namespaces: [storefront]`, so a namespace of our own is
outside their reach and the engine under test is the only one with an opinion.

Everything this module creates carries `compliance.oscal/probe-run=<run id>`, and teardown is a
delete by that label. No identifier from a component definition reaches a delete.

A pass is the absence of a warning, so the probe proves it can see
------------------------------------------------------------------

The weakness of reading verdicts from warnings is that silence is the pass signal, and a
misconfigured probe is also silent: a selector that matches nothing, a policy that failed to
compile, a `matchConstraints` that does not cover the resource. That failure mode hands back a
clean bill of health for a cluster nobody looked at, which is the most expensive way for this
code to be wrong.

So each check is first put to a **canary**: an object that is known to violate it. If the canary
does not produce the check's warning, the probe does not know how to see that check and says so
-- every subject for it becomes `error`. A check with no canary on disk is `error` too. There is
deliberately no flag to skip this.
"""

import datetime
import pathlib
import re
import time
import uuid
from typing import Any, Dict, List, Optional, Sequence, Tuple

import yaml

from c2p.common.err import C2PError
from c2p.common.logging import getLogger
from plugins_public.plugins.policy_template import validate_identifier

from .client import PROBE_RUN_LABEL, KubeClient

logger = getLogger(__name__)

POLICY_REPORT_RESOURCES = 'policyreports.wgpolicyk8s.io,clusterpolicyreports.wgpolicyk8s.io'

RESULT_PASS = 'pass'
RESULT_FAIL = 'fail'
RESULT_ERROR = 'error'

VAP_BINDING_RESOURCE = 'validatingadmissionpolicybindings.admissionregistration.k8s.io'

DEFAULT_PROBE_RESOURCES = ['deployments', 'statefulsets']

CANARY_SUFFIX = '.canary.yaml'

# A binding does not take effect the moment it is created: the API server compiles and caches
# admission policies, and a submission in between is evaluated against the previous set. Sleeping
# a fixed time instead would either waste it or, worse, measure the wrong set and call the result
# a pass. The canary already asks exactly the right question -- "does this check report yet?" --
# so it is also the settle signal, and it is polled rather than waited on.
#
# Measured: on an idle single-node cluster the first attempt usually succeeds; on the loaded demo
# cluster one check of four needed a second or two. Omitting this entirely reported that check as
# `error` for every workload, which is the canary doing its job but not a run anyone wants.
DEFAULT_SETTLE_TRIES = 30
SETTLE_DELAY_SECONDS = 1.0

# `Warning: Validation failed for ValidatingAdmissionPolicy 'x' with binding 'y': message`
WARNING_PATTERN = re.compile(
    r"^Warning: Validation failed for ValidatingAdmissionPolicy '(?P<policy>[^']*)' "
    r"with binding '(?P<binding>[^']*)': (?P<message>.*)$"
)

# Fields that describe where an object has been rather than what it is. They have to go before
# the copy is submitted somewhere else: the namespace and name are owned by the live object, and
# `last-applied-configuration` is a second copy of the whole object that no engine reads.
PROVENANCE_FIELDS = (
    'uid',
    'resourceVersion',
    'generation',
    'creationTimestamp',
    'managedFields',
    'ownerReferences',
    'selfLink',
    'finalizers',
)
PROVENANCE_ANNOTATIONS = (
    'kubectl.kubernetes.io/last-applied-configuration',
    'deployment.kubernetes.io/revision',
)

# Enough of a rejection to tell a reader which engine spoke, without copying a submitted object
# into the report. Server-side apply does not echo the object, but an engine is free to quote it
# back in its own message, and that message ends up in Assessment Results.
REASON_LIMIT = 300


def _now() -> str:
    return datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')


def _summarise(stderr: str) -> str:
    """One line of an engine's message, capped. Never the whole stream."""
    lines = [line for line in stderr.splitlines() if line.strip() != '' and not line.startswith('Warning:')]
    first = lines[0].strip() if lines != [] else 'no message'
    return first if len(first) <= REASON_LIMIT else first[:REASON_LIMIT] + '... [truncated]'


def collect_policy_reports(
    client: KubeClient,
    check_ids: Optional[Sequence[str]] = None,
) -> List[Dict[str, Any]]:
    """PolicyReports and ClusterPolicyReports, ready to hand to a reporting engine's plugin.

    `check_ids` is how the caller says which engine is collecting, and omitting it is almost
    always wrong when more than one engine reports here. Every Kyverno policy family writes to
    the same PolicyReport API, and a plugin keys its observations by `results[].policy`, which
    is the check id -- so handing the unfiltered set to two engines' plugins makes each of them
    report the other's checks as well. The Assessment Results then carries two observations per
    check with the same subjects counted twice, which is invisible in a posture that indexes by
    check id and wrong in anything that counts.

    Filtering here rather than in the plugin keeps the evidence on disk engine-specific, so a
    later `result2oscal` over `results/<engine>/` stays a pure function of what was collected.
    """
    reports = client.get_json(POLICY_REPORT_RESOURCES, all_namespaces=True)
    items = reports.get('items', [])
    if check_ids == None:
        logger.info(f'Collected {len(items)} policy report(s), unfiltered')
        return items

    wanted = set(check_ids)
    kept = []
    for item in items:
        results = [r for r in item.get('results', []) if r.get('policy') in wanted]
        if results != []:
            kept.append({**item, 'results': results})
    logger.info(f'Collected {len(kept)} policy report(s) covering {len(wanted)} check(s)')
    return kept


def probe_admission_policies(
    client: KubeClient,
    check_ids: Sequence[str],
    namespace: str,
    canary_dir: str,
    resources: Sequence[str] = tuple(DEFAULT_PROBE_RESOURCES),
    run_id: Optional[str] = None,
    settle_tries: int = DEFAULT_SETTLE_TRIES,
) -> List[Dict[str, Any]]:
    """Ask an admission-only engine what it thinks of the objects already in `namespace`.

    Creates a scratch namespace and one `Warn` binding per check, submits each object there as
    a server-side dry run, and reads the verdicts out of the API warnings. Existing policies
    and bindings are never read, modified or deleted.

    `canary_dir` holds `<check_id>.canary.yaml` per check: an object that must fail the check.
    A check whose canary does not fail is reported as `error`, because this probe cannot
    distinguish "the object passed" from "nothing evaluated the object".

    What is left behind if this process is killed: a namespace containing nothing, and bindings
    that deny nothing. Enforcement is untouched throughout.
    """
    for check_id in check_ids:
        validate_identifier('Check_Id', check_id)

    run = run_id if run_id != None else uuid.uuid4().hex[:8]
    scratch = f'c2p-probe-{run}'
    validate_identifier('probe namespace', scratch)

    objects = []
    for resource in resources:
        objects = objects + client.names(resource, namespace=namespace)
    if objects == []:
        logger.warning(f'No {", ".join(resources)} in namespace {namespace}: nothing to probe')
        return []

    try:
        client.create_namespace(scratch, {PROBE_RUN_LABEL: run, 'kubernetes.io/metadata.name': scratch})
        for check_id in check_ids:
            client.apply_manifest(_warn_binding(check_id, scratch, run))
        logger.info(f'Probe {run}: {len(check_ids)} Warn binding(s) added, nothing removed')

        visible, canary_reasons = _canary(client, check_ids, canary_dir, scratch, run, settle_tries)
        outcomes: List[Dict[str, Any]] = []
        for reference in objects:
            outcomes = outcomes + _probe_object(
                client, check_ids, reference, namespace, scratch, run, visible, canary_reasons
            )
    finally:
        client.delete_by_label(VAP_BINDING_RESOURCE, PROBE_RUN_LABEL, run)
        client.delete_by_label('namespaces', PROBE_RUN_LABEL, run)
        logger.info(f'Probe {run}: removed what this run created, by label')

    logger.info(f'Probed {len(check_ids)} check(s) against {len(objects)} object(s)')
    return outcomes


def _warn_binding(check_id: str, scratch: str, run: str) -> str:
    """A second binding for a policy c2p generated, reporting instead of denying.

    Named after the run as well as the check so that two probes cannot collide, and so the
    warnings this run reads can be told apart from any other Warn binding on the same policy.
    """
    return (
        'apiVersion: admissionregistration.k8s.io/v1\n'
        'kind: ValidatingAdmissionPolicyBinding\n'
        'metadata:\n'
        f'  name: c2p-probe-{run}-{check_id}\n'
        '  labels:\n'
        f'    {PROBE_RUN_LABEL}: {run}\n'
        'spec:\n'
        f'  policyName: {check_id}\n'
        '  validationActions: [Warn]\n'
        '  matchResources:\n'
        '    namespaceSelector:\n'
        '      matchLabels:\n'
        f'        kubernetes.io/metadata.name: {scratch}\n'
    )


def _warnings_by_check(stderr: str, check_ids: Sequence[str], run: str) -> Dict[str, List[str]]:
    """Group this run's warnings by check.

    Keyed on the binding name as well as the policy name: another Warn binding on the same
    policy, left by someone else, would otherwise be read as this probe's own measurement.
    One failed validation expression is one warning, so a check with several expressions
    contributes several messages.
    """
    found: Dict[str, List[str]] = {check_id: [] for check_id in check_ids}
    for line in stderr.splitlines():
        match = WARNING_PATTERN.match(line.strip())
        if match == None:
            continue
        if match.group('binding') != f'c2p-probe-{run}-{match.group("policy")}':
            continue
        if match.group('policy') in found:
            found[match.group('policy')].append(match.group('message').strip())
    return found


def _canary(
    client: KubeClient,
    check_ids: Sequence[str],
    canary_dir: str,
    scratch: str,
    run: str,
    settle_tries: int = DEFAULT_SETTLE_TRIES,
) -> Tuple[Dict[str, bool], Dict[str, str]]:
    """Establish, per check, that this probe can see the check fail.

    Until that is shown, the absence of a warning means nothing. Each canary is submitted on
    its own: a canary is written to break one check, and sending them together would let one
    object's warnings vouch for another check's visibility.
    """
    visible: Dict[str, bool] = {}
    reasons: Dict[str, str] = {}
    for check_id in check_ids:
        path = pathlib.Path(canary_dir) / f'{check_id}{CANARY_SUFFIX}'
        if not path.is_file():
            visible[check_id] = False
            reasons[check_id] = (
                f'no canary at {path}, so a pass could not be told apart from an unevaluated '
                f'object; write an object that violates {check_id}'
            )
            continue
        manifest = _retarget(path.read_text(), scratch)
        completed, seen = _canary_until_it_reports(client, manifest, check_id, scratch, run, settle_tries)
        visible[check_id] = seen != []
        if seen != []:
            continue
        # Two different faults, and they call for different repairs, so they are not merged.
        # A canary the API server refuses outright never reaches admission -- an object that is
        # invalid rather than non-compliant, which is easy to write by accident because the
        # properties a check objects to and the properties Kubernetes forbids overlap. A canary
        # that is admitted in silence means the check is not watching this namespace.
        if completed.returncode != 0:
            reasons[check_id] = (
                f'the canary for {check_id} was rejected before any policy evaluated it, so it '
                f'cannot show that this check is visible; fix the canary object: '
                f'{_summarise(completed.stderr)}'
            )
        else:
            reasons[check_id] = (
                f'the canary for {check_id} was admitted without this check reporting it, so '
                f'this probe cannot see the check after {settle_tries} attempt(s): its binding, '
                f'the policy, or the policy\'s matchConstraints did not take effect for {path.name}'
            )
    seeing = len([c for c in check_ids if visible[c]])
    logger.info(f'Probe {run}: canaries confirm {seeing} of {len(check_ids)} check(s) are visible')
    return visible, reasons


def _canary_until_it_reports(client, manifest, check_id, scratch, run, settle_tries):
    """Submit the canary until the check reports it, or until the attempts run out.

    A rejection is not retried: that is a canary which is invalid rather than one whose binding
    has not landed, and waiting will not change it.
    """
    completed = None
    for attempt in range(max(settle_tries, 1)):
        completed = client.dry_run_apply(manifest, namespace=scratch)
        if completed.returncode != 0:
            return completed, []
        seen = _warnings_by_check(completed.stderr, [check_id], run)[check_id]
        if seen != []:
            if attempt != 0:
                logger.info(f'Probe {run}: {check_id} became visible on attempt {attempt + 1}')
            return completed, seen
        if attempt != max(settle_tries, 1) - 1:
            time.sleep(SETTLE_DELAY_SECONDS)
    return completed, []


def _retarget(manifest: str, namespace: str) -> str:
    """Strip where the object has been, and point the copy at the scratch namespace.

    Submitting a `kubectl get -o yaml` dump unaltered does not work: kubectl refuses on the
    namespace mismatch, and `status`, `managedFields` and `resourceVersion` belong to the live
    object rather than to the question being asked of it.
    """
    document = yaml.safe_load(manifest)
    if not isinstance(document, dict) or 'metadata' not in document:
        raise C2PError('A probe target has to be a single Kubernetes object with metadata')
    document.pop('status', None)
    metadata = document['metadata']
    for field in PROVENANCE_FIELDS:
        metadata.pop(field, None)
    annotations = metadata.get('annotations', {})
    for annotation in PROVENANCE_ANNOTATIONS:
        annotations.pop(annotation, None)
    if annotations == {}:
        metadata.pop('annotations', None)
    metadata['namespace'] = namespace
    return yaml.safe_dump(document, sort_keys=False)


def _probe_object(
    client: KubeClient,
    check_ids: Sequence[str],
    reference: str,
    namespace: str,
    scratch: str,
    run: str,
    visible: Dict[str, bool],
    canary_reasons: Dict[str, str],
) -> List[Dict[str, Any]]:
    """One submission, one verdict per check."""
    manifest = client.get_manifest(reference, namespace=namespace)
    if manifest == None:
        logger.warning(f'{reference} in {namespace} could not be read: skipped')
        return []

    completed = client.dry_run_apply(_retarget(manifest, scratch), namespace=scratch)
    rejected = completed.returncode != 0
    found = _warnings_by_check(completed.stderr, check_ids, run)

    outcomes = []
    for check_id in check_ids:
        if not visible[check_id]:
            result, reason = RESULT_ERROR, canary_reasons[check_id]
        elif rejected:
            # Nothing in the scratch namespace should deny, so this is something the probe does
            # not model -- a cluster-wide webhook, a quota, a conversion failure. The engine
            # under test never gave a verdict, so neither do we.
            result = RESULT_ERROR
            reason = (
                f'the probe submission for {reference} was rejected, so {check_id} never '
                f'evaluated it: {_summarise(completed.stderr)}'
            )
        elif found[check_id] != []:
            result, reason = RESULT_FAIL, '; '.join(found[check_id])
        else:
            result = RESULT_PASS
            reason = f'admitted with {check_id} reporting, and the canary for it was reported as failing'
        outcomes.append(
            {
                'checkId': check_id,
                'title': f'{reference} in {namespace}',
                'resourceId': f'{reference}/{namespace}',
                'result': result,
                'reason': reason,
                'evaluatedOn': _now(),
                'props': {'engine.object': check_id, 'engine.action': 'warn'},
            }
        )
    return outcomes
