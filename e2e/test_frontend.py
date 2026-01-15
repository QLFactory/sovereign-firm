#!/usr/bin/env python3
"""
End-to-end tests for Sovereign Firm frontend using Playwright.
Tests the console UI, WebSocket connections, and workflow interactions.
"""

from playwright.sync_api import sync_playwright
import time
import json

def test_frontend():
    """Test the Sovereign Firm frontend console."""

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        print("=" * 60)
        print("SOVEREIGN FIRM FRONTEND E2E TESTS")
        print("=" * 60)

        # Test 1: Page loads correctly
        print("\n[TEST 1] Page Load...")
        page.goto('http://localhost:3000')
        page.wait_for_load_state('domcontentloaded')
        page.wait_for_timeout(2000)  # Wait for React to hydrate

        # Take screenshot of initial state
        page.screenshot(path='/tmp/sovereign-firm-initial.png', full_page=True)
        print("  ✓ Screenshot saved to /tmp/sovereign-firm-initial.png")

        # Verify basic page structure
        title = page.title()
        print(f"  ✓ Page title: {title}")

        # Test 2: Check main UI elements exist
        print("\n[TEST 2] UI Elements...")

        # Check for Project Pod header
        project_pod = page.locator('text=Project Pod')
        if project_pod.count() > 0:
            print("  ✓ Project Pod header found")
        else:
            print("  ✗ Project Pod header NOT found")

        # Check for Chat/Spec tabs
        chat_btn = page.locator('button:text("Chat")')
        spec_btn = page.locator('button:text("Spec")')
        if chat_btn.count() > 0:
            print("  ✓ Chat tab found")
        if spec_btn.count() > 0:
            print("  ✓ Spec tab found")

        # Check for Preview/Terminal tabs
        preview_btn = page.locator('button:text("Preview")')
        terminal_btn = page.locator('button:text("Terminal")')
        if preview_btn.count() > 0:
            print("  ✓ Preview tab found")
        if terminal_btn.count() > 0:
            print("  ✓ Terminal tab found")

        # Check for message input
        message_input = page.locator('input[placeholder*="Type message"]')
        if message_input.count() > 0:
            print("  ✓ Message input found")
        else:
            print("  ✗ Message input NOT found")

        # Check for Send button
        send_btn = page.locator('button:text("Send")')
        if send_btn.count() > 0:
            print("  ✓ Send button found")
        else:
            print("  ✗ Send button NOT found")

        # Test 3: Check status indicator
        print("\n[TEST 3] Status Indicator...")
        status_indicator = page.locator('text=Initializing')
        if status_indicator.count() > 0:
            print("  ✓ Status indicator shows 'Initializing...'")

        # Test 4: Test input interaction
        print("\n[TEST 4] Input Interaction...")
        if message_input.count() > 0:
            message_input.fill("Hello, test message!")
            value = message_input.input_value()
            if value == "Hello, test message!":
                print("  ✓ Can type in message input")
            else:
                print(f"  ✗ Input value mismatch: {value}")

            # Clear for next test
            message_input.clear()

        # Test 5: Check responsive layout
        print("\n[TEST 5] Layout Structure...")

        # Check for two-column layout
        main_container = page.locator('main')
        if main_container.count() > 0:
            print("  ✓ Main container found")

        # Check left panel (chat)
        left_panel = page.locator('.w-1\\/3')
        if left_panel.count() > 0:
            print("  ✓ Left panel (1/3 width) found")

        # Check right panel (preview)
        right_panel = page.locator('.flex-1')
        if right_panel.count() > 0:
            print("  ✓ Right panel (flex-1) found")

        # Test 6: Console logs check
        print("\n[TEST 6] Console Logs...")
        console_messages = []
        page.on('console', lambda msg: console_messages.append(msg.text))

        # Trigger a reload to capture any console output
        page.reload()
        page.wait_for_load_state('domcontentloaded')
        page.wait_for_timeout(2000)

        if console_messages:
            print(f"  ⚠ Console messages captured: {len(console_messages)}")
            for msg in console_messages[:5]:  # Show first 5
                print(f"    - {msg[:80]}...")
        else:
            print("  ✓ No console errors")

        # Test 7: Take final screenshot
        print("\n[TEST 7] Final Screenshot...")
        page.screenshot(path='/tmp/sovereign-firm-final.png', full_page=True)
        print("  ✓ Final screenshot saved to /tmp/sovereign-firm-final.png")

        # Print page content for debugging
        print("\n[DEBUG] Page Structure:")
        html = page.content()
        # Check for key elements in HTML
        checks = [
            ('WebSocket', 'ws://' in html or 'wss://' in html or 'WebSocket' in html),
            ('React root', 'id="root"' in html),
            ('Tailwind classes', 'bg-zinc' in html or 'text-white' in html),
        ]
        for name, found in checks:
            status = "✓" if found else "✗"
            print(f"  {status} {name}")

        browser.close()

        print("\n" + "=" * 60)
        print("E2E TESTS COMPLETED")
        print("=" * 60)
        print("\nScreenshots:")
        print("  - /tmp/sovereign-firm-initial.png")
        print("  - /tmp/sovereign-firm-final.png")

if __name__ == "__main__":
    test_frontend()
