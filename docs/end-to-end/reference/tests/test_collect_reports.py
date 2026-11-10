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
"""Collecting from a reporting engine, where collection is a read rather than a probe."""

from platform_k8s import collect


class FakeReportingClient:
    """Returns one recorded PolicyReport whose results name the given policies."""

    def __init__(self, *policies: str) -> None:
        self.policies = policies

    def get_json(self, resources, all_namespaces=False):
        return {
            'items': [
                {
                    'apiVersion': 'wgpolicyk8s.io/v1alpha2',
                    'kind': 'PolicyReport',
                    'scope': {'kind': 'Deployment', 'name': 'web', 'namespace': 'storefront'},
                    'results': [{'policy': p, 'result': 'pass'} for p in self.policies],
                }
            ]
        }


def test_collecting_for_one_engine_drops_another_engines_checks():
    """Two engines report to the same PolicyReport API, so collection has to be scoped.

    Unfiltered, each engine's plugin keys observations by `results[].policy` and therefore
    reports the other engine's checks as well -- two observations per check, with the same
    subjects counted twice.
    """
    client = FakeReportingClient('kyverno-dsl-a', 'kyverno-cel-a')

    kept = collect.collect_policy_reports(client, ['kyverno-dsl-a'])

    assert [r['policy'] for r in kept[0]['results']] == ['kyverno-dsl-a']


def test_a_report_with_no_wanted_check_is_dropped_entirely():
    client = FakeReportingClient('kyverno-cel-a')

    assert collect.collect_policy_reports(client, ['kyverno-dsl-a']) == []


def test_the_report_scope_survives_filtering():
    """A subject comes from the report's `scope` when a result names no resource, so filtering
    the results must not discard the report around them."""
    client = FakeReportingClient('kyverno-dsl-a', 'kyverno-cel-a')

    kept = collect.collect_policy_reports(client, ['kyverno-dsl-a'])

    assert kept[0]['scope']['name'] == 'web'


def test_omitting_check_ids_collects_everything():
    """Still correct for a cluster where only one engine reports here."""
    client = FakeReportingClient('kyverno-dsl-a', 'kyverno-cel-a')

    kept = collect.collect_policy_reports(client)

    assert [r['policy'] for r in kept[0]['results']] == ['kyverno-dsl-a', 'kyverno-cel-a']


class FakeFailingClient:
    """get_json as the rewritten client implements it, for the two kinds of failure."""

    def __init__(self, stderr: str) -> None:
        self.stderr = stderr

    def get_json(self, resources, namespace=None, all_namespaces=False):
        from c2p.common.err import C2PError

        if 'the server doesn\'t have a resource type' in self.stderr or 'NotFound' in self.stderr:
            return {'items': []}
        raise C2PError(f'kubectl get {resources} failed, so this run cannot be trusted: {self.stderr}')


def test_an_engine_that_is_not_installed_reports_nothing():
    client = FakeFailingClient('error: the server doesn\'t have a resource type "policyreports"')

    assert collect.collect_policy_reports(client, ['a']) == []


def test_a_cluster_that_could_not_be_read_is_not_an_empty_result_set():
    """The failure this prevents: an authentication error or the wrong context collected as
    "no findings", which renders as a uniformly not-assessed posture that looks like a finding
    rather than like a broken run."""
    import pytest

    from c2p.common.err import C2PError

    client = FakeFailingClient('error: You must be logged in to the server (Unauthorized)')

    with pytest.raises(C2PError, match='cannot be trusted'):
        collect.collect_policy_reports(client, ['a'])
