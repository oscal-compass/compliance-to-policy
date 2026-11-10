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
"""The wgpolicyk8s.io PolicyReport API, read once for every engine that reports through it.

Both Kyverno policy families report here -- `kyverno.io/v1 ClusterPolicy` and the
`policies.kyverno.io/v1` CEL types -- and both run a background scan, so for both of them
collection is a read and a violation that already existed is visible. The generated policy
documents differ by API group; the evidence does not, and duplicating this parser once per
engine is how two engines start disagreeing about what `warn` means.

`policy` on a result is the check id: one generated policy document per check is the
convention of `policy_template`, so the engine's own name for the object it evaluated is
already the check that the component definition declared.
"""

from datetime import datetime, timezone
from typing import Any, Dict, List

from c2p.common.utils import get_datetime, get_dict_safely
from c2p.framework.models.pvp_result import ObservationByCheck, ResultEnum, Subject

POLICY_REPORT_API = 'wgpolicyk8s.io/v1alpha2'
POLICY_REPORT_KINDS = ['PolicyReport', 'ClusterPolicyReport']

# A background-scan `fail` is a violation that already exists, which is where most real
# findings live. `warn` is an object that was admitted when the policy objected to it -- the
# control is not merely violated, it was not enforced -- so it is not folded into Failure.
RESULT_DICTIONARY = {
    'pass': ResultEnum.Pass,
    'fail': ResultEnum.Failure,
    'warn': ResultEnum.Warning,
    'skip': ResultEnum.Error,
    'error': ResultEnum.Error,
}


def reports(data: Any) -> List[Dict[str, Any]]:
    """Accept either a list of report objects or a `kubectl get -o json` envelope."""
    items = data['items'] if isinstance(data, dict) else data
    return [
        item
        for item in items
        if get_dict_safely(item, 'apiVersion') == POLICY_REPORT_API
        and get_dict_safely(item, 'kind') in POLICY_REPORT_KINDS
    ]


def results_with_scope(data: Any) -> List[Dict[str, Any]]:
    """Flatten every report's results, carrying the report's subject down with each one.

    `results[].resources` is routinely absent on these reports -- the PolicyReport names its
    subject once, at the report level, in `scope`. Reading only `resources` yields a posture
    where every subject is an unnamed resource, which is the part a person actually reads.
    """
    results = []
    for report in reports(data):
        scope = get_dict_safely(report, 'scope', {})
        for result in get_dict_safely(report, 'results', []):
            if get_dict_safely(result, 'policy') == None:
                continue
            results.append({**result, '_scope': scope})
    return results


def to_subject(result: Dict[str, Any]) -> Subject:
    resources = get_dict_safely(result, 'resources', [])
    resource = resources[0] if resources != [] and resources != None else get_dict_safely(result, '_scope', {})
    kind = get_dict_safely(resource, 'kind', 'Resource')
    api_version = get_dict_safely(resource, 'apiVersion', 'v1')
    name = get_dict_safely(resource, 'name', '-')
    namespace = get_dict_safely(resource, 'namespace', '-')
    raw_result = get_dict_safely(result, 'result')
    return Subject(
        title=f'{kind} {namespace}/{name}',
        type='resource',
        resource_id=f'{api_version}/{kind}/{namespace}/{name}',
        result=RESULT_DICTIONARY.get(raw_result, ResultEnum.Error),
        evaluated_on=evaluated_on(result),
        reason=get_dict_safely(result, 'message', 'no message'),
        props=[],
    )


def evaluated_on(result: Dict[str, Any]) -> datetime:
    """The report's own timestamp, not collection time: when the engine evaluated it."""
    seconds = get_dict_safely(result, ['timestamp', 'seconds'])
    if seconds == None:
        return get_datetime()
    return datetime.fromtimestamp(int(seconds), tz=timezone.utc)


def observations_by_check(data: Any, methods: List[str]) -> List[ObservationByCheck]:
    """One observation per check, with every object the engine judged under that check."""
    results = results_with_scope(data)
    observations = []
    for check_id in sorted({r['policy'] for r in results}):
        observation = ObservationByCheck(check_id=check_id, methods=methods, collected=get_datetime())
        observation.subjects = [to_subject(r) for r in results if r['policy'] == check_id]
        observations.append(observation)
    return observations
