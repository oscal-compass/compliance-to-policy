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
from typing import Any, Collection, Dict, List, Optional

import yaml

from jinja2 import Environment, StrictUndefined
from jinja2.sandbox import SandboxedEnvironment

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


# Every character YAML 1.1 treats as a line break. A value that carries one can leave the
# comment it was written into and become a document of its own -- and OSCAL validation is no
# help here: trestle refuses LF in a prop value but accepts CR, U+2028 and U+0085 (measured
# with trestle 3.12.4).
_LINE_BREAKS = re.compile(r'[\r\n\x0b\x0c\x85\u2028\u2029]+')


def _oneline(value: Any) -> str:
    """One line of text, for a value that is going into a comment rather than a field."""
    return _LINE_BREAKS.sub(' ', str(value)).strip()


def environment() -> Environment:
    """A sandboxed environment, because a template is executable and this one is not ours.

    Jinja2's default `Environment` gives a template the full Python object graph: a template
    can walk an attribute chain to a module and call it, so rendering one is running it. In
    the authoring flow this plugin was written for the templates are produced by an agent from
    a component definition, a catalog and whatever documents it was pointed at, which is
    untrusted input by any reading. Rendering happens on a developer's machine or in CI, where
    there is a kubeconfig and there are secrets.

    Reading the generated policy does not substitute for this. The reading happens *after* the
    render, and the skill tells the reader to read the output rather than the template.

    `StrictUndefined` for the second half of the same problem: a mistyped `params.foo` is an
    empty string under the default `Undefined`, so a check silently asserts nothing and every
    workload passes it.
    """
    env = SandboxedEnvironment(keep_trailing_newline=True, undefined=StrictUndefined)
    env.filters['quote'] = _quote
    env.filters['tojson'] = _tojson
    env.filters['oneline'] = _oneline
    return env


def render_context(rule_set: RuleSet, parameters: Dict[str, str], engine: str) -> Dict[str, Any]:
    return {
        'engine': engine,
        'rule': {'id': rule_set.rule_id, 'description': rule_set.rule_description},
        'check': {'id': rule_set.check_id, 'description': rule_set.check_description},
        'params': parameters,
    }


def validate_rendered(text: str, check_id: str, allowed_kinds: Collection[str]) -> None:
    """Refuse a rendered document set that is not the one the check asked for.

    Rendering substitutes values from a component definition into a template, and both come
    from an agent in the authoring flow this plugin was written for. A value carrying a line
    break or a quote can close the construct it was written into and open another, so what a
    template says it emits and what it actually emitted are different questions. This asks
    the second one, of the text that is about to be written.

    Not a substitute for quoting in the template -- it is the check that quoting happened.
    """
    try:
        documents = list(yaml.safe_load_all(text))
    except yaml.YAMLError as e:
        raise C2PError(
            f'Check {check_id} rendered to text that is not valid YAML ({e.__class__.__name__}). '
            f'A value substituted into the template most likely carried a quote or a line break.'
        )
    documents = [d for d in documents if d != None]
    if documents == []:
        raise C2PError(f'Check {check_id} rendered no document at all.')
    for document in documents:
        if not isinstance(document, dict):
            raise C2PError(f'Check {check_id} rendered a {type(document).__name__}, not a Kubernetes object.')
        kind = document.get('kind')
        if kind not in allowed_kinds:
            raise C2PError(
                f'Check {check_id} rendered a {kind!r} document, which this engine does not '
                f'generate (expected one of {sorted(allowed_kinds)}). A value substituted into '
                f'the template has escaped the field it was written into.'
            )
        name = document.get('metadata', {}).get('name') if isinstance(document.get('metadata'), dict) else None
        if name != check_id:
            raise C2PError(
                f'Check {check_id} rendered a {kind} named {name!r}. Every document generated '
                f'for a check carries the check id as its name, so this one was not generated '
                f'by the template on purpose.'
            )
    kinds = [d.get('kind') for d in documents]
    if len(set(kinds)) != len(kinds):
        raise C2PError(f'Check {check_id} rendered more than one {kinds} document, which no template does.')


def render(
    policy: Policy,
    template_dir: str,
    deliverable_dir: str,
    engine: str,
    allowed_kinds: Collection[str],
    suffix: str = TEMPLATE_SUFFIX,
) -> List[pathlib.Path]:
    """Render `<template_dir>/<Check_Id><suffix>` to `<deliverable_dir>/<Rule_Id>/<Check_Id>.yaml`.

    The output path carries the rule as well as the check so that a rule implemented by
    several engines does not collapse into one file, and so the generated tree can be applied
    per rule.

    `allowed_kinds` is the set of Kubernetes kinds this engine generates. It has no default:
    a plugin that cannot say what it emits cannot have its output checked, and the check is
    what stands between a component definition and a document nobody asked for.
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
        template = env.from_string(template_path.read_text())
        rendered = template.render(render_context(rule_set, parameters, engine))
        validate_rendered(rendered, rule_set.check_id, allowed_kinds)
        out_path.write_text(rendered)
        written.append(out_path)
        logger.info(f'Generated {out_path}')
    return written
