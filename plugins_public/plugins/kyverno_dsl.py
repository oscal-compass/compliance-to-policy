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
"""Kyverno's earlier policy API: kyverno.io/v1 ClusterPolicy, one document per check.

Separate from `kyverno.py`, which also generates ClusterPolicy, because the two differ in
where the authored decision lives. `kyverno.py` copies a directory tree per rule and
substitutes a flat namespace of parameters into whatever is inside it -- the shape Kyverno DSL
policies in the wild already have. This plugin renders one template per **check**, with the
rule, the check and the engine in the render context, which is what lets a generated policy
carry its own provenance and lets one rule be implemented by several engines without the
output paths colliding. Both are kept: the older convention is still how most existing
template sets are organised, and retiring it would break them.

The policy is deprecated in Kyverno 1.19 and removed in 1.20. That is the argument for
authoring the control rather than the policy: the control definition does not change when an
engine is retired, and `kyverno_cel` generates the replacement from the same component
definition.

Results come from the shared `policy_report` reader: ClusterPolicy reports through the same
wgpolicyk8s.io API as the CEL types.
"""

from typing import List, Optional

from pydantic.v1 import Field

from c2p.common.err import C2PError
from c2p.common.logging import getLogger
from c2p.framework.models import Policy, PVPResult, RawResult
from c2p.framework.plugin_spec import PluginConfig, PluginSpec
from plugins_public.plugins import policy_report, policy_template

logger = getLogger(__name__)

ENGINE = 'kyverno-dsl'
ENGINE_API = 'kyverno.io/v1 ClusterPolicy'

# What a template for this engine is allowed to have produced. Checked after rendering, so a
# value from the component definition cannot add a document of its own.
GENERATED_KINDS: List[str] = ['ClusterPolicy']

# A background scan, so an existing violation is visible without re-submitting the object.
METHODS: List[str] = ['TEST-AUTOMATED']


class PluginConfigKyvernoDsl(PluginConfig):
    # Optional because the two halves of this plugin have different needs. Generating policy
    # reads templates and writes files; converting results back into OSCAL does neither, and is
    # a function of the component definition and what the engine reported. Requiring a template
    # directory in order to read results would force a caller that only collects to invent one.
    policy_template_dir: Optional[str] = Field(None, title='Path to Policy template directory')
    deliverable_policy_dir: Optional[str] = Field(None, title='Path to deliverable (generated) policy directory')


class PluginKyvernoDsl(PluginSpec):

    def __init__(self, config: Optional[PluginConfigKyvernoDsl] = None) -> None:
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
            GENERATED_KINDS,
        )

    def generate_pvp_result(self, raw_result: RawResult) -> PVPResult:
        pvp_result = PVPResult()
        pvp_result.observations_by_check = policy_report.observations_by_check(raw_result.data, METHODS)
        return pvp_result
