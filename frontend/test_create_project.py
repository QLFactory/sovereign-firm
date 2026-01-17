"""
Sovereign Firm - Create Project E2E Test
Tests creating a project from the UI and monitoring workflow execution
"""
from playwright.sync_api import sync_playwright
import os
import time
import json

os.makedirs('/tmp/sovereign-firm-tests', exist_ok=True)

def test_create_project():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={'width': 1920, 'height': 1080})

        # Capture console logs
        console_logs = []
        page.on('console', lambda msg: console_logs.append(f"[{msg.type}] {msg.text}"))

        # Capture network requests
        api_responses = []
        def handle_response(response):
            if '/api/pods' in response.url:
                try:
                    api_responses.append({
                        'url': response.url,
                        'status': response.status,
                        'body': response.text() if response.status < 400 else response.text()
                    })
                except:
                    pass
        page.on('response', handle_response)

        print("=" * 60)
        print("SOVEREIGN FIRM - CREATE PROJECT TEST")
        print("=" * 60)

        # Step 1: Navigate to dashboard
        print("\n[STEP 1] Navigating to dashboard...")
        page.goto('http://localhost:3000/dashboard')
        page.wait_for_load_state('networkidle')
        page.wait_for_timeout(1000)
        print("✓ Dashboard loaded")

        # Step 2: Click New Project button
        print("\n[STEP 2] Opening create project modal...")
        new_project_btn = page.locator('button:has-text("New Project")').first
        new_project_btn.click()
        page.wait_for_timeout(500)
        page.screenshot(path='/tmp/sovereign-firm-tests/create-01-modal-open.png')
        print("✓ Modal opened")

        # Step 3: Fill out the form
        print("\n[STEP 3] Filling out project form...")

        # Project name
        name_input = page.locator('input[placeholder*="TaskFlow"]')
        name_input.fill("E2ETestApp")
        print("  - Project name: E2ETestApp")

        # Description
        desc_input = page.locator('textarea[placeholder*="Describe"]')
        desc_input.fill("A comprehensive task management application with user authentication, real-time updates, and team collaboration features.")
        print("  - Description filled")

        # Select tech stack (use defaults: React, Node.js, PostgreSQL, AWS)
        print("  - Using default tech stack: React, Node.js, PostgreSQL, AWS")

        # Ensure all artifact toggles are checked
        checkboxes = page.locator('input[type="checkbox"]').all()
        for checkbox in checkboxes:
            if not checkbox.is_checked():
                checkbox.check()
        print("  - All artifact toggles enabled")

        page.screenshot(path='/tmp/sovereign-firm-tests/create-02-form-filled.png')
        print("✓ Form filled")

        # Step 4: Submit the form
        print("\n[STEP 4] Submitting project...")
        submit_btn = page.locator('button:has-text("Create Project")')
        submit_btn.click()

        # Wait for API response
        page.wait_for_timeout(3000)
        page.screenshot(path='/tmp/sovereign-firm-tests/create-03-after-submit.png')

        # Check for errors
        if api_responses:
            last_response = api_responses[-1]
            print(f"  - API Response: {last_response['status']}")
            if last_response['status'] < 400:
                try:
                    response_data = json.loads(last_response['body'])
                    workflow_id = response_data.get('workflow_id', 'unknown')
                    print(f"  - Workflow ID: {workflow_id}")
                    print("✓ Project created successfully!")
                except:
                    print(f"  - Response: {last_response['body'][:200]}")
            else:
                print(f"✗ API Error: {last_response['body'][:200]}")
        else:
            print("  - No API response captured (may have used cached data)")

        # Step 5: Check if redirected to console or project view
        print("\n[STEP 5] Checking post-creation state...")
        page.wait_for_timeout(2000)
        page.screenshot(path='/tmp/sovereign-firm-tests/create-04-result.png', full_page=True)

        # Check for project card or console
        current_url = page.url
        print(f"  - Current URL: {current_url}")

        # Look for success indicators
        phase_badge = page.locator('.badge:has-text("INTAKE"), .badge:has-text("SIZING"), .badge:has-text("PLANNING")')
        if phase_badge.count() > 0:
            print(f"✓ Project phase badge visible: {phase_badge.first.text_content()}")
        else:
            print("  - Phase badge not found (may still be initializing)")

        # Check for PodConsole or artifacts view
        console_view = page.locator('[class*="console"], [class*="terminal"]')
        if console_view.count() > 0:
            print("✓ Console view loaded")
        else:
            print("  - Console view not visible yet")

        # Step 6: Poll for workflow progress
        print("\n[STEP 6] Monitoring workflow progress...")
        for i in range(5):
            page.wait_for_timeout(2000)
            page.screenshot(path=f'/tmp/sovereign-firm-tests/create-05-progress-{i+1}.png')

            # Check phase changes
            badges = page.locator('.badge').all()
            current_phases = [b.text_content() for b in badges if b.text_content() in
                           ['INTAKE', 'SIZING', 'PLANNING', 'ARCHITECTURE', 'DEVELOPMENT',
                            'TESTING', 'DEPLOYMENT', 'OPERATIONS', 'HANDOFF', 'COMPLETE', 'FAILED']]
            if current_phases:
                print(f"  - Progress check {i+1}: Phases visible: {current_phases}")

        # Summary
        print("\n" + "=" * 60)
        print("TEST SUMMARY")
        print("=" * 60)
        print("\nScreenshots saved:")
        for f in sorted(os.listdir('/tmp/sovereign-firm-tests')):
            if f.startswith('create-'):
                print(f"  - /tmp/sovereign-firm-tests/{f}")

        if console_logs:
            errors = [log for log in console_logs if '[error]' in log.lower()]
            if errors:
                print(f"\nConsole errors detected: {len(errors)}")
                for e in errors[:5]:
                    print(f"  {e}")
            else:
                print("\n✓ No console errors")

        print("\nAPI Responses:")
        for resp in api_responses:
            print(f"  - {resp['url']}: {resp['status']}")

        browser.close()

if __name__ == '__main__':
    test_create_project()
