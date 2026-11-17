# -*- mode:python; coding:utf-8 -*-
# Copyright 2024 JC Company / OSCAL Compass Contributors
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

"""
C2P Plugin for GitHub Actions DevSecOps Pipeline.

This plugin bridges OSCAL Component Definitions and GitHub Actions pipeline
results, enabling automated compliance assessment for DevSecOps pipelines.

Supported tools:
  SAST:    Semgrep, SpotBugs/FindSecBugs, Hadolint, CIS Java Check
  SCA:     Grype (SBOM + Image), Trivy CIS, Dependency-Track
  DAST:    OWASP ZAP
  Secrets: Gitleaks
  SBOM:    CycloneDX
  Test:    JUnit, Terratest
  Runtime: Falco, GuardDuty
  CI/CD:   GitHub Actions, Cosign, Renovate Bot
"""

import json
import xml.etree.ElementTree as ET
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional

from c2p.framework.models.policy import Policy, RuleSet
from c2p.framework.models.pvp_result import (
    Link,
    ObservationByCheck,
    Property,
    PVPResult,
    ResultEnum,
    Subject,
)
from c2p.framework.models.raw_result import RawResult
from c2p.framework.plugin_spec import PluginConfig, PluginSpec

# ── 툴 → OSCAL 통제 매핑 ────────────────────────────────────────────────────
# 근거: payment-api Component Definition (component-definitions/payment-api-components)
#       NIST SP 800-53 Rev5 High Baseline

TOOL_TO_CONTROLS: Dict[str, List[str]] = {
    # Secret Detection
    'gitleaks': ['ia-5', 'si-3'],
    # SAST
    'semgrep': ['si-10', 'sa-11', 'si-3', 'sc-13'],
    'spotbugs': ['si-10', 'sa-11', 'si-3', 'sc-13'],
    'hadolint': ['cm-7', 'cm-2'],
    'cis-java': ['cm-7', 'sc-8', 'sc-13'],
    # SCA / CVE
    'grype': ['ra-5', 'si-2', 'sr-3', 'si-3'],
    'trivy': ['ra-5', 'si-2', 'sr-3', 'si-3'],
    'dependency-track': ['ra-5', 'si-2', 'sr-3'],
    # DAST
    'zap': ['sa-11', 'si-10', 'ac-3'],
    # SBOM
    'cyclonedx': ['sr-3', 'ra-5'],
    # IaC / Config
    'checkov': ['cm-2', 'cm-7', 'sc-8', 'sc-28', 'au-9', 'ac-6'],
    'kyverno': ['cm-7', 'ac-3', 'sr-3', 'cm-2'],
    # CI/CD
    'github-actions': ['sa-15', 'sa-11', 'sr-3', 'cm-3'],
    'cosign': ['sr-3', 'si-3', 'cm-2'],
    'renovate': ['si-2', 'sr-3', 'cm-3'],
    # Runtime
    'falco': ['si-3', 'au-2', 'ac-6', 'cm-7'],
    'guardduty': ['si-3', 'ra-5'],
    # Test
    'junit': ['sa-11'],
    'terratest': ['cm-2', 'cm-3'],
}

# ── 파일 형식 → 파서 매핑 ────────────────────────────────────────────────────
SARIF_TOOLS = {'gitleaks', 'semgrep', 'hadolint'}
JSON_TOOLS = {'grype', 'trivy', 'dependency-track', 'cyclonedx'}
XML_TOOLS = {'spotbugs', 'junit'}
TEXT_TOOLS = {'cis-java'}
HTML_TOOLS = {'zap'}


class GitHubActionsPluginConfig(PluginConfig):
    """Configuration for GitHub Actions C2P Plugin."""

    pipeline_run_url: Optional[str] = None
    commit_sha: Optional[str] = None
    actor: Optional[str] = None


class GitHubActionsPlugin(PluginSpec):
    """
    C2P Plugin for GitHub Actions DevSecOps Pipeline.

    Maps GitHub Actions pipeline tool results to OSCAL Assessment Results,
    enabling automated continuous compliance for DevSecOps pipelines aligned
    with NIST SP 800-53, SSDF SP 800-218, and SP 800-204D.
    """

    def __init__(self, config: Optional[GitHubActionsPluginConfig] = None):
        self.config = config or GitHubActionsPluginConfig()

    def generate_pvp_policy(self, policy: Policy) -> Dict[str, Any]:
        """
        OSCAL Component Definition → GitHub Actions job configuration.

        입력: Policy (OSCAL에서 추출한 rule_sets)
        출력: 각 툴의 GitHub Actions 설정 딕셔너리

        예: sa-11 통제 → Semgrep job 설정 반환
        """
        pvp_policy = {}

        for rule_set in policy.rule_sets or []:
            tool = rule_set.rule_id.lower()
            controls = TOOL_TO_CONTROLS.get(tool, [])

            pvp_policy[tool] = {
                'rule_id': rule_set.rule_id,
                'check_id': rule_set.check_id,
                'description': rule_set.rule_description or '',
                'controls': controls,
                'config': self._get_tool_config(tool, rule_set),
            }

        return pvp_policy

    def _get_tool_config(self, tool: str, rule_set: RuleSet) -> Dict[str, Any]:
        """툴별 GitHub Actions 기본 설정 반환."""
        configs = {
            'semgrep': {
                'action': 'semgrep/semgrep-action@v1',
                'ruleset': 'p/java p/owasp-top-ten p/secrets p/sql-injection',
                'fail_on': 'error',
            },
            'gitleaks': {
                'action': 'gitleaks/gitleaks-action@v2',
                'fetch_depth': 0,
            },
            'spotbugs': {
                'maven_goal': 'spotbugs:check',
                'threshold': 'High',
                'effort': 'Max',
            },
            'grype': {
                'action': 'anchore/scan-action@v4',
                'fail_build': False,
                'severity_cutoff': 'critical',
            },
            'trivy': {
                'compliance': 'docker-cis-1.6',
                'format': 'json',
            },
            'zap': {
                'action': 'zaproxy/action-full-scan@v0.10.0',
                'fail_build': False,
            },
            'hadolint': {
                'action': 'hadolint/hadolint-action@v3.1.0',
                'failure_threshold': 'error',
            },
            'junit': {
                'maven_goal': 'verify',
                'coverage_threshold': 0.60,
            },
        }
        return configs.get(tool, {})

    def generate_pvp_result(self, raw_result: RawResult) -> PVPResult:
        """
        GitHub Actions 파이프라인 결과 → OSCAL Assessment Results.

        입력: RawResult (툴 결과 파일 내용 + 메타데이터)
        출력: PVPResult (OSCAL Assessment Results 형태)

        additional_props에서 'tool' 키로 어떤 툴의 결과인지 판단.
        """
        tool = raw_result.additional_props.get('tool', '').lower()
        check_id = raw_result.additional_props.get('check_id', tool)
        filepath = raw_result.metadata.filepath or ''
        collected = datetime.now(timezone.utc)

        # 툴별 파서 선택
        if tool in SARIF_TOOLS:
            result, reason, props = self._parse_sarif(raw_result.data, tool)
        elif tool in JSON_TOOLS:
            result, reason, props = self._parse_json(raw_result.data, tool)
        elif tool in XML_TOOLS:
            result, reason, props = self._parse_xml(raw_result.data, tool)
        elif tool in TEXT_TOOLS:
            result, reason, props = self._parse_text(raw_result.data, tool)
        else:
            result = ResultEnum.Error
            reason = f'Unknown tool: {tool}'
            props = []

        # 관련 통제 목록
        controls = TOOL_TO_CONTROLS.get(tool, [check_id])

        observations = []
        for control_id in controls:
            obs = ObservationByCheck(
                check_id=control_id,
                title=f'{tool.upper()} - {control_id.upper()}',
                description=f'Automated assessment of {control_id} via {tool} in GitHub Actions pipeline',
                methods=['TEST-AUTOMATED'],
                collected=collected,
                subjects=[
                    Subject(
                        title=f'{tool.upper()} scan result',
                        type='tool',
                        resource_id=tool,
                        result=result,
                        reason=reason,
                        evaluated_on=collected,
                        props=props,
                    )
                ],
                relevant_evidences=(
                    [
                        Link(
                            description=f'{tool} scan result file',
                            href=filepath or f'{tool}-results',
                        )
                    ]
                    if filepath
                    else None
                ),
                props=[
                    Property(name='tool', value=tool),
                    Property(name='pipeline-run', value=self.config.pipeline_run_url or ''),
                    Property(name='commit-sha', value=self.config.commit_sha or ''),
                ]
                + (props or []),
            )
            observations.append(obs)

        return PVPResult(observations_by_check=observations)

    # ── 파서들 ──────────────────────────────────────────────────────────────

    def _parse_sarif(self, data: Any, tool: str) -> tuple[ResultEnum, str, List[Property]]:
        """SARIF 형식 파싱 (Semgrep, Gitleaks, Hadolint)."""
        try:
            if isinstance(data, str):
                data = json.loads(data)

            runs = data.get('runs', [])
            total_results = 0
            error_count = 0
            warning_count = 0

            for run in runs:
                results = run.get('results', [])
                total_results += len(results)
                for r in results:
                    level = r.get('level', 'warning')
                    if level == 'error':
                        error_count += 1
                    else:
                        warning_count += 1

            props = [
                Property(name='total-findings', value=str(total_results)),
                Property(name='error-findings', value=str(error_count)),
                Property(name='warning-findings', value=str(warning_count)),
            ]

            if error_count > 0:
                return (
                    ResultEnum.Failure,
                    f'{tool}: {error_count} error(s), {warning_count} warning(s) found',
                    props,
                )
            elif total_results > 0:
                return (
                    ResultEnum.Pass,
                    f'{tool}: {warning_count} warning(s) found (no blocking errors)',
                    props,
                )
            else:
                return ResultEnum.Pass, f'{tool}: No findings', props

        except Exception as e:
            return ResultEnum.Error, f'{tool}: Failed to parse SARIF - {e}', []

    def _parse_json(self, data: Any, tool: str) -> tuple[ResultEnum, str, List[Property]]:
        """JSON 형식 파싱 (Grype, Trivy, Dependency-Track)."""
        try:
            if isinstance(data, str):
                data = json.loads(data)

            if tool == 'grype':
                return self._parse_grype_json(data)
            elif tool == 'trivy':
                return self._parse_trivy_json(data)
            else:
                # 일반 JSON - matches 또는 vulnerabilities 필드 탐색
                matches = data.get('matches', data.get('vulnerabilities', []))
                count = len(matches)
                props = [Property(name='total-findings', value=str(count))]
                if count > 0:
                    return ResultEnum.Failure, f'{tool}: {count} finding(s)', props
                return ResultEnum.Pass, f'{tool}: No findings', props

        except Exception as e:
            return ResultEnum.Error, f'{tool}: Failed to parse JSON - {e}', []

    def _parse_grype_json(self, data: Any) -> tuple[ResultEnum, str, List[Property]]:
        """Grype CVE 스캔 결과 파싱."""
        matches = data.get('matches', [])
        severity_counts: Dict[str, int] = {}

        for match in matches:
            severity = match.get('vulnerability', {}).get('severity', 'Unknown')
            severity_counts[severity] = severity_counts.get(severity, 0) + 1

        critical = severity_counts.get('Critical', 0)
        high = severity_counts.get('High', 0)
        medium = severity_counts.get('Medium', 0)
        total = len(matches)

        props = [
            Property(name='total-cves', value=str(total)),
            Property(name='critical-cves', value=str(critical)),
            Property(name='high-cves', value=str(high)),
            Property(name='medium-cves', value=str(medium)),
        ]

        reason = f'Grype: {total} CVE(s) found (Critical:{critical}, High:{high}, Medium:{medium})'

        if critical > 0:
            return ResultEnum.Failure, reason, props
        elif total > 0:
            return ResultEnum.Pass, reason, props
        return ResultEnum.Pass, 'Grype: No CVEs found', props

    def _parse_trivy_json(self, data: Any) -> tuple[ResultEnum, str, List[Property]]:
        """Trivy CIS Benchmark 결과 파싱."""
        summary = data.get('Summary', {})
        fail_count = summary.get('failCount', 0)
        pass_count = summary.get('passCount', 0)
        total = fail_count + pass_count

        props = [
            Property(name='cis-pass', value=str(pass_count)),
            Property(name='cis-fail', value=str(fail_count)),
            Property(name='cis-total', value=str(total)),
        ]

        if fail_count > 0:
            return (
                ResultEnum.Failure,
                f'Trivy CIS: {fail_count} failure(s), {pass_count} pass(es)',
                props,
            )
        return ResultEnum.Pass, f'Trivy CIS: All {pass_count} checks passed', props

    def _parse_xml(self, data: Any, tool: str) -> tuple[ResultEnum, str, List[Property]]:
        """XML 형식 파싱 (SpotBugs, JUnit)."""
        try:
            if isinstance(data, str):
                root = ET.fromstring(data)
            else:
                return ResultEnum.Error, f'{tool}: XML data must be string', []

            if tool == 'spotbugs':
                return self._parse_spotbugs_xml(root)
            elif tool == 'junit':
                return self._parse_junit_xml(root)
            else:
                return ResultEnum.Pass, f'{tool}: Parsed successfully', []

        except Exception as e:
            return ResultEnum.Error, f'{tool}: Failed to parse XML - {e}', []

    def _parse_spotbugs_xml(self, root: ET.Element) -> tuple[ResultEnum, str, List[Property]]:
        """SpotBugs XML 결과 파싱."""
        bugs = root.findall('.//BugInstance')
        high_bugs = [b for b in bugs if b.get('priority') in ('1', '2')]
        total = len(bugs)

        props = [
            Property(name='total-bugs', value=str(total)),
            Property(name='high-priority-bugs', value=str(len(high_bugs))),
        ]

        if len(high_bugs) > 0:
            return (
                ResultEnum.Failure,
                f'SpotBugs: {len(high_bugs)} high-priority bug(s) found',
                props,
            )
        return ResultEnum.Pass, f'SpotBugs: {total} total findings (no high priority)', props

    def _parse_junit_xml(self, root: ET.Element) -> tuple[ResultEnum, str, List[Property]]:
        """JUnit XML 결과 파싱."""
        # testsuite 또는 testsuites 처리
        if root.tag == 'testsuites':
            suites = root.findall('testsuite')
        else:
            suites = [root]

        total = 0
        failures = 0
        errors = 0

        for suite in suites:
            total += int(suite.get('tests', 0))
            failures += int(suite.get('failures', 0))
            errors += int(suite.get('errors', 0))

        props = [
            Property(name='total-tests', value=str(total)),
            Property(name='failures', value=str(failures)),
            Property(name='errors', value=str(errors)),
        ]

        if failures > 0 or errors > 0:
            return (
                ResultEnum.Failure,
                f'JUnit: {failures} failure(s), {errors} error(s) out of {total} tests',
                props,
            )
        return ResultEnum.Pass, f'JUnit: All {total} tests passed', props

    def _parse_text(self, data: Any, tool: str) -> tuple[ResultEnum, str, List[Property]]:
        """텍스트 형식 파싱 (CIS Java Check)."""
        try:
            if not isinstance(data, str):
                data = str(data)

            pass_count = data.count('✅ PASS')
            fail_count = data.count('❌ FAIL')
            warn_count = data.count('⚠️  WARN')

            props = [
                Property(name='pass-count', value=str(pass_count)),
                Property(name='fail-count', value=str(fail_count)),
                Property(name='warn-count', value=str(warn_count)),
            ]

            if fail_count > 0:
                return (
                    ResultEnum.Failure,
                    f'CIS Java: {fail_count} FAIL, {warn_count} WARN, {pass_count} PASS',
                    props,
                )
            return (
                ResultEnum.Pass,
                f'CIS Java: {pass_count} PASS, {warn_count} WARN (no failures)',
                props,
            )

        except Exception as e:
            return ResultEnum.Error, f'{tool}: Failed to parse text - {e}', []
