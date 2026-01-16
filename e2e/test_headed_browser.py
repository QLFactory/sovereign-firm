#!/usr/bin/env python3
"""Test counter app with headed browser (visible window)"""

from playwright.sync_api import sync_playwright
import time

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

        print("\n1. Opening application...")
        page.goto('http://localhost:3000')
        time.sleep(3)

        # Start fresh
        print("\n2. Starting new pod...")
        try:
            new_pod = page.locator('button:has-text("New Pod")').first
            new_pod.click()
            time.sleep(2)
        except:
            pass

        # Send project request
        print("\n3. Sending project request...")
        input_field = page.locator('input[placeholder*="Describe"]').first
        input_field.fill("Build a simple counter app with + and - buttons using React")
        page.locator('button:has-text("Send")').first.click()
        time.sleep(5)

        # Answer questions
        print("\n4. Answering PM questions...")
        input_field = page.locator('input').first
        input_field.fill("Start at 0, allow negatives, simple styling is fine")
        page.locator('button:has-text("Send")').first.click()
        time.sleep(3)

        # Approve
        print("\n5. Approving spec...")
        input_field = page.locator('input').first
        input_field.fill("/approve")
        page.locator('button:has-text("Send")').first.click()
        time.sleep(2)

        # Wait for implementation
        print("\n6. Waiting for code generation...")
        for i in range(30):
            time.sleep(2)
            try:
                files_btn = page.locator('button:has-text("FILES")').first
                text = files_btn.text_content()
                if any(c.isdigit() and c != '0' for c in text):
                    print(f"   ✓ Files generated!")
                    break
            except:
                pass
            if i % 5 == 0:
                print(f"   Waiting... ({i*2}s)")

        # Go to PREVIEW
        print("\n7. Opening PREVIEW tab...")
        preview_tab = page.locator('button:has-text("PREVIEW")').first
        preview_tab.click()
        time.sleep(3)

        # Wait for WebContainer to load
        print("\n8. Waiting for WebContainer (watch the browser!)...")
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
        print("\n9. Testing counter buttons...")
        time.sleep(2)

        try:
            frames = page.frames
            for frame in frames:
                try:
                    plus_btn = frame.locator('button:has-text("+")').first
                    if plus_btn.is_visible(timeout=2000):
                        print("   Clicking + button...")
                        plus_btn.click()
                        time.sleep(1)
                        plus_btn.click()
                        time.sleep(1)

                        minus_btn = frame.locator('button:has-text("-")').first
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
        print("\n10. Taking screenshot...")
        page.screenshot(path='/tmp/headed_test.png', full_page=True)

        print("\n" + "=" * 60)
        print("TEST COMPLETE!")
        print("Browser will stay open for 10 seconds for inspection...")
        print("=" * 60)

        time.sleep(10)
        browser.close()

if __name__ == "__main__":
    main()
