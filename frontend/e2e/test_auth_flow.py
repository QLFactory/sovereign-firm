#!/usr/bin/env python3
"""
E2E Authentication Flow Tests for Sovereign Firm

Tests the complete authentication flow:
- Landing page navigation
- User registration
- Dashboard access (protected route)
- Logout functionality
- Login with existing account
- Protected route redirect

Usage:
    # Run headless (CI mode)
    python e2e/test_auth_flow.py

    # Run with browser visible (headed mode)
    python e2e/test_auth_flow.py --headed

    # Run with slow motion for debugging
    python e2e/test_auth_flow.py --headed --slow
"""

import argparse
import random
import string
import sys
import time
from playwright.sync_api import sync_playwright

# Test configuration
BASE_URL = "http://localhost:3000"
SCREENSHOT_DIR = "e2e/screenshots"


def generate_random_email():
    """Generate a random email for testing."""
    random_str = ''.join(random.choices(string.ascii_lowercase + string.digits, k=8))
    return f"test_{random_str}@example.com"


def generate_random_tenant():
    """Generate a random tenant name for testing."""
    random_suffix = ''.join(random.choices(string.ascii_lowercase + string.digits, k=6))
    return f"Test Company {random_suffix}"


class AuthFlowTest:
    def __init__(self, headed=False, slow_mo=0):
        self.headed = headed
        self.slow_mo = slow_mo
        self.test_email = generate_random_email()
        self.test_password = "TestPassword123!"
        self.test_tenant = generate_random_tenant()
        self.passed = 0
        self.failed = 0

    def log(self, message, status="info"):
        icons = {"pass": "✓", "fail": "✗", "info": "→"}
        icon = icons.get(status, "→")
        print(f"   {icon} {message}")

    def run(self):
        print(f"\n{'='*60}")
        print("Sovereign Firm - E2E Auth Flow Tests")
        print(f"{'='*60}")
        print(f"Test Email:  {self.test_email}")
        print(f"Test Tenant: {self.test_tenant}")
        print(f"Mode:        {'Headed' if self.headed else 'Headless'}")
        print()

        with sync_playwright() as p:
            browser = p.chromium.launch(
                headless=not self.headed,
                slow_mo=self.slow_mo
            )
            context = browser.new_context(viewport={"width": 1280, "height": 800})
            page = context.new_page()

            # Capture console errors
            errors = []
            page.on("console", lambda msg: errors.append(msg.text) if msg.type == "error" else None)

            try:
                self.test_landing_page(page)
                self.test_register_page(page)
                self.test_dashboard_access(page)
                self.test_logout(page)
                self.test_login(page)
                self.test_protected_route_redirect(page)

                # Print results
                print(f"\n{'='*60}")
                if self.failed == 0:
                    print(f"✓ ALL TESTS PASSED! ({self.passed}/{self.passed + self.failed})")
                else:
                    print(f"✗ TESTS FAILED: {self.failed} failed, {self.passed} passed")
                print(f"{'='*60}\n")

                # Save final screenshot
                page.screenshot(path=f"{SCREENSHOT_DIR}/final_state.png")
                print(f"Screenshots saved to {SCREENSHOT_DIR}/")

                return self.failed == 0

            except Exception as e:
                print(f"\n✗ TEST ERROR: {str(e)}")
                page.screenshot(path=f"{SCREENSHOT_DIR}/error.png")
                if errors:
                    print("\nBrowser console errors:")
                    for err in errors:
                        print(f"  - {err}")
                return False
            finally:
                browser.close()

    def test_landing_page(self, page):
        """Test 1: Landing page has auth navigation links"""
        print("\n1. Testing Landing Page...")

        page.goto(BASE_URL)
        page.wait_for_load_state("networkidle")

        # Check for Sign In and Get Started in navigation
        nav = page.locator("nav")
        sign_in = nav.locator("a:has-text('Sign In')")
        get_started = nav.locator("a:has-text('Get Started')")

        if sign_in.is_visible() and get_started.is_visible():
            self.log("Landing page loaded with auth links", "pass")
            self.passed += 1
        else:
            self.log("Auth links not visible in navigation", "fail")
            self.failed += 1

        page.screenshot(path=f"{SCREENSHOT_DIR}/01_landing_page.png")

    def test_register_page(self, page):
        """Test 2: User registration flow"""
        print("\n2. Testing Registration...")

        page.goto(f"{BASE_URL}/register")
        page.wait_for_load_state("networkidle")
        page.wait_for_selector("input#tenantName", timeout=5000)

        # Fill registration form
        page.fill("input#tenantName", self.test_tenant)
        page.fill("input#email", self.test_email)
        page.fill("input#password", self.test_password)
        page.fill("input#confirmPassword", self.test_password)

        self.log("Registration form filled", "pass")
        page.screenshot(path=f"{SCREENSHOT_DIR}/02_register_form.png")

        # Submit form
        page.click("button[type='submit']")

        # Wait for redirect to dashboard
        try:
            page.wait_for_url("**/dashboard**", timeout=10000)
            self.log("Registration successful - redirected to dashboard", "pass")
            self.passed += 1
            page.screenshot(path=f"{SCREENSHOT_DIR}/03_post_register.png")
        except Exception:
            error_msg = page.locator("[class*='rose-glow']").first
            if error_msg.is_visible():
                self.log(f"Registration failed: {error_msg.text_content()}", "fail")
            else:
                self.log("Registration failed: unknown error", "fail")
            self.failed += 1
            page.screenshot(path=f"{SCREENSHOT_DIR}/03_register_error.png")
            raise

    def test_dashboard_access(self, page):
        """Test 3: Dashboard shows user info"""
        print("\n3. Testing Dashboard Access...")

        # Verify sidebar is present
        page.wait_for_selector("aside", timeout=5000)

        # Check for user info in sidebar
        sidebar = page.locator("aside")
        sidebar_text = sidebar.text_content()

        if self.test_email[0].upper() in sidebar_text or self.test_email in sidebar_text:
            self.log("Dashboard shows user info", "pass")
            self.passed += 1
        else:
            self.log("User info not visible in sidebar", "fail")
            self.failed += 1

        # Check logout button
        logout_btn = page.locator("button[title='Sign out']")
        if logout_btn.is_visible():
            self.log("Logout button visible", "pass")
            self.passed += 1
        else:
            self.log("Logout button not visible", "fail")
            self.failed += 1

        page.screenshot(path=f"{SCREENSHOT_DIR}/04_dashboard.png")

    def test_logout(self, page):
        """Test 4: Logout functionality"""
        print("\n4. Testing Logout...")

        logout_btn = page.locator("button[title='Sign out']")
        logout_btn.click()

        try:
            page.wait_for_url("**/login**", timeout=5000)
            self.log("Logout successful - redirected to login", "pass")
            self.passed += 1
            page.screenshot(path=f"{SCREENSHOT_DIR}/05_post_logout.png")
        except Exception:
            self.log("Logout failed - not redirected to login", "fail")
            self.failed += 1
            raise

    def test_login(self, page):
        """Test 5: Login with existing credentials"""
        print("\n5. Testing Login...")

        page.wait_for_selector("input#email", timeout=5000)

        # Fill login form
        page.fill("input#email", self.test_email)
        page.fill("input#password", self.test_password)

        page.screenshot(path=f"{SCREENSHOT_DIR}/06_login_form.png")

        # Submit
        page.click("button[type='submit']")

        try:
            page.wait_for_url("**/dashboard**", timeout=10000)
            self.log("Login successful - redirected to dashboard", "pass")
            self.passed += 1
            page.screenshot(path=f"{SCREENSHOT_DIR}/07_post_login.png")
        except Exception:
            error_msg = page.locator("[class*='rose-glow']").first
            if error_msg.is_visible():
                self.log(f"Login failed: {error_msg.text_content()}", "fail")
            else:
                self.log("Login failed: unknown error", "fail")
            self.failed += 1
            raise

    def test_protected_route_redirect(self, page):
        """Test 6: Protected routes redirect unauthenticated users"""
        print("\n6. Testing Protected Route Redirect...")

        # Logout first
        logout_btn = page.locator("button[title='Sign out']")
        if logout_btn.is_visible():
            logout_btn.click()
            page.wait_for_url("**/login**", timeout=5000)

        # Clear localStorage
        page.evaluate("localStorage.clear()")

        # Try to access dashboard directly
        page.goto(f"{BASE_URL}/dashboard")
        page.wait_for_load_state("networkidle")
        time.sleep(1)  # Wait for auth check

        if "/login" in page.url:
            self.log("Unauthenticated users redirected to login", "pass")
            self.passed += 1
        else:
            self.log(f"Expected redirect to login, got: {page.url}", "fail")
            self.failed += 1

        page.screenshot(path=f"{SCREENSHOT_DIR}/08_protected_redirect.png")


def main():
    parser = argparse.ArgumentParser(description="Run E2E auth flow tests")
    parser.add_argument("--headed", action="store_true", help="Run with browser visible")
    parser.add_argument("--slow", action="store_true", help="Run with slow motion (500ms)")
    args = parser.parse_args()

    # Create screenshots directory
    import os
    os.makedirs(SCREENSHOT_DIR, exist_ok=True)

    # Run tests
    slow_mo = 500 if args.slow else 0
    test = AuthFlowTest(headed=args.headed, slow_mo=slow_mo)
    success = test.run()

    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
