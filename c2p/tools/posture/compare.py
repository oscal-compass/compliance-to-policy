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
"""Two postures, side by side, per control.

What to look at is not the controls that change state -- it is the ones that stop having a
state at all. A control with no assessment capability behind it is not compliant and not
non-compliant; there is no evidence either way, and a posture that quietly omits it is how an
estate convinces itself it is fine.

This is why the comparison cannot be done on findings, or on any view that lists only controls
with results: the interesting cell is an absence, and such a view cannot render an absence.
"""

from typing import Dict, List, Optional

from pydantic.v1 import Field

from c2p.tools.posture.model import STATE_NOT_ASSESSED, Posture, PostureBaseModel

NOT_REPORTED = 'not reported'


class ControlComparison(PostureBaseModel):
    control_id: str = Field(..., alias='controlId')
    baseline: Optional[str]
    other: Optional[str]
    dropped: bool = False


class Comparison(PostureBaseModel):
    controls: List[ControlComparison] = []
    dropped: List[ControlComparison] = []


def states(posture: Posture) -> Dict[str, str]:
    """Control id -> state. A control absent from the posture entirely is absent here too."""
    return {control.control_id: control.state for component in posture.components for control in component.controls}


def compare(baseline: Posture, other: Posture) -> Comparison:
    """Compare `other` against `baseline`, control by control.

    A control counts as dropped when the baseline reported a state for it and `other` either
    does not mention it or could not assess it. Both are the same thing to a reader: the
    control is no longer being answered.
    """
    baseline_states = states(baseline)
    other_states = states(other)

    comparison = Comparison()
    for control_id in sorted(set(baseline_states) | set(other_states)):
        in_baseline = baseline_states.get(control_id)
        in_other = other_states.get(control_id)
        dropped = (
            in_baseline != None
            and in_baseline != STATE_NOT_ASSESSED
            and (in_other == None or in_other == STATE_NOT_ASSESSED)
        )
        row = ControlComparison(control_id=control_id, baseline=in_baseline, other=in_other, dropped=dropped)
        comparison.controls.append(row)
        if dropped:
            comparison.dropped.append(row)
    return comparison


def render_text(
    comparison: Comparison,
    baseline_label: str = 'baseline',
    other_label: str = 'compared',
    condition: str = 'in the compared posture',
    state_labels: Optional[Dict[str, str]] = None,
) -> str:
    """A terminal table, with the absences named in words underneath.

    The labels are parameters because this output gets quoted -- in a document, in a recording,
    in a build that pins callouts to a regular expression over it. A caller that depends on the
    exact wording should pass it rather than inherit it.
    """
    labels = state_labels if state_labels != None else {}

    def cell(state):
        if state == None or state == STATE_NOT_ASSESSED:
            return f'— {NOT_REPORTED} —'
        return labels.get(state, state)

    width = max([len(row.control_id) for row in comparison.controls], default=10) + 2
    header = 'control'.ljust(width) + baseline_label.ljust(24) + other_label
    lines = [header, '-' * len(header)]
    for row in comparison.controls:
        lines.append(row.control_id.ljust(width) + cell(row.baseline).ljust(24) + cell(row.other))

    if comparison.dropped != []:
        lines.append('')
        lines.append(f'{len(comparison.dropped)} control(s) stop being reported {condition}:')
        for row in comparison.dropped:
            lines.append(f'  {row.control_id}: was {cell(row.baseline)}, now absent')
    return '\n'.join(lines) + '\n'
