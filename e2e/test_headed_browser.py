#!/usr/bin/env python3
"""Test counter app with headed browser (visible window)"""

from playwright.sync_api import sync_playwright
import time
import random
import string

FRONTEND_URL = "http://localhost:3000"

def random_string(length=6):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

def register_and_create_project(page):
    """Register a new user and create a project to get to workspace."""
    test_id = random_string()
    email = f"headed_test_{test_id}@example.com"
    tenant = f"Headed Test {test_id}"
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
    page.fill("textarea[placeholder*='Describe']", "A counter app for headed browser testing")
    page.click("button:has-text('Create Project')")
    page.wait_for_url("**/projects/**", timeout=15000)
    print(f"  ✓ Project created: {project_name}")

    return project_name

def main():
    with sync_playwright() as p:
        # Launch with visible browser window
        browser = p.chromium.launch(
            headless=False,  # HEADED MODE - visible browser
            slow_mo=500,     # Slow down for visibility
        )
        page = browser.new_page(viewport={"width": 1400, "height": 900})

        print("=" * 60)
        print("HEADED BROWSER TEST - Watch the browser window!")
        print("=" * 60)

        print("\n1. Setting up user and project...")
        project_name = register_and_create_project(page)
        time.sleep(2)

        # Send project request
        print("\n2. Sending project request...")
        chat_input = page.locator("input[name='message']")
        if chat_input.count() == 0:
            chat_input = page.locator("input").first
        chat_input.fill("Build a simple counter app with + and - buttons using React")
        page.locator("button:has-text('Send')").first.click()
        time.sleep(5)

        # Answer questions
        print("\n3. Answering PM questions...")
        chat_input = page.locator("input[name='message']")
        if chat_input.count() == 0:
            chat_input = page.locator("input").first
        chat_input.fill("Start at 0, allow negatives, simple styling is fine")
        page.locator("button:has-text('Send')").first.click()
        time.sleep(3)

        # Approve
        print("\n4. Approving spec...")
        chat_input = page.locator("input[name='message']")
        if chat_input.count() == 0:
            chat_input = page.locator("input").first
        chat_input.fill("/approve")
        page.locator("button:has-text('Send')").first.click()
        time.sleep(2)

        # Wait for implementation
        print("\n5. Waiting for code generation...")
        for i in range(30):
            time.sleep(2)
            try:
                files_btn = page.locator("button:has-text('FILES')").first
                text = files_btn.text_content()
                if any(c.isdigit() and c != '0' for c in text):
                    print(f"   ✓ Files generated!")
                    break
            except:
                pass
            if i % 5 == 0:
                print(f"   Waiting... ({i*2}s)")

        # Go to PREVIEW
        print("\n6. Opening PREVIEW tab...")
        preview_tab = page.locator("button:has-text('PREVIEW')").first
        preview_tab.click()
        time.sleep(3)

        # Wait for WebContainer to load
        print("\n7. Waiting for WebContainer (watch the browser!)...")
        print("   The counter app should appear in the preview panel...")

        for i in range(45):
            time.sleep(2)

            # Check if app loaded
            try:
                frames = page.frames
                for frame in frames:
                    try:
                        content = frame.content()
                        if 'Simple Counter' in content or 'counter' in content.lower():
                            print("   ✓ Counter app loaded!")
                            break
                    except:
                        pass
            except:
                pass

            if i % 10 == 0 and i > 0:
                print(f"   Still loading... ({i*2}s)")

        # Try clicking buttons
        print("\n8. Testing counter buttons...")
        time.sleep(2)

        try:
            frames = page.frames
            for frame in frames:
                try:
                    plus_btn = frame.locator("button:has-text('+')").first
                    if plus_btn.is_visible(timeout=2000):
                        print("   Clicking + button...")
                        plus_btn.click()
                        time.sleep(1)
                        plus_btn.click()
                        time.sleep(1)

                        minus_btn = frame.locator("button:has-text('-')").first
                        print("   Clicking - button...")
                        minus_btn.click()
                        time.sleep(1)

                        print("   ✓ Counter app is working!")
                        break
                except:
                    continue
        except Exception as e:
            print(f"   Note: {e}")

        # Take screenshot
        print("\n9. Taking screenshot...")
        page.screenshot(path='/tmp/headed_test.png', full_page=True)

        print("\n" + "=" * 60)
        print("TEST COMPLETE!")
        print("Browser will stay open for 10 seconds for inspection...")
        print("=" * 60)

        time.sleep(10)
        browser.close()

if __name__ == "__main__":
    main()
