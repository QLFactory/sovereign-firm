#!/usr/bin/env python3
"""
Full workflow test with live Sandpack preview.
Creates a project, goes through the workflow, and captures the live preview.
"""

from playwright.sync_api import sync_playwright
import requests
import time
import sys

BASE_URL = "http://localhost:8080"
FRONTEND_URL = "http://localhost:3000"

def test_live_preview_workflow():
    """Test complete workflow with live preview."""

    print("=" * 60)
    print("LIVE PREVIEW WORKFLOW TEST")
    print("=" * 60)

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Step 1: Load frontend
        print("\n[STEP 1] Loading frontend...")
        page.goto(FRONTEND_URL)
        page.wait_for_load_state('domcontentloaded')
        page.wait_for_timeout(3000)

        page.screenshot(path='/tmp/workflow-01-initial.png', full_page=True)
        print("  ✓ Frontend loaded")

        # Step 2: Wait for pod to start and get workflow ID from page
        print("\n[STEP 2] Waiting for pod to start...")
        page.wait_for_timeout(3000)

        # Check status shows PM Agent Active
        status = page.locator('text=PM Agent Active')
        if status.count() > 0:
            print("  ✓ Pod started - PM Agent Active")
        else:
            print("  ⚠ Waiting for pod...")
            page.wait_for_timeout(5000)

        page.screenshot(path='/tmp/workflow-02-pod-started.png', full_page=True)

        # Step 3: Send a message describing the app
        print("\n[STEP 3] Describing app to PM Agent...")
        input_field = page.locator('input[placeholder*="Describe"]')
        if input_field.count() == 0:
            input_field = page.locator('input')

        input_field.fill("Build a simple counter app with + and - buttons")
        page.wait_for_timeout(500)

        send_btn = page.locator('button:text("Send")')
        send_btn.click()
        print("  ✓ Sent: Build a simple counter app with + and - buttons")

        # Step 4: Wait for PM response
        print("\n[STEP 4] Waiting for PM Agent response...")
        page.wait_for_timeout(8000)  # Wait for LLM response

        page.screenshot(path='/tmp/workflow-03-pm-response.png', full_page=True)

        # Check for PM response in chat
        pm_message = page.locator('text=PM Agent')
        if pm_message.count() > 0:
            print("  ✓ PM Agent responded")
        else:
            print("  ⚠ Checking chat history...")

        # Step 5: Send approval
        print("\n[STEP 5] Sending /approve...")
        input_field.fill("/approve")
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
            impl_badge = page.locator('text=IMPLEMENTATION')
            review_badge = page.locator('text=REVIEW')

            if impl_badge.count() > 0:
                print(f"  ✓ Phase: IMPLEMENTATION (after {(i+1)*2}s)")
                page.screenshot(path='/tmp/workflow-05-implementation.png', full_page=True)

            if review_badge.count() > 0:
                print(f"  ✓ Phase: REVIEW - Code generated! (after {(i+1)*2}s)")
                break

            # Check FILES tab for generated files
            files_tab = page.locator('button:text("FILES")')
            if files_tab.count() > 0:
                # Check if file count badge exists
                file_count = page.locator('button:text("FILES") span')
                if file_count.count() > 0:
                    count_text = file_count.text_content()
                    if count_text and int(count_text) > 0:
                        print(f"  ✓ Files generated: {count_text}")
                        break

        page.screenshot(path='/tmp/workflow-06-code-generated.png', full_page=True)

        # Step 7: View the FILES tab
        print("\n[STEP 7] Viewing generated files...")
        files_tab = page.locator('button:text("FILES")')
        if files_tab.count() > 0:
            files_tab.click()
            page.wait_for_timeout(1000)

            page.screenshot(path='/tmp/workflow-07-files-list.png', full_page=True)
            print("  ✓ FILES tab opened")

            # List the files
            file_buttons = page.locator('button.font-mono')
            file_count = file_buttons.count()
            print(f"  ✓ Found {file_count} files:")
            for i in range(min(file_count, 10)):
                file_name = file_buttons.nth(i).text_content()
                print(f"    - {file_name}")

        # Step 8: View the PREVIEW tab
        print("\n[STEP 8] Viewing live preview...")
        preview_tab = page.locator('button:text("PREVIEW")')
        if preview_tab.count() > 0:
            preview_tab.click()
            page.wait_for_timeout(3000)  # Wait for Sandpack to render

            page.screenshot(path='/tmp/workflow-08-preview.png', full_page=True)
            print("  ✓ PREVIEW tab opened")

            # Check for Sandpack iframe
            sandpack_preview = page.locator('.sp-preview iframe, iframe[title*="Sandpack"]')
            if sandpack_preview.count() > 0:
                print("  ✓ Sandpack preview rendering")

        # Step 9: View the CODE tab
        print("\n[STEP 9] Viewing code editor...")
        code_tab = page.locator('button:text("CODE")')
        if code_tab.count() > 0:
            code_tab.click()
            page.wait_for_timeout(2000)

            page.screenshot(path='/tmp/workflow-09-code-editor.png', full_page=True)
            print("  ✓ CODE tab opened with editor")

        # Step 10: Final summary
        print("\n[STEP 10] Capturing final state...")

        # Go back to PREVIEW for final screenshot
        preview_tab = page.locator('button:text("PREVIEW")')
        if preview_tab.count() > 0:
            preview_tab.click()
            page.wait_for_timeout(2000)

        page.screenshot(path='/tmp/workflow-10-final.png', full_page=True)

        browser.close()

    print("\n" + "=" * 60)
    print("WORKFLOW TEST COMPLETED")
    print("=" * 60)
    print("\nScreenshots saved:")
    print("  /tmp/workflow-01-initial.png      - Initial load")
    print("  /tmp/workflow-02-pod-started.png  - Pod started")
    print("  /tmp/workflow-03-pm-response.png  - PM Agent response")
    print("  /tmp/workflow-04-approve-sent.png - Approval sent")
    print("  /tmp/workflow-05-implementation.png - Implementation phase")
    print("  /tmp/workflow-06-code-generated.png - Code generated")
    print("  /tmp/workflow-07-files-list.png   - Generated files list")
    print("  /tmp/workflow-08-preview.png      - Live preview")
    print("  /tmp/workflow-09-code-editor.png  - Code editor")
    print("  /tmp/workflow-10-final.png        - Final state")

if __name__ == "__main__":
    test_live_preview_workflow()
