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
"""One policy file per check, rendered from a template named after the check.

Shared by the policy engines whose generated artifact is a single document per check, which
is every CEL-based engine here. The older convention of copying a directory tree per rule and
rendering whatever is inside it is kept in `kyverno.py`, because the Kyverno DSL templates in
the wild are shaped that way.

The render context is deliberately structured rather than a flat namespace of parameters: a
generated policy has to carry its own provenance -- which rule and check produced it, by which
engine -- and with a flat namespace a template cannot name any of that.
"""

import json
import pathlib
import re
from typing import Any, Dict, List, Optional

from jinja2 import Environment

from c2p.common.err import C2PError
from c2p.common.logging import getLogger
from c2p.framework.models import Policy
from c2p.framework.models.policy import RuleSet

logger = getLogger(__name__)

TEMPLATE_SUFFIX = '.yaml.tmpl'

# A rule id and a check id both become a path segment here and an object name on the cluster, so
# they are constrained to what Kubernetes accepts as a name. This is not belt-and-braces: the
# ids come from a component definition, which in the authoring flow this plugin was written for
# is produced by an agent, and a value like `../../x` would otherwise send a generated file
# outside the deliverable directory -- over whatever is already there, since rendering writes
# unconditionally. Rejecting at generation time is also kinder than the alternative, which is a
# tree that generates cleanly and is refused by the API server on apply.
DNS_1123_LABEL = re.compile(r'^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')
DNS_1123_MAX = 253


def validate_identifier(kind: str, value: Any) -> str:
    """Accept an identifier that is safe as both a path segment and a Kubernetes object name."""
    if not isinstance(value, str) or value == '' or len(value) > DNS_1123_MAX:
        raise C2PError(f'{kind} must be a non-empty string of at most {DNS_1123_MAX} characters, got {value!r}')
    if DNS_1123_LABEL.match(value) == None:
        raise C2PError(
            f'{kind} {value!r} is not a valid Kubernetes name: lower-case letters, digits and '
            f'"-" only, starting and ending with a letter or digit. It is used as a directory '
            f'name and as the generated object\'s name, so it cannot be an arbitrary string.'
        )
    return value


def _contained(root: pathlib.Path, candidate: pathlib.Path) -> pathlib.Path:
    """The second line of defence: resolve, then prove the result is still under `root`."""
    resolved_root = root.resolve()
    resolved = candidate.resolve()
    if resolved != resolved_root and resolved_root not in resolved.parents:
        raise C2PError(f'{candidate} resolves outside {root}, which generation will not do')
    return resolved


def _quote(value: Any) -> str:
    """A double-quoted scalar, for a template that needs the value to survive as a YAML string."""
    return json.dumps(str(value))


def _tojson(value: Any) -> str:
    """Compact JSON, which for a list of strings is also a valid CEL list literal.

    No space after the separators: the same parameter has to render identically here and in
    any other engine's template, and a CEL expression is compared as a string.
    """
    return json.dumps(value, separators=(',', ':'))


def environment() -> Environment:
    env = Environment(keep_trailing_newline=True)
    env.filters['quote'] = _quote
    env.filters['tojson'] = _tojson
    return env


def render_context(rule_set: RuleSet, parameters: Dict[str, str], engine: str) -> Dict[str, Any]:
    return {
        'engine': engine,
        'rule': {'id': rule_set.rule_id, 'description': rule_set.rule_description},
        'check': {'id': rule_set.check_id, 'description': rule_set.check_description},
        'params': parameters,
    }


def render(
    policy: Policy,
    template_dir: str,
    deliverable_dir: str,
    engine: str,
    suffix: str = TEMPLATE_SUFFIX,
) -> List[pathlib.Path]:
    """Render `<template_dir>/<Check_Id><suffix>` to `<deliverable_dir>/<Rule_Id>/<Check_Id>.yaml`.

    The output path carries the rule as well as the check so that a rule implemented by
    several engines does not collapse into one file, and so the generated tree can be applied
    per rule.
    """
    parameters = {p.id: p.value for p in (policy.parameters if policy.parameters != None else [])}
    env = environment()
    written = []
    for rule_set in policy.rule_sets:
        validate_identifier('Rule_Id', rule_set.rule_id)
        validate_identifier('Check_Id', rule_set.check_id)
        template_path = _contained(
            pathlib.Path(template_dir), pathlib.Path(template_dir) / f'{rule_set.check_id}{suffix}'
        )
        if not template_path.exists():
            raise C2PError(
                f'No template for check {rule_set.check_id} at {template_path}. Every check the '
                f'component definition declares for this engine needs one, or the generated '
                f'policy set silently misses a control.'
            )
        out_dir = pathlib.Path(deliverable_dir)
        out_dir.mkdir(parents=True, exist_ok=True)
        out_path = out_dir / rule_set.rule_id / f'{rule_set.check_id}.yaml'
        out_path.parent.mkdir(parents=True, exist_ok=True)
        _contained(out_dir, out_path)
        template = env.from_string(template_path.open('r').read())
        out_path.write_text(template.render(render_context(rule_set, parameters, engine)))
        written.append(out_path)
        logger.info(f'Generated {out_path}')
    return written
