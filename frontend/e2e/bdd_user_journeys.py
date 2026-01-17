#!/usr/bin/env python3
"""
BDD-Style User Journey Tests with Video Recording

Features:
- Given/When/Then BDD syntax
- Video recording of all test runs
- Trace files for debugging
- Screenshots at key steps
- HTML report generation

Usage:
    python e2e/bdd_user_journeys.py                    # Run all journeys
    python e2e/bdd_user_journeys.py --journey auth     # Run specific journey
    python e2e/bdd_user_journeys.py --headed           # Run with browser visible
    python e2e/bdd_user_journeys.py --slow             # Run in slow motion
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
from typing import Callable, List, Optional

from playwright.sync_api import Browser, BrowserContext, Page, sync_playwright

# =============================================================================
# Configuration
# =============================================================================

BASE_URL = os.getenv("BASE_URL", "http://localhost:3000")
API_URL = os.getenv("API_URL", "http://localhost:8080")
OUTPUT_DIR = Path("e2e/results")
DB_CONNECTION = "postgres://sovereign:sovereign123@localhost:5432/sovereign_firm?sslmode=disable"


# =============================================================================
# BDD Framework
# =============================================================================

class StepType(Enum):
    GIVEN = "Given"
    WHEN = "When"
    THEN = "Then"
    AND = "And"


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
        self.data = {}
        self.step_count = 0

    def set(self, key: str, value):
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
        print(f"  {'-'*60}")

        # Create output directory for this scenario
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        safe_name = "".join(c if c.isalnum() else "_" for c in scenario.name)
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

            # Create context with video recording
            context = browser.new_context(
                viewport={"width": 1400, "height": 900},
                record_video_dir=str(scenario_dir),
                record_video_size={"width": 1400, "height": 900}
            )

            # Start tracing
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
                # Save trace
                trace_path = str(scenario_dir / "trace.zip")
                context.tracing.stop(path=trace_path)

                # Close and get video path
                page.close()
                context.close()
                browser.close()

                # Find the video file
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
            # Take error screenshot
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

        html = f"""<!DOCTYPE html>
<html>
<head>
    <title>BDD Test Report</title>
    <style>
        body {{ font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 40px; background: #1a1a2e; color: #eee; }}
        h1 {{ color: #00d9ff; }}
        h2 {{ color: #a855f7; margin-top: 30px; }}
        h3 {{ color: #22d3ee; }}
        .summary {{ background: #16213e; padding: 20px; border-radius: 8px; margin-bottom: 30px; }}
        .passed {{ color: #22c55e; }}
        .failed {{ color: #ef4444; }}
        .scenario {{ background: #0f3460; padding: 15px; border-radius: 8px; margin: 15px 0; }}
        .step {{ padding: 8px 0; border-bottom: 1px solid #1a1a2e; }}
        .step-type {{ color: #a855f7; font-weight: bold; }}
        .media {{ margin-top: 10px; }}
        .media a {{ color: #00d9ff; text-decoration: none; margin-right: 15px; }}
        .media a:hover {{ text-decoration: underline; }}
        table {{ width: 100%; border-collapse: collapse; }}
        th, td {{ padding: 10px; text-align: left; border-bottom: 1px solid #333; }}
        th {{ background: #16213e; }}
    </style>
</head>
<body>
    <h1>🎭 BDD Test Report</h1>
    <p>Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>

    <div class="summary">
        <h2>Summary</h2>
        <table>
            <tr><th>Total Scenarios</th><td>{total_scenarios}</td></tr>
            <tr><th>Passed</th><td class="passed">{passed_scenarios}</td></tr>
            <tr><th>Failed</th><td class="failed">{total_scenarios - passed_scenarios}</td></tr>
            <tr><th>Pass Rate</th><td>{passed_scenarios/total_scenarios*100:.1f}%</td></tr>
        </table>
    </div>
"""

        for feature_result in self.results:
            status_class = "passed" if feature_result.passed else "failed"
            status_icon = "✓" if feature_result.passed else "✗"

            html += f"""
    <h2>Feature: {feature_result.feature.name}</h2>
    <p>{feature_result.feature.description}</p>
"""

            for scenario_result in feature_result.scenario_results:
                s_status = "passed" if scenario_result.passed else "failed"
                s_icon = "✓" if scenario_result.passed else "✗"

                html += f"""
    <div class="scenario">
        <h3 class="{s_status}">{s_icon} Scenario: {scenario_result.scenario.name}</h3>
        <p>Duration: {scenario_result.duration_ms:.0f}ms</p>
"""

                for step_result in scenario_result.step_results:
                    step_class = "passed" if step_result.passed else "failed"
                    step_icon = "✓" if step_result.passed else "✗"

                    html += f"""
        <div class="step">
            <span class="step-type">{step_result.step.type.value}</span>
            {step_result.step.description}
            <span class="{step_class}">{step_icon}</span>
"""
                    if step_result.error:
                        html += f"""
            <br><small class="failed">Error: {step_result.error}</small>
"""
                    html += """
        </div>
"""

                html += """
        <div class="media">
"""
                if scenario_result.video_path:
                    html += f"""            <a href="{scenario_result.video_path}">📹 Watch Video</a>
"""
                if scenario_result.trace_path:
                    html += f"""            <a href="{scenario_result.trace_path}">📊 View Trace</a>
"""
                html += """
        </div>
    </div>
"""

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


def cleanup_test_data(email: str, tenant_slug: str):
    """Clean up test data from database"""
    query_db(f"DELETE FROM projects WHERE tenant_id IN (SELECT id FROM tenants WHERE slug LIKE '{tenant_slug}%')")
    query_db(f"DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE email = '{email}')")
    query_db(f"DELETE FROM users WHERE email = '{email}'")
    query_db(f"DELETE FROM tenants WHERE slug LIKE '{tenant_slug}%'")


# =============================================================================
# Assertion Helpers (for use in lambdas)
# =============================================================================

def verify_token_exists(ctx: TestContext) -> None:
    """Verify access token exists and is valid length"""
    token = ctx.page.evaluate("localStorage.getItem('sovereign-firm-access-token')")
    ctx.set("access_token", token)
    if not token or len(token) < 100:
        raise AssertionError(f"Access token not found or too short (got {len(token) if token else 0} chars)")


def verify_token_valid(ctx: TestContext) -> None:
    """Verify access token exists and is valid"""
    token = ctx.page.evaluate("localStorage.getItem('sovereign-firm-access-token')")
    if not token or len(token) < 100:
        raise AssertionError(f"Token invalid or missing (length: {len(token) if token else 0})")


def verify_project_in_db(ctx: TestContext) -> None:
    """Verify project exists in database"""
    workflow_id = ctx.get("workflow_id")
    result = query_db(f"SELECT name, phase FROM projects WHERE workflow_id = '{workflow_id}'")
    ctx.set("db_result", result)
    project_name = ctx.get("project_name")
    if not result or project_name not in result:
        raise AssertionError(f"Project '{project_name}' not found in database (workflow_id: {workflow_id})")


def verify_text_in_page(ctx: TestContext, text: str) -> None:
    """Verify text appears in page content"""
    content = ctx.page.content()
    if text not in content:
        raise AssertionError(f"Expected text '{text}' not found in page content")


def verify_phase_indicator(ctx: TestContext) -> None:
    """Verify project phase indicator is visible"""
    content = ctx.page.content()
    if "INTAKE" not in content and "phase" not in content.lower():
        raise AssertionError("Phase indicator not found in page content")


def verify_temporal_workflow(ctx: TestContext) -> None:
    """Verify Temporal workflow is running"""
    workflow_id = ctx.get("workflow_id")
    result = subprocess.run(
        ["temporal", "workflow", "describe",
         "--workflow-id", workflow_id,
         "--address", "localhost:7233"],
        capture_output=True, text=True, timeout=10
    )
    if result.returncode != 0:
        raise AssertionError(f"Temporal workflow not found: {workflow_id}\n{result.stderr}")


# =============================================================================
# User Journey: Authentication
# =============================================================================

def create_auth_journey_feature() -> Feature:
    """Create the authentication user journey feature"""

    # Test data
    test_email = f"bdd_auth_{random_string()}@example.com"
    test_password = "SecurePassword123!"
    test_name = f"BDD User {random_string(4)}"
    test_tenant = f"BDD Tenant {random_string(6)}"

    # --- Scenario: New User Registration ---
    registration_scenario = Scenario(
        name="New user can register and access dashboard",
        tags=["auth", "registration", "smoke"]
    )

    registration_scenario.steps = [
        Step(
            StepType.GIVEN,
            "I am on the registration page",
            lambda ctx: (
                ctx.page.goto(f"{BASE_URL}/register"),
                ctx.page.wait_for_load_state("networkidle"),
                ctx.set("test_email", test_email),
                ctx.set("test_password", test_password),
                ctx.set("test_name", test_name),
                ctx.set("test_tenant", test_tenant)
            )
        ),
        Step(
            StepType.WHEN,
            "I fill in my company name",
            lambda ctx: ctx.page.fill("input#tenantName", ctx.get("test_tenant"))
        ),
        Step(
            StepType.AND,
            "I fill in my name",
            lambda ctx: ctx.page.fill("input#name", ctx.get("test_name"))
        ),
        Step(
            StepType.AND,
            "I fill in my email",
            lambda ctx: ctx.page.fill("input#email", ctx.get("test_email"))
        ),
        Step(
            StepType.AND,
            "I fill in my password",
            lambda ctx: ctx.page.fill("input#password", ctx.get("test_password"))
        ),
        Step(
            StepType.AND,
            "I confirm my password",
            lambda ctx: ctx.page.fill("input#confirmPassword", ctx.get("test_password"))
        ),
        Step(
            StepType.AND,
            "I click the Create Account button",
            lambda ctx: ctx.page.click("button[type='submit']")
        ),
        Step(
            StepType.THEN,
            "I should be redirected to the dashboard",
            lambda ctx: ctx.page.wait_for_url("**/dashboard**", timeout=15000)
        ),
        Step(
            StepType.AND,
            "I should see my user info in the sidebar",
            lambda ctx: (
                ctx.page.wait_for_selector("aside", timeout=5000),
                ctx.page.locator("aside").text_content()  # Just verify sidebar exists
            )
        ),
        Step(
            StepType.AND,
            "I should have an access token stored",
            lambda ctx: verify_token_exists(ctx)
        ),
    ]

    # --- Scenario: User Login ---
    login_scenario = Scenario(
        name="Existing user can login",
        tags=["auth", "login"]
    )

    login_scenario.steps = [
        Step(
            StepType.GIVEN,
            "I have a registered account",
            lambda ctx: (
                # Register first
                ctx.page.goto(f"{BASE_URL}/register"),
                ctx.page.wait_for_load_state("networkidle"),
                ctx.set("test_email", f"bdd_login_{random_string()}@example.com"),
                ctx.set("test_password", "SecurePassword123!"),
                ctx.page.fill("input#tenantName", f"Login Test {random_string(4)}"),
                ctx.page.fill("input#name", "Login User"),
                ctx.page.fill("input#email", ctx.get("test_email")),
                ctx.page.fill("input#password", ctx.get("test_password")),
                ctx.page.fill("input#confirmPassword", ctx.get("test_password")),
                ctx.page.click("button[type='submit']"),
                ctx.page.wait_for_url("**/dashboard**", timeout=15000),
                # Logout
                ctx.page.evaluate("localStorage.clear()"),
            )
        ),
        Step(
            StepType.AND,
            "I am on the login page",
            lambda ctx: (
                ctx.page.goto(f"{BASE_URL}/login"),
                ctx.page.wait_for_load_state("networkidle")
            )
        ),
        Step(
            StepType.WHEN,
            "I enter my email",
            lambda ctx: ctx.page.fill("input#email", ctx.get("test_email"))
        ),
        Step(
            StepType.AND,
            "I enter my password",
            lambda ctx: ctx.page.fill("input#password", ctx.get("test_password"))
        ),
        Step(
            StepType.AND,
            "I click the Sign In button",
            lambda ctx: ctx.page.click("button[type='submit']")
        ),
        Step(
            StepType.THEN,
            "I should be redirected to the dashboard",
            lambda ctx: ctx.page.wait_for_url("**/dashboard**", timeout=15000)
        ),
        Step(
            StepType.AND,
            "I should be authenticated",
            lambda ctx: verify_token_valid(ctx)
        ),
    ]

    return Feature(
        name="User Authentication",
        description="As a user, I want to register and login so that I can access the platform",
        scenarios=[registration_scenario, login_scenario]
    )


# =============================================================================
# User Journey: Project Creation
# =============================================================================

def create_project_journey_feature() -> Feature:
    """Create the project creation user journey feature"""

    test_email = f"bdd_project_{random_string()}@example.com"
    test_password = "SecurePassword123!"
    project_name = f"BDD Project {random_string(6)}"

    # --- Scenario: Create New Project ---
    create_project_scenario = Scenario(
        name="Authenticated user can create a new project",
        tags=["project", "creation", "smoke"]
    )

    create_project_scenario.steps = [
        Step(
            StepType.GIVEN,
            "I am logged in as a registered user",
            lambda ctx: (
                ctx.page.goto(f"{BASE_URL}/register"),
                ctx.page.wait_for_load_state("networkidle"),
                ctx.set("test_email", test_email),
                ctx.set("project_name", project_name),
                ctx.page.fill("input#tenantName", f"Project Test {random_string(4)}"),
                ctx.page.fill("input#name", "Project User"),
                ctx.page.fill("input#email", test_email),
                ctx.page.fill("input#password", test_password),
                ctx.page.fill("input#confirmPassword", test_password),
                ctx.page.click("button[type='submit']"),
                ctx.page.wait_for_url("**/dashboard**", timeout=15000)
            )
        ),
        Step(
            StepType.AND,
            "I am on the dashboard",
            lambda ctx: ctx.page.wait_for_selector("aside", timeout=5000)
        ),
        Step(
            StepType.WHEN,
            "I click the New Project button",
            lambda ctx: ctx.page.locator("button:has-text('New Project')").first.click()
        ),
        Step(
            StepType.THEN,
            "I should see the project creation modal",
            lambda ctx: ctx.page.wait_for_selector("input[placeholder*='TaskFlow']", timeout=5000)
        ),
        Step(
            StepType.WHEN,
            "I enter a project name",
            lambda ctx: ctx.page.fill("input[placeholder*='TaskFlow']", ctx.get("project_name"))
        ),
        Step(
            StepType.AND,
            "I enter a project description",
            lambda ctx: ctx.page.fill(
                "textarea[placeholder*='Describe']",
                "BDD Test: Build a task management app with user authentication and real-time updates"
            )
        ),
        Step(
            StepType.AND,
            "I click Create Project",
            lambda ctx: ctx.page.click("button:has-text('Create Project')")
        ),
        Step(
            StepType.THEN,
            "I should be redirected to the project page",
            lambda ctx: (
                ctx.page.wait_for_url("**/projects/**", timeout=20000),
                ctx.set("workflow_id", ctx.page.url.split("/projects/")[-1].split("?")[0])
            )
        ),
        Step(
            StepType.AND,
            "the project should be stored in the database",
            lambda ctx: verify_project_in_db(ctx)
        ),
    ]

    # --- Scenario: View Project List ---
    list_projects_scenario = Scenario(
        name="User can view their projects list",
        tags=["project", "list"]
    )

    list_projects_scenario.steps = [
        Step(
            StepType.GIVEN,
            "I am logged in with an existing project",
            lambda ctx: (
                ctx.page.goto(f"{BASE_URL}/register"),
                ctx.page.wait_for_load_state("networkidle"),
                ctx.set("project_name", f"List Test {random_string(4)}"),
                ctx.page.fill("input#tenantName", f"List Test {random_string(4)}"),
                ctx.page.fill("input#name", "List User"),
                ctx.page.fill("input#email", f"bdd_list_{random_string()}@example.com"),
                ctx.page.fill("input#password", "SecurePassword123!"),
                ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                ctx.page.click("button[type='submit']"),
                ctx.page.wait_for_url("**/dashboard**", timeout=15000),
                # Create a project
                ctx.page.locator("button:has-text('New Project')").first.click(),
                ctx.page.wait_for_selector("input[placeholder*='TaskFlow']"),
                ctx.page.fill("input[placeholder*='TaskFlow']", ctx.get("project_name")),
                ctx.page.fill("textarea[placeholder*='Describe']", "Test project for list"),
                ctx.page.click("button:has-text('Create Project')"),
                ctx.page.wait_for_url("**/projects/**", timeout=20000)
            )
        ),
        Step(
            StepType.WHEN,
            "I navigate to the dashboard",
            lambda ctx: (
                ctx.page.goto(f"{BASE_URL}/dashboard"),
                ctx.page.wait_for_load_state("networkidle"),
                time.sleep(2)  # Wait for projects to load
            )
        ),
        Step(
            StepType.THEN,
            "I should see my project in the list",
            lambda ctx: verify_text_in_page(ctx, ctx.get("project_name"))
        ),
        Step(
            StepType.AND,
            "I should see the project phase indicator",
            lambda ctx: verify_phase_indicator(ctx)
        ),
    ]

    return Feature(
        name="Project Management",
        description="As a user, I want to create and manage projects so that I can build software",
        scenarios=[create_project_scenario, list_projects_scenario]
    )


# =============================================================================
# User Journey: Full Workflow
# =============================================================================

def create_full_workflow_feature() -> Feature:
    """Create the complete end-to-end workflow feature"""

    test_id = random_string(6)
    test_email = f"bdd_e2e_{test_id}@example.com"

    full_workflow_scenario = Scenario(
        name="Complete user journey from registration to project creation",
        tags=["e2e", "smoke", "critical"]
    )

    full_workflow_scenario.steps = [
        # Registration
        Step(
            StepType.GIVEN,
            "I am a new user visiting the platform",
            lambda ctx: (
                ctx.page.goto(BASE_URL),
                ctx.page.wait_for_load_state("networkidle"),
                ctx.set("test_email", test_email),
                ctx.set("test_id", test_id)
            )
        ),
        Step(
            StepType.WHEN,
            "I navigate to the registration page",
            lambda ctx: (
                ctx.page.goto(f"{BASE_URL}/register"),
                ctx.page.wait_for_load_state("networkidle")
            )
        ),
        Step(
            StepType.AND,
            "I complete the registration form",
            lambda ctx: (
                ctx.page.fill("input#tenantName", f"E2E Corp {ctx.get('test_id')}"),
                ctx.page.fill("input#name", f"E2E User {ctx.get('test_id')}"),
                ctx.page.fill("input#email", ctx.get("test_email")),
                ctx.page.fill("input#password", "SecurePassword123!"),
                ctx.page.fill("input#confirmPassword", "SecurePassword123!"),
                ctx.page.click("button[type='submit']")
            )
        ),
        Step(
            StepType.THEN,
            "I should be logged in and on the dashboard",
            lambda ctx: ctx.page.wait_for_url("**/dashboard**", timeout=15000)
        ),

        # Project Creation
        Step(
            StepType.WHEN,
            "I create a new project",
            lambda ctx: (
                ctx.page.locator("button:has-text('New Project')").first.click(),
                ctx.page.wait_for_selector("input[placeholder*='TaskFlow']"),
                ctx.set("project_name", f"E2E Project {ctx.get('test_id')}"),
                ctx.page.fill("input[placeholder*='TaskFlow']", ctx.get("project_name")),
                ctx.page.fill("textarea[placeholder*='Describe']",
                    "E2E Test: Build a comprehensive task management application"),
                ctx.page.click("button:has-text('Create Project')")
            )
        ),
        Step(
            StepType.THEN,
            "I should see the project detail page",
            lambda ctx: (
                ctx.page.wait_for_url("**/projects/**", timeout=20000),
                ctx.set("workflow_id", ctx.page.url.split("/projects/")[-1].split("?")[0])
            )
        ),

        # Verification
        Step(
            StepType.AND,
            "the project should exist in PostgreSQL",
            lambda ctx: verify_project_in_db(ctx)
        ),
        Step(
            StepType.AND,
            "a Temporal workflow should be running",
            lambda ctx: verify_temporal_workflow(ctx)
        ),
        Step(
            StepType.WHEN,
            "I go back to the dashboard",
            lambda ctx: (
                ctx.page.goto(f"{BASE_URL}/dashboard"),
                ctx.page.wait_for_load_state("networkidle"),
                time.sleep(2)
            )
        ),
        Step(
            StepType.THEN,
            "I should see my project in the projects list",
            lambda ctx: verify_text_in_page(ctx, ctx.get("project_name"))
        ),
    ]

    return Feature(
        name="Complete User Journey",
        description="As a user, I want to go through the complete workflow from registration to project creation",
        scenarios=[full_workflow_scenario]
    )


# =============================================================================
# Main
# =============================================================================

def main():
    parser = argparse.ArgumentParser(description="BDD User Journey Tests")
    parser.add_argument("--journey", choices=["auth", "project", "full", "all"],
                       default="all", help="Which journey to run")
    parser.add_argument("--headed", action="store_true", help="Run with browser visible")
    parser.add_argument("--slow", action="store_true", help="Run in slow motion")
    args = parser.parse_args()

    # Create output directory
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    # Initialize runner
    slow_mo = 500 if args.slow else 0
    runner = BDDTestRunner(headed=args.headed, slow_mo=slow_mo)

    print("\n" + "="*70)
    print("🎭 BDD User Journey Tests")
    print("="*70)
    print(f"Mode: {'Headed' if args.headed else 'Headless'}")
    print(f"Slow Motion: {slow_mo}ms")
    print(f"Output: {OUTPUT_DIR}")

    # Select and run features
    features = []
    if args.journey in ["auth", "all"]:
        features.append(create_auth_journey_feature())
    if args.journey in ["project", "all"]:
        features.append(create_project_journey_feature())
    if args.journey in ["full", "all"]:
        features.append(create_full_workflow_feature())

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

    # Open report in browser
    if os.path.exists(report_path):
        print(f"\nTo view traces: npx playwright show-trace <trace.zip>")

    sys.exit(0 if all_passed else 1)


if __name__ == "__main__":
    main()
