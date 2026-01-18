Feature: Project Creation
  As a user
  I want to create new projects
  So that I can build applications with AI assistance

  Background:
    Given I am logged in as "test@example.com"

  Scenario: Create a new project with default settings
    Given I am on the dashboard page
    When I click "New Project"
    And I fill in "Project Name" with "Counter App"
    And I fill in "Description" with "A simple counter application"
    And I click "Create Project"
    Then I should see the workspace page
    And the project phase should be "INTAKE"

  Scenario: Create a full-stack project
    Given I am on the dashboard page
    When I click "New Project"
    And I fill in "Project Name" with "E-Commerce App"
    And I fill in "Description" with "Full e-commerce platform with cart and checkout"
    And I select "React" for frontend
    And I select "Node.js" for backend
    And I select "PostgreSQL" for database
    And I enable "Full Stack" option
    And I click "Create Project"
    Then I should see the workspace page
    And the project phase should be "INTAKE"

  Scenario: Cannot create project without name
    Given I am on the dashboard page
    When I click "New Project"
    And I fill in "Description" with "Some description"
    Then the "Create Project" button should be disabled

  Scenario: View project list after creation
    Given I have created a project named "Test Project"
    When I navigate to the projects page
    Then I should see "Test Project" in the project list
    And the project status should show "In Progress"

  Scenario: Open existing project from dashboard
    Given I have created a project named "My Counter"
    When I navigate to the projects page
    And I click on the project "My Counter"
    Then I should see the workspace page
    And the page title should contain "My Counter"
