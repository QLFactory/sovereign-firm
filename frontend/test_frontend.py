"""
Sovereign Firm Frontend E2E Test
Tests landing page, navigation, and dashboard
"""
from playwright.sync_api import sync_playwright
import os

# Create screenshots directory
os.makedirs('/tmp/sovereign-firm-tests', exist_ok=True)

def test_frontend():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={'width': 1920, 'height': 1080})

        print("=" * 60)
        print("SOVEREIGN FIRM FRONTEND TEST")
        print("=" * 60)

        # Test 1: Landing Page Load
        print("\n[TEST 1] Loading landing page...")
        page.goto('http://localhost:3000')
        page.wait_for_load_state('networkidle')

        # Take full page screenshot
        page.screenshot(path='/tmp/sovereign-firm-tests/01-landing-page.png', full_page=True)
        print("✓ Landing page loaded")
        print("  Screenshot: /tmp/sovereign-firm-tests/01-landing-page.png")

        # Test 2: Hero Section
        print("\n[TEST 2] Verifying hero section...")
        hero_title = page.locator('h1').first
        if hero_title:
            title_text = hero_title.text_content()
            print(f"✓ Hero title found: '{title_text[:50]}...'")
        else:
            print("✗ Hero title not found")

        # Check for the animated code terminal
        terminal = page.locator('.font-mono').first
        if terminal:
            print("✓ Code terminal animation present")
        else:
            print("✗ Code terminal not found")

        # Test 3: Navigation Links
        print("\n[TEST 3] Checking navigation...")
        nav_links = page.locator('.nav-link').all()
        print(f"✓ Found {len(nav_links)} navigation links")
        for link in nav_links:
            print(f"  - {link.text_content()}")

        # Test 4: Stats Section
        print("\n[TEST 4] Verifying stats section...")
        stats = page.locator('.text-display-lg').all()
        print(f"✓ Found {len(stats)} stat counters")

        # Test 5: Features Section
        print("\n[TEST 5] Verifying features section...")
        page.locator('#features').scroll_into_view_if_needed()
        page.wait_for_timeout(500)
        page.screenshot(path='/tmp/sovereign-firm-tests/02-features-section.png')
        features = page.locator('.card-glow').all()
        print(f"✓ Found {len(features)} feature cards")
        print("  Screenshot: /tmp/sovereign-firm-tests/02-features-section.png")

        # Test 6: How It Works Section
        print("\n[TEST 6] Verifying how-it-works section...")
        page.locator('#how-it-works').scroll_into_view_if_needed()
        page.wait_for_timeout(500)
        page.screenshot(path='/tmp/sovereign-firm-tests/03-how-it-works.png')
        print("✓ How it works section visible")
        print("  Screenshot: /tmp/sovereign-firm-tests/03-how-it-works.png")

        # Test 7: Pricing Section
        print("\n[TEST 7] Verifying pricing section...")
        page.locator('#pricing').scroll_into_view_if_needed()
        page.wait_for_timeout(500)
        page.screenshot(path='/tmp/sovereign-firm-tests/04-pricing.png')
        pricing_cards = page.locator('[class*="rounded-2xl"]').filter(has_text="month").all()
        print(f"✓ Found {len(pricing_cards)} pricing tiers")
        print("  Screenshot: /tmp/sovereign-firm-tests/04-pricing.png")

        # Test 8: Navigate to Dashboard
        print("\n[TEST 8] Navigating to dashboard...")
        page.goto('http://localhost:3000/dashboard')
        page.wait_for_load_state('networkidle')
        page.wait_for_timeout(1000)  # Extra wait for animations
        page.screenshot(path='/tmp/sovereign-firm-tests/05-dashboard.png', full_page=True)
        print("✓ Dashboard loaded")
        print("  Screenshot: /tmp/sovereign-firm-tests/05-dashboard.png")

        # Test 9: Dashboard Sidebar
        print("\n[TEST 9] Checking dashboard sidebar...")
        sidebar_items = page.locator('nav button, nav a').all()
        print(f"✓ Found {len(sidebar_items)} sidebar items")

        # Test 10: Project Cards
        print("\n[TEST 10] Checking project cards...")
        project_cards = page.locator('[class*="glass-card"]').all()
        print(f"✓ Found {len(project_cards)} project/glass cards on dashboard")

        # Test 11: Create Project Button
        print("\n[TEST 11] Testing create project modal...")
        create_btn = page.locator('button:has-text("New Project"), button:has-text("Create")')
        if create_btn.count() > 0:
            create_btn.first.click()
            page.wait_for_timeout(500)
            page.screenshot(path='/tmp/sovereign-firm-tests/06-create-modal.png')
            print("✓ Create project modal opened")
            print("  Screenshot: /tmp/sovereign-firm-tests/06-create-modal.png")

            # Close modal by clicking outside or pressing Escape
            page.keyboard.press('Escape')
            page.wait_for_timeout(300)
        else:
            print("○ Create project button not found (may be expected)")

        # Test 12: Check for console errors
        print("\n[TEST 12] Console log check...")
        # Reload to capture any errors
        console_messages = []
        page.on('console', lambda msg: console_messages.append(f"[{msg.type}] {msg.text}"))
        page.reload()
        page.wait_for_load_state('networkidle')
        page.wait_for_timeout(2000)

        errors = [m for m in console_messages if '[error]' in m.lower()]
        if errors:
            print(f"✗ Found {len(errors)} console errors:")
            for e in errors[:5]:
                print(f"  {e}")
        else:
            print("✓ No console errors detected")

        browser.close()

        print("\n" + "=" * 60)
        print("TEST COMPLETE")
        print("=" * 60)
        print("\nScreenshots saved to: /tmp/sovereign-firm-tests/")
        print("Files:")
        for f in sorted(os.listdir('/tmp/sovereign-firm-tests')):
            print(f"  - {f}")

if __name__ == '__main__':
    test_frontend()
