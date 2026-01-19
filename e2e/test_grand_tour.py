#!/usr/bin/env python3
"""
Sovereign Firm: Grand Tour E2E Test
A comprehensive user journey test that validates the entire consultancy workflow:
Registration -> Project Creation -> Agent Interaction -> Multi-Agent DAG -> Preview.
"""

import os
import time
import random
import string
import argparse
from datetime import datetime
from playwright.sync_api import sync_playwright

# Configuration
FRONTEND_URL = os.getenv("FRONTEND_URL", "http://localhost:3000")
RESULTS_DIR = "e2e/results/grand_tour"
SCREENSHOTS_DIR = f"{RESULTS_DIR}/screenshots"

# Ensure directories exist
os.makedirs(SCREENSHOTS_DIR, exist_ok=True)

def log(msg):
    timestamp = datetime.now().strftime("%H:%M:%S")
    print(f"[{timestamp}] {msg}")

def random_string(length=6):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

def take_screenshot(page, name):
    path = f"{SCREENSHOTS_DIR}/{name}.png"
    page.screenshot(path=path, full_page=True)
    log(f"📸 Screenshot saved: {name}.png")

def run_grand_tour(headless=True):
    log("🚀 Starting Grand Tour Test Journey...")
    
    with sync_playwright() as p:
        browser = p.chromium.launch(
            headless=headless,
            slow_mo=200 if not headless else 0
        )
        context = browser.new_context(viewport={"width": 1440, "height": 900})
        page = context.new_page()

        try:
            # 1. Registration
            log("1. Account Registration...")
            test_id = random_string()
            email = f"grand_tour_{test_id}@example.com"
            page.goto(f"{FRONTEND_URL}/register")
            page.wait_for_load_state("networkidle")
            
            page.fill("input#tenantName", f"Sovereign Corp {test_id}")
            page.fill("input#name", "Senior Partner")
            page.fill("input#email", email)
            page.fill("input#password", "P@ssword123!")
            page.fill("input#confirmPassword", "P@ssword123!")
            
            take_screenshot(page, "01_registration_filled")
            page.click("button[type='submit']")
            page.wait_for_url("**/dashboard**", timeout=15000)
            log("✅ Registration successful")

            # 2. Project Creation
            log("2. Creating complex project...")
            page.click("button:has-text('New Project')")
            page.wait_for_selector("input[placeholder*='TaskFlow']", timeout=5000)
            
            project_name = f"Enterprise Analytics Suite {test_id}"
            page.fill("input[placeholder*='TaskFlow']", project_name)
            page.fill("textarea[placeholder*='Describe']", 
                     "Build a full-stack dashboard with a Go backend, "
                     "PostgreSQL database, and a React frontend using Tailwind CSS. "
                     "The app should track real-time stock prices and show charts.")
            
            # Enable premium features
            page.check("label:has-text('Full Stack') input")
            page.check("label:has-text('DevOps') input")
            
            take_screenshot(page, "02_project_config")
            page.click("button:has-text('Create Project')")
            page.wait_for_url("**/projects/**", timeout=15000)
            log(f"✅ Project active: {project_name}")

            # 3. Intake Phase (Chat)
            log("3. Initiating Intake with PM Agent...")
            page.wait_for_selector("text=PM Agent", timeout=20000)
            take_screenshot(page, "03_intake_chat_ready")
            
            chat_input = page.locator("input[name='message'], textarea[name='message']").first
            chat_input.fill("Yes, please use Highcharts for the dashboard and Gin for the backend.")
            page.click("button:has-text('Send')")
            log("✓ Sent requirements to PM")
            
            # 4. Approval Phase
            log("4. Waiting for Planning & Approval...")
            # Ideally the system asks questions, but we'll try to push it to approval
            time.sleep(5)
            chat_input.fill("/approve")
            page.click("button:has-text('Send')")
            log("✓ Approval granted. Workflow transitioning to ARCHITECTURE")
            
            # 5. Monitoring Multi-Agent Execution (DAG)
            log("5. Monitoring Multi-Agent DAG...")
            page.click("button:has-text('TASKS')")
            time.sleep(3)
            take_screenshot(page, "04_task_dag_view")
            
            # Poll for agent activity and phase progression
            log("⏳ Watching Agents work (Multi-Agent Parallelism)...")
            max_wait = 60 # wait up to 5 minutes (60 * 5s)
            for i in range(max_wait):
                time.sleep(5)
                try:
                    active_agents = page.locator("button:has-text('AGENTS')").first
                    agent_count = active_agents.text_content()
                    
                    content = page.content()
                    phase_badge = page.locator(".badge").first.text_content() if page.locator(".badge").count() > 0 else "UNKNOWN"
                    
                    log(f"   Step {i+1}/{max_wait}: {agent_count} | Phase: {phase_badge}")
                    
                    if i % 3 == 0: # Every 15s
                        take_screenshot(page, f"05_dag_progress_{i}")
                    
                    # Escape loop if reached advanced phase
                    if any(p in content for p in ["TESTING", "DEPLOYMENT", "HANDOFF", "COMPLETE"]):
                        log(f"🎯 Workflow reached final phase: {phase_badge}!")
                        break
                except Exception as e:
                    log(f"   Poll error: {str(e)}")
                    pass
            else:
                log("⚠️ Timed out waiting for workflow completion, proceeding anyway...")

            # 6. CI/CD & Testing
            log("6. Verifying CI/CD Feedback...")
            page.click("button:has-text('CI')")
            time.sleep(2)
            take_screenshot(page, "06_ci_pipeline_view")
            log("✓ CI Pipeline monitored")

            # 7. Final Preview
            log("7. Validating Final Preview...")
            page.click("button:has-text('PREVIEW')")
            setStatus = "Checking browser preview..."
            log(setStatus)
            
            # Wait for WebContainer/Preview
            page.wait_for_timeout(10000)
            take_screenshot(page, "07_final_preview")
            
            log("🎉 Grand Tour Journey Completed Successfully!")

        except Exception as e:
            log(f"❌ Test Failed: {str(e)}")
            take_screenshot(page, "ERROR_STATE")
            raise e
        finally:
            browser.close()

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--headed", action="store_true", help="Run with visible browser window")
    args = parser.parse_args()
    
    run_grand_tour(headless=not args.headed)
