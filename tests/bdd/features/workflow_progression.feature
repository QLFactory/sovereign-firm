Feature: Workflow Progression
  As a user
  I want to progress through project phases
  So that my application gets built step by step

  Background:
    Given I am logged in as "test@example.com"
    And I have a project in "INTAKE" phase

  Scenario: Progress from INTAKE to SIZING
    Given the PM agent is gathering requirements
    When I describe my project as "Build a todo app with tasks and priorities"
    And I send the message "/approve"
    Then the project phase should transition to "SIZING"
    And I should see the complexity estimation

  Scenario: Progress through PLANNING phase
    Given the project is in "SIZING" phase
    When the sizing is complete
    Then the project phase should transition to "PLANNING"
    And I should see the project plan
    And I should see effort estimates

  Scenario: Progress through ARCHITECTURE phase
    Given the project is in "PLANNING" phase
    When the planning is complete
    Then the project phase should transition to "ARCHITECTURE"
    And I should see the tech stack selection
    And I should see the system design

  Scenario: Progress through DEVELOPMENT phase
    Given the project is in "ARCHITECTURE" phase
    When the architecture is complete
    Then the project phase should transition to "DEVELOPMENT"
    And I should see code being generated
    And I should see the file browser populated

  Scenario: Progress through TESTING phase
    Given the project is in "DEVELOPMENT" phase
    When the development is complete
    Then the project phase should transition to "TESTING"
    And I should see test files being generated
    And I should see test results

  Scenario: Complete project delivery
    Given the project is in "TESTING" phase
    And all tests pass
    When the testing is complete
    Then the project phase should transition to "DEPLOYMENT"
    And I should see deployment artifacts

  Scenario: Handle test failures with retry
    Given the project is in "TESTING" phase
    When tests fail
    Then the QA agent should regenerate tests
    And the attempt count should increase
    And I should see the retry indicator

  Scenario: Three-strike rule for failures
    Given the project is in "DEVELOPMENT" phase
    When code generation fails 3 times
    Then the project should be marked as "FAILED"
    And I should see an error message

  Scenario: View phase history
    Given the project has progressed through multiple phases
    When I view the project details
    Then I should see the complete phase history
    And each phase transition should show timestamps
