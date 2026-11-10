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

from c2p.tools.posture import compare
from c2p.tools.posture.model import (
    STATE_NOT_ASSESSED,
    STATE_NOT_SATISFIED,
    STATE_SATISFIED,
    Component,
    Control,
    Coverage,
    Posture,
)


def a_posture(controls):
    return Posture(
        components=[
            Component(
                title='target',
                controls=[
                    Control(
                        control_id=control_id,
                        state=state,
                        reason='pass',
                        coverage=Coverage(expected_checks=1, assessed_checks=1),
                    )
                    for control_id, state in controls
                ],
            )
        ]
    )


def test_a_control_that_stops_being_assessed_is_reported_as_dropped():
    """Not a state change -- the absence is the finding. An engine that cannot evaluate a
    control leaves it neither compliant nor violated."""
    baseline = a_posture([('ac-6', STATE_SATISFIED), ('si-7', STATE_NOT_SATISFIED)])
    other = a_posture([('ac-6', STATE_SATISFIED), ('si-7', STATE_NOT_ASSESSED)])

    comparison = compare.compare(baseline, other)

    assert [row.control_id for row in comparison.dropped] == ['si-7']


def test_a_control_missing_from_the_other_posture_counts_as_dropped():
    """Absent and not-assessed read the same to a person: the control is no longer answered."""
    baseline = a_posture([('ac-6', STATE_SATISFIED), ('cm-7.5', STATE_SATISFIED)])
    other = a_posture([('ac-6', STATE_SATISFIED)])

    comparison = compare.compare(baseline, other)

    assert [row.control_id for row in comparison.dropped] == ['cm-7.5']


def test_a_control_that_was_never_assessed_in_the_baseline_has_not_dropped():
    baseline = a_posture([('ac-6', STATE_NOT_ASSESSED)])
    other = a_posture([('ac-6', STATE_NOT_ASSESSED)])

    assert compare.compare(baseline, other).dropped == []


def test_a_state_change_is_not_a_drop():
    baseline = a_posture([('ac-6', STATE_SATISFIED)])
    other = a_posture([('ac-6', STATE_NOT_SATISFIED)])

    comparison = compare.compare(baseline, other)

    assert comparison.dropped == []
    assert comparison.controls[0].other == STATE_NOT_SATISFIED


def test_the_text_view_names_the_absences_in_words():
    """The table alone does not land: a reader scanning columns sees a blank, not a claim."""
    baseline = a_posture([('ac-6', STATE_SATISFIED), ('si-7.15', STATE_NOT_SATISFIED)])
    other = a_posture([('ac-6', STATE_SATISFIED)])

    rendered = compare.render_text(compare.compare(baseline, other), condition='with only VAP')

    assert 'not reported' in rendered
    assert '1 control(s) stop being reported with only VAP:' in rendered
    assert 'si-7.15: was not-satisfied, now absent' in rendered
