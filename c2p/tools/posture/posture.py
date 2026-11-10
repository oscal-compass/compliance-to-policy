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
"""Assessment Results + Component Definition (+ Catalog) -> c2p-posture/v1.

A pure function of its inputs: it reads files and touches nothing else, so it needs no
credentials and holds no connection to whatever system was assessed.

Two properties are deliberate.

The expected rule/check set comes from the **component definition, not the assessment
results**. An Assessment Results cannot contain what was never assessed, so a posture derived
from it alone can only ever show what some engine happened to report -- which is how an estate
convinces itself it is fine. Expecting from the CD is what makes `not-assessed` and `coverage`
computable at all.

Findings are **not** read. Finding generation in the wild omits passing controls, so a posture
built from findings silently drops every green row.
"""

from typing import Dict, List, Optional, Sequence, Tuple, Union

from trestle.common.list_utils import as_list
from trestle.oscal.assessment_results import AssessmentResults, Observation
from trestle.oscal.catalog import Catalog
from trestle.oscal.common import Property
from trestle.oscal.component import ComponentDefinition

from c2p.common.oscal import is_component_type_validation
from c2p.framework import oscal_utils
from c2p.framework.oscal_utils import group_props_by_remarks
from c2p.tools.posture.model import (
    REASON_FAIL,
    REASON_OTHER,
    REASON_PASS,
    RESULT_ERROR,
    RESULT_FAIL,
    RESULT_NOT_ASSESSED,
    RESULT_PASS,
    RESULT_WARNING,
    STATE_NOT_ASSESSED,
    STATE_NOT_SATISFIED,
    STATE_SATISFIED,
    Check,
    Component,
    Control,
    Coverage,
    Link,
    Posture,
    Rule,
    Source,
    Subject,
    Summary,
)

# Subject props promoted to fields; anything else is carried through in `props` verbatim.
_PROMOTED_SUBJECT_PROPS = ['result', 'reason', 'resource-id', 'evaluated-on']

# This project emits `failure`; other producers emit `fail`. One vocabulary from here on.
_RESULT_ALIASES = {
    'failure': RESULT_FAIL,
    'fail': RESULT_FAIL,
    'pass': RESULT_PASS,
    'warning': RESULT_WARNING,
    'error': RESULT_ERROR,
}


def _prop(props: Optional[List[Property]], name: str) -> Optional[str]:
    """Look up a prop by name, ignoring namespace: some producers stamp a trestle ns on every prop."""
    p = next(filter(lambda x: x.name == name, as_list(props)), None)
    return p.value if p != None else None


def _normalize_result(value: Optional[str]) -> str:
    if value == None:
        return RESULT_ERROR
    return _RESULT_ALIASES.get(value.lower(), RESULT_ERROR)


def control_titles(catalog: Optional[Catalog]) -> Dict[str, str]:
    """Control id -> title, recursively, so a posture row can name the control it is about."""
    titles: Dict[str, str] = {}

    def walk_controls(controls) -> None:
        for control in as_list(controls):
            titles[control.id] = control.title
            walk_controls(control.controls)

    def walk_groups(groups) -> None:
        for group in as_list(groups):
            walk_controls(group.controls)
            walk_groups(group.groups)

    if catalog != None:
        walk_controls(catalog.controls)
        walk_groups(catalog.groups)
    return titles


def expected_checks(component_definition: ComponentDefinition) -> Dict[str, List[Dict[str, str]]]:
    """Rule_Id -> every check the component definition expects for it, one per PVP.

    This is the set a posture is accountable for. A check in here with no matching observation
    is reported as `not-assessed` rather than omitted.
    """
    expected: Dict[str, List[Dict[str, str]]] = {}
    for component in as_list(component_definition.components):
        if not is_component_type_validation(component.type):
            continue
        for rule_set in group_props_by_remarks(component):
            rule_id = rule_set.get('Rule_Id')
            if rule_id == None:
                continue
            expected.setdefault(rule_id, []).append(
                {
                    'engine': component.title,
                    'check_id': rule_set.get('Check_Id'),
                    'check_description': rule_set.get('Check_Description'),
                    'rule_description': rule_set.get('Rule_Description'),
                }
            )
    return expected


def index_observations(
    assessment_results: Union[AssessmentResults, Sequence[AssessmentResults]],
) -> Tuple[Dict[str, List[Observation]], Dict[str, List[Observation]]]:
    """Index observations by check id and by rule id.

    Several Assessment Results are accepted because one of them is often all a run can
    produce: a c2p pipeline assesses one PVP at a time, so a cluster with three policy engines
    yields three documents. The posture is the place they come back together, since it is
    already the only component that knows the whole expected set of checks.

    Both indexes are needed: an `assessment-check-id` prop identifies the check directly, and
    an Assessment Results carrying only `assessment-rule-id` has to be matched by rule.
    """
    by_check: Dict[str, List[Observation]] = {}
    by_rule: Dict[str, List[Observation]] = {}
    documents = [assessment_results] if isinstance(assessment_results, AssessmentResults) else list(assessment_results)
    for result in [r for document in documents for r in as_list(document.results)]:
        for observation in as_list(result.observations):
            check_id = _prop(observation.props, 'assessment-check-id')
            rule_id = _prop(observation.props, 'assessment-rule-id')
            if check_id != None:
                by_check.setdefault(check_id, []).append(observation)
            if rule_id != None:
                by_rule.setdefault(rule_id, []).append(observation)
    return by_check, by_rule


def _subjects_of(observations: List[Observation]) -> List[Subject]:
    subjects = []
    for observation in observations:
        for subject in as_list(observation.subjects):
            props = {p.name: p.value for p in as_list(subject.props) if p.name not in _PROMOTED_SUBJECT_PROPS}
            subjects.append(
                Subject(
                    title=subject.title if subject.title != None else subject.subject_uuid,
                    resource_id=_prop(subject.props, 'resource-id'),
                    result=_normalize_result(_prop(subject.props, 'result')),
                    reason=_prop(subject.props, 'reason'),
                    evaluated_on=_prop(subject.props, 'evaluated-on'),
                    props=props if props != {} else None,
                )
            )
    return subjects


def _evidence_of(observations: List[Observation]) -> List[Link]:
    evidence = []
    seen = set()
    for observation in observations:
        for relevant in as_list(observation.relevant_evidence):
            href = relevant.href if relevant.href != None else ''
            if href in seen:
                continue
            seen.add(href)
            evidence.append(Link(description=relevant.description, href=href))
    return evidence


# A verdict is something an engine decided about a subject. `error` is not one: it is the
# engine, or the probe in front of it, saying it could not decide -- a probe that could not
# reach a check, a `skip` from a matching PolicyException. Neither is evidence of compliance
# nor of its absence, so neither is counted as either. They are reported as coverage: the
# control was not fully assessed, which is a different fact from the control being violated
# and is the fact that is true.
VERDICTS = [RESULT_FAIL, RESULT_WARNING, RESULT_PASS]


def _check_result(subjects: List[Subject]) -> str:
    """The strongest verdict any subject produced, or `error` if none produced one."""
    results = [s.result for s in subjects]
    if results == []:
        return RESULT_NOT_ASSESSED
    for candidate in VERDICTS:
        if candidate in results:
            return candidate
    return RESULT_ERROR


def _undecided(subjects: List[Subject]) -> List[Subject]:
    return [s for s in subjects if s.result not in VERDICTS]


def _check_remarks(result: str, subjects: List[Subject]) -> str:
    if result == RESULT_NOT_ASSESSED:
        return 'no evidence covered this check'
    undecided = len(_undecided(subjects))
    if result == RESULT_ERROR:
        return f'{undecided} of {len(subjects)} subject(s) could not be evaluated'
    counts = {
        'failing': len([s for s in subjects if s.result == RESULT_FAIL]),
        'non-compliant but unscored': len([s for s in subjects if s.result == RESULT_WARNING]),
        'passing': len([s for s in subjects if s.result == RESULT_PASS]),
    }
    named = ', '.join(f'{n} {label}' for label, n in counts.items() if n != 0)
    if undecided != 0:
        named = f'{named}, {undecided} not evaluated'
    return f'{named}, of {len(subjects)} subject(s)'


def _roll_up(checks: List[Check]) -> Tuple[str, str, str]:
    """Control state, OSCAL reason classifier, and human remarks, from the control's checks.

    A denied non-compliant request is `pass` evidence -- the control was effective -- so this
    does not try to re-derive compliance from subjects; it aggregates what the engines said.
    """
    reported = [c for c in checks if c.result != RESULT_NOT_ASSESSED]
    # A check that errored reported no verdict, so it is not a check this control can be
    # judged on. It stays in `expected`, which is what makes the gap visible as coverage.
    decided = [c for c in reported if c.result != RESULT_ERROR]
    subjects = [s for c in reported for s in c.subjects]
    if checks == []:
        return STATE_NOT_ASSESSED, REASON_OTHER, 'no PVP declares a check for this control'
    if reported == []:
        remarks = f'no evidence from any of the {len(checks)} check(s) this control expects'
        return STATE_NOT_ASSESSED, REASON_OTHER, remarks
    if decided == []:
        remarks = f'all {len(reported)} reporting check(s) failed to evaluate, of {len(checks)} expected'
        return STATE_NOT_ASSESSED, REASON_OTHER, remarks

    coverage = f'{len(decided)} of {len(checks)} expected check(s) reported'
    undecided = _undecided(subjects)
    gap = f', {len(undecided)} subject(s) could not be evaluated' if undecided != [] else ''

    failing = [s for s in subjects if s.result == RESULT_FAIL]
    if failing != []:
        remarks = f'{len(failing)} of {len(subjects)} subject(s) failing, {coverage}{gap}'
        return STATE_NOT_SATISFIED, REASON_FAIL, remarks

    warned = [s for s in subjects if s.result == RESULT_WARNING]
    if warned != []:
        remarks = (
            f'{len(warned)} of {len(subjects)} subject(s) non-compliant, reported by the engine '
            f'as a warning rather than scored, {coverage}{gap}'
        )
        return STATE_NOT_SATISFIED, REASON_OTHER, remarks

    passing = [s for s in subjects if s.result == RESULT_PASS]
    if undecided != []:
        # Something passed and something could not be evaluated. Neither "satisfied" nor
        # "not-satisfied" is true of the control, and saying either would invent a verdict the
        # evidence does not carry.
        remarks = f'{len(passing)} of {len(subjects)} subject(s) passing, {coverage}{gap}'
        return STATE_NOT_ASSESSED, REASON_OTHER, remarks

    return STATE_SATISFIED, REASON_PASS, f'all {len(subjects)} subject(s) passing, {coverage}'


def _build_check(
    rule_id: str,
    candidate: Dict[str, str],
    candidates: List[Dict[str, str]],
    by_check: Dict[str, List[Observation]],
    by_rule: Dict[str, List[Observation]],
) -> Check:
    check_id = candidate['check_id']
    observations = by_check.get(check_id, []) if check_id != None else []
    # Attributing a rule-keyed observation is only unambiguous when the rule has a single
    # check; with one check per engine, an AR without `assessment-check-id` cannot say which
    # engine reported, and guessing would invent coverage the evidence does not support.
    if observations == [] and len(candidates) == 1:
        observations = by_rule.get(rule_id, [])
    subjects = _subjects_of(observations)
    result = _check_result(subjects)
    return Check(
        check_id=check_id if result != RESULT_NOT_ASSESSED else None,
        engine=candidate['engine'],
        result=result,
        remarks=_check_remarks(result, subjects),
        evidence=_evidence_of(observations),
        subjects=subjects,
    )


def _build_rules(
    implemented_requirement,
    expected: Dict[str, List[Dict[str, str]]],
    by_check: Dict[str, List[Observation]],
    by_rule: Dict[str, List[Observation]],
) -> List[Rule]:
    rules = []
    for prop in as_list(implemented_requirement.props):
        if prop.name != 'Rule_Id':
            continue
        rule_id = prop.value
        candidates = expected.get(rule_id, [])
        rule = Rule(
            rule_id=rule_id,
            description=next((c['rule_description'] for c in candidates if c['rule_description']), None),
        )
        for candidate in candidates:
            rule.checks.append(_build_check(rule_id, candidate, candidates, by_check, by_rule))
        rules.append(rule)
    return rules


def _dedupe_evidence(checks: List[Check]) -> List[Link]:
    evidence = []
    seen = set()
    for check in checks:
        for link in check.evidence:
            if link.href in seen:
                continue
            seen.add(link.href)
            evidence.append(link)
    return evidence


def _build_control(control_id: str, title: Optional[str], rules: List[Rule]) -> Control:
    checks = [c for r in rules for c in r.checks]
    state, reason, remarks = _roll_up(checks)
    assessed = len([c for c in checks if c.result not in [RESULT_NOT_ASSESSED, RESULT_ERROR]])
    return Control(
        control_id=control_id,
        title=title,
        state=state,
        reason=reason,
        remarks=remarks,
        coverage=Coverage(expected_checks=len(checks), assessed_checks=assessed),
        evidence=_dedupe_evidence(checks),
        rules=rules,
    )


def _tally(summary: Summary, state: str) -> None:
    if state == STATE_SATISFIED:
        summary.satisfied = summary.satisfied + 1
    elif state == STATE_NOT_SATISFIED:
        summary.not_satisfied = summary.not_satisfied + 1
    else:
        summary.not_assessed = summary.not_assessed + 1


def generate(
    assessment_results: Union[AssessmentResults, Sequence[AssessmentResults]],
    component_definition: ComponentDefinition,
    catalog: Optional[Catalog] = None,
    framework: Optional[str] = None,
    source: Optional[Source] = None,
) -> Posture:
    """Build a c2p-posture/v1 document from OSCAL artifacts on disk.

    The catalog is optional and only titles the controls; the accountable set of checks comes
    from the component definition either way.
    """
    titles = control_titles(catalog)
    expected = expected_checks(component_definition)
    by_check, by_rule = index_observations(assessment_results)

    components = []
    summary = Summary()
    for component in as_list(component_definition.components):
        if is_component_type_validation(component.type):
            continue
        rendered = Component(title=component.title)
        for control_implementation in as_list(component.control_implementations):
            for implemented_requirement in as_list(control_implementation.implemented_requirements):
                control_id = implemented_requirement.control_id
                rules = _build_rules(implemented_requirement, expected, by_check, by_rule)
                control = _build_control(control_id, titles.get(control_id), rules)
                rendered.controls.append(control)
                _tally(summary, control.state)
        components.append(rendered)

    return Posture(
        framework=framework,
        generated_at=oscal_utils.get_datetime_str(),
        source=source,
        components=components,
        summary=summary,
    )
