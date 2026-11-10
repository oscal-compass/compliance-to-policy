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
"""The admission probe, against a recorded cluster rather than a real one.

Two properties carry the design, so most of these tests are about them rather than about the
returned verdicts:

* **It adds and never removes.** The probe must not delete, patch or read any binding it did not
  create, and what it does delete must be selected by the label it wrote during this run. A test
  that only checked the output would pass for code that unbinds half the cluster first.
* **A pass is earned.** The verdict for an admitted object is the absence of a warning, which a
  misconfigured probe also produces. Every pass therefore has to be backed by a canary the check
  was observed to reject.

The behaviour here was first measured against Kubernetes v1.34.0; see `platform_k8s/collect.py`
for those numbers. These tests pin the logic that interprets it.
"""

import pathlib
import subprocess

import pytest
import yaml
from platform_k8s import collect

from c2p.common.err import C2PError

RUN = 'r1'
CHECKS = ['check-a', 'check-b']


def _warning(check_id: str, run: str = RUN, binding: str = None) -> str:
    bound = binding if binding != None else f'c2p-probe-{run}-{check_id}'
    return (
        f"Warning: Validation failed for ValidatingAdmissionPolicy '{check_id}' "
        f"with binding '{bound}': {check_id} says no"
    )


def _canaries(directory: pathlib.Path, check_ids=CHECKS) -> str:
    """One canary file per check, named so the fake cluster can tell which check it breaks."""
    for check_id in check_ids:
        path = directory / f'{check_id}{collect.CANARY_SUFFIX}'
        path.write_text(f'apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: canary-{check_id}\nspec: {{}}\n')
    return str(directory)


class FakeClient:
    """A cluster that records what was asked of it.

    `object_fails` and `canary_fails` say which checks report a failure for which submission;
    `reject` makes a submission fail outright, as an engine or the API server would.
    """

    def __init__(self, object_fails=None, canary_fails=None, reject=None, foreign_warnings=(), run=RUN):
        self.object_fails = object_fails if object_fails != None else {}
        self.canary_fails = set(canary_fails) if canary_fails != None else set(CHECKS)
        self.reject = reject if reject != None else {}
        self.foreign_warnings = foreign_warnings
        self.run = run
        self.calls = []
        self.submissions = []

    def names(self, resources, namespace=None):
        self.calls.append(f'names {resources}')
        return ['deployment.apps/target'] if resources == 'deployments' else []

    def get_manifest(self, reference, namespace=None):
        self.calls.append(f'get_manifest {reference}')
        return (
            'apiVersion: apps/v1\nkind: Deployment\n'
            'metadata:\n'
            '  name: target\n'
            '  namespace: storefront\n'
            '  uid: 0b7d674e\n'
            '  resourceVersion: "9967"\n'
            '  managedFields: [{manager: kubectl}]\n'
            '  annotations:\n'
            '    kubectl.kubernetes.io/last-applied-configuration: |\n'
            '      {"spec":{"env":[{"name":"POSTGRES_PASSWORD","value":"not-a-real-password"}]}}\n'
            'spec: {}\n'
            'status: {replicas: 1}\n'
        )

    def create_namespace(self, name, labels):
        self.calls.append(f'create_namespace {name} {sorted(labels.items())}')

    def apply_manifest(self, manifest, namespace=None):
        document = yaml.safe_load(manifest)
        self.calls.append(f"apply_manifest {document['kind']}/{document['metadata']['name']}")

    def delete_by_label(self, resources, label, value):
        self.calls.append(f'delete_by_label {resources} {label}={value}')

    def dry_run_apply(self, manifest, namespace=None):
        document = yaml.safe_load(manifest)
        name = document['metadata']['name']
        self.submissions.append(f'{name} -> {namespace}')
        if name in self.reject:
            return subprocess.CompletedProcess(args=[], returncode=1, stdout='', stderr=self.reject[name])
        if name.startswith('canary-'):
            check_id = name[len('canary-') :]
            failing = [check_id] if check_id in self.canary_fails else []
        else:
            failing = sorted(self.object_fails.get(name, []))
        lines = [_warning(c, self.run) for c in failing] + list(self.foreign_warnings)
        return subprocess.CompletedProcess(args=[], returncode=0, stdout='', stderr='\n'.join(lines))


def _probe(client, canary_dir, check_ids=CHECKS, settle_tries=1):
    return collect.probe_admission_policies(
        client,
        check_ids,
        namespace='storefront',
        canary_dir=canary_dir,
        resources=['deployments'],
        run_id=RUN,
        settle_tries=settle_tries,
    )


def _by_check(outcomes):
    return {o['checkId']: o['result'] for o in outcomes}


# --- it adds, and never removes --------------------------------------------------------------


def test_the_probe_deletes_nothing_by_name_and_only_its_own_label(tmp_path):
    """The failure this guards against is a probe that unbinds a policy to isolate another, which
    leaves the cluster unenforced for as long as the probe runs and for ever if it is killed."""
    client = FakeClient()

    _probe(client, _canaries(tmp_path))

    deletes = [c for c in client.calls if c.startswith('delete')]
    assert deletes == [
        f'delete_by_label {collect.VAP_BINDING_RESOURCE} {collect.PROBE_RUN_LABEL}={RUN}',
        f'delete_by_label namespaces {collect.PROBE_RUN_LABEL}={RUN}',
    ]
    assert not any('delete_resource' in call for call in client.calls)


def test_the_only_bindings_touched_are_the_ones_this_run_created(tmp_path):
    client = FakeClient()

    _probe(client, _canaries(tmp_path))

    applied = [c for c in client.calls if c.startswith('apply_manifest')]
    assert applied == [
        'apply_manifest ValidatingAdmissionPolicyBinding/c2p-probe-r1-check-a',
        'apply_manifest ValidatingAdmissionPolicyBinding/c2p-probe-r1-check-b',
    ]


def test_every_submission_goes_to_the_scratch_namespace(tmp_path):
    client = FakeClient()

    _probe(client, _canaries(tmp_path))

    assert [s.split(' -> ')[1] for s in client.submissions] == [f'c2p-probe-{RUN}'] * 3


def test_what_this_run_created_is_removed_even_when_the_probe_raises(tmp_path):
    client = FakeClient()
    client.get_manifest = lambda reference, namespace=None: (_ for _ in ()).throw(C2PError('boom'))

    with pytest.raises(C2PError):
        _probe(client, _canaries(tmp_path))

    assert [c for c in client.calls if c.startswith('delete_by_label')] == [
        f'delete_by_label {collect.VAP_BINDING_RESOURCE} {collect.PROBE_RUN_LABEL}={RUN}',
        f'delete_by_label namespaces {collect.PROBE_RUN_LABEL}={RUN}',
    ]


def test_one_submission_answers_every_check(tmp_path):
    """With Deny bindings a submission reveals one verdict, which is why the old probe isolated
    each check in turn. Reporting bindings do not short-circuit, so the object is sent once."""
    client = FakeClient(object_fails={'target': CHECKS})

    outcomes = _probe(client, _canaries(tmp_path))

    assert len([s for s in client.submissions if s.startswith('target')]) == 1
    assert _by_check(outcomes) == {'check-a': 'fail', 'check-b': 'fail'}


# --- a pass is earned ------------------------------------------------------------------------


def test_an_admitted_object_passes_a_check_whose_canary_was_rejected(tmp_path):
    client = FakeClient(object_fails={'target': ['check-b']})

    outcomes = _probe(client, _canaries(tmp_path))

    assert _by_check(outcomes) == {'check-a': 'pass', 'check-b': 'fail'}


def test_a_check_whose_canary_is_admitted_in_silence_is_an_error_never_a_pass(tmp_path):
    """The whole point. A binding that matched nothing, a policy that did not compile and a
    compliant object are indistinguishable from the probe's side, so the canary decides."""
    client = FakeClient(canary_fails=['check-b'])

    outcomes = _probe(client, _canaries(tmp_path))

    assert _by_check(outcomes) == {'check-a': 'error', 'check-b': 'pass'}
    reason = [o['reason'] for o in outcomes if o['checkId'] == 'check-a'][0]
    assert 'admitted without this check reporting it' in reason


def test_a_canary_is_retried_until_the_binding_has_propagated(monkeypatch, tmp_path):
    """The API server compiles and caches admission policies, so a binding created a moment ago
    is not yet in force. Without this, a check whose binding had not landed was reported as
    `error` for every workload -- which is the canary refusing to guess, but not a usable run."""
    monkeypatch.setattr(collect, 'SETTLE_DELAY_SECONDS', 0)
    client = FakeClient(canary_fails=[])
    attempts = {'n': 0}
    inner = client.dry_run_apply

    def flaky(manifest, namespace=None):
        if yaml.safe_load(manifest)['metadata']['name'].startswith('canary-'):
            attempts['n'] = attempts['n'] + 1
            if attempts['n'] >= 3:
                client.canary_fails = set(CHECKS)
        return inner(manifest, namespace=namespace)

    client.dry_run_apply = flaky

    outcomes = _probe(client, _canaries(tmp_path), settle_tries=5)

    assert attempts['n'] >= 3
    assert _by_check(outcomes) == {'check-a': 'pass', 'check-b': 'pass'}


def test_a_canary_that_never_reports_is_still_an_error_after_the_retries(monkeypatch, tmp_path):
    monkeypatch.setattr(collect, 'SETTLE_DELAY_SECONDS', 0)
    client = FakeClient(canary_fails=[])

    outcomes = _probe(client, _canaries(tmp_path), settle_tries=3)

    assert _by_check(outcomes) == {'check-a': 'error', 'check-b': 'error'}
    assert 'after 3 attempt(s)' in outcomes[0]['reason']


def test_a_rejected_canary_is_not_retried(monkeypatch, tmp_path):
    """A canary the API server refuses is invalid, not late. Waiting cannot fix it."""
    monkeypatch.setattr(collect, 'SETTLE_DELAY_SECONDS', 0)
    client = FakeClient(reject={'canary-check-a': 'The Deployment is invalid: cannot set both'})

    _probe(client, _canaries(tmp_path), settle_tries=9)

    assert len([s for s in client.submissions if s.startswith('canary-check-a')]) == 1


def test_a_check_with_no_canary_on_disk_is_an_error(tmp_path):
    client = FakeClient()

    outcomes = _probe(client, _canaries(tmp_path, ['check-b']))

    assert _by_check(outcomes)['check-a'] == 'error'
    assert 'no canary at' in [o['reason'] for o in outcomes if o['checkId'] == 'check-a'][0]


def test_a_canary_the_cluster_refuses_is_a_different_error_from_a_silent_one(tmp_path):
    """An object that is invalid rather than non-compliant never reaches admission. Easy to write
    by accident, because what a check objects to and what Kubernetes forbids overlap."""
    client = FakeClient(reject={'canary-check-a': 'The Deployment is invalid: cannot set both'})

    outcomes = _probe(client, _canaries(tmp_path))

    reason = [o['reason'] for o in outcomes if o['checkId'] == 'check-a'][0]
    assert 'rejected before any policy evaluated it' in reason
    assert 'fix the canary object' in reason


def test_a_warning_from_someone_elses_binding_is_not_this_probes_measurement(tmp_path):
    """Another Warn binding on the same policy would otherwise be read as the probe's own, and a
    check nobody in this run observed would report a failure."""
    client = FakeClient(foreign_warnings=[_warning('check-a', binding='someone-elses-binding')])

    outcomes = _probe(client, _canaries(tmp_path, []))

    assert _by_check(outcomes) == {'check-a': 'error', 'check-b': 'error'}


def test_a_rejected_submission_is_an_error_for_every_check(tmp_path):
    """Nothing in the scratch namespace should deny, so a rejection is something the probe does
    not model. The engine under test gave no verdict, so neither does the probe."""
    client = FakeClient(reject={'target': 'admission webhook "other.svc" denied the request'})

    outcomes = _probe(client, _canaries(tmp_path))

    assert _by_check(outcomes) == {'check-a': 'error', 'check-b': 'error'}
    assert 'never evaluated it' in outcomes[0]['reason']


# --- what leaves the cluster in a report ------------------------------------------------------


def test_the_submitted_copy_carries_no_provenance_and_no_second_copy_of_itself(tmp_path):
    client = FakeClient()

    _probe(client, _canaries(tmp_path))
    retargeted = collect._retarget(client.get_manifest('deployment.apps/target'), 'scratch')
    document = yaml.safe_load(retargeted)

    assert document['metadata']['namespace'] == 'scratch'
    assert 'status' not in document
    for field in ['uid', 'resourceVersion', 'managedFields', 'annotations']:
        assert field not in document['metadata'], field
    assert 'not-a-real-password' not in retargeted


def test_a_rejection_reason_is_summarised_rather_than_copied(tmp_path):
    """An engine is free to quote the submitted object back in its message, and that message is
    what reaches Assessment Results."""
    noisy = 'denied: ' + 'x' * 5000 + '\nand another line'
    client = FakeClient(reject={'target': noisy})

    outcomes = _probe(client, _canaries(tmp_path))

    for outcome in outcomes:
        assert len(outcome['reason']) < collect.REASON_LIMIT + 200
        assert 'and another line' not in outcome['reason']


def test_a_check_id_that_is_not_a_kubernetes_name_is_refused_before_anything_is_created(tmp_path):
    client = FakeClient()

    with pytest.raises(C2PError, match='not a valid Kubernetes name'):
        _probe(client, _canaries(tmp_path), check_ids=['../../escaped'])

    assert client.calls == []


def test_no_objects_to_probe_is_not_an_error(tmp_path):
    client = FakeClient()
    client.names = lambda resources, namespace=None: []

    assert _probe(client, _canaries(tmp_path)) == []
