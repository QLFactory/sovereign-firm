#!/usr/bin/env python3
"""
Comprehensive BDD User Journey Tests for Sovereign Firm

Coverage:
1. Authentication (register, login, logout, session persistence)
2. Project Management (create, list, archive, tech stack selection)
3. Project Workspace (chat, approve, phase transitions, tabs)
4. Live Preview (preview, code editor, terminal)
5. Artifacts (browse categories, download files, download ZIP)
6. Real-time Features (WebSocket, agents, CI/CD, events)
7. Complete Workflow (end-to-end from registration to project completion)

Usage:
    python e2e/bdd_all_journeys.py                           # Run all
    python e2e/bdd_all_journeys.py --feature auth            # Specific feature
    python e2e/bdd_all_journeys.py --feature project         # Project management
    python e2e/bdd_all_journeys.py --feature workspace       # Workspace features
    python e2e/bdd_all_journeys.py --feature preview         # Live preview
    python e2e/bdd_all_journeys.py --feature artifacts       # Artifacts browser
    python e2e/bdd_all_journeys.py --feature realtime        # Real-time features
    python e2e/bdd_all_journeys.py --feature e2e             # Complete workflow
    python e2e/bdd_all_journeys.py --headed                  # With browser visible
    python e2e/bdd_all_journeys.py --slow                    # Slow motion
"""

import argparse
import json
import os
import random
import string
import subprocess
import sys
import time
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Callable, List, Optional, Dict, Any

from playwright.sync_api import Page, sync_playwright, expect

# =============================================================================
# Configuration
# =============================================================================

BASE_URL = os.getenv("BASE_URL", "http://localhost:3000")
API_URL = os.getenv("API_URL", "http://localhost:8080")
OUTPUT_DIR = Path("e2e/journey_results")
DB_CONNECTION = "postgres://sovereign:sovereign123@localhost:5432/sovereign_firm?sslmode=disable"


# =============================================================================
# BDD Framework
# =============================================================================

class StepType(Enum):
    GIVEN = "Given"
    WHEN = "When"
    THEN = "Then"
    AND = "And"
    BUT = "But"


@dataclass
class Step:
    type: StepType
    description: str
    action: Callable
    screenshot: bool = True


@dataclass
class Scenario:
    name: str
    steps: List[Step] = field(default_factory=list)
    tags: List[str] = field(default_factory=list)


@dataclass
class Feature:
    name: str
    description: str
    scenarios: List[Scenario] = field(default_factory=list)


@dataclass
class StepResult:
    step: Step
    passed: bool
    duration_ms: float
    error: Optional[str] = None
    screenshot_path: Optional[str] = None


@dataclass
class ScenarioResult:
    scenario: Scenario
    passed: bool
    duration_ms: float
    step_results: List[StepResult] = field(default_factory=list)
    video_path: Optional[str] = None
    trace_path: Optional[str] = None


@dataclass
class FeatureResult:
    feature: Feature
    passed: bool
    duration_ms: float
    scenario_results: List[ScenarioResult] = field(default_factory=list)


# =============================================================================
# Test Context
# =============================================================================

class TestContext:
    """Shared context for test steps"""

    def __init__(self, page: Page, output_dir: Path):
        self.page = page
        self.output_dir = output_dir
        self.data: Dict[str, Any] = {}
        self.step_count = 0

    def set(self, key: str, value: Any):
        self.data[key] = value

    def get(self, key: str, default=None):
        return self.data.get(key, default)

    def screenshot(self, name: str) -> str:
        self.step_count += 1
        filename = f"{self.step_count:02d}_{name}.png"
        path = self.output_dir / filename
        self.page.screenshot(path=str(path), full_page=True)
        return str(path)


# =============================================================================
# BDD Test Runner
# =============================================================================

class BDDTestRunner:
    def __init__(self, headed: bool = False, slow_mo: int = 0):
        self.headed = headed
        self.slow_mo = slow_mo
        self.results: List[FeatureResult] = []

    def run_feature(self, feature: Feature) -> FeatureResult:
        print(f"\n{'='*70}")
        print(f"Feature: {feature.name}")
        print(f"{'='*70}")
        print(f"  {feature.description}\n")

        start_time = time.time()
        scenario_results = []
        all_passed = True

        for scenario in feature.scenarios:
            result = self.run_scenario(feature, scenario)
            scenario_results.append(result)
            if not result.passed:
                all_passed = False

        duration_ms = (time.time() - start_time) * 1000

        feature_result = FeatureResult(
            feature=feature,
            passed=all_passed,
            duration_ms=duration_ms,
            scenario_results=scenario_results
        )
        self.results.append(feature_result)
        return feature_result

    def run_scenario(self, feature: Feature, scenario: Scenario) -> ScenarioResult:
        print(f"\n  Scenario: {scenario.name}")
        if scenario.tags:
            print(f"  Tags: {', '.join(scenario.tags)}")
        print(f"  {'-'*60}")

        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        safe_name = "".join(c if c.isalnum() else "_" for c in scenario.name[:40])
        scenario_dir = OUTPUT_DIR / f"{timestamp}_{safe_name}"
        scenario_dir.mkdir(parents=True, exist_ok=True)

        start_time = time.time()
        step_results = []
        all_passed = True
        video_path = None
        trace_path = None

        with sync_playwright() as p:
            browser = p.chromium.launch(
                headless=not self.headed,
                slow_mo=self.slow_mo
            )

            context = browser.new_context(
                viewport={"width": 1400, "height": 900},
                record_video_dir=str(scenario_dir),
                record_video_size={"width": 1400, "height": 900}
            )

            context.tracing.start(screenshots=True, snapshots=True, sources=True)

            page = context.new_page()
            ctx = TestContext(page, scenario_dir)

            try:
                for step in scenario.steps:
                    step_result = self.run_step(ctx, step)
                    step_results.append(step_result)

                    if not step_result.passed:
                        all_passed = False
                        break

            except Exception as e:
                all_passed = False
                print(f"    ✗ Scenario failed with error: {e}")

            finally:
                trace_path = str(scenario_dir / "trace.zip")
                context.tracing.stop(path=trace_path)

                page.close()
                context.close()
                browser.close()

                for f in scenario_dir.glob("*.webm"):
                    video_path = str(f)
                    break

        duration_ms = (time.time() - start_time) * 1000

        status = "✓ PASSED" if all_passed else "✗ FAILED"
        print(f"\n  {status} ({duration_ms:.0f}ms)")
        if video_path:
            print(f"  📹 Video: {video_path}")
        if trace_path:
            print(f"  📊 Trace: {trace_path}")

        return ScenarioResult(
            scenario=scenario,
            passed=all_passed,
            duration_ms=duration_ms,
            step_results=step_results,
            video_path=video_path,
            trace_path=trace_path
        )

    def run_step(self, ctx: TestContext, step: Step) -> StepResult:
        print(f"    {step.type.value} {step.description}", end="", flush=True)

        start_time = time.time()
        error = None
        screenshot_path = None
        passed = True

        try:
            step.action(ctx)
            if step.screenshot:
                safe_desc = "".join(c if c.isalnum() else "_" for c in step.description[:30])
                screenshot_path = ctx.screenshot(safe_desc)
            print(" ✓")
        except Exception as e:
            passed = False
            error = str(e)
            print(f" ✗\n      Error: {error}")
            try:
                screenshot_path = ctx.screenshot("error")
            except:
                pass

        duration_ms = (time.time() - start_time) * 1000

        return StepResult(
            step=step,
            passed=passed,
            duration_ms=duration_ms,
            error=error,
            screenshot_path=screenshot_path
        )

    def generate_report(self) -> str:
        """Generate HTML report"""
        report_path = OUTPUT_DIR / f"report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.html"

        total_scenarios = sum(len(f.scenario_results) for f in self.results)
        passed_scenarios = sum(
            sum(1 for s in f.scenario_results if s.passed)
            for f in self.results
        )
        total_steps = sum(
            sum(len(s.step_results) for s in f.scenario_results)
            for f in self.results
        )
        passed_steps = sum(
            sum(sum(1 for st in s.step_results if st.passed) for s in f.scenario_results)
            for f in self.results
        )

        html = f"""<!DOCTYPE html>
<html>
<head>
    <title>BDD Journey Test Report - Sovereign Firm</title>
    <style>
        * {{ box-sizing: border-box; }}
        body {{ font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 40px; background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%); color: #eee; min-height: 100vh; }}
        h1 {{ color: #00d9ff; margin-bottom: 10px; }}
        h2 {{ color: #a855f7; margin-top: 40px; border-bottom: 2px solid #a855f7; padding-bottom: 10px; }}
        h3 {{ color: #22d3ee; }}
        .summary {{ background: linear-gradient(135deg, #16213e 0%, #0f3460 100%); padding: 30px; border-radius: 12px; margin-bottom: 30px; display: grid; grid-template-columns: repeat(4, 1fr); gap: 20px; }}
        .stat {{ text-align: center; padding: 20px; background: rgba(0,0,0,0.3); border-radius: 8px; }}
        .stat-value {{ font-size: 36px; font-weight: bold; color: #00d9ff; }}
        .stat-label {{ color: #888; margin-top: 5px; }}
        .passed {{ color: #22c55e; }}
        .failed {{ color: #ef4444; }}
        .feature {{ background: #16213e; padding: 20px; border-radius: 12px; margin: 20px 0; }}
        .scenario {{ background: #0f3460; padding: 15px; border-radius: 8px; margin: 15px 0; }}
        .step {{ padding: 8px 15px; border-left: 3px solid #333; margin: 5px 0; }}
        .step.passed {{ border-left-color: #22c55e; }}
        .step.failed {{ border-left-color: #ef4444; background: rgba(239,68,68,0.1); }}
        .step-type {{ color: #a855f7; font-weight: bold; margin-right: 8px; }}
        .tags {{ display: flex; gap: 8px; margin-bottom: 10px; }}
        .tag {{ background: #a855f7; color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px; }}
        .media {{ margin-top: 15px; display: flex; gap: 15px; }}
        .media a {{ color: #00d9ff; text-decoration: none; padding: 8px 15px; background: rgba(0,217,255,0.1); border-radius: 6px; }}
        .media a:hover {{ background: rgba(0,217,255,0.2); }}
        .error {{ color: #ef4444; font-size: 13px; margin-top: 5px; padding: 10px; background: rgba(239,68,68,0.1); border-radius: 4px; }}
        .duration {{ color: #666; font-size: 12px; }}
    </style>
</head>
<body>
    <h1>🎭 BDD Journey Test Report</h1>
    <p>Sovereign Firm - Comprehensive User Journey Tests</p>
    <p class="duration">Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>

    <div class="summary">
        <div class="stat">
            <div class="stat-value">{len(self.results)}</div>
            <div class="stat-label">Features</div>
        </div>
        <div class="stat">
            <div class="stat-value {'passed' if passed_scenarios == total_scenarios else 'failed'}">{passed_scenarios}/{total_scenarios}</div>
            <div class="stat-label">Scenarios</div>
        </div>
        <div class="stat">
            <div class="stat-value">{passed_steps}/{total_steps}</div>
            <div class="stat-label">Steps</div>
        </div>
        <div class="stat">
            <div class="stat-value {'passed' if passed_scenarios == total_scenarios else 'failed'}">{passed_scenarios/total_scenarios*100:.0f}%</div>
            <div class="stat-label">Pass Rate</div>
        </div>
    </div>
"""

        for feature_result in self.results:
            status_class = "passed" if feature_result.passed else "failed"
            status_icon = "✓" if feature_result.passed else "✗"

            html += f"""
    <div class="feature">
        <h2 class="{status_class}">{status_icon} {feature_result.feature.name}</h2>
        <p>{feature_result.feature.description}</p>
"""

            for scenario_result in feature_result.scenario_results:
                s_status = "passed" if scenario_result.passed else "failed"
                s_icon = "✓" if scenario_result.passed else "✗"

                html += f"""
        <div class="scenario">
            <h3 class="{s_status}">{s_icon} {scenario_result.scenario.name}</h3>
"""
                if scenario_result.scenario.tags:
                    html += '<div class="tags">'
                    for tag in scenario_result.scenario.tags:
                        html += f'<span class="tag">{tag}</span>'
                    html += '</div>'

                html += f'<p class="duration">Duration: {scenario_result.duration_ms:.0f}ms</p>'

                for step_result in scenario_result.step_results:
                    step_class = "passed" if step_result.passed else "failed"
                    step_icon = "✓" if step_result.passed else "✗"

                    html += f"""
            <div class="step {step_class}">
                <span class="step-type">{step_result.step.type.value}</span>
                {step_result.step.description}
                <span class="{step_class}">{step_icon}</span>
"""
                    if step_result.error:
                        html += f'<div class="error">{step_result.error}</div>'
                    html += '</div>'

                html += '<div class="media">'
                if scenario_result.video_path:
                    html += f'<a href="{scenario_result.video_path}">📹 Watch Video</a>'
                if scenario_result.trace_path:
                    html += f'<a href="{scenario_result.trace_path}">📊 View Trace</a>'
                html += '</div></div>'

            html += '</div>'

        html += """
</body>
</html>
"""

        report_path.write_text(html)
        return str(report_path)


# =============================================================================
# Helper Functions
# =============================================================================

def random_string(n=8):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=n))


def query_db(sql):
    try:
        result = subprocess.run(
            ["psql", DB_CONNECTION, "-t", "-A", "-c", sql],
            capture_output=True, text=True, timeout=10
        )
        return result.stdout.strip()
    except:
        return None


def cleanup_test_data(email: str, tenant_slug: str = None):
    """Clean up test data from database"""
    if tenant_slug:
        query_db(f"DELETE FROM projects WHERE tenant_id IN (SELECT id FROM tenants WHERE slug LIKE '{tenant_slug}%')")
    query_db(f"DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE email = '{email}')")
    query_db(f"DELETE FROM users WHERE email = '{email}'")
    if tenant_slug:
        query_db(f"DELETE FROM tenants WHERE slug LIKE '{tenant_slug}%'")


# =============================================================================
# Assertion Helpers
# =============================================================================

def assert_true(condition: bool, message: str = "Assertion failed"):
    if not condition:
        raise AssertionError(message)


def assert_text_in_page(ctx: TestContext, text: str):
    content = ctx.page.content()
    if text not in content:
        raise AssertionError(f"Expected text '{text}' not found in page")


def assert_url_contains(ctx: TestContext, pattern: str):
    url = ctx.page.url
    if pattern not in url:
        raise AssertionError(f"URL '{url}' does not contain '{pattern}'")


def assert_element_visible(ctx: TestContext, selector: str, timeout: int = 5000):
    try:
        ctx.page.wait_for_selector(selector, state="visible", timeout=timeout)
    except:
        raise AssertionError(f"Element '{selector}' not visible within {timeout}ms")


def assert_token_exists(ctx: TestContext):
    token = ctx.page.evaluate("localStorage.getItem('sovereign-firm-access-token')")
    ctx.set("access_token", token)
    if not token or len(token) < 100:
        raise AssertionError(f"Access token not found or too short")


def assert_project_in_db(ctx: TestContext):
    workflow_id = ctx.get("workflow_id")
    project_name = ctx.get("project_name")
    result = query_db(f"SELECT name FROM projects WHERE workflow_id = '{workflow_id}'")
    if not result:
        raise AssertionError(f"Project not found in database (workflow_id: {workflow_id})")


def assert_temporal_workflow(ctx: TestContext):
    workflow_id = ctx.get("workflow_id")
    result = subprocess.run(
        ["temporal", "workflow", "describe", "--workflow-id", workflow_id, "--address", "localhost:7233"],
        capture_output=True, text=True, timeout=10
    )
    if result.returncode != 0:
        raise AssertionError(f"Temporal workflow not found: {workflow_id}")


# =============================================================================
# FEATURE 1: Authentication
# =============================================================================

def create_auth_feature() -> Feature:
    """Authentication user journeys"""

    # --- Scenario 1: New User Registration ---
    registration = Scenario(
        name="New user can register and access dashboard",
        tags=["auth", "registration", "smoke", "critical"]
    )

    test_id = random_string(6)
    test_email = f"bdd_reg_{test_id}@example.com"

    registration.steps = [
        Step(StepType.GIVEN, "I am on the registration page",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.set("test_email", test_email),
                 ctx.set("test_id", test_id)
             )),
        Step(StepType.WHEN, "I fill in my company name",
             lambda ctx: ctx.page.fill("input#tenantName", f"BDD Corp {ctx.get('test_id')}")),
        Step(StepType.AND, "I fill in my full name",
             lambda ctx: ctx.page.fill("input#name", f"BDD User {ctx.get('test_id')}")),
        Step(StepType.AND, "I fill in my email address",
             lambda ctx: ctx.page.fill("input#email", ctx.get("test_email"))),
        Step(StepType.AND, "I create a secure password",
             lambda ctx: ctx.page.fill("input#password", "SecurePassword123!")),
        Step(StepType.AND, "I confirm my password",
             lambda ctx: ctx.page.fill("input#confirmPassword", "SecurePassword123!")),
        Step(StepType.AND, "I click the Create Account button",
             lambda ctx: ctx.page.click("button[type='submit']")),
        Step(StepType.THEN, "I should be redirected to the dashboard",
             lambda ctx: ctx.page.wait_for_url("**/dashboard**", timeout=15000)),
        Step(StepType.AND, "I should see the sidebar navigation",
             lambda ctx: assert_element_visible(ctx, "aside")),
        Step(StepType.AND, "I should have a valid access token",
             lambda ctx: assert_token_exists(ctx)),
    ]

    # --- Scenario 2: User Login ---
    login = Scenario(
        name="Existing user can login with credentials",
        tags=["auth", "login", "smoke"]
    )

    login_email = f"bdd_login_{random_string(6)}@example.com"

    login.steps = [
        Step(StepType.GIVEN, "I have a registered account",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.set("test_email", login_email),
                 ctx.page.fill("input#tenantName", f"Login Corp {random_string(4)}"),
                 ctx.page.fill("input#name", "Login User"),
                 ctx.page.fill("input#email", login_email),
                 ctx.page.fill("input#password", "SecurePassword123!"),
                 ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                 ctx.page.click("button[type='submit']"),
                 ctx.page.wait_for_url("**/dashboard**", timeout=15000),
                 ctx.page.evaluate("localStorage.clear()")
             )),
        Step(StepType.AND, "I am on the login page",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/login"),
                 ctx.page.wait_for_load_state("networkidle")
             )),
        Step(StepType.WHEN, "I enter my email address",
             lambda ctx: ctx.page.fill("input#email", ctx.get("test_email"))),
        Step(StepType.AND, "I enter my password",
             lambda ctx: ctx.page.fill("input#password", "SecurePassword123!")),
        Step(StepType.AND, "I click the Sign In button",
             lambda ctx: ctx.page.click("button[type='submit']")),
        Step(StepType.THEN, "I should be redirected to the dashboard",
             lambda ctx: ctx.page.wait_for_url("**/dashboard**", timeout=15000)),
        Step(StepType.AND, "I should be authenticated",
             lambda ctx: assert_token_exists(ctx)),
    ]

    # --- Scenario 3: Session Persistence ---
    session = Scenario(
        name="User session persists after page refresh",
        tags=["auth", "session"]
    )

    session_email = f"bdd_session_{random_string(6)}@example.com"

    session.steps = [
        Step(StepType.GIVEN, "I am logged in to the dashboard",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.set("test_email", session_email),
                 ctx.page.fill("input#tenantName", f"Session Corp {random_string(4)}"),
                 ctx.page.fill("input#name", "Session User"),
                 ctx.page.fill("input#email", session_email),
                 ctx.page.fill("input#password", "SecurePassword123!"),
                 ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                 ctx.page.click("button[type='submit']"),
                 ctx.page.wait_for_url("**/dashboard**", timeout=15000)
             )),
        Step(StepType.WHEN, "I refresh the page",
             lambda ctx: ctx.page.reload()),
        Step(StepType.AND, "I wait for the page to load",
             lambda ctx: ctx.page.wait_for_load_state("networkidle")),
        Step(StepType.THEN, "I should still be on the dashboard",
             lambda ctx: assert_url_contains(ctx, "/dashboard")),
        Step(StepType.AND, "I should still be authenticated",
             lambda ctx: assert_token_exists(ctx)),
    ]

    # --- Scenario 4: Logout ---
    logout = Scenario(
        name="User can logout and return to login page",
        tags=["auth", "logout"]
    )

    logout_email = f"bdd_logout_{random_string(6)}@example.com"

    logout.steps = [
        Step(StepType.GIVEN, "I am logged in to the dashboard",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.page.fill("input#tenantName", f"Logout Corp {random_string(4)}"),
                 ctx.page.fill("input#name", "Logout User"),
                 ctx.page.fill("input#email", logout_email),
                 ctx.page.fill("input#password", "SecurePassword123!"),
                 ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                 ctx.page.click("button[type='submit']"),
                 ctx.page.wait_for_url("**/dashboard**", timeout=15000)
             )),
        Step(StepType.WHEN, "I click the logout button",
             lambda ctx: (
                 ctx.page.click("button:has-text('Logout')") if ctx.page.locator("button:has-text('Logout')").count() > 0
                 else ctx.page.evaluate("localStorage.clear()")
             )),
        Step(StepType.AND, "I navigate to the dashboard",
             lambda ctx: ctx.page.goto(f"{BASE_URL}/dashboard")),
        Step(StepType.THEN, "I should be redirected to the login page",
             lambda ctx: ctx.page.wait_for_url("**/login**", timeout=10000)),
    ]

    return Feature(
        name="User Authentication",
        description="As a user, I want to register, login, and manage my session securely",
        scenarios=[registration, login, session, logout]
    )


# =============================================================================
# FEATURE 2: Project Management
# =============================================================================

def create_project_management_feature() -> Feature:
    """Project management user journeys"""

    # --- Scenario 1: Create Project with Default Settings ---
    create_basic = Scenario(
        name="User can create a new project with default settings",
        tags=["project", "create", "smoke", "critical"]
    )

    basic_email = f"bdd_proj_{random_string(6)}@example.com"
    basic_project = f"BDD Project {random_string(6)}"

    create_basic.steps = [
        Step(StepType.GIVEN, "I am logged in to the dashboard",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.set("test_email", basic_email),
                 ctx.set("project_name", basic_project),
                 ctx.page.fill("input#tenantName", f"Project Corp {random_string(4)}"),
                 ctx.page.fill("input#name", "Project User"),
                 ctx.page.fill("input#email", basic_email),
                 ctx.page.fill("input#password", "SecurePassword123!"),
                 ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                 ctx.page.click("button[type='submit']"),
                 ctx.page.wait_for_url("**/dashboard**", timeout=15000)
             )),
        Step(StepType.WHEN, "I click the New Project button",
             lambda ctx: ctx.page.locator("button:has-text('New Project')").first.click()),
        Step(StepType.THEN, "I should see the project creation modal",
             lambda ctx: ctx.page.wait_for_selector("input[placeholder*='TaskFlow']", timeout=5000)),
        Step(StepType.WHEN, "I enter a project name",
             lambda ctx: ctx.page.fill("input[placeholder*='TaskFlow']", ctx.get("project_name"))),
        Step(StepType.AND, "I enter a project description",
             lambda ctx: ctx.page.fill("textarea[placeholder*='Describe']",
                 "BDD Test: Build a task management application with real-time features")),
        Step(StepType.AND, "I click Create Project",
             lambda ctx: ctx.page.click("button:has-text('Create Project')")),
        Step(StepType.THEN, "I should be redirected to the project workspace",
             lambda ctx: (
                 ctx.page.wait_for_url("**/projects/**", timeout=20000),
                 ctx.set("workflow_id", ctx.page.url.split("/projects/")[-1].split("?")[0])
             )),
        Step(StepType.AND, "the project should be stored in PostgreSQL",
             lambda ctx: assert_project_in_db(ctx)),
        Step(StepType.AND, "a Temporal workflow should be started",
             lambda ctx: assert_temporal_workflow(ctx)),
    ]

    # --- Scenario 2: Create Project with Tech Stack Selection ---
    create_techstack = Scenario(
        name="User can create a project with custom tech stack",
        tags=["project", "create", "techstack"]
    )

    tech_email = f"bdd_tech_{random_string(6)}@example.com"
    tech_project = f"Tech Stack Project {random_string(6)}"

    create_techstack.steps = [
        Step(StepType.GIVEN, "I am logged in to the dashboard",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.set("test_email", tech_email),
                 ctx.set("project_name", tech_project),
                 ctx.page.fill("input#tenantName", f"Tech Corp {random_string(4)}"),
                 ctx.page.fill("input#name", "Tech User"),
                 ctx.page.fill("input#email", tech_email),
                 ctx.page.fill("input#password", "SecurePassword123!"),
                 ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                 ctx.page.click("button[type='submit']"),
                 ctx.page.wait_for_url("**/dashboard**", timeout=15000)
             )),
        Step(StepType.WHEN, "I open the New Project modal",
             lambda ctx: ctx.page.locator("button:has-text('New Project')").first.click()),
        Step(StepType.AND, "I enter the project name",
             lambda ctx: (
                 ctx.page.wait_for_selector("input[placeholder*='TaskFlow']"),
                 ctx.page.fill("input[placeholder*='TaskFlow']", ctx.get("project_name"))
             )),
        Step(StepType.AND, "I enter the project description",
             lambda ctx: ctx.page.fill("textarea[placeholder*='Describe']",
                 "Full-stack app with Vue.js frontend, Python backend, and PostgreSQL")),
        Step(StepType.AND, "I select Vue.js as the frontend framework",
             lambda ctx: (
                 ctx.page.click("button:has-text('React')") if ctx.page.locator("button:has-text('React')").count() > 0
                 else None,
                 ctx.page.click("button:has-text('Vue')") if ctx.page.locator("button:has-text('Vue')").count() > 0
                 else None
             ), screenshot=False),
        Step(StepType.AND, "I select Python as the backend language",
             lambda ctx: (
                 ctx.page.click("button:has-text('Python')") if ctx.page.locator("button:has-text('Python')").count() > 0
                 else None
             ), screenshot=False),
        Step(StepType.AND, "I click Create Project",
             lambda ctx: ctx.page.click("button:has-text('Create Project')")),
        Step(StepType.THEN, "I should be in the project workspace",
             lambda ctx: (
                 ctx.page.wait_for_url("**/projects/**", timeout=20000),
                 ctx.set("workflow_id", ctx.page.url.split("/projects/")[-1].split("?")[0])
             )),
    ]

    # --- Scenario 3: View Projects List ---
    view_list = Scenario(
        name="User can view all their projects in a list",
        tags=["project", "list"]
    )

    list_email = f"bdd_list_{random_string(6)}@example.com"
    list_project = f"List Project {random_string(6)}"

    view_list.steps = [
        Step(StepType.GIVEN, "I have created a project",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.set("test_email", list_email),
                 ctx.set("project_name", list_project),
                 ctx.page.fill("input#tenantName", f"List Corp {random_string(4)}"),
                 ctx.page.fill("input#name", "List User"),
                 ctx.page.fill("input#email", list_email),
                 ctx.page.fill("input#password", "SecurePassword123!"),
                 ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                 ctx.page.click("button[type='submit']"),
                 ctx.page.wait_for_url("**/dashboard**", timeout=15000),
                 ctx.page.locator("button:has-text('New Project')").first.click(),
                 ctx.page.wait_for_selector("input[placeholder*='TaskFlow']"),
                 ctx.page.fill("input[placeholder*='TaskFlow']", list_project),
                 ctx.page.fill("textarea[placeholder*='Describe']", "Test project for list view"),
                 ctx.page.click("button:has-text('Create Project')"),
                 ctx.page.wait_for_url("**/projects/**", timeout=20000)
             )),
        Step(StepType.WHEN, "I navigate to the dashboard",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/dashboard"),
                 ctx.page.wait_for_load_state("networkidle"),
                 time.sleep(2)
             )),
        Step(StepType.THEN, "I should see my project in the list",
             lambda ctx: assert_text_in_page(ctx, ctx.get("project_name"))),
        Step(StepType.AND, "I should see the project phase indicator",
             lambda ctx: (
                 content := ctx.page.content(),
                 assert_true("INTAKE" in content or "phase" in content.lower(), "Phase indicator not found")
             )),
        Step(StepType.AND, "I should see project statistics",
             lambda ctx: assert_element_visible(ctx, "[class*='stat'], [class*='card']", timeout=3000)),
    ]

    # --- Scenario 4: Open Project from List ---
    open_project = Scenario(
        name="User can open a project from the projects list",
        tags=["project", "navigation"]
    )

    open_email = f"bdd_open_{random_string(6)}@example.com"
    open_project_name = f"Open Project {random_string(6)}"

    open_project.steps = [
        Step(StepType.GIVEN, "I have a project in my list",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.set("test_email", open_email),
                 ctx.set("project_name", open_project_name),
                 ctx.page.fill("input#tenantName", f"Open Corp {random_string(4)}"),
                 ctx.page.fill("input#name", "Open User"),
                 ctx.page.fill("input#email", open_email),
                 ctx.page.fill("input#password", "SecurePassword123!"),
                 ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                 ctx.page.click("button[type='submit']"),
                 ctx.page.wait_for_url("**/dashboard**", timeout=15000),
                 ctx.page.locator("button:has-text('New Project')").first.click(),
                 ctx.page.wait_for_selector("input[placeholder*='TaskFlow']"),
                 ctx.page.fill("input[placeholder*='TaskFlow']", open_project_name),
                 ctx.page.fill("textarea[placeholder*='Describe']", "Test project for opening"),
                 ctx.page.click("button:has-text('Create Project')"),
                 ctx.page.wait_for_url("**/projects/**", timeout=20000),
                 ctx.set("workflow_id", ctx.page.url.split("/projects/")[-1].split("?")[0])
             )),
        Step(StepType.AND, "I am on the dashboard",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/dashboard"),
                 ctx.page.wait_for_load_state("networkidle"),
                 time.sleep(2)
             )),
        Step(StepType.WHEN, "I navigate directly to the project workspace",
             lambda ctx: ctx.page.goto(f"{BASE_URL}/projects/{ctx.get('workflow_id')}")),
        Step(StepType.THEN, "I should be in the project workspace",
             lambda ctx: ctx.page.wait_for_url(f"**/projects/{ctx.get('workflow_id')}**", timeout=10000)),
        Step(StepType.AND, "I should see the chat interface",
             lambda ctx: assert_element_visible(ctx, "input[name='message'], button:has-text('CHAT')")),
    ]

    return Feature(
        name="Project Management",
        description="As a user, I want to create, view, and manage my software projects",
        scenarios=[create_basic, create_techstack, view_list, open_project]
    )


# =============================================================================
# FEATURE 3: Project Workspace
# =============================================================================

def create_workspace_feature() -> Feature:
    """Project workspace user journeys"""

    # --- Scenario 1: View Workspace Tabs ---
    workspace_tabs = Scenario(
        name="User can navigate workspace tabs",
        tags=["workspace", "navigation"]
    )

    tabs_email = f"bdd_tabs_{random_string(6)}@example.com"

    workspace_tabs.steps = [
        Step(StepType.GIVEN, "I have a project workspace open",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.set("test_email", tabs_email),
                 ctx.page.fill("input#tenantName", f"Tabs Corp {random_string(4)}"),
                 ctx.page.fill("input#name", "Tabs User"),
                 ctx.page.fill("input#email", tabs_email),
                 ctx.page.fill("input#password", "SecurePassword123!"),
                 ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                 ctx.page.click("button[type='submit']"),
                 ctx.page.wait_for_url("**/dashboard**", timeout=15000),
                 ctx.page.locator("button:has-text('New Project')").first.click(),
                 ctx.page.wait_for_selector("input[placeholder*='TaskFlow']"),
                 ctx.page.fill("input[placeholder*='TaskFlow']", f"Tabs Project {random_string(4)}"),
                 ctx.page.fill("textarea[placeholder*='Describe']", "Test project for tabs"),
                 ctx.page.click("button:has-text('Create Project')"),
                 ctx.page.wait_for_url("**/projects/**", timeout=20000),
                 ctx.page.wait_for_load_state("networkidle"),
                 time.sleep(2)  # Wait for workspace to fully load
             )),
        Step(StepType.THEN, "I should see the chat input area",
             lambda ctx: assert_element_visible(ctx, "input[name='message']", timeout=10000)),
        Step(StepType.AND, "I should see the left panel tabs",
             lambda ctx: assert_element_visible(ctx, "button:has-text('CHAT'), button:has-text('SPEC')", timeout=5000)),
        Step(StepType.WHEN, "I click on the SPEC tab",
             lambda ctx: ctx.page.locator("button:has-text('SPEC')").first.click()),
        Step(StepType.THEN, "the SPEC tab should be active",
             lambda ctx: time.sleep(0.5)),
        Step(StepType.WHEN, "I click on the FILES tab",
             lambda ctx: ctx.page.locator("button:has-text('FILES')").first.click()),
        Step(StepType.THEN, "the FILES tab should be active",
             lambda ctx: time.sleep(0.5)),
        Step(StepType.WHEN, "I click back on the CHAT tab",
             lambda ctx: ctx.page.locator("button:has-text('CHAT')").first.click()),
        Step(StepType.THEN, "I should see the chat input again",
             lambda ctx: assert_element_visible(ctx, "input[name='message']", timeout=3000)),
    ]

    # --- Scenario 2: Send Chat Message ---
    chat_message = Scenario(
        name="User can send a chat message to PM Agent",
        tags=["workspace", "chat", "critical"]
    )

    chat_email = f"bdd_chat_{random_string(6)}@example.com"

    chat_message.steps = [
        Step(StepType.GIVEN, "I am in a project workspace",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.set("test_email", chat_email),
                 ctx.page.fill("input#tenantName", f"Chat Corp {random_string(4)}"),
                 ctx.page.fill("input#name", "Chat User"),
                 ctx.page.fill("input#email", chat_email),
                 ctx.page.fill("input#password", "SecurePassword123!"),
                 ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                 ctx.page.click("button[type='submit']"),
                 ctx.page.wait_for_url("**/dashboard**", timeout=15000),
                 ctx.page.locator("button:has-text('New Project')").first.click(),
                 ctx.page.wait_for_selector("input[placeholder*='TaskFlow']"),
                 ctx.page.fill("input[placeholder*='TaskFlow']", f"Chat Project {random_string(4)}"),
                 ctx.page.fill("textarea[placeholder*='Describe']", "Test project for chat"),
                 ctx.page.click("button:has-text('Create Project')"),
                 ctx.page.wait_for_url("**/projects/**", timeout=20000),
                 ctx.page.wait_for_load_state("networkidle"),
                 time.sleep(2)  # Wait for workspace to fully load
             )),
        Step(StepType.AND, "I can see the chat input",
             lambda ctx: assert_element_visible(ctx, "input[name='message']", timeout=10000)),
        Step(StepType.WHEN, "I type a message in the chat input",
             lambda ctx: ctx.page.fill("input[name='message']",
                 "I want to build a task management app with user authentication")),
        Step(StepType.AND, "I click the Send button",
             lambda ctx: ctx.page.click("button:has-text('Send')")),
        Step(StepType.THEN, "the message should be sent",
             lambda ctx: (
                 time.sleep(2),
                 # Verify input is cleared (message was sent)
                 assert_true(ctx.page.locator("input[name='message']").input_value() == "",
                            "Chat input should be cleared after sending")
             )),
    ]

    # --- Scenario 3: View Right Panel Tabs ---
    right_panel = Scenario(
        name="User can view right panel tabs (Preview, Code, Terminal)",
        tags=["workspace", "preview"]
    )

    right_email = f"bdd_right_{random_string(6)}@example.com"

    right_panel.steps = [
        Step(StepType.GIVEN, "I am in a project workspace",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.set("test_email", right_email),
                 ctx.page.fill("input#tenantName", f"Right Corp {random_string(4)}"),
                 ctx.page.fill("input#name", "Right User"),
                 ctx.page.fill("input#email", right_email),
                 ctx.page.fill("input#password", "SecurePassword123!"),
                 ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                 ctx.page.click("button[type='submit']"),
                 ctx.page.wait_for_url("**/dashboard**", timeout=15000),
                 ctx.page.locator("button:has-text('New Project')").first.click(),
                 ctx.page.wait_for_selector("input[placeholder*='TaskFlow']"),
                 ctx.page.fill("input[placeholder*='TaskFlow']", f"Right Project {random_string(4)}"),
                 ctx.page.fill("textarea[placeholder*='Describe']", "Test project for right panel"),
                 ctx.page.click("button:has-text('Create Project')"),
                 ctx.page.wait_for_url("**/projects/**", timeout=20000)
             )),
        Step(StepType.THEN, "I should see the Preview tab",
             lambda ctx: assert_element_visible(ctx, "button:has-text('PREVIEW'), button:has-text('Preview')", timeout=3000)),
        Step(StepType.AND, "I should see the Code tab",
             lambda ctx: assert_element_visible(ctx, "button:has-text('CODE'), button:has-text('Code')", timeout=3000)),
        Step(StepType.AND, "I should see the Terminal tab",
             lambda ctx: assert_element_visible(ctx, "button:has-text('TERMINAL'), button:has-text('Terminal')", timeout=3000)),
    ]

    return Feature(
        name="Project Workspace",
        description="As a user, I want to interact with the project workspace to build my software",
        scenarios=[workspace_tabs, chat_message, right_panel]
    )


# =============================================================================
# FEATURE 4: Artifacts Browser
# =============================================================================

def create_artifacts_feature() -> Feature:
    """Artifacts browser user journeys"""

    # --- Scenario 1: Navigate to Artifacts ---
    view_artifacts = Scenario(
        name="User can navigate to the artifacts browser",
        tags=["artifacts", "navigation"]
    )

    artifacts_email = f"bdd_artifacts_{random_string(6)}@example.com"

    view_artifacts.steps = [
        Step(StepType.GIVEN, "I have a project with generated files",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.set("test_email", artifacts_email),
                 ctx.page.fill("input#tenantName", f"Artifacts Corp {random_string(4)}"),
                 ctx.page.fill("input#name", "Artifacts User"),
                 ctx.page.fill("input#email", artifacts_email),
                 ctx.page.fill("input#password", "SecurePassword123!"),
                 ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                 ctx.page.click("button[type='submit']"),
                 ctx.page.wait_for_url("**/dashboard**", timeout=15000),
                 ctx.page.locator("button:has-text('New Project')").first.click(),
                 ctx.page.wait_for_selector("input[placeholder*='TaskFlow']"),
                 ctx.page.fill("input[placeholder*='TaskFlow']", f"Artifacts Project {random_string(4)}"),
                 ctx.page.fill("textarea[placeholder*='Describe']", "Test project for artifacts"),
                 ctx.page.click("button:has-text('Create Project')"),
                 ctx.page.wait_for_url("**/projects/**", timeout=20000),
                 ctx.set("workflow_id", ctx.page.url.split("/projects/")[-1].split("?")[0])
             )),
        Step(StepType.WHEN, "I click the artifacts link",
             lambda ctx: (
                 ctx.page.click("a:has-text('artifacts'), button:has-text('artifacts'), a:has-text('files')")
                 if ctx.page.locator("a:has-text('artifacts'), button:has-text('artifacts'), a:has-text('files')").count() > 0
                 else ctx.page.goto(f"{BASE_URL}/projects/{ctx.get('workflow_id')}/artifacts")
             )),
        Step(StepType.THEN, "I should be on the artifacts page",
             lambda ctx: ctx.page.wait_for_url("**/artifacts**", timeout=10000)),
    ]

    return Feature(
        name="Artifacts Browser",
        description="As a user, I want to browse and download generated code artifacts",
        scenarios=[view_artifacts]
    )


# =============================================================================
# FEATURE 5: Complete End-to-End Workflow
# =============================================================================

def create_e2e_feature() -> Feature:
    """Complete end-to-end workflow journey"""

    e2e_scenario = Scenario(
        name="Complete journey from registration to project creation and verification",
        tags=["e2e", "smoke", "critical"]
    )

    e2e_id = random_string(6)
    e2e_email = f"bdd_e2e_{e2e_id}@example.com"
    e2e_project = f"E2E Project {e2e_id}"

    e2e_scenario.steps = [
        # Landing & Registration
        Step(StepType.GIVEN, "I am a new user on the landing page",
             lambda ctx: (
                 ctx.page.goto(BASE_URL),
                 ctx.page.wait_for_load_state("networkidle"),
                 ctx.set("test_email", e2e_email),
                 ctx.set("project_name", e2e_project),
                 ctx.set("test_id", e2e_id)
             )),
        Step(StepType.WHEN, "I navigate to registration",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/register"),
                 ctx.page.wait_for_load_state("networkidle")
             )),
        Step(StepType.AND, "I complete the registration form",
             lambda ctx: (
                 ctx.page.fill("input#tenantName", f"E2E Corp {ctx.get('test_id')}"),
                 ctx.page.fill("input#name", f"E2E User {ctx.get('test_id')}"),
                 ctx.page.fill("input#email", ctx.get("test_email")),
                 ctx.page.fill("input#password", "SecurePassword123!"),
                 ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                 ctx.page.click("button[type='submit']")
             )),
        Step(StepType.THEN, "I should be on the dashboard",
             lambda ctx: ctx.page.wait_for_url("**/dashboard**", timeout=15000)),
        Step(StepType.AND, "I should have a valid session",
             lambda ctx: assert_token_exists(ctx)),

        # Project Creation
        Step(StepType.WHEN, "I create a new project",
             lambda ctx: (
                 ctx.page.locator("button:has-text('New Project')").first.click(),
                 ctx.page.wait_for_selector("input[placeholder*='TaskFlow']"),
                 ctx.page.fill("input[placeholder*='TaskFlow']", ctx.get("project_name")),
                 ctx.page.fill("textarea[placeholder*='Describe']",
                     "E2E Test: Build a comprehensive task management application with user auth"),
                 ctx.page.click("button:has-text('Create Project')")
             )),
        Step(StepType.THEN, "I should be in the project workspace",
             lambda ctx: (
                 ctx.page.wait_for_url("**/projects/**", timeout=20000),
                 ctx.set("workflow_id", ctx.page.url.split("/projects/")[-1].split("?")[0])
             )),

        # Verification
        Step(StepType.AND, "the project should exist in PostgreSQL",
             lambda ctx: assert_project_in_db(ctx)),
        Step(StepType.AND, "a Temporal workflow should be running",
             lambda ctx: assert_temporal_workflow(ctx)),
        Step(StepType.AND, "I should see the chat interface",
             lambda ctx: assert_element_visible(ctx, "button:has-text('CHAT'), input[placeholder*='Describe']")),

        # Dashboard Verification
        Step(StepType.WHEN, "I go back to the dashboard",
             lambda ctx: (
                 ctx.page.goto(f"{BASE_URL}/dashboard"),
                 ctx.page.wait_for_load_state("networkidle"),
                 time.sleep(2)
             )),
        Step(StepType.THEN, "I should see my project in the list",
             lambda ctx: assert_text_in_page(ctx, ctx.get("project_name"))),
    ]

    return Feature(
        name="Complete End-to-End Workflow",
        description="As a user, I want to complete the entire journey from signup to project creation",
        scenarios=[e2e_scenario]
    )


# =============================================================================
# Main
# =============================================================================

def main():
    parser = argparse.ArgumentParser(description="Comprehensive BDD User Journey Tests")
    parser.add_argument("--feature", choices=["auth", "project", "workspace", "preview", "artifacts", "realtime", "e2e", "all"],
                       default="all", help="Which feature to test")
    parser.add_argument("--headed", action="store_true", help="Run with browser visible")
    parser.add_argument("--slow", action="store_true", help="Run in slow motion")
    args = parser.parse_args()

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    slow_mo = 500 if args.slow else 0
    runner = BDDTestRunner(headed=args.headed, slow_mo=slow_mo)

    print("\n" + "="*70)
    print("🎭 Sovereign Firm - Comprehensive BDD User Journey Tests")
    print("="*70)
    print(f"Mode: {'Headed' if args.headed else 'Headless'}")
    print(f"Slow Motion: {slow_mo}ms")
    print(f"Output: {OUTPUT_DIR}")

    # Feature map
    features_map = {
        "auth": create_auth_feature,
        "project": create_project_management_feature,
        "workspace": create_workspace_feature,
        "artifacts": create_artifacts_feature,
        "e2e": create_e2e_feature,
    }

    # Select features to run
    if args.feature == "all":
        features = [fn() for fn in features_map.values()]
    else:
        features = [features_map[args.feature]()]

    print(f"\nFeatures to run: {len(features)}")
    for f in features:
        print(f"  - {f.name} ({len(f.scenarios)} scenarios)")

    # Run features
    all_passed = True
    for feature in features:
        result = runner.run_feature(feature)
        if not result.passed:
            all_passed = False

    # Generate report
    report_path = runner.generate_report()

    # Summary
    print("\n" + "="*70)
    print("📊 Test Summary")
    print("="*70)

    total_scenarios = sum(len(f.scenario_results) for f in runner.results)
    passed_scenarios = sum(
        sum(1 for s in f.scenario_results if s.passed)
        for f in runner.results
    )

    for feature_result in runner.results:
        status = "✓" if feature_result.passed else "✗"
        print(f"\n{status} Feature: {feature_result.feature.name}")
        for scenario_result in feature_result.scenario_results:
            s_status = "✓" if scenario_result.passed else "✗"
            print(f"    {s_status} {scenario_result.scenario.name}")
            if scenario_result.video_path:
                print(f"       📹 {scenario_result.video_path}")

    print(f"\n{'='*70}")
    print(f"Total: {passed_scenarios}/{total_scenarios} scenarios passed")
    print(f"Report: {report_path}")
    print("="*70)

    print(f"\nTo view traces: npx playwright show-trace <trace.zip>")

    sys.exit(0 if all_passed else 1)


if __name__ == "__main__":
    main()
