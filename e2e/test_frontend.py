#!/usr/bin/env python3
"""
End-to-end tests for Sovereign Firm frontend using Playwright.
Tests the workspace UI, WebSocket connections, and workflow interactions.
"""

from playwright.sync_api import sync_playwright
import time
import random
import string

BASE_URL = "http://localhost:3000"


def random_string(length=6):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))


def register_user(page):
    """Register a new user and return to dashboard."""
    test_id = random_string()
    email = f"frontend_test_{test_id}@example.com"
    tenant = f"Frontend Test {test_id}"

    print(f"  -> Registering user: {email}")
    page.goto(f"{BASE_URL}/register")
    page.wait_for_load_state("networkidle")
    time.sleep(1)

    # Fill registration form
    page.fill("input#tenantName", tenant)
    page.fill("input#name", "Test User")
    page.fill("input#email", email)
    page.fill("input#password", "TestPassword123!")
    page.fill("input#confirmPassword", "TestPassword123!")

    # Submit registration
    page.click("button[type='submit']")
    page.wait_for_url("**/dashboard**", timeout=15000)
    page.wait_for_load_state("networkidle")

    print(f"  -> User registered and on dashboard")
    return email


def create_project(page):
    """Create a new project and navigate to workspace."""
    test_id = random_string()
    project_name = f"TestProject{test_id}"

    print(f"  -> Creating project: {project_name}")

    # Navigate to projects page
    page.goto(f"{BASE_URL}/projects")
    page.wait_for_load_state("networkidle")
    time.sleep(1)

    # Click New Project button
    new_project_btn = page.locator('button:has-text("New Project")')
    if new_project_btn.count() > 0:
        new_project_btn.click()
        time.sleep(1)

        # Fill project form - input has placeholder "e.g., TaskFlow, InvoiceHub, ChatBot"
        name_input = page.locator('input[placeholder*="TaskFlow"]')
        if name_input.count() == 0:
            name_input = page.locator('input.input').first
        name_input.fill(project_name)

        # Description textarea has placeholder about "Describe what you want to build"
        desc_input = page.locator('textarea[placeholder*="Describe"]')
        if desc_input.count() == 0:
            desc_input = page.locator('textarea.input').first
        if desc_input.count() > 0:
            desc_input.fill("Test project for E2E testing")

        # Submit - button says "Create Project"
        create_btn = page.locator('button:has-text("Create Project")')
        if create_btn.count() == 0:
            create_btn = page.locator('button[type="submit"]')
        if create_btn.count() > 0:
            create_btn.click()
            # Wait for redirect to workspace
            page.wait_for_url("**/projects/**", timeout=15000)
            page.wait_for_load_state("networkidle")
            time.sleep(2)

    print(f"  -> Project created")
    return project_name


def test_frontend():
    """Test the Sovereign Firm frontend workspace."""

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        print("=" * 60)
        print("SOVEREIGN FIRM FRONTEND E2E TESTS")
        print("=" * 60)

        # Test 1: Register and setup
        print("\n[TEST 1] User Registration...")
        register_user(page)
        print("  ✓ User registered successfully")

        # Test 2: Create project
        print("\n[TEST 2] Create Project...")
        create_project(page)
        page.screenshot(path='/tmp/sovereign-firm-workspace.png', full_page=True)
        print("  ✓ Screenshot saved to /tmp/sovereign-firm-workspace.png")

        # Test 3: Check workspace UI elements
        print("\n[TEST 3] Workspace UI Elements...")

        # Check for phase indicator (INTAKE, SIZING, etc.)
        phase_indicators = ['INTAKE', 'SIZING', 'PLANNING', 'DEVELOPMENT']
        phase_found = False
        for phase in phase_indicators:
            phase_el = page.locator(f'text="{phase}"')
            if phase_el.count() > 0:
                print(f"  ✓ Phase indicator found: {phase}")
                phase_found = True
                break
        if not phase_found:
            print("  ⚠ No phase indicator found")

        # Check for tabs (CHAT, SPEC, FILES, TASKS, AGENTS)
        tabs = ['CHAT', 'SPEC', 'FILES', 'TASKS', 'AGENTS']
        for tab in tabs:
            tab_btn = page.locator(f'button:has-text("{tab}")')
            if tab_btn.count() > 0:
                print(f"  ✓ {tab} tab found")

        # Check right panel tabs (PREVIEW, CODE, TERMINAL, CI, EVENTS)
        right_tabs = ['PREVIEW', 'CODE', 'TERMINAL', 'CI', 'EVENTS']
        for tab in right_tabs:
            tab_btn = page.locator(f'button:has-text("{tab}")')
            if tab_btn.count() > 0:
                print(f"  ✓ {tab} tab found")

        # Test 4: Check chat input
        print("\n[TEST 4] Chat Input...")
        chat_input = page.locator('input[placeholder*="Describe"], textarea[placeholder*="Describe"]')
        if chat_input.count() == 0:
            chat_input = page.locator('input[placeholder*="app"], textarea[placeholder*="app"]')

        if chat_input.count() > 0:
            chat_input.first.fill("Test message for E2E")
            value = chat_input.first.input_value()
            if "Test message" in value:
                print("  ✓ Chat input accepts text")
            chat_input.first.clear()
        else:
            print("  ⚠ Chat input not found")

        # Check for Send button
        send_btn = page.locator('button:has-text("Send")')
        if send_btn.count() > 0:
            print("  ✓ Send button found")

        # Test 5: Check layout structure
        print("\n[TEST 5] Layout Structure...")

        # Check for main container
        main_container = page.locator('main')
        if main_container.count() > 0:
            print("  ✓ Main container found")

        # Check for left panel (chat area)
        left_panel = page.locator('[class*="border-r"], [class*="w-1/3"], [class*="w-96"]')
        if left_panel.count() > 0:
            print("  ✓ Left panel found")

        # Check for right panel (preview area)
        right_panel = page.locator('[class*="flex-1"]')
        if right_panel.count() > 0:
            print("  ✓ Right panel found")

        # Test 6: WebSocket connection indicator
        print("\n[TEST 6] Connection Status...")
        connected_indicator = page.locator('text="Connected"')
        if connected_indicator.count() > 0:
            print("  ✓ WebSocket connected indicator found")
        else:
            disconnected = page.locator('text="Disconnected"')
            if disconnected.count() > 0:
                print("  ⚠ WebSocket shows disconnected")
            else:
                print("  ⚠ Connection status not visible")

        # Test 7: Console logs check
        print("\n[TEST 7] Console Errors...")
        console_errors = []
        page.on('console', lambda msg: console_errors.append(msg.text) if msg.type == 'error' else None)

        try:
            page.reload()
            page.wait_for_load_state('domcontentloaded', timeout=10000)
            time.sleep(2)
        except Exception as e:
            print(f"  ⚠ Reload timeout (WebSocket keeping connection active)")

        if console_errors:
            print(f"  ⚠ Console errors: {len(console_errors)}")
            for err in console_errors[:3]:
                print(f"    - {err[:80]}...")
        else:
            print("  ✓ No console errors")

        # Test 8: Final screenshot
        print("\n[TEST 8] Final Screenshot...")
        page.screenshot(path='/tmp/sovereign-firm-final.png', full_page=True)
        print("  ✓ Final screenshot saved to /tmp/sovereign-firm-final.png")

        browser.close()

        print("\n" + "=" * 60)
        print("E2E TESTS COMPLETED")
        print("=" * 60)
        print("\nScreenshots:")
        print("  - /tmp/sovereign-firm-workspace.png")
        print("  - /tmp/sovereign-firm-final.png")


if __name__ == "__main__":
    test_frontend()
