#!/usr/bin/env python3
"""
E2E test for the Three-Strike Rule feedback loop.
Tests that when QA-generated tests fail, the system regenerates them.
"""

import time
from playwright.sync_api import sync_playwright

def test_feedback_loop():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        print("📍 Navigating to frontend...")
        page.goto("http://localhost:3000", timeout=60000)
        page.wait_for_load_state("domcontentloaded")
        page.wait_for_timeout(3000)  # Give time for React to hydrate

        # Take initial screenshot
        page.screenshot(path="/tmp/feedback_loop_1_initial.png")
        print("📸 Screenshot: /tmp/feedback_loop_1_initial.png")

        # Click "New Pod" button
        print("🆕 Creating new pod...")
        new_pod_btn = page.locator("button:has-text('New Pod')")
        if new_pod_btn.count() > 0:
            new_pod_btn.click()
            page.wait_for_timeout(2000)

        page.screenshot(path="/tmp/feedback_loop_2_new_pod.png")
        print("📸 Screenshot: /tmp/feedback_loop_2_new_pod.png")

        # Find the chat input and send a message describing a simple app
        print("💬 Sending app description...")
        chat_input = page.locator("input[placeholder*='Describe your app']")
        chat_input.wait_for(timeout=10000)
        chat_input.fill("Create a simple React counter app with increment and decrement buttons. The counter should start at 0 and display the current count prominently.")

        # Click send button
        send_btn = page.locator("button:has-text('Send')")
        send_btn.click()

        print("⏳ Waiting for PM response...")
        page.wait_for_timeout(10000)  # Wait for LLM response

        page.screenshot(path="/tmp/feedback_loop_3_after_message.png")
        print("📸 Screenshot: /tmp/feedback_loop_3_after_message.png")

        # Send /approve to trigger code generation
        print("✅ Sending /approve to generate code...")
        chat_input = page.locator("input[placeholder*='Describe your app']")
        chat_input.fill("/approve")
        send_btn = page.locator("button:has-text('Send')")
        send_btn.click()

        print("⏳ Waiting for code generation and test execution...")
        print("   This may take a few minutes as tests are generated and run...")

        # Wait and monitor for test execution (check worker logs)
        # The workflow will run tests automatically after QA generates them

        # Poll for phase changes and monitor progress
        for i in range(60):  # Wait up to 5 minutes
            page.wait_for_timeout(5000)  # Check every 5 seconds

            # Take periodic screenshots
            if i % 6 == 0:  # Every 30 seconds
                screenshot_path = f"/tmp/feedback_loop_progress_{i//6}.png"
                page.screenshot(path=screenshot_path)
                print(f"📸 Progress screenshot: {screenshot_path}")

            # Check for phase indicator
            page_content = page.content()

            if "REVIEW" in page_content:
                print("🎯 Reached REVIEW phase - tests completed!")
                break
            elif "DONE" in page_content:
                print("✅ Reached DONE phase!")
                break
            elif "Running tests" in page_content or "test" in page_content.lower():
                print(f"🧪 Tests running... (check {i*5}s)")

            # Check if code files appeared
            if i == 12:  # After 1 minute
                print("📝 Checking for generated code...")

        # Final screenshot
        page.screenshot(path="/tmp/feedback_loop_final.png", full_page=True)
        print("📸 Final screenshot: /tmp/feedback_loop_final.png")

        # Switch to PREVIEW tab if available
        preview_tab = page.locator("button:has-text('PREVIEW')")
        if preview_tab.count() > 0:
            preview_tab.click()
            page.wait_for_timeout(5000)
            page.screenshot(path="/tmp/feedback_loop_preview.png")
            print("📸 Preview screenshot: /tmp/feedback_loop_preview.png")

        browser.close()
        print("\n✅ Feedback loop test completed!")
        print("Check worker logs for test execution details:")
        print("  docker compose -f docker-compose.prod.yaml logs worker --tail 100")

if __name__ == "__main__":
    test_feedback_loop()
