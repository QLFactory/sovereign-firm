#!/usr/bin/env python3
"""
Full workflow test with live Sandpack preview.
Creates a project, goes through the workflow, and captures the live preview.
"""

from playwright.sync_api import sync_playwright
import requests
import time
import sys
import random
import string

BASE_URL = "http://localhost:8080"
FRONTEND_URL = "http://localhost:3000"

def random_string(length=6):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

def register_and_create_project(page):
    """Register a new user and create a project to get to workspace."""
    test_id = random_string()
    email = f"preview_test_{test_id}@example.com"
    tenant = f"Preview Test {test_id}"
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
    page.fill("textarea[placeholder*='Describe']", "A counter app for testing live preview")
    page.click("button:has-text('Create Project')")
    page.wait_for_url("**/projects/**", timeout=15000)
    print(f"  ✓ Project created: {project_name}")

    return project_name

def test_live_preview_workflow():
    """Test complete workflow with live preview."""

    print("=" * 60)
    print("LIVE PREVIEW WORKFLOW TEST")
    print("=" * 60)

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Step 1: Register and create project
        print("\n[STEP 1] Setting up user and project...")
        project_name = register_and_create_project(page)
        page.wait_for_load_state('networkidle')
        page.wait_for_timeout(2000)

        page.screenshot(path='/tmp/workflow-01-workspace.png', full_page=True)
        print("  ✓ Workspace loaded")

        # Step 2: Verify we're in the workspace
        print("\n[STEP 2] Verifying workspace...")
        chat_input = page.locator("input[name='message']")
        if chat_input.count() > 0:
            print("  ✓ Chat input found")
        else:
            print("  ⚠ Chat input not found, looking for alternatives...")
            chat_input = page.locator("input").first

        page.screenshot(path='/tmp/workflow-02-ready.png', full_page=True)

        # Step 3: Send a message describing the app
        print("\n[STEP 3] Describing app to PM Agent...")
        chat_input.fill("Build a simple counter app with + and - buttons")
        page.wait_for_timeout(500)

        send_btn = page.locator("button:has-text('Send')")
        send_btn.click()
        print("  ✓ Sent: Build a simple counter app with + and - buttons")

        # Step 4: Wait for PM response
        print("\n[STEP 4] Waiting for PM Agent response...")
        page.wait_for_timeout(8000)  # Wait for LLM response

        page.screenshot(path='/tmp/workflow-03-pm-response.png', full_page=True)

        # Check for PM response in chat
        pm_message = page.locator("text=PM Agent")
        if pm_message.count() > 0:
            print("  ✓ PM Agent responded")
        else:
            print("  ⚠ Checking chat history...")

        # Step 5: Send approval
        print("\n[STEP 5] Sending /approve...")
        chat_input = page.locator("input[name='message']")
        if chat_input.count() == 0:
            chat_input = page.locator("input").first
        chat_input.fill("/approve")
        send_btn = page.locator("button:has-text('Send')")
        send_btn.click()
        print("  ✓ Sent: /approve")

        page.screenshot(path='/tmp/workflow-04-approve-sent.png', full_page=True)

        # Step 6: Wait for code generation
        print("\n[STEP 6] Waiting for code generation...")
        print("  ⏳ This may take 15-30 seconds...")

        # Poll for implementation phase
        for i in range(20):
            page.wait_for_timeout(2000)

            # Check for phase change
            content = page.content()

            if "DEVELOPMENT" in content:
                print(f"  ✓ Phase: DEVELOPMENT (after {(i+1)*2}s)")
                page.screenshot(path='/tmp/workflow-05-development.png', full_page=True)

            if "TESTING" in content or "DEPLOYMENT" in content:
                print(f"  ✓ Phase progressed! (after {(i+1)*2}s)")
                break

            # Check FILES tab for generated files
            files_tab = page.locator("button:has-text('FILES')").first
            try:
                files_text = files_tab.text_content()
                if files_text and any(c.isdigit() and c != '0' for c in files_text):
                    print(f"  ✓ Files generated")
                    break
            except:
                pass

        page.screenshot(path='/tmp/workflow-06-code-generated.png', full_page=True)

        # Step 7: View the FILES tab
        print("\n[STEP 7] Viewing generated files...")
        files_tab = page.locator("button:has-text('FILES')").first
        if files_tab.count() > 0:
            files_tab.click()
            page.wait_for_timeout(1000)

            page.screenshot(path='/tmp/workflow-07-files-list.png', full_page=True)
            print("  ✓ FILES tab opened")

            # List the files
            file_buttons = page.locator("button.font-mono")
            file_count = file_buttons.count()
            print(f"  ✓ Found {file_count} files")

        # Step 8: View the PREVIEW tab
        print("\n[STEP 8] Viewing live preview...")
        preview_tab = page.locator("button:has-text('PREVIEW')").first
        if preview_tab.count() > 0:
            preview_tab.click()
            page.wait_for_timeout(3000)  # Wait for preview to render

            page.screenshot(path='/tmp/workflow-08-preview.png', full_page=True)
            print("  ✓ PREVIEW tab opened")

        # Step 9: View the CODE tab
        print("\n[STEP 9] Viewing code editor...")
        code_tab = page.locator("button:has-text('CODE')").first
        if code_tab.count() > 0:
            code_tab.click()
            page.wait_for_timeout(2000)

            page.screenshot(path='/tmp/workflow-09-code-editor.png', full_page=True)
            print("  ✓ CODE tab opened")

        # Step 10: Final summary
        print("\n[STEP 10] Capturing final state...")

        # Go back to PREVIEW for final screenshot
        preview_tab = page.locator("button:has-text('PREVIEW')").first
        if preview_tab.count() > 0:
            preview_tab.click()
            page.wait_for_timeout(2000)

        page.screenshot(path='/tmp/workflow-10-final.png', full_page=True)

        browser.close()

    print("\n" + "=" * 60)
    print("WORKFLOW TEST COMPLETED")
    print("=" * 60)
    print("\nScreenshots saved:")
    print("  /tmp/workflow-01-workspace.png   - Workspace loaded")
    print("  /tmp/workflow-02-ready.png       - Ready to chat")
    print("  /tmp/workflow-03-pm-response.png - PM Agent response")
    print("  /tmp/workflow-04-approve-sent.png - Approval sent")
    print("  /tmp/workflow-05-development.png - Development phase")
    print("  /tmp/workflow-06-code-generated.png - Code generated")
    print("  /tmp/workflow-07-files-list.png   - Generated files list")
    print("  /tmp/workflow-08-preview.png      - Live preview")
    print("  /tmp/workflow-09-code-editor.png  - Code editor")
    print("  /tmp/workflow-10-final.png        - Final state")

if __name__ == "__main__":
    test_live_preview_workflow()
