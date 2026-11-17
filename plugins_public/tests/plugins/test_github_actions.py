# -*- mode:python; coding:utf-8 -*-
# Copyright 2024 JC Company / OSCAL Compass Contributors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0

"""
Tests for C2P GitHub Actions Plugin.

Tests cover:
  - generate_pvp_policy(): OSCAL Policy → GitHub Actions config
  - generate_pvp_result(): Pipeline results → OSCAL Assessment Results
    - SARIF (Semgrep, Gitleaks, Hadolint)
    - JSON (Grype, Trivy)
    - XML (SpotBugs, JUnit)
    - Text (CIS Java)
"""

import json
from datetime import datetime, timezone

import pytest

from c2p.framework.models.policy import Policy, RuleSet
from c2p.framework.models.pvp_result import ResultEnum
from c2p.framework.models.raw_result import Metadata, RawResult
from c2p.framework.plugin_spec import PluginSpec
from plugins_public.plugins.github_actions import (
    TOOL_TO_CONTROLS,
    GitHubActionsPlugin,
    GitHubActionsPluginConfig,
)

# ── Fixtures ─────────────────────────────────────────────────────────────────


@pytest.fixture
def plugin():
    """기본 플러그인 인스턴스."""
    return GitHubActionsPlugin()


@pytest.fixture
def plugin_with_config():
    """설정이 있는 플러그인 인스턴스."""
    config = GitHubActionsPluginConfig(
        pipeline_run_url='https://github.com/s1ns3nz0/jccompany-payment-api/actions/runs/123',
        commit_sha='abc123def456',
        actor='s1ns3nz0',
    )
    return GitHubActionsPlugin(config=config)


@pytest.fixture
def sarif_no_findings():
    """발견 없는 SARIF 데이터."""
    return {'runs': [{'results': []}]}


@pytest.fixture
def sarif_with_error():
    """에러 발견이 있는 SARIF 데이터."""
    return {
        'runs': [
            {
                'results': [
                    {'level': 'error', 'ruleId': 'spring-actuator-dangerous-endpoints'},
                    {'level': 'warning', 'ruleId': 'test-warning'},
                ]
            }
        ]
    }


@pytest.fixture
def sarif_warnings_only():
    """경고만 있는 SARIF 데이터."""
    return {
        'runs': [
            {
                'results': [
                    {'level': 'warning', 'ruleId': 'some-warning'},
                ]
            }
        ]
    }


@pytest.fixture
def grype_no_cves():
    """CVE 없는 Grype 결과."""
    return {'matches': []}


@pytest.fixture
def grype_critical_cves():
    """Critical CVE가 있는 Grype 결과."""
    return {
        'matches': [
            {'vulnerability': {'id': 'CVE-2021-44228', 'severity': 'Critical'}},
            {'vulnerability': {'id': 'CVE-2022-1234', 'severity': 'High'}},
            {'vulnerability': {'id': 'CVE-2023-5678', 'severity': 'Medium'}},
        ]
    }


@pytest.fixture
def grype_high_only():
    """High CVE만 있는 Grype 결과 (Critical 없음)."""
    return {
        'matches': [
            {'vulnerability': {'id': 'CVE-2022-9999', 'severity': 'High'}},
        ]
    }


@pytest.fixture
def trivy_cis_pass():
    """CIS Benchmark 전체 통과 Trivy 결과."""
    return {'Summary': {'passCount': 20, 'failCount': 0}}


@pytest.fixture
def trivy_cis_fail():
    """CIS Benchmark 실패가 있는 Trivy 결과."""
    return {'Summary': {'passCount': 15, 'failCount': 5}}


@pytest.fixture
def spotbugs_no_bugs():
    """버그 없는 SpotBugs XML."""
    return '<BugCollection version="4.8.3"><SummaryHTML/></BugCollection>'


@pytest.fixture
def spotbugs_high_bugs():
    """High priority 버그가 있는 SpotBugs XML."""
    return '''<BugCollection version="4.8.3">
        <BugInstance type="SQL_INJECTION" priority="1" rank="1">
            <Class classname="com.jccompany.PaymentService"/>
        </BugInstance>
        <BugInstance type="WEAK_CRYPTO" priority="2" rank="5">
            <Class classname="com.jccompany.SecurityConfig"/>
        </BugInstance>
    </BugCollection>'''


@pytest.fixture
def junit_all_pass():
    """모든 테스트 통과 JUnit XML."""
    return '''<testsuite tests="11" failures="0" errors="0" time="1.234">
        <testcase name="testAC3UnauthenticatedBlocked" classname="PaymentApiTest"/>
        <testcase name="testAC3AuthenticatedAccess" classname="PaymentApiTest"/>
    </testsuite>'''


@pytest.fixture
def junit_with_failures():
    """실패가 있는 JUnit XML."""
    return '''<testsuite tests="11" failures="2" errors="0" time="1.234">
        <testcase name="testAC3UnauthenticatedBlocked" classname="PaymentApiTest">
            <failure>Expected 401 but got 200</failure>
        </testcase>
    </testsuite>'''


@pytest.fixture
def cis_java_all_pass():
    """모든 항목 통과 CIS Java 결과."""
    return """✅ PASS | 원격 디버깅 비활성화 (CM-7)
✅ PASS | JMX 원격 접속 비활성화 (CM-7)
✅ PASS | 비루트 사용자 실행 확인 (CM-7, AC-6)
✅ PASS | 약한 암호화 알고리즘 미사용 (SC-13)
✅ PASS | H2 콘솔 비활성화 확인 (CM-7)
✅ PASS | 디버그 모드 비활성화 확인 (CM-7)
⚠️  WARN | Actuator 엔드포인트 노출 | info endpoint (CM-7)
✅ PASS | 로깅 설정 존재 (AU-2)
✅ PASS | SNAPSHOT 버전 미사용 (SR-3)"""


@pytest.fixture
def cis_java_with_failures():
    """실패가 있는 CIS Java 결과."""
    return """✅ PASS | 원격 디버깅 비활성화 (CM-7)
❌ FAIL | TLS 버전 미명시 | 운영환경에서 TLS 1.2+ 명시 필요 (SC-8)
❌ FAIL | H2 콘솔 비활성화 | 운영환경 H2 콘솔 노출 위험 (CM-7)
⚠️  WARN | Actuator 엔드포인트 노출 | info endpoint (CM-7)"""


def make_raw_result(tool: str, data, filepath: str = '') -> RawResult:
    """테스트용 RawResult 생성 헬퍼."""
    return RawResult(
        metadata=Metadata(filepath=filepath or f'{tool}-results'),
        data=data,
        additional_props={'tool': tool, 'check_id': TOOL_TO_CONTROLS[tool][0]},
    )


# ── 기본 구조 테스트 ──────────────────────────────────────────────────────────


class TestPluginSpec:
    """플러그인 스펙 준수 테스트."""

    def test_inherits_plugin_spec(self, plugin):
        """PluginSpec 상속 확인."""
        assert isinstance(plugin, PluginSpec)

    def test_has_required_methods(self, plugin):
        """필수 메서드 존재 확인."""
        assert hasattr(plugin, 'generate_pvp_policy')
        assert hasattr(plugin, 'generate_pvp_result')

    def test_default_config(self, plugin):
        """기본 설정 확인."""
        assert plugin.config is not None
        assert plugin.config.pipeline_run_url is None

    def test_custom_config(self, plugin_with_config):
        """커스텀 설정 확인."""
        assert plugin_with_config.config.commit_sha == 'abc123def456'
        assert plugin_with_config.config.actor == 's1ns3nz0'


class TestToolMapping:
    """TOOL_TO_CONTROLS 매핑 테스트."""

    def test_all_required_tools_present(self):
        """필수 툴이 모두 매핑됐는지 확인."""
        required = [
            'gitleaks',
            'semgrep',
            'spotbugs',
            'hadolint',
            'cis-java',
            'grype',
            'trivy',
            'zap',
            'cyclonedx',
            'junit',
        ]
        for tool in required:
            assert tool in TOOL_TO_CONTROLS, f'{tool} missing from TOOL_TO_CONTROLS'

    def test_controls_are_nist_format(self):
        """통제 ID가 NIST 형식인지 확인 (예: ac-3, sa-11)."""
        import re

        pattern = re.compile(r'^[a-z]{2}-\d+$')
        for tool, controls in TOOL_TO_CONTROLS.items():
            for control in controls:
                assert pattern.match(control), f"Invalid control ID '{control}' for tool '{tool}'"

    def test_semgrep_controls(self):
        """Semgrep 통제 매핑 확인."""
        controls = TOOL_TO_CONTROLS['semgrep']
        assert 'sa-11' in controls  # Developer Testing
        assert 'si-10' in controls  # Input Validation
        assert 'sc-13' in controls  # Cryptographic Protection

    def test_grype_controls(self):
        """Grype 통제 매핑 확인."""
        controls = TOOL_TO_CONTROLS['grype']
        assert 'ra-5' in controls  # Vulnerability Scanning
        assert 'si-2' in controls  # Flaw Remediation
        assert 'sr-3' in controls  # Supply Chain

    def test_gitleaks_controls(self):
        """Gitleaks 통제 매핑 확인."""
        controls = TOOL_TO_CONTROLS['gitleaks']
        assert 'ia-5' in controls  # Authenticator Management
        assert 'si-3' in controls  # Malicious Code Protection


# ── generate_pvp_policy() 테스트 ──────────────────────────────────────────────


class TestGeneratePvpPolicy:
    """generate_pvp_policy() 테스트."""

    def test_single_tool(self, plugin):
        """단일 툴 정책 생성."""
        policy = Policy(rule_sets=[RuleSet(rule_id='semgrep', check_id='sa-11', rule_description='SAST')])
        result = plugin.generate_pvp_policy(policy)
        assert 'semgrep' in result
        assert result['semgrep']['rule_id'] == 'semgrep'
        assert result['semgrep']['check_id'] == 'sa-11'

    def test_multiple_tools(self, plugin):
        """복수 툴 정책 생성."""
        policy = Policy(
            rule_sets=[
                RuleSet(rule_id='semgrep', check_id='sa-11'),
                RuleSet(rule_id='grype', check_id='ra-5'),
                RuleSet(rule_id='gitleaks', check_id='ia-5'),
            ]
        )
        result = plugin.generate_pvp_policy(policy)
        assert len(result) == 3
        assert 'semgrep' in result
        assert 'grype' in result
        assert 'gitleaks' in result

    def test_controls_included(self, plugin):
        """통제 목록이 결과에 포함되는지 확인."""
        policy = Policy(rule_sets=[RuleSet(rule_id='grype', check_id='ra-5')])
        result = plugin.generate_pvp_policy(policy)
        assert 'ra-5' in result['grype']['controls']
        assert 'sr-3' in result['grype']['controls']

    def test_tool_config_included(self, plugin):
        """툴 설정이 결과에 포함되는지 확인."""
        policy = Policy(rule_sets=[RuleSet(rule_id='semgrep', check_id='sa-11')])
        result = plugin.generate_pvp_policy(policy)
        config = result['semgrep']['config']
        assert 'action' in config
        assert 'semgrep/semgrep-action' in config['action']

    def test_empty_policy(self, plugin):
        """빈 정책 처리."""
        policy = Policy(rule_sets=[])
        result = plugin.generate_pvp_policy(policy)
        assert result == {}

    def test_unknown_tool(self, plugin):
        """알 수 없는 툴 처리 (오류 없이 처리)."""
        policy = Policy(rule_sets=[RuleSet(rule_id='unknown-tool', check_id='sa-11')])
        result = plugin.generate_pvp_policy(policy)
        assert 'unknown-tool' in result
        assert result['unknown-tool']['controls'] == []


# ── generate_pvp_result() SARIF 테스트 ────────────────────────────────────────


class TestParseSarif:
    """SARIF 형식 파싱 테스트 (Semgrep, Gitleaks, Hadolint)."""

    def test_semgrep_no_findings(self, plugin, sarif_no_findings):
        """Semgrep 발견 없음 → Pass."""
        raw = make_raw_result('semgrep', sarif_no_findings, 'semgrep.sarif')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Pass

    def test_semgrep_error_findings(self, plugin, sarif_with_error):
        """Semgrep 에러 발견 → Failure."""
        raw = make_raw_result('semgrep', sarif_with_error, 'semgrep.sarif')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Failure

    def test_semgrep_warnings_only(self, plugin, sarif_warnings_only):
        """Semgrep 경고만 → Pass (blocking 아님)."""
        raw = make_raw_result('semgrep', sarif_warnings_only, 'semgrep.sarif')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Pass

    def test_gitleaks_no_secrets(self, plugin, sarif_no_findings):
        """Gitleaks 시크릿 없음 → Pass."""
        raw = make_raw_result('gitleaks', sarif_no_findings, 'gitleaks.sarif')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Pass

    def test_observations_cover_all_controls(self, plugin, sarif_no_findings):
        """Semgrep 결과가 모든 관련 통제를 커버하는지 확인."""
        raw = make_raw_result('semgrep', sarif_no_findings)
        result = plugin.generate_pvp_result(raw)
        check_ids = {obs.check_id for obs in result.observations_by_check}
        for control in TOOL_TO_CONTROLS['semgrep']:
            assert control in check_ids

    def test_evidence_link_included(self, plugin, sarif_no_findings):
        """증거 파일 링크가 포함되는지 확인."""
        raw = make_raw_result('semgrep', sarif_no_findings, 'semgrep-results.sarif')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.relevant_evidences is not None
        assert obs.relevant_evidences[0].href == 'semgrep-results.sarif'

    def test_method_is_test_automated(self, plugin, sarif_no_findings):
        """방법론이 TEST-AUTOMATED인지 확인 (OSCAL 요구사항)."""
        raw = make_raw_result('semgrep', sarif_no_findings)
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert 'TEST-AUTOMATED' in obs.methods

    def test_properties_include_tool_name(self, plugin, sarif_no_findings):
        """툴 이름이 properties에 포함되는지 확인."""
        raw = make_raw_result('semgrep', sarif_no_findings)
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        prop_names = {p.name for p in (obs.props or [])}
        assert 'tool' in prop_names


# ── generate_pvp_result() JSON 테스트 ────────────────────────────────────────


class TestParseGrypeJson:
    """Grype JSON 파싱 테스트."""

    def test_no_cves_pass(self, plugin, grype_no_cves):
        """CVE 없음 → Pass."""
        raw = make_raw_result('grype', grype_no_cves, 'grype.json')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Pass

    def test_critical_cves_failure(self, plugin, grype_critical_cves):
        """Critical CVE → Failure."""
        raw = make_raw_result('grype', grype_critical_cves, 'grype.json')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Failure

    def test_high_only_pass(self, plugin, grype_high_only):
        """High CVE만 있음 → Pass (Critical만 차단)."""
        raw = make_raw_result('grype', grype_high_only, 'grype.json')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Pass

    def test_severity_counts_in_props(self, plugin, grype_critical_cves):
        """CVE 심각도 카운트가 properties에 포함되는지 확인."""
        raw = make_raw_result('grype', grype_critical_cves, 'grype.json')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        prop_values = {p.name: p.value for p in (obs.props or [])}
        assert 'critical-cves' in prop_values
        assert prop_values['critical-cves'] == '1'
        assert prop_values['high-cves'] == '1'


class TestParseTrivyJson:
    """Trivy CIS JSON 파싱 테스트."""

    def test_all_pass(self, plugin, trivy_cis_pass):
        """CIS 전체 통과 → Pass."""
        raw = make_raw_result('trivy', trivy_cis_pass, 'trivy.json')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Pass

    def test_with_failures(self, plugin, trivy_cis_fail):
        """CIS 실패 있음 → Failure."""
        raw = make_raw_result('trivy', trivy_cis_fail, 'trivy.json')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Failure

    def test_cis_counts_in_props(self, plugin, trivy_cis_fail):
        """CIS 카운트가 properties에 포함되는지 확인."""
        raw = make_raw_result('trivy', trivy_cis_fail, 'trivy.json')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        prop_values = {p.name: p.value for p in (obs.props or [])}
        assert prop_values['cis-fail'] == '5'
        assert prop_values['cis-pass'] == '15'


# ── generate_pvp_result() XML 테스트 ────────────────────────────────────────


class TestParseSpotbugsXml:
    """SpotBugs XML 파싱 테스트."""

    def test_no_bugs_pass(self, plugin, spotbugs_no_bugs):
        """버그 없음 → Pass."""
        raw = make_raw_result('spotbugs', spotbugs_no_bugs, 'spotbugs.xml')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Pass

    def test_high_bugs_failure(self, plugin, spotbugs_high_bugs):
        """High priority 버그 → Failure."""
        raw = make_raw_result('spotbugs', spotbugs_high_bugs, 'spotbugs.xml')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Failure

    def test_bug_counts_in_props(self, plugin, spotbugs_high_bugs):
        """버그 카운트가 properties에 포함되는지 확인."""
        raw = make_raw_result('spotbugs', spotbugs_high_bugs, 'spotbugs.xml')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        prop_values = {p.name: p.value for p in (obs.props or [])}
        assert prop_values['high-priority-bugs'] == '2'


class TestParseJunitXml:
    """JUnit XML 파싱 테스트."""

    def test_all_pass(self, plugin, junit_all_pass):
        """모든 테스트 통과 → Pass."""
        raw = make_raw_result('junit', junit_all_pass, 'junit.xml')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Pass

    def test_with_failures(self, plugin, junit_with_failures):
        """테스트 실패 → Failure."""
        raw = make_raw_result('junit', junit_with_failures, 'junit.xml')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Failure

    def test_test_counts_in_props(self, plugin, junit_all_pass):
        """테스트 카운트가 properties에 포함되는지 확인."""
        raw = make_raw_result('junit', junit_all_pass, 'junit.xml')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        prop_values = {p.name: p.value for p in (obs.props or [])}
        assert prop_values['total-tests'] == '11'
        assert prop_values['failures'] == '0'


# ── generate_pvp_result() Text 테스트 ────────────────────────────────────────


class TestParseCisJavaText:
    """CIS Java Check 텍스트 파싱 테스트."""

    def test_all_pass(self, plugin, cis_java_all_pass):
        """모든 항목 통과 → Pass."""
        raw = make_raw_result('cis-java', cis_java_all_pass, 'cis-java.txt')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Pass

    def test_with_failures(self, plugin, cis_java_with_failures):
        """실패 항목 있음 → Failure."""
        raw = make_raw_result('cis-java', cis_java_with_failures, 'cis-java.txt')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        assert obs.subjects[0].result == ResultEnum.Failure

    def test_counts_in_props(self, plugin, cis_java_with_failures):
        """카운트가 properties에 포함되는지 확인."""
        raw = make_raw_result('cis-java', cis_java_with_failures, 'cis-java.txt')
        result = plugin.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        prop_values = {p.name: p.value for p in (obs.props or [])}
        assert prop_values['fail-count'] == '2'
        assert prop_values['warn-count'] == '1'
        assert prop_values['pass-count'] == '1'


# ── 통합 테스트 ───────────────────────────────────────────────────────────────


class TestIntegration:
    """전체 파이프라인 통합 테스트."""

    def test_full_pipeline_assessment(self, plugin_with_config):
        """전체 파이프라인 결과로 Assessment Results 생성."""
        pipeline_results = [
            ('gitleaks', {'runs': [{'results': []}]}, 'gitleaks.sarif'),
            ('semgrep', {'runs': [{'results': [{'level': 'warning'}]}]}, 'semgrep.sarif'),
            ('grype', {'matches': []}, 'grype.json'),
            ('trivy', {'Summary': {'passCount': 20, 'failCount': 0}}, 'trivy.json'),
        ]

        all_observations = []
        for tool, data, filepath in pipeline_results:
            raw = make_raw_result(tool, data, filepath)
            pvp_result = plugin_with_config.generate_pvp_result(raw)
            all_observations.extend(pvp_result.observations_by_check)

        # 모든 주요 통제가 커버되는지 확인
        covered_controls = {obs.check_id for obs in all_observations}
        assert 'ia-5' in covered_controls  # Gitleaks
        assert 'sa-11' in covered_controls  # Semgrep
        assert 'ra-5' in covered_controls  # Grype
        assert 'si-3' in covered_controls  # Trivy

    def test_config_propagated_to_observations(self, plugin_with_config):
        """설정값이 observations에 전파되는지 확인."""
        raw = make_raw_result('semgrep', {'runs': [{'results': []}]})
        result = plugin_with_config.generate_pvp_result(raw)
        obs = result.observations_by_check[0]
        prop_values = {p.name: p.value for p in (obs.props or [])}
        assert prop_values['commit-sha'] == 'abc123def456'
        assert 'jccompany-payment-api' in prop_values['pipeline-run']
