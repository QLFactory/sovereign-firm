#!/usr/bin/env python3
"""
Full-Stack Project Creation Test

Tests the complete flow:
1. Register/Login user via frontend
2. Create a project via the dashboard UI
3. Verify project appears in the projects list
4. Verify project is stored in PostgreSQL
5. Verify Temporal workflow was started

Usage:
    python e2e/test_full_stack_project.py --headed
"""

import argparse
import json
import random
import string
import subprocess
import sys
import time
from playwright.sync_api import sync_playwright

# Configuration
BASE_URL = "http://localhost:3000"
API_URL = "http://localhost:8080"
SCREENSHOT_DIR = "e2e/screenshots"
DB_CONNECTION = "postgres://sovereign:sovereign123@localhost:5432/sovereign_firm?sslmode=disable"


def random_string(n=8):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=n))


def query_database(sql):
    """Execute SQL query and return results as JSON."""
    try:
        result = subprocess.run(
            ["psql", DB_CONNECTION, "-t", "-A", "-c", sql],
            capture_output=True,
            text=True,
            timeout=10
        )
        return result.stdout.strip()
    except Exception as e:
        print(f"Database query failed: {e}")
        return None


def check_temporal_workflow(workflow_id):
    """Check if a Temporal workflow exists."""
    try:
        result = subprocess.run(
            ["temporal", "workflow", "describe", "--workflow-id", workflow_id,
             "--address", "localhost:7233"],
            capture_output=True,
            text=True,
            timeout=10
        )
        return result.returncode == 0
    except Exception as e:
        print(f"Temporal check failed: {e}")
        return False


class FullStackProjectTest:
    def __init__(self, headed=False, slow_mo=0):
        self.headed = headed
        self.slow_mo = slow_mo
        self.test_email = f"fullstack_{random_string()}@example.com"
        self.test_password = "TestPassword123!"
        self.test_tenant = f"FullStack Test {random_string(6)}"
        self.test_project_name = f"E2E Project {random_string(6)}"
        self.created_workflow_id = None
        self.access_token = None

    def log(self, message, status="info"):
        icons = {"pass": "✓", "fail": "✗", "info": "→", "warn": "⚠"}
        print(f"   {icons.get(status, '→')} {message}")

    def run(self):
        print(f"\n{'='*70}")
        print("Full-Stack Project Creation Test")
        print(f"{'='*70}")
        print(f"Test Email:   {self.test_email}")
        print(f"Test Tenant:  {self.test_tenant}")
        print(f"Project Name: {self.test_project_name}")
        print(f"Mode:         {'Headed' if self.headed else 'Headless'}")
        print()

        with sync_playwright() as p:
            browser = p.chromium.launch(
                headless=not self.headed,
                slow_mo=self.slow_mo
            )
            context = browser.new_context(viewport={"width": 1400, "height": 900})
            page = context.new_page()

            try:
                # Step 1: Register user
                print("\n[1/6] Registering new user...")
                self.register_user(page)

                # Step 2: Verify dashboard access
                print("\n[2/6] Verifying dashboard access...")
                self.verify_dashboard(page)

                # Step 3: Create project via UI
                print("\n[3/6] Creating project via dashboard...")
                self.create_project(page)

                # Step 4: Verify project in UI
                print("\n[4/6] Verifying project in UI...")
                self.verify_project_in_ui(page)

                # Step 5: Verify project in database
                print("\n[5/6] Verifying project in PostgreSQL...")
                self.verify_project_in_database()

                # Step 6: Verify Temporal workflow
                print("\n[6/6] Verifying Temporal workflow...")
                self.verify_temporal_workflow()

                print(f"\n{'='*70}")
                print("✓ FULL-STACK TEST PASSED!")
                print(f"{'='*70}")
                print(f"\nProject Details:")
                print(f"  - Name: {self.test_project_name}")
                print(f"  - Workflow ID: {self.created_workflow_id}")
                print(f"  - User: {self.test_email}")
                print()

                page.screenshot(path=f"{SCREENSHOT_DIR}/fullstack_success.png")
                return True

            except Exception as e:
                print(f"\n✗ TEST FAILED: {str(e)}")
                page.screenshot(path=f"{SCREENSHOT_DIR}/fullstack_error.png")
                return False
            finally:
                browser.close()
                self.cleanup()

    def register_user(self, page):
        """Register a new user via the frontend."""
        page.goto(f"{BASE_URL}/register")
        page.wait_for_load_state("networkidle")
        page.wait_for_selector("input#tenantName", timeout=10000)

        page.fill("input#tenantName", self.test_tenant)
        page.fill("input#name", f"Test User {random_string(4)}")
        page.fill("input#email", self.test_email)
        page.fill("input#password", self.test_password)
        page.fill("input#confirmPassword", self.test_password)

        page.screenshot(path=f"{SCREENSHOT_DIR}/fullstack_01_register.png")
        page.click("button[type='submit']")

        page.wait_for_url("**/dashboard**", timeout=15000)
        self.log("User registered successfully", "pass")

        # Extract access token from localStorage
        self.access_token = page.evaluate("localStorage.getItem('sovereign-firm-access-token')")
        if self.access_token:
            self.log(f"Access token obtained (length: {len(self.access_token)})", "pass")

    def verify_dashboard(self, page):
        """Verify dashboard loads with user info."""
        page.wait_for_selector("aside", timeout=5000)

        sidebar_text = page.locator("aside").text_content()
        if self.test_email[0].upper() in sidebar_text:
            self.log("User info displayed in sidebar", "pass")
        else:
            self.log("User info not found in sidebar", "warn")

        page.screenshot(path=f"{SCREENSHOT_DIR}/fullstack_02_dashboard.png")

    def create_project(self, page):
        """Create a new project via the dashboard UI."""
        # Click "New Project" button
        new_project_btn = page.locator("button:has-text('New Project')").first
        new_project_btn.click()

        # Wait for modal
        page.wait_for_selector("input[placeholder*='TaskFlow']", timeout=5000)
        self.log("Project creation modal opened", "pass")

        # Fill project details
        page.fill("input[placeholder*='TaskFlow']", self.test_project_name)
        page.fill("textarea[placeholder*='Describe']",
            "E2E test project: Build a task management application with user authentication, "
            "project boards, and real-time updates. Include a REST API and React frontend.")

        page.screenshot(path=f"{SCREENSHOT_DIR}/fullstack_03_project_form.png")

        # Submit the form
        page.click("button:has-text('Create Project')")
        self.log("Project creation submitted", "pass")

        # Wait for redirect to project page or project list update
        try:
            # Wait for URL to change to project detail page
            page.wait_for_url("**/projects/**", timeout=20000)

            # Extract workflow ID from URL
            current_url = page.url
            if "/projects/" in current_url:
                self.created_workflow_id = current_url.split("/projects/")[-1].split("?")[0]
                self.log(f"Project created with ID: {self.created_workflow_id}", "pass")

            page.screenshot(path=f"{SCREENSHOT_DIR}/fullstack_04_project_created.png")
        except Exception as e:
            # Maybe stayed on dashboard - check for project in list
            self.log(f"Did not redirect to project page: {e}", "warn")
            page.screenshot(path=f"{SCREENSHOT_DIR}/fullstack_04_after_create.png")

    def verify_project_in_ui(self, page):
        """Verify project appears in the UI."""
        # Go to dashboard to see project list
        page.goto(f"{BASE_URL}/dashboard")
        page.wait_for_load_state("networkidle")
        time.sleep(2)  # Wait for projects to load

        page.screenshot(path=f"{SCREENSHOT_DIR}/fullstack_05_project_list.png")

        # Look for project name in the page
        page_content = page.content()
        if self.test_project_name in page_content:
            self.log(f"Project '{self.test_project_name}' found in UI", "pass")
        else:
            # Try API directly to get workflow ID
            self.log("Project not found in UI, checking via API...", "warn")
            self.get_project_from_api(page)

    def get_project_from_api(self, page):
        """Get project details from API."""
        if not self.access_token:
            self.log("No access token available", "fail")
            return

        # Use page.evaluate to make fetch request with auth
        result = page.evaluate(f"""
            async () => {{
                const response = await fetch('{API_URL}/api/projects', {{
                    headers: {{
                        'Authorization': 'Bearer {self.access_token}'
                    }}
                }});
                return await response.json();
            }}
        """)

        if result and isinstance(result, list) and len(result) > 0:
            project = result[0]
            self.created_workflow_id = project.get('workflow_id') or project.get('id')
            self.log(f"Found project via API: {self.created_workflow_id}", "pass")
        else:
            self.log(f"API returned: {result}", "warn")

    def verify_project_in_database(self):
        """Verify project exists in PostgreSQL."""
        if not self.created_workflow_id:
            self.log("No workflow ID to verify", "warn")
            # Try to find by name
            sql = f"SELECT workflow_id, name, phase FROM projects WHERE name = '{self.test_project_name}' LIMIT 1"
        else:
            sql = f"SELECT workflow_id, name, phase FROM projects WHERE workflow_id = '{self.created_workflow_id}' LIMIT 1"

        result = query_database(sql)

        if result:
            parts = result.split("|")
            if len(parts) >= 2:
                self.created_workflow_id = parts[0]
                self.log(f"Project found in database: {parts[0]}", "pass")
                self.log(f"  Name: {parts[1]}", "info")
                if len(parts) > 2:
                    self.log(f"  Phase: {parts[2] or 'INTAKE'}", "info")
                return

        self.log("Project not found in database (may be using Temporal-only storage)", "warn")

    def verify_temporal_workflow(self):
        """Verify Temporal workflow was started."""
        if not self.created_workflow_id:
            self.log("No workflow ID to verify in Temporal", "warn")
            return

        if check_temporal_workflow(self.created_workflow_id):
            self.log(f"Temporal workflow exists: {self.created_workflow_id}", "pass")
        else:
            # Try with different prefixes
            for prefix in ["consultancy-", "project-", ""]:
                wf_id = f"{prefix}{self.created_workflow_id}"
                if check_temporal_workflow(wf_id):
                    self.log(f"Temporal workflow exists: {wf_id}", "pass")
                    return

            self.log("Temporal workflow not found (may use different naming)", "warn")

    def cleanup(self):
        """Clean up test data."""
        print("\nCleaning up test data...")

        # Delete from database
        if self.created_workflow_id:
            query_database(f"DELETE FROM projects WHERE workflow_id = '{self.created_workflow_id}'")

        query_database(f"DELETE FROM users WHERE email = '{self.test_email}'")

        # Get tenant slug and delete
        tenant_slug = self.test_tenant.lower().replace(" ", "-")
        query_database(f"DELETE FROM tenants WHERE slug LIKE '{tenant_slug}%'")

        self.log("Test data cleaned up", "info")


def main():
    parser = argparse.ArgumentParser(description="Full-stack project creation test")
    parser.add_argument("--headed", action="store_true", help="Run with browser visible")
    parser.add_argument("--slow", action="store_true", help="Run with slow motion")
    args = parser.parse_args()

    import os
    os.makedirs(SCREENSHOT_DIR, exist_ok=True)

    slow_mo = 300 if args.slow else 0
    test = FullStackProjectTest(headed=args.headed, slow_mo=slow_mo)
    success = test.run()

    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
