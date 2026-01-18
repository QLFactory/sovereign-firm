"""
Step definitions for project creation feature.
"""

import pytest
from pytest_bdd import given, when, then, scenarios, parsers

# Load scenarios from feature file
scenarios("../features/project_creation.feature")


# =============================================================================
# Given Steps
# =============================================================================

@given(parsers.parse('I am logged in as "{email}"'))
def logged_in_user(test_helpers, email):
    """Log in a test user."""
    test_helpers.register_and_login(email)


@given("I am on the dashboard page")
def on_dashboard_page(page, base_url):
    """Navigate to dashboard page."""
    page.goto(f"{base_url}/projects")
    page.wait_for_load_state("networkidle")


@given(parsers.parse('I have created a project named "{project_name}"'))
def created_project(test_helpers, project_name):
    """Create a project for testing."""
    test_helpers.navigate_to("/projects")
    test_helpers.create_project(project_name, "Test project description")


# =============================================================================
# When Steps
# =============================================================================

@when(parsers.parse('I click "{button_text}"'))
def click_button(page, button_text):
    """Click a button by text."""
    page.click(f"button:has-text('{button_text}')")
    page.wait_for_load_state("networkidle")


@when(parsers.parse('I fill in "{field_name}" with "{value}"'))
def fill_in_field(page, field_name, value):
    """Fill in a form field."""
    field_mapping = {
        "Project Name": "[name='project_name'], input[placeholder*='Project']",
        "Description": "[name='initial_message'], textarea[placeholder*='Describe']",
    }
    selector = field_mapping.get(field_name, f"[name='{field_name.lower().replace(' ', '_')}']")
    page.fill(selector, value)


@when(parsers.parse('I select "{option}" for {field}'))
def select_option(page, option, field):
    """Select an option in a dropdown."""
    field_mapping = {
        "frontend": "select[name='preferred_frontend']",
        "backend": "select[name='preferred_backend']",
        "database": "select[name='preferred_database']",
        "cloud": "select[name='preferred_cloud']",
    }
    selector = field_mapping.get(field.lower(), f"select[name='preferred_{field.lower()}']")
    page.select_option(selector, label=option)


@when(parsers.parse('I enable "{option}" option'))
def enable_option(page, option):
    """Enable a checkbox option."""
    option_mapping = {
        "Full Stack": "enable_full_stack",
        "DevOps": "enable_deployment",
        "SRE": "enable_sre",
    }
    name = option_mapping.get(option, option.lower().replace(" ", "_"))
    checkbox = page.locator(f"input[name='{name}'], input[type='checkbox']:near(:text('{option}'))")
    if not checkbox.is_checked():
        checkbox.click()


@when("I navigate to the projects page")
def navigate_to_projects(page, base_url):
    """Navigate to projects page."""
    page.goto(f"{base_url}/projects")
    page.wait_for_load_state("networkidle")


@when(parsers.parse('I click on the project "{project_name}"'))
def click_project(page, project_name):
    """Click on a project card."""
    page.click(f"text={project_name}")
    page.wait_for_load_state("networkidle")


# =============================================================================
# Then Steps
# =============================================================================

@then("I should see the workspace page")
def see_workspace_page(page):
    """Verify we're on the workspace page."""
    page.wait_for_url("**/projects/**")
    assert "/projects/" in page.url


@then(parsers.parse('the project phase should be "{phase}"'))
def verify_phase(page, phase):
    """Verify the project phase."""
    # Look for phase indicator
    phase_element = page.locator(f"text={phase}")
    assert phase_element.is_visible() or page.locator(f".badge:has-text('{phase}')").is_visible()


@then(parsers.parse('the "{button}" button should be disabled'))
def button_disabled(page, button):
    """Verify a button is disabled."""
    button_element = page.locator(f"button:has-text('{button}')")
    assert button_element.is_disabled()


@then(parsers.parse('I should see "{text}" in the project list'))
def see_in_project_list(page, text):
    """Verify text appears in project list."""
    page.wait_for_selector(f"text={text}")
    assert page.locator(f"text={text}").is_visible()


@then(parsers.parse('the project status should show "{status}"'))
def verify_project_status(page, status):
    """Verify project status indicator."""
    status_element = page.locator(f"text={status}")
    assert status_element.is_visible()


@then(parsers.parse('the page title should contain "{text}"'))
def verify_page_title(page, text):
    """Verify page title contains text."""
    title = page.title()
    assert text in title or page.locator(f"h1:has-text('{text}'), h2:has-text('{text}')").is_visible()
