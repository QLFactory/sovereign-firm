#!/usr/bin/env python3
"""
Test the WebContainer-based live preview in the frontend workspace.
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
    email = f"preview_test_{test_id}@example.com"
    tenant = f"Preview Test {test_id}"

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

    print(f"  -> User registered")
    return email


def create_project(page):
    """Create a new project and navigate to workspace."""
    test_id = random_string()
    project_name = f"PreviewTest{test_id}"

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
            desc_input.fill("Test project for preview testing")

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


def test_sandpack_preview():
    """Test that preview tabs and workspace load correctly."""

    print("=" * 60)
    print("PREVIEW & WORKSPACE E2E TEST")
    print("=" * 60)

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Test 1: Setup
        print("\n[TEST 1] Setup...")
        register_user(page)
        create_project(page)
        time.sleep(2)

        # Take initial screenshot
        page.screenshot(path='/tmp/sandpack-initial.png', full_page=True)
        print("  ✓ Screenshot: /tmp/sandpack-initial.png")

        # Test 2: Check workspace UI structure
        print("\n[TEST 2] Check Workspace UI...")

        # Check for phase badge
        phase_found = False
        for phase in ['INTAKE', 'SIZING', 'PLANNING', 'DEVELOPMENT']:
            phase_badge = page.locator(f'text="{phase}"')
            if phase_badge.count() > 0:
                print(f"  ✓ Phase badge found: {phase}")
                phase_found = True
                break
        if not phase_found:
            print("  ⚠ No phase badge found")

        # Check for left panel tabs (CHAT, SPEC, FILES, TASKS, AGENTS)
        left_tabs = ['CHAT', 'SPEC', 'FILES', 'TASKS', 'AGENTS']
        found_left = 0
        for tab in left_tabs:
            tab_btn = page.locator(f'button:has-text("{tab}")')
            if tab_btn.count() > 0:
                print(f"  ✓ {tab} tab found")
                found_left += 1
        print(f"  -> Found {found_left}/{len(left_tabs)} left panel tabs")

        # Check right panel tabs (PREVIEW, CODE, TERMINAL, CI, EVENTS)
        right_tabs = ['PREVIEW', 'CODE', 'TERMINAL', 'CI', 'EVENTS']
        found_right = 0
        for tab in right_tabs:
            tab_btn = page.locator(f'button:has-text("{tab}")')
            if tab_btn.count() > 0:
                print(f"  ✓ {tab} tab found")
                found_right += 1
        print(f"  -> Found {found_right}/{len(right_tabs)} right panel tabs")

        # Test 3: Check Preview Panel
        print("\n[TEST 3] Check Preview Panel...")

        # Click PREVIEW tab (the one with the play icon, not the sub-tab)
        preview_tab = page.locator('button:has-text("▶ PREVIEW")').first
        if preview_tab.count() == 0:
            preview_tab = page.locator('button:has-text("PREVIEW")').first
        if preview_tab.count() > 0:
            preview_tab.click()
            time.sleep(2)
            print("  ✓ Clicked PREVIEW tab")

        # Check for WebContainer or iframe
        iframe = page.locator('iframe')
        webcontainer_text = page.locator('text="Booting WebContainer"')
        preview_content = page.locator('[class*="preview"], [class*="Preview"]')

        if iframe.count() > 0:
            print(f"  ✓ Found {iframe.count()} iframe(s) in preview")
        elif webcontainer_text.count() > 0:
            print("  ✓ WebContainer is booting")
        elif preview_content.count() > 0:
            print("  ✓ Preview container found")
        else:
            print("  ⚠ Preview content not detected yet")

        page.screenshot(path='/tmp/sandpack-preview.png', full_page=True)
        print("  ✓ Screenshot: /tmp/sandpack-preview.png")

        # Test 4: Test Tab Navigation
        print("\n[TEST 4] Test Tab Navigation...")

        # Click CODE tab (use icon to differentiate)
        code_tab = page.locator('button:has-text("CODE")').first
        if code_tab.count() > 0:
            code_tab.click()
            time.sleep(1)
            print("  ✓ Clicked CODE tab")

            # Check for code editor elements
            code_editor = page.locator('[class*="editor"], [class*="Editor"], [class*="monaco"]')
            if code_editor.count() > 0:
                print("  ✓ Code editor area visible")

            page.screenshot(path='/tmp/sandpack-code.png', full_page=True)
            print("  ✓ Screenshot: /tmp/sandpack-code.png")

        # Click TERMINAL tab
        terminal_tab = page.locator('button:has-text("TERMINAL")').first
        if terminal_tab.count() > 0:
            terminal_tab.click()
            time.sleep(1)
            print("  ✓ Clicked TERMINAL tab")

            page.screenshot(path='/tmp/sandpack-terminal.png', full_page=True)
            print("  ✓ Screenshot: /tmp/sandpack-terminal.png")

        # Click back to PREVIEW
        preview_tab = page.locator('button:has-text("▶ PREVIEW")').first
        if preview_tab.count() == 0:
            preview_tab = page.locator('button:has-text("PREVIEW")').first
        if preview_tab.count() > 0:
            preview_tab.click()
            time.sleep(1)
            print("  ✓ Clicked back to PREVIEW tab")

        # Test 5: Test Chat Input
        print("\n[TEST 5] Test Chat Input...")

        # Find chat input
        chat_input = page.locator('input[placeholder*="Describe"], input[placeholder*="app"]')
        if chat_input.count() == 0:
            chat_input = page.locator('textarea[placeholder*="Describe"], textarea[placeholder*="app"]')

        if chat_input.count() > 0:
            chat_input.first.fill("Build me a todo app")
            value = chat_input.first.input_value()
            if "todo" in value.lower():
                print("  ✓ Chat input accepts text")
            chat_input.first.clear()
        else:
            # List all inputs for debugging
            all_inputs = page.locator('input, textarea')
            print(f"  ⚠ Chat input not found, found {all_inputs.count()} input field(s)")

        # Check Send button
        send_btn = page.locator('button:has-text("Send")')
        if send_btn.count() > 0:
            print("  ✓ Send button found")

        # Test 6: Check connection status
        print("\n[TEST 6] Connection Status...")
        connected = page.locator('text="Connected"')
        if connected.count() > 0:
            print("  ✓ WebSocket connected")
        else:
            print("  ⚠ Connection status not visible")

        # Final screenshot
        page.screenshot(path='/tmp/sandpack-final.png', full_page=True)
        print("\n  ✓ Final screenshot: /tmp/sandpack-final.png")

        browser.close()

    print("\n" + "=" * 60)
    print("PREVIEW & WORKSPACE TEST COMPLETED")
    print("=" * 60)
    print("\nScreenshots saved to /tmp/sandpack-*.png")


if __name__ == "__main__":
    test_sandpack_preview()
