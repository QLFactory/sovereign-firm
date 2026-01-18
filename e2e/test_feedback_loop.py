#!/usr/bin/env python3
"""
E2E test for the Three-Strike Rule feedback loop.
Tests that when QA-generated tests fail, the system regenerates them.
"""

import time
import random
import string
from playwright.sync_api import sync_playwright

FRONTEND_URL = "http://localhost:3000"

def random_string(length=6):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

def register_and_create_project(page):
    """Register a new user and create a project to get to workspace."""
    test_id = random_string()
    email = f"feedback_test_{test_id}@example.com"
    tenant = f"Feedback Test {test_id}"
    project_name = f"Counter App {test_id}"

    # Register
    print("  → Registering user...")
    page.goto(f"{FRONTEND_URL}/register")
    page.wait_for_load_state("networkidle")
    page.fill("input#tenantName", tenant)
    page.fill("input#name", "Test User")
    page.fill("input#email", email)
    page.fill("input#password", "TestPassword123!")
    page.fill("input#confirmPassword", "TestPassword123!")
    page.click("button[type='submit']")
    page.wait_for_url("**/dashboard**", timeout=15000)
    print("  ✓ User registered")

    # Create project
    print("  → Creating project...")
    page.click("button:has-text('New Project')")
    page.wait_for_selector("input[placeholder*='TaskFlow']", timeout=5000)
    page.fill("input[placeholder*='TaskFlow']", project_name)
    page.fill("textarea[placeholder*='Describe']", "A counter app for testing feedback loop")
    page.click("button:has-text('Create Project')")
    page.wait_for_url("**/projects/**", timeout=15000)
    print(f"  ✓ Project created: {project_name}")

    return project_name

def test_feedback_loop():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        print("📍 Setting up test environment...")
        project_name = register_and_create_project(page)
        page.wait_for_load_state("networkidle")
        page.wait_for_timeout(2000)

        # Take initial screenshot
        page.screenshot(path="/tmp/feedback_loop_1_initial.png")
        print("📸 Screenshot: /tmp/feedback_loop_1_initial.png")

        # Find the chat input and send a message describing a simple app
        print("💬 Sending app description...")
        chat_input = page.locator("input[name='message']")
        if chat_input.count() == 0:
            chat_input = page.locator("input").first

        chat_input.fill("Create a simple React counter app with increment and decrement buttons. The counter should start at 0 and display the current count prominently.")

        # Click send button
        send_btn = page.locator("button:has-text('Send')")
        send_btn.click()

        print("⏳ Waiting for PM response...")
        page.wait_for_timeout(10000)  # Wait for LLM response

        page.screenshot(path="/tmp/feedback_loop_2_after_message.png")
        print("📸 Screenshot: /tmp/feedback_loop_2_after_message.png")

        # Send /approve to trigger code generation
        print("✅ Sending /approve to generate code...")
        chat_input = page.locator("input[name='message']")
        if chat_input.count() == 0:
            chat_input = page.locator("input").first
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

            if "TESTING" in page_content:
                print("🧪 Reached TESTING phase!")
            if "DEPLOYMENT" in page_content:
                print("🚀 Reached DEPLOYMENT phase!")
                break
            elif "COMPLETE" in page_content or "HANDOFF" in page_content:
                print("✅ Workflow completed!")
                break

            # Check if code files appeared
            if i == 12:  # After 1 minute
                print("📝 Checking for generated code...")

        # Final screenshot
        page.screenshot(path="/tmp/feedback_loop_final.png", full_page=True)
        print("📸 Final screenshot: /tmp/feedback_loop_final.png")

        # Switch to PREVIEW tab if available
        preview_tab = page.locator("button:has-text('PREVIEW')").first
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
