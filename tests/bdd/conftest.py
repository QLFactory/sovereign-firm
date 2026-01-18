"""
BDD Test Configuration and Fixtures
"""

import os
import pytest
from typing import Generator

# Optional Playwright import for E2E tests
try:
    from playwright.sync_api import Page, Browser, BrowserContext, Playwright, sync_playwright
    PLAYWRIGHT_AVAILABLE = True
except ImportError:
    PLAYWRIGHT_AVAILABLE = False
    Page = None
    Browser = None
    BrowserContext = None


# =============================================================================
# Configuration
# =============================================================================

BASE_URL = os.environ.get("TEST_BASE_URL", "http://localhost:3000")
API_URL = os.environ.get("TEST_API_URL", "http://localhost:8080")


# =============================================================================
# Fixtures
# =============================================================================

@pytest.fixture(scope="session")
def base_url() -> str:
    """Return the base URL for the frontend."""
    return BASE_URL


@pytest.fixture(scope="session")
def api_url() -> str:
    """Return the base URL for the API."""
    return API_URL


@pytest.fixture(scope="session")
def playwright_instance():
    """Create a Playwright instance for the test session."""
    if not PLAYWRIGHT_AVAILABLE:
        pytest.skip("Playwright not available")

    with sync_playwright() as playwright:
        yield playwright


@pytest.fixture(scope="session")
def browser(playwright_instance) -> Generator:
    """Create a browser instance for the test session."""
    if not PLAYWRIGHT_AVAILABLE:
        pytest.skip("Playwright not available")

    browser = playwright_instance.chromium.launch(
        headless=os.environ.get("HEADLESS", "true").lower() == "true"
    )
    yield browser
    browser.close()


@pytest.fixture
def context(browser) -> Generator:
    """Create a new browser context for each test."""
    if not PLAYWRIGHT_AVAILABLE:
        pytest.skip("Playwright not available")

    context = browser.new_context(
        viewport={"width": 1280, "height": 720},
        base_url=BASE_URL
    )
    yield context
    context.close()


@pytest.fixture
def page(context) -> Generator:
    """Create a new page for each test."""
    if not PLAYWRIGHT_AVAILABLE:
        pytest.skip("Playwright not available")

    page = context.new_page()
    yield page
    page.close()


# =============================================================================
# Test Data Fixtures
# =============================================================================

@pytest.fixture
def test_user() -> dict:
    """Return test user credentials."""
    return {
        "email": "test@example.com",
        "password": "testpassword123",
        "name": "Test User"
    }


@pytest.fixture
def test_project() -> dict:
    """Return test project configuration."""
    return {
        "project_name": "Test Counter App",
        "initial_message": "Build a simple counter app with increment and decrement buttons",
        "enable_full_stack": True,
        "enable_deployment": False,
        "enable_sre": False,
        "preferred_frontend": "react",
        "preferred_backend": "nodejs",
        "preferred_database": "postgresql",
        "preferred_cloud": "aws"
    }


@pytest.fixture
def brownfield_project() -> dict:
    """Return brownfield project configuration."""
    return {
        "name": "Legacy App Import",
        "description": "Importing existing legacy application",
        "repo_url": "https://github.com/example/legacy-app.git",
        "local_path": None
    }


# =============================================================================
# Helper Class
# =============================================================================

class TestHelpers:
    """Helper methods for BDD tests."""

    def __init__(self, page=None, base_url=None, api_url=None):
        self.page = page
        self.base_url = base_url
        self.api_url = api_url

    def navigate_to(self, path: str) -> None:
        """Navigate to a path."""
        if self.page:
            self.page.goto(path)

    def click_button(self, text: str) -> None:
        """Click a button by text."""
        if self.page:
            self.page.click(f"button:has-text('{text}')")

    def fill_input(self, selector: str, value: str) -> None:
        """Fill an input field."""
        if self.page:
            self.page.fill(selector, value)

    def wait_for_text(self, text: str, timeout: int = 5000) -> None:
        """Wait for text to appear on page."""
        if self.page:
            self.page.wait_for_selector(f"text={text}", timeout=timeout)

    def register_and_login(self, email: str, password: str = "testpassword") -> None:
        """Register and login a test user."""
        if not self.page:
            return

        self.page.goto("/auth/register")
        self.page.fill("[name='email']", email)
        self.page.fill("[name='password']", password)
        self.page.click("button[type='submit']")

        # Wait for redirect
        self.page.wait_for_url("**/projects**", timeout=10000)

    def create_project(self, name: str, description: str) -> str:
        """Create a new project and return its ID."""
        if not self.page:
            return ""

        self.click_button("New Project")
        self.fill_input("[name='project_name']", name)
        self.fill_input("[name='initial_message']", description)
        self.click_button("Create Project")

        # Wait for redirect and extract project ID from URL
        self.page.wait_for_url("**/projects/**", timeout=10000)
        return self.page.url.split("/")[-1]

    def get_current_phase(self) -> str:
        """Get the current project phase from the UI."""
        if not self.page:
            return ""

        phase_element = self.page.locator("[data-testid='current-phase']")
        if phase_element.count() > 0:
            return phase_element.inner_text()

        # Fallback to badge element
        badge = self.page.locator(".badge").first
        if badge.count() > 0:
            return badge.inner_text()

        return ""


@pytest.fixture
def test_helpers(page, base_url, api_url) -> TestHelpers:
    """Return test helper instance."""
    return TestHelpers(page, base_url, api_url)


# =============================================================================
# Hooks
# =============================================================================

def pytest_bdd_step_error(request, feature, scenario, step, step_func, step_func_args, exception):
    """Handle step errors - take screenshot on failure."""
    if PLAYWRIGHT_AVAILABLE:
        page = request.getfixturevalue("page") if "page" in request.fixturenames else None
        if page:
            screenshot_dir = "tests/bdd/screenshots"
            os.makedirs(screenshot_dir, exist_ok=True)
            screenshot_path = f"{screenshot_dir}/{scenario.name}_{step.name}.png"
            page.screenshot(path=screenshot_path)
            print(f"Screenshot saved: {screenshot_path}")
