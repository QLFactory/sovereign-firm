Feature: Brownfield Import
  As a user
  I want to import existing projects
  So that I can enhance them with AI assistance

  Background:
    Given I am logged in as "test@example.com"
    And I am on the projects page

  Scenario: Open import modal
    When I click "Import Existing"
    Then I should see the import modal
    And I should see the Git Repository option
    And I should see the Local Directory option

  Scenario: Import from Git repository
    When I click "Import Existing"
    And I select "Git Repository" as source
    And I fill in "Project Name" with "Legacy App"
    And I fill in "Description" with "Importing legacy codebase"
    And I fill in "Repository URL" with "https://github.com/example/app.git"
    And I click "Start Import"
    Then I should see the analysis progress
    And I should see "Analyzing Codebase" message
    And I should see the file count updating

  Scenario: Import from local directory
    When I click "Import Existing"
    And I select "Local Directory" as source
    And I fill in "Project Name" with "Local Project"
    And I fill in "Description" with "Importing local project"
    And I fill in "Local Path" with "/path/to/project"
    And I click "Start Import"
    Then I should see the analysis progress

  Scenario: View import progress indicators
    Given I have started importing a project
    When the analysis is running
    Then I should see "Files Indexed" count
    And I should see "Code Chunks" count
    And I should see "Symbols Found" count
    And the counts should increase over time

  Scenario: Complete import successfully
    Given I have started importing a project
    When the import completes successfully
    Then I should see "Import Complete" message
    And I should be redirected to the workspace
    And the project phase should be "PLANNING"
    And the PM agent should have context from the analysis

  Scenario: Handle import failure
    Given I have started importing a project
    When the import fails
    Then I should see an error message
    And I should see "Try Again" button
    And I should be able to retry the import

  Scenario: Validate required fields
    When I click "Import Existing"
    And I leave "Project Name" empty
    Then the "Start Import" button should be disabled

  Scenario: Validate Git URL format
    When I click "Import Existing"
    And I select "Git Repository" as source
    And I fill in "Repository URL" with "not-a-valid-url"
    Then I should see a validation error
    And the "Start Import" button should be disabled

  Scenario: Import with technology detection
    Given I have imported a React project
    When the analysis is complete
    Then the detected tech stack should include "React"
    And the detected tech stack should include "JavaScript" or "TypeScript"
    And the architect agent should use this context

  Scenario: Import monorepo project
    Given I have imported a monorepo project
    When the analysis is complete
    Then I should see multiple packages detected
    And the PM agent should understand the monorepo structure
