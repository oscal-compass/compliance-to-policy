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
"""Kubernetes' built-in ValidatingAdmissionPolicy -- admissionregistration.k8s.io/v1.

The generate half is ordinary. The result half is not, and the reason is worth stating because
it shapes the interface: **VAP has no report API and evaluates nothing retroactively.** There
is no object to read to learn what it thinks of a workload that already exists; the only way
to find out is to re-submit that workload with a server-side dry run and see whether admission
rejects it.

That probing needs a cluster, and C2P does not ship it. It lives in the reference
implementation under `docs/end-to-end/reference/platform_k8s`, which is also where the
credentials and the decision to write to a cluster live. This plugin only converts what such a
probe observed, which is what lets the same results be replayed from disk without a cluster --
and what lets a different probe, or a different operator's script, feed it.

Two consequences a reader should expect. VAP cannot express a control that needs I/O -- a
registry lookup, a transparency log -- so for such a rule there is simply no check declared
for this engine, and the posture reports the gap instead of a false pass. And a probe reads a
pass as the *absence* of a complaint, which a misconfigured probe produces too, so an outcome
it could not stand behind arrives here as `error`, never as `pass`.
"""

from datetime import datetime
from typing import Any, Dict, List, Optional

from pydantic.v1 import Field

from c2p.common.err import C2PError
from c2p.common.logging import getLogger
from c2p.common.utils import get_datetime, get_dict_safely
from c2p.framework.models import Policy, PVPResult, RawResult
from c2p.framework.models.pvp_result import (
    ObservationByCheck,
    Property,
    ResultEnum,
    Subject,
)
from c2p.framework.plugin_spec import PluginConfig, PluginSpec
from plugins_public.plugins import policy_template

logger = getLogger(__name__)

ENGINE = 'vap-cel'
ENGINE_API = 'admissionregistration.k8s.io/v1 ValidatingAdmissionPolicy'

# `pass` means the probe submitted the object, this check was observed to be reporting, and it
# said nothing. `error` means the verdict is unknown -- the probe could not show that it sees
# this check, or the submission was rejected by something the probe does not model -- and must
# not be reported as either compliant or violated.
RESULT_DICTIONARY = {
    'pass': ResultEnum.Pass,
    'fail': ResultEnum.Failure,
    'warning': ResultEnum.Warning,
    'error': ResultEnum.Error,
}


class PluginConfigVap(PluginConfig):
    # Optional because the two halves of this plugin have different needs. Generating policy
    # reads templates and writes files; converting results back into OSCAL does neither, and is
    # a function of the component definition and what the engine reported. Requiring a template
    # directory in order to read results would force a caller that only collects to invent one.
    policy_template_dir: Optional[str] = Field(None, title='Path to Policy template directory')
    deliverable_policy_dir: Optional[str] = Field(None, title='Path to deliverable (generated) policy directory')


class PluginVap(PluginSpec):

    def __init__(self, config: Optional[PluginConfigVap] = None) -> None:
        super().__init__()
        self.config = config

    def generate_pvp_policy(self, policy: Policy):
        if self.config.policy_template_dir == None or self.config.deliverable_policy_dir == None:
            raise C2PError(
                'Generating policy needs policy_template_dir and deliverable_policy_dir. They are '
                'optional only so that converting results does not have to supply them.'
            )
        policy_template.render(
            policy,
            self.config.policy_template_dir,
            self.config.deliverable_policy_dir,
            ENGINE,
        )

    def generate_pvp_result(self, raw_result: RawResult) -> PVPResult:
        """Convert probe outcomes, one per (check, object), into observations per check.

        Each outcome is `{checkId, title, resourceId, result, reason, evaluatedOn, props}` --
        the shape the reference implementation's probe produces.
        """
        outcomes = [o for o in _outcomes(raw_result.data) if get_dict_safely(o, 'checkId') != None]

        observations: List[ObservationByCheck] = []
        for check_id in sorted({o['checkId'] for o in outcomes}):
            observation = ObservationByCheck(check_id=check_id, methods=['TEST-AUTOMATED'], collected=get_datetime())
            observation.subjects = [_to_subject(o) for o in outcomes if o['checkId'] == check_id]
            observations.append(observation)

        pvp_result = PVPResult()
        pvp_result.observations_by_check = observations
        return pvp_result


def _outcomes(data: Any) -> List[Dict[str, Any]]:
    """Accept a bare list of outcomes, or a c2p-evidence/v1 document's subjects."""
    if isinstance(data, dict) and 'observations' in data:
        flattened = []
        for observation in get_dict_safely(data, 'observations', []):
            check_id = get_dict_safely(observation, 'checkId')
            for subject in get_dict_safely(observation, 'subjects', []):
                flattened.append({**subject, 'checkId': check_id})
        return flattened
    return data if isinstance(data, list) else []


def _to_subject(outcome: Dict[str, Any]) -> Subject:
    resource_id = get_dict_safely(outcome, 'resourceId', 'unknown')
    return Subject(
        title=get_dict_safely(outcome, 'title', resource_id),
        type='resource',
        resource_id=resource_id,
        result=RESULT_DICTIONARY.get(get_dict_safely(outcome, 'result'), ResultEnum.Error),
        evaluated_on=_evaluated_on(outcome),
        reason=get_dict_safely(outcome, 'reason', 'no message'),
        props=[Property(name=k, value=v) for k, v in get_dict_safely(outcome, 'props', {}).items()],
    )


def _evaluated_on(outcome: Dict[str, Any]) -> datetime:
    evaluated_on = get_dict_safely(outcome, 'evaluatedOn')
    if evaluated_on == None:
        return get_datetime()
    return datetime.fromisoformat(evaluated_on.replace('Z', '+00:00'))
