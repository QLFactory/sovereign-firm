"""Quick check of current UI state"""
from playwright.sync_api import sync_playwright
import os

os.makedirs('/tmp/sovereign-firm-tests', exist_ok=True)

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page(viewport={'width': 1920, 'height': 1080})

    # Go to dashboard
    page.goto('http://localhost:3000/dashboard')
    page.wait_for_load_state('networkidle')
    page.wait_for_timeout(2000)

    # Screenshot current state
    page.screenshot(path='/tmp/sovereign-firm-tests/progress-current.png', full_page=True)
    print("Screenshot saved: /tmp/sovereign-firm-tests/progress-current.png")

    # Check for the E2ETestApp project
    project_link = page.locator('button:has-text("E2ETestApp")')
    if project_link.count() > 0:
        project_link.click()
        page.wait_for_timeout(2000)
        page.screenshot(path='/tmp/sovereign-firm-tests/progress-console.png', full_page=True)
        print("Console screenshot: /tmp/sovereign-firm-tests/progress-console.png")

    browser.close()
