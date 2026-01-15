#!/usr/bin/env python3
"""
Test the new Sandpack-based live preview in the frontend.
"""

from playwright.sync_api import sync_playwright
import time

def test_sandpack_preview():
    """Test that Sandpack preview loads correctly."""

    print("=" * 60)
    print("SANDPACK PREVIEW E2E TEST")
    print("=" * 60)

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        print("\n[TEST 1] Load Frontend...")
        page.goto('http://localhost:3000')
        page.wait_for_load_state('domcontentloaded')
        page.wait_for_timeout(3000)  # Wait for React + Sandpack to load

        # Take initial screenshot
        page.screenshot(path='/tmp/sandpack-initial.png', full_page=True)
        print("  ✓ Screenshot: /tmp/sandpack-initial.png")

        # Check for main UI elements
        print("\n[TEST 2] Check UI Structure...")

        # Left panel
        project_pod = page.locator('text=Project Pod')
        if project_pod.count() > 0:
            print("  ✓ Project Pod header found")

        # Check for new tabs (CHAT, SPEC, FILES)
        chat_tab = page.locator('button:text("CHAT")')
        spec_tab = page.locator('button:text("SPEC")')
        files_tab = page.locator('button:text("FILES")')

        if chat_tab.count() > 0:
            print("  ✓ CHAT tab found")
        if spec_tab.count() > 0:
            print("  ✓ SPEC tab found")
        if files_tab.count() > 0:
            print("  ✓ FILES tab found")

        # Right panel tabs
        preview_tab = page.locator('button:text("PREVIEW")')
        code_tab = page.locator('button:text("CODE")')
        terminal_tab = page.locator('button:text("TERMINAL")')

        if preview_tab.count() > 0:
            print("  ✓ PREVIEW tab found")
        if code_tab.count() > 0:
            print("  ✓ CODE tab found")
        if terminal_tab.count() > 0:
            print("  ✓ TERMINAL tab found")

        # Check for phase badge
        phase_badge = page.locator('text=DISCOVERY')
        if phase_badge.count() > 0:
            print("  ✓ Phase badge (DISCOVERY) found")

        print("\n[TEST 3] Check Sandpack Preview...")

        # Wait for Sandpack to load
        page.wait_for_timeout(3000)

        # Check for Sandpack iframe or preview
        sandpack_preview = page.locator('iframe[title*="Sandpack"]')
        sandpack_container = page.locator('.sp-preview')

        if sandpack_preview.count() > 0 or sandpack_container.count() > 0:
            print("  ✓ Sandpack preview container found")
        else:
            # Try alternative selectors
            preview_frame = page.locator('iframe')
            print(f"  ⚠ Found {preview_frame.count()} iframe(s)")

        # Take screenshot of preview state
        page.screenshot(path='/tmp/sandpack-preview.png', full_page=True)
        print("  ✓ Screenshot: /tmp/sandpack-preview.png")

        print("\n[TEST 4] Test Tab Navigation...")

        # Click CODE tab
        if code_tab.count() > 0:
            code_tab.click()
            page.wait_for_timeout(1000)
            print("  ✓ Clicked CODE tab")

            # Check for code editor
            code_editor = page.locator('.sp-code-editor, .cm-editor, [class*="editor"]')
            if code_editor.count() > 0:
                print("  ✓ Code editor visible")

            page.screenshot(path='/tmp/sandpack-code.png', full_page=True)
            print("  ✓ Screenshot: /tmp/sandpack-code.png")

        # Click TERMINAL tab
        if terminal_tab.count() > 0:
            terminal_tab.click()
            page.wait_for_timeout(500)
            print("  ✓ Clicked TERMINAL tab")

            page.screenshot(path='/tmp/sandpack-terminal.png', full_page=True)
            print("  ✓ Screenshot: /tmp/sandpack-terminal.png")

        # Click back to PREVIEW
        if preview_tab.count() > 0:
            preview_tab.click()
            page.wait_for_timeout(500)
            print("  ✓ Clicked back to PREVIEW tab")

        print("\n[TEST 5] Test Chat Input...")

        # Find and test input
        input_field = page.locator('input[placeholder*="Describe your app"]')
        if input_field.count() > 0:
            input_field.fill("Build me a todo app")
            value = input_field.input_value()
            if value == "Build me a todo app":
                print("  ✓ Input accepts text")
            input_field.clear()
        else:
            # Try alternative placeholder
            input_field = page.locator('input')
            if input_field.count() > 0:
                print(f"  ⚠ Found {input_field.count()} input field(s)")

        # Final screenshot
        page.screenshot(path='/tmp/sandpack-final.png', full_page=True)
        print("\n  ✓ Final screenshot: /tmp/sandpack-final.png")

        browser.close()

    print("\n" + "=" * 60)
    print("SANDPACK PREVIEW TEST COMPLETED")
    print("=" * 60)
    print("\nScreenshots saved to /tmp/sandpack-*.png")

if __name__ == "__main__":
    test_sandpack_preview()
