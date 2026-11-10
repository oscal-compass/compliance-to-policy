# -*- mode:python; coding:utf-8 -*-

# Copyright 2024 IBM Corporation
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

import pathlib
import re
from typing import Dict, List, Optional

from pydantic.v1 import BaseModel
from trestle import __version__ as TRESTLE_VERSION
from trestle.common.list_utils import as_list
from trestle.oscal import OSCAL_VERSION
from trestle.oscal.assessment_results import (
    AssessmentResults,
    ImportAp,
    Observation,
    Result,
)
from trestle.oscal.catalog import Catalog
from trestle.oscal.catalog import Model as CatalogRoot
from trestle.oscal.common import (
    Link,
    Metadata,
    Property,
    RelevantEvidence,
    SubjectReference,
)
from trestle.oscal.component import ComponentDefinition
from trestle.oscal.component import Model as ComponentDefinitionRoot
from trestle.oscal.profile import Model as ProfileRoot
from trestle.oscal.profile import Profile

from c2p.common.logging import getLogger
from c2p.common.oscal import is_component_type_validation
from c2p.common.utils import get_dict_safely
from c2p.framework import oscal_utils
from c2p.framework.models.c2p_config import C2PConfig
from c2p.framework.models.policy import Parameter, Policy, RuleSet
from c2p.framework.models.pvp_result import PVPResult, set_defaults

RuleId = str

logger = getLogger(__name__)

# `Parameter_Id`, `Parameter_Id_2`, `Parameter_Id_3`, ... -- a component definition carries
# several parameters for one rule by numbering the columns, so the suffix has to be read
# rather than assumed absent.
_PARAMETER_ID_PATTERN = re.compile(r'^Parameter_Id(?P<suffix>_\d+)?$')


def _set_parameter_values(component) -> Dict[str, str]:
    """param-id -> value, taken from the component's control implementations.

    trestle's csv-to-oscal-cd writes parameter *values* into `set-parameters` and leaves only
    ids and descriptions in the component props, so a reader that looks only at props finds
    parameters with nothing in them.
    """
    values: Dict[str, str] = {}
    for control_implementation in as_list(component.control_implementations):
        for set_parameter in as_list(control_implementation.set_parameters):
            if set_parameter.values != []:
                values[set_parameter.param_id] = set_parameter.values[0]
    return values


def _parameters_of_group(prop_group: Dict[str, str], values: Dict[str, str]) -> List[Parameter]:
    parameters = []
    for name in prop_group:
        matched = _PARAMETER_ID_PATTERN.match(name)
        if matched == None:
            continue
        suffix = matched.group('suffix') if matched.group('suffix') != None else ''
        parameter_id = prop_group[name]
        # set-parameters wins: it is the resolved value. The props are the authored fallback.
        value = values.get(parameter_id)
        if value == None:
            value = get_dict_safely(prop_group, f'Parameter_Value_Default{suffix}')
        if value == None:
            value = get_dict_safely(prop_group, f'Parameter_Value_Alternatives{suffix}')
        if value == None:
            logger.warning(f'Parameter {parameter_id} has no value in set-parameters or in props')
            value = ''
        parameters.append(
            Parameter(
                id=parameter_id,
                description=get_dict_safely(prop_group, f'Parameter_Description{suffix}'),
                value=value,
            )
        )
    return parameters


class _RuleSet(BaseModel):
    effective_rule_id: str
    effective_check_id: str
    rule_id: str
    rule_description: Optional[str]
    check_id: Optional[str]
    check_description: Optional[str]
    raw: Optional[Dict[str, str]]


class C2P:
    def __init__(self, c2p_config: C2PConfig):
        self._c2p_config = c2p_config
        if c2p_config.compliance.catalog:
            catalog = Catalog.oscal_read(pathlib.Path(c2p_config.compliance.catalog))
            self._catalog_root: CatalogRoot = CatalogRoot(catalog=catalog)
        if c2p_config.compliance.profile:
            profile = Profile.oscal_read(pathlib.Path(c2p_config.compliance.profile))
            self._profile_root: ProfileRoot = ProfileRoot(profile=profile)
        cdef = ComponentDefinition.oscal_read(pathlib.Path(c2p_config.compliance.component_definition))
        self._component_root: ComponentDefinitionRoot = ComponentDefinitionRoot(component_definition=cdef)

    def set_pvp_result(self, pvp_result: PVPResult):
        self._c2p_config.pvp_result = pvp_result

    def add_pvp_result(self, pvp_result: PVPResult):
        """Add one PVP's results to be aggregated into a single Assessment Results.

        A cluster typically runs more than one policy engine, and each is assessed on its own
        -- different API, different collection. Emitting one Assessment Results per engine
        makes the artifact an accident of how the assessment was organised rather than a
        statement about the system, and anything reading it then has to know to look for
        several files. Call this once per engine instead.
        """
        if self._c2p_config.pvp_results == None:
            self._c2p_config.pvp_results = []
        self._c2p_config.pvp_results.append(pvp_result)

    def result_to_oscal(self) -> AssessmentResults:
        pvp_results = [set_defaults(r) for r in self._collected_pvp_results()]
        timestamp = oscal_utils.get_datetime_str()
        metadata = Metadata(
            title=self._c2p_config.result_title,
            oscal_version=OSCAL_VERSION,
            version=TRESTLE_VERSION,
            last_modified=timestamp,
        )
        import_ap = ImportAp(href='https://not-available-for-now')
        value = AssessmentResults(
            uuid=oscal_utils.uuid(),
            metadata=metadata,
            import_ap=import_ap,
            results=[self._get_result(pvp_results)],
        )
        return value

    def get_policy(self) -> Policy:
        return Policy(rule_sets=self.get_rule_sets(), parameters=self.get_parameters())

    def get_rule_sets(self) -> List[RuleSet]:
        _rule_sets = self._get_rule_sets()

        def _conv(x: _RuleSet):
            return RuleSet(
                rule_id=x.effective_rule_id,
                rule_description=x.rule_description,
                check_id=x.effective_check_id,
                check_description=x.check_description,
                raw=x.raw,
            )

        return list(map(_conv, _rule_sets))

    def get_parameters(self) -> List[Parameter]:
        return self._get_parameters()

    def _get_rule_sets(self, every_pvp: bool = False) -> List[_RuleSet]:
        """The rule sets an observation's check id can be matched against.

        Aggregating several PVPs means matching against every validation component, not only
        the configured one. Check ids are unique per engine, which is what makes that safe --
        a duplicate across engines would be ambiguous, so it is reported.
        """
        rule_sets: List[Dict[str, str]] = []
        for comp in self._component_root.component_definition.components:
            if not is_component_type_validation(comp.type):
                continue
            if every_pvp or comp.title == self._c2p_config.pvp_name:
                rule_sets = rule_sets + oscal_utils.group_props_by_remarks(comp)
        if every_pvp:
            check_ids = [r.get('Check_Id') for r in rule_sets if r.get('Check_Id') != None]
            duplicated = {c for c in check_ids if check_ids.count(c) > 1}
            if duplicated != set():
                logger.warning(
                    f'Check_Id used by more than one validation component: {sorted(duplicated)}. '
                    f'An observation carrying such a check id cannot be attributed to one engine.'
                )

        def _conv(x: Dict[str, str]) -> _RuleSet:
            return _RuleSet(
                rule_id=get_dict_safely(x, 'Rule_Id'),
                rule_description=get_dict_safely(x, 'Rule_Description'),
                check_id=get_dict_safely(x, 'Check_Id'),
                check_description=get_dict_safely(x, 'Check_Description'),
                effective_rule_id=get_dict_safely(x, self._c2p_config.compliance.rule_id_column),
                effective_check_id=get_dict_safely(x, self._c2p_config.compliance.check_id_column),
                raw=x,
            )

        return list(map(_conv, filter(lambda x: 'Rule_Id' in x, rule_sets)))

    def _find_rule_set(self, check_id: str, rule_sets: List[_RuleSet]) -> Optional[_RuleSet]:
        return next(filter(lambda x: x.effective_check_id == check_id, rule_sets), None)

    def _get_parameters(self) -> List[Parameter]:
        parameters: List[Parameter] = []
        for component in self._component_root.component_definition.components:
            if is_component_type_validation(component.type):
                continue
            values = _set_parameter_values(component)
            for prop_group in oscal_utils.group_props_by_remarks(component):
                parameters = parameters + _parameters_of_group(prop_group, values)
        return parameters

    def _find_parameter(self, id: str, parameters: List[Parameter]) -> Optional[Parameter]:
        return next(filter(lambda x: x.id == id, parameters), None)

    def _collected_pvp_results(self) -> List[PVPResult]:
        if self._c2p_config.pvp_results not in [None, []]:
            return self._c2p_config.pvp_results
        return [self._c2p_config.pvp_result]

    def _get_result(self, pvp_results: List[PVPResult]) -> Result:
        """Return one result covering every PVP whose results were collected."""
        findings = [f for r in pvp_results if r.findings != None for f in r.findings]
        links = [l for r in pvp_results if r.links != None for l in r.links]
        # local_definitions describes the inventory the subjects belong to, which is a property
        # of the system rather than of the engine that looked at it, so the first engine to
        # describe it speaks for all of them.
        local_definitions = next((r.local_definitions for r in pvp_results if r.local_definitions != None), None)

        result = Result(
            uuid=oscal_utils.uuid(),
            title=self._c2p_config.result_title,
            description=self._c2p_config.result_description,
            start=oscal_utils.get_datetime_str(),
            observations=self._get_observations(pvp_results),
            reviewed_controls=oscal_utils.reviewed_controls(self._component_root.component_definition),
            local_definitions=local_definitions,
            findings=findings if findings != [] else None,
        )
        if links != []:
            result.links = list(map(lambda x: Link(href=x.href, text=x.description), links))
        if self._c2p_config.result_labels != None:
            result.props = list(map(lambda x: Property(name='label', value=x), self._c2p_config.result_labels))
        return result

    def _get_observations(self, pvp_results: List[PVPResult]) -> List[Observation]:
        aggregating = len(pvp_results) > 1
        rule_sets = self._get_rule_sets(every_pvp=aggregating)
        observations = []
        for observation in [o for r in pvp_results for o in r.observations_by_check]:
            rule_set = self._find_rule_set(observation.check_id, rule_sets)
            if rule_set != None:
                subjects = []
                for subject in observation.subjects:
                    props = []
                    oscal_utils.add_prop(props, 'resource-id', subject, ['resource_id'])
                    oscal_utils.add_prop(props, 'result', subject, ['result'])
                    oscal_utils.add_prop(props, 'evaluated-on', subject, ['evaluated_on'])
                    oscal_utils.add_prop(props, 'reason', subject, ['reason'])
                    s = SubjectReference(
                        subject_uuid=subject.subject_uuid if subject.subject_uuid else oscal_utils.uuid(),
                        title=subject.title,
                        type=subject.type,
                        props=props,
                    )
                    subjects.append(s)

                relevant_evidences = []
                if observation.relevant_evidences != None:
                    for rel in observation.relevant_evidences:
                        relevant_evidences.append(RelevantEvidence(href=rel.href, description=rel.description))

                props = []
                oscal_utils.add_prop(props, 'assessment-rule-id', rule_set.effective_rule_id, [])
                # Which check, not only which rule. One rule is routinely implemented by
                # several PVPs, so a rule id alone cannot say which engine reported, and a
                # reader of the Assessment Results has no way to attribute the observation.
                oscal_utils.add_prop(props, 'assessment-check-id', rule_set.effective_check_id, [])
                if observation.props != None:
                    props = props + observation.props
                o = Observation(
                    uuid=observation.uuid if observation.uuid else oscal_utils.uuid(),
                    title=observation.title,
                    description=observation.title,
                    methods=observation.methods,
                    props=props,
                    subjects=subjects,
                    collected=observation.collected,
                )
                if len(relevant_evidences) > 0:
                    o.relevant_evidence = relevant_evidences
                observations.append(o)
        return observations
