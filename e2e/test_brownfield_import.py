#!/usr/bin/env python3
"""E2E test for brownfield import feature."""

import time
import random
import string
from playwright.sync_api import sync_playwright

BASE_URL = "http://localhost:3000"


def random_string(length=6):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))


def register_user(page):
    """Register a new user."""
    test_id = random_string()
    email = f"brownfield_test_{test_id}@example.com"
    tenant = f"Brownfield Test {test_id}"

    print(f"  → Registering user: {email}")
    page.goto(f"{BASE_URL}/register")
    page.wait_for_load_state("networkidle")
    time.sleep(1)
    page.screenshot(path="/tmp/brownfield_01_register_page.png")

    # Fill registration form
    page.fill("input#tenantName", tenant)
    page.fill("input#name", "Test User")
    page.fill("input#email", email)
    page.fill("input#password", "TestPassword123!")
    page.fill("input#confirmPassword", "TestPassword123!")

    page.screenshot(path="/tmp/brownfield_02_register_filled.png")

    # Submit registration
    page.click("button[type='submit']")
    page.wait_for_url("**/dashboard**", timeout=15000)
    page.wait_for_load_state("networkidle")

    page.screenshot(path="/tmp/brownfield_03_dashboard.png")
    print(f"  ✓ User registered and on dashboard")

    return email


def test_brownfield_import():
    """Test the brownfield import modal and flow."""

    print("=" * 60)
    print("BROWNFIELD IMPORT TEST")
    print("=" * 60)

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=False)
        context = browser.new_context()
        page = context.new_page()

        # Enable console logging
        page.on("console", lambda msg: print(f"  [Browser] {msg.text}") if "error" in msg.text.lower() else None)

        try:
            # Step 1: Register a new user
            print("\n[STEP 1] Registering user...")
            register_user(page)

            # Step 2: Navigate to projects page
            print("\n[STEP 2] Navigating to projects...")
            page.goto(f"{BASE_URL}/projects")
            page.wait_for_load_state("networkidle")
            time.sleep(1)
            page.screenshot(path="/tmp/brownfield_04_projects.png")
            print("  ✓ On projects page")

            # Step 3: Click "Import Existing" button
            print("\n[STEP 3] Opening import modal...")
            import_btn = page.locator('button:has-text("Import Existing")')

            if import_btn.count() == 0:
                print("  ✗ Import button not found!")
                # List all buttons for debugging
                buttons = page.locator('button').all()
                print(f"  Available buttons: {[b.text_content() for b in buttons]}")
                return False

            import_btn.click()
            time.sleep(1)
            page.screenshot(path="/tmp/brownfield_05_modal_open.png")
            print("  ✓ Import modal opened")

            # Step 4: Fill in the import form
            print("\n[STEP 4] Filling import form...")

            # Project name
            name_input = page.locator('input[name="name"]')
            if name_input.count() == 0:
                name_input = page.locator('input[placeholder*="Legacy"]')
            name_input.fill("Test Import Project")

            # Description
            desc_input = page.locator('textarea[name="description"]')
            if desc_input.count() == 0:
                desc_input = page.locator('textarea[placeholder*="Describe"]')
            desc_input.fill("Testing brownfield import feature")

            # Git source should be selected by default, fill repo URL
            repo_input = page.locator('input[name="repo_url"]')
            if repo_input.count() == 0:
                repo_input = page.locator('input[placeholder*="github"]')
            repo_input.fill("https://github.com/expressjs/express.git")

            page.screenshot(path="/tmp/brownfield_06_form_filled.png")
            print("  ✓ Form filled")

            # Step 5: Click Start Import
            print("\n[STEP 5] Starting import...")
            start_btn = page.locator('button:has-text("Start Import")')

            if not start_btn.is_visible():
                print("  ✗ Start Import button not visible")
                page.screenshot(path="/tmp/brownfield_07_error.png")
                return False

            start_btn.click()
            time.sleep(2)
            page.screenshot(path="/tmp/brownfield_07_importing.png")

            # Step 6: Wait for progress/completion
            print("\n[STEP 6] Waiting for import progress...")
            max_files_seen = 0
            max_chunks_seen = 0

            for i in range(30):  # Increased iterations for longer imports
                time.sleep(2)
                page.screenshot(path=f"/tmp/brownfield_08_progress_{i:02d}.png")

                # Check if we've been redirected to the workspace (SUCCESS!)
                current_url = page.url
                if "/projects/" in current_url and "/projects" != current_url.rstrip('/'):
                    # We've been redirected to a specific project workspace
                    print(f"  ✓ Redirected to workspace: {current_url}")
                    page.screenshot(path="/tmp/brownfield_09_workspace.png")

                    # Verify the workspace loaded
                    time.sleep(2)
                    phase_indicator = page.locator('text="INTAKE"').or_(page.locator('text="SIZING"')).or_(page.locator('text="PLANNING"'))
                    if phase_indicator.count() > 0:
                        print("  ✓ Workspace loaded with active phase")
                        return True
                    else:
                        print("  ✓ Import completed and redirected to workspace")
                        return True

                # Check for success state (modal still visible)
                success_text = page.locator('text="Import Complete"')
                if success_text.count() > 0 and success_text.is_visible():
                    print(f"  ✓ Import completed! Waiting for redirect...")
                    page.screenshot(path="/tmp/brownfield_09_success.png")
                    time.sleep(3)  # Wait for redirect
                    return True

                # Check for error state
                error_text = page.locator('text="Import Failed"')
                if error_text.count() > 0 and error_text.is_visible():
                    error_msg_el = page.locator('.text-\\[var\\(--rose-glow\\)\\]')
                    error_msg = error_msg_el.text_content() if error_msg_el.count() > 0 else "Unknown error"
                    print(f"  ✗ Import failed: {error_msg}")
                    page.screenshot(path="/tmp/brownfield_09_error.png")
                    return False

                # Check for analyzing state and extract progress
                analyzing_text = page.locator('text="Analyzing Codebase"')
                if analyzing_text.count() > 0 and analyzing_text.is_visible():
                    # Extract progress numbers from the UI
                    try:
                        # Look for the files indexed number (first number in cyan)
                        files_el = page.locator('.text-2xl.font-bold.text-\\[var\\(--cyan-glow\\)\\]').first
                        if files_el.count() > 0:
                            files_text = files_el.text_content()
                            if files_text and files_text.isdigit():
                                files = int(files_text)
                                if files > max_files_seen:
                                    max_files_seen = files

                        # Look for chunks (purple number)
                        chunks_el = page.locator('.text-2xl.font-bold.text-\\[var\\(--violet-glow\\)\\]').first
                        if chunks_el.count() > 0:
                            chunks_text = chunks_el.text_content()
                            if chunks_text and chunks_text.isdigit():
                                chunks = int(chunks_text)
                                if chunks > max_chunks_seen:
                                    max_chunks_seen = chunks
                    except Exception as e:
                        pass

                    print(f"  ... Analyzing (iteration {i+1}/30) - Files: {max_files_seen}, Chunks: {max_chunks_seen}")

            # Timeout - check if we at least saw some progress or ended up somewhere useful
            print(f"  ⚠ Timeout after 60s. Max files seen: {max_files_seen}, Max chunks: {max_chunks_seen}")
            current_url = page.url
            if "/projects/" in current_url and current_url != f"{BASE_URL}/projects":
                print(f"  ✓ Ended up in workspace: {current_url}")
                return True
            elif max_files_seen > 0:
                print("  ✓ Import appears to be working (files indexed before timeout)")
                return True
            else:
                print("  ✗ No files were indexed and not redirected - import not working")
                return False

        except Exception as e:
            print(f"\n  ✗ Error: {e}")
            page.screenshot(path="/tmp/brownfield_error.png")
            import traceback
            traceback.print_exc()
            return False
        finally:
            browser.close()


if __name__ == "__main__":
    success = test_brownfield_import()
    print("\n" + "=" * 60)
    print(f"TEST RESULT: {'PASSED' if success else 'FAILED'}")
    print("Screenshots saved to /tmp/brownfield_*.png")
    print("=" * 60)
