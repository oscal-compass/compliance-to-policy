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
"""Renderers over c2p-posture/v1. Pure views: they compute no status of their own."""

from typing import List

from jinja2 import Environment, Template

from c2p.tools.posture.model import (
    RESULT_NOT_ASSESSED,
    RESULT_PASS,
    STATE_NOT_ASSESSED,
    STATE_SATISFIED,
    Check,
    Posture,
    Rule,
)
from c2p.tools.posture.template import HTML_TEMPLATE, TEMPLATE

_ICONS = {
    STATE_SATISFIED: ':white_check_mark:',
    RESULT_PASS: ':white_check_mark:',
    STATE_NOT_ASSESSED: ':large_orange_diamond:',
    RESULT_NOT_ASSESSED: ':large_orange_diamond:',
}


def icon(state: str) -> str:
    """Three outcomes need three marks: a control nothing could assess is not a failure."""
    return _ICONS.get(state, ':x:')


def flatten_checks(rules: List[Rule]) -> List[Check]:
    return [check for rule in rules for check in rule.checks]


def render_markdown(posture: Posture) -> str:
    template = Template(source=TEMPLATE)
    return template.render(posture=posture, icon=icon, flatten_checks=flatten_checks)


def render_html(posture: Posture) -> str:
    """Every control, every check, every subject with the engine's own message."""
    return Environment(autoescape=True).from_string(HTML_TEMPLATE).render(posture=posture)


def render_json(posture: Posture) -> str:
    return posture.serialize_json_bytes(pretty=True).decode('utf-8')
