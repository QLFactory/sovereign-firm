#!/usr/bin/env python3
"""
Full end-to-end test for Sovereign Firm: API + Frontend workflow.
Creates a project via API and verifies UI updates.
"""

from playwright.sync_api import sync_playwright
import requests
import json
import time
import sys

BASE_URL = "http://localhost:8080"
FRONTEND_URL = "http://localhost:3000"

def test_full_workflow():
    """Test complete workflow from API to UI."""

    print("=" * 60)
    print("SOVEREIGN FIRM FULL E2E TEST")
    print("=" * 60)

    # Test 1: Health Check
    print("\n[TEST 1] Health Check...")
    try:
        resp = requests.get(f"{BASE_URL}/health", timeout=5)
        health = resp.json()
        if health.get("status") == "healthy":
            print("  ✓ Orchestrator is healthy")
            print(f"    Streaming: {health.get('streaming', False)}")
        else:
            print(f"  ✗ Unhealthy: {health}")
            return False
    except Exception as e:
        print(f"  ✗ Health check failed: {e}")
        return False

    # Test 2: Create Project via API
    print("\n[TEST 2] Create Project via API...")
    project_name = f"e2e-test-{int(time.time())}"
    try:
        resp = requests.post(
            f"{BASE_URL}/api/pods",
            json={"project_name": project_name},
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        if resp.status_code == 200:
            data = resp.json()
            workflow_id = data.get("workflow_id")
            run_id = data.get("run_id")
            print(f"  ✓ Project created")
            print(f"    Workflow ID: {workflow_id}")
            print(f"    Run ID: {run_id}")
        else:
            print(f"  ✗ Create failed: {resp.status_code} - {resp.text}")
            return False
    except Exception as e:
        print(f"  ✗ Create failed: {e}")
        return False

    # Test 3: Get Project Status
    print("\n[TEST 3] Get Project Status...")
    time.sleep(1)  # Wait for workflow to initialize
    try:
        resp = requests.get(f"{BASE_URL}/api/pods/{workflow_id}", timeout=10)
        if resp.status_code == 200:
            state = resp.json()
            print(f"  ✓ Status retrieved")
            print(f"    Phase: {state.get('phase', 'unknown')}")
            print(f"    Chat history length: {len(state.get('chat_history', ''))}")
        else:
            print(f"  ✗ Status failed: {resp.status_code} - {resp.text}")
    except Exception as e:
        print(f"  ✗ Status failed: {e}")

    # Test 4: Send Message
    print("\n[TEST 4] Send Message...")
    try:
        resp = requests.post(
            f"{BASE_URL}/api/pods/{workflow_id}/message",
            json={"message": "Build a hello world React app"},
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        if resp.status_code == 200:
            print("  ✓ Message sent")
        else:
            print(f"  ✗ Message failed: {resp.status_code} - {resp.text}")
    except Exception as e:
        print(f"  ✗ Message failed: {e}")

    # Test 5: Wait for PM Response
    print("\n[TEST 5] Wait for PM Response...")
    time.sleep(5)  # Wait for LLM to respond
    try:
        resp = requests.get(f"{BASE_URL}/api/pods/{workflow_id}", timeout=10)
        if resp.status_code == 200:
            state = resp.json()
            chat_history = state.get("chat_history", "")
            if "PM:" in chat_history:
                print("  ✓ PM responded")
                # Show first PM response
                lines = chat_history.split("\n")
                pm_lines = [l for l in lines if l.startswith("PM:")]
                if pm_lines:
                    print(f"    PM: {pm_lines[0][:60]}...")
            else:
                print("  ⚠ No PM response yet")
        else:
            print(f"  ✗ Status failed: {resp.status_code}")
    except Exception as e:
        print(f"  ✗ Status failed: {e}")

    # Test 6: Frontend with Playwright
    print("\n[TEST 6] Frontend UI Test...")
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        try:
            page.goto(FRONTEND_URL)
            page.wait_for_load_state('domcontentloaded')
            page.wait_for_timeout(2000)

            # Check UI elements
            project_pod = page.locator('text=Project Pod')
            send_btn = page.locator('button:text("Send")')
            message_input = page.locator('input[placeholder*="Type message"]')

            if project_pod.count() > 0 and send_btn.count() > 0:
                print("  ✓ UI elements present")

                # Test input interaction
                message_input.fill("Test from Playwright")
                if message_input.input_value() == "Test from Playwright":
                    print("  ✓ Input interaction works")
                message_input.clear()

            # Take screenshot
            page.screenshot(path='/tmp/e2e-workflow-test.png', full_page=True)
            print("  ✓ Screenshot saved: /tmp/e2e-workflow-test.png")

        except Exception as e:
            print(f"  ✗ Frontend test failed: {e}")
        finally:
            browser.close()

    # Test 7: Send Approve and Check Implementation
    print("\n[TEST 7] Approve and Check Implementation...")
    try:
        # Send approve
        resp = requests.post(
            f"{BASE_URL}/api/pods/{workflow_id}/message",
            json={"message": "/approve"},
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        if resp.status_code == 200:
            print("  ✓ Approval sent")

        # Wait for implementation
        print("  ⏳ Waiting for implementation (15s)...")
        time.sleep(15)

        # Check state
        resp = requests.get(f"{BASE_URL}/api/pods/{workflow_id}", timeout=10)
        if resp.status_code == 200:
            state = resp.json()
            phase = state.get("phase", "unknown")
            code_files = state.get("code_files", {})

            print(f"  ✓ Current phase: {phase}")
            if code_files:
                print(f"  ✓ Code files generated: {len(code_files)}")
                for filename in list(code_files.keys())[:5]:
                    print(f"    - {filename}")
            else:
                print("  ⚠ No code files yet")
    except Exception as e:
        print(f"  ✗ Implementation check failed: {e}")

    print("\n" + "=" * 60)
    print("E2E TEST COMPLETED")
    print("=" * 60)
    print(f"\nWorkflow: {workflow_id}")
    print("View in Temporal UI: http://localhost:8088")

    return True

if __name__ == "__main__":
    success = test_full_workflow()
    sys.exit(0 if success else 1)
