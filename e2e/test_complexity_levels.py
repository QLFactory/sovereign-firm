#!/usr/bin/env python3
"""Complexity test - run with different prompts"""

from playwright.sync_api import sync_playwright
import time
import sys
import random
import string

FRONTEND_URL = "http://localhost:3000"

# Get prompt from command line or use default
PROMPT = sys.argv[1] if len(sys.argv) > 1 else "Create a simple Hello World page"
LEVEL = sys.argv[2] if len(sys.argv) > 2 else "1"

def random_string(length=6):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

def register_and_create_project(page, project_name):
    """Register a new user and create a project to get to workspace."""
    test_id = random_string()
    email = f"complexity_test_{test_id}@example.com"
    tenant = f"Complexity Test {test_id}"

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
    page.fill("textarea[placeholder*='Describe']", f"Complexity level {LEVEL} test project")
    page.click("button:has-text('Create Project')")
    page.wait_for_url("**/projects/**", timeout=15000)
    print(f"  ✓ Project created: {project_name}")

    return project_name

def main():
    with sync_playwright() as p:
        browser = p.chromium.launch(
            headless=False,
            slow_mo=300,
        )
        page = browser.new_page(viewport={"width": 1400, "height": 900})

        print("=" * 70)
        print(f"COMPLEXITY TEST - LEVEL {LEVEL}")
        print(f"Prompt: {PROMPT[:60]}...")
        print("=" * 70)

        # Setup
        print("\n1. Setting up user and project...")
        project_name = f"Level {LEVEL} Test {random_string()}"
        register_and_create_project(page, project_name)
        time.sleep(2)

        # Send project request
        print(f"\n2. Sending request...")
        chat_input = page.locator("input[name='message']")
        if chat_input.count() == 0:
            chat_input = page.locator("input").first
        chat_input.fill(PROMPT)
        page.locator("button:has-text('Send')").first.click()
        time.sleep(5)

        # Wait for PM response and answer
        print("\n3. Waiting for PM Agent...")
        time.sleep(3)

        # Quick answer + approve
        print("\n4. Answering and approving...")
        chat_input = page.locator("input[name='message']")
        if chat_input.count() == 0:
            chat_input = page.locator("input").first
        chat_input.fill("Keep it simple, use defaults. Looks good!")
        page.locator("button:has-text('Send')").first.click()
        time.sleep(3)

        chat_input = page.locator("input[name='message']")
        if chat_input.count() == 0:
            chat_input = page.locator("input").first
        chat_input.fill("/approve")
        page.locator("button:has-text('Send')").first.click()
        time.sleep(2)

        # Wait for code generation
        print("\n5. Waiting for code generation...")
        files_generated = False
        start_time = time.time()

        for i in range(60):  # Up to 2 minutes
            time.sleep(2)
            try:
                files_btn = page.locator("button:has-text('FILES')").first
                text = files_btn.text_content()
                if any(c.isdigit() and c != '0' for c in text):
                    elapsed = time.time() - start_time
                    print(f"   Files generated in {elapsed:.1f}s")
                    files_generated = True
                    break
            except:
                pass
            if i % 10 == 0 and i > 0:
                print(f"   Waiting... ({i*2}s)")

        if not files_generated:
            print("   Timeout waiting for files")

        # Check FILES tab
        print("\n6. Checking generated files...")
        files_tab = page.locator("button:has-text('FILES')").first
        files_tab.click()
        time.sleep(1)

        try:
            file_items = page.locator("button.font-mono").all()
            print(f"   Generated {len(file_items)} files:")
            for item in file_items[:8]:
                print(f"     - {item.text_content()}")
        except:
            pass

        page.screenshot(path=f'/tmp/level{LEVEL}_files.png')

        # Go to PREVIEW
        print("\n7. Opening PREVIEW...")
        preview_tab = page.locator("button:has-text('PREVIEW')").first
        preview_tab.click()
        time.sleep(3)

        # Wait for WebContainer
        print("\n8. Waiting for WebContainer...")
        for i in range(45):
            time.sleep(2)
            if i % 15 == 0 and i > 0:
                print(f"   Loading... ({i*2}s)")
                page.screenshot(path=f'/tmp/level{LEVEL}_preview_{i}.png')

        page.screenshot(path=f'/tmp/level{LEVEL}_preview_final.png')

        # Try to interact with preview
        print("\n9. Testing preview...")
        try:
            frames = page.frames
            for frame in frames[1:]:
                try:
                    content = frame.content()
                    if len(content) > 500:
                        print(f"   Preview has content ({len(content)} chars)")
                        # Try to find buttons
                        buttons = frame.locator("button").all()
                        if buttons:
                            print(f"   Found {len(buttons)} buttons")
                            buttons[0].click()
                            time.sleep(0.5)
                            print("   Clicked first button!")
                        break
                except:
                    continue
        except Exception as e:
            print(f"   Preview check: {e}")

        print("\n" + "=" * 70)
        print(f"LEVEL {LEVEL} TEST COMPLETE")
        print("Browser stays open for 10 seconds...")
        print("=" * 70)

        time.sleep(10)
        browser.close()

if __name__ == "__main__":
    main()
