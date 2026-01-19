# User Journeys Documentation

**Last Updated:** January 19, 2026
**Platform:** Sovereign Firm - AI Software Consultancy

---

## Overview

This document details all primary user journeys through the Sovereign Firm platform. Each journey represents a complete workflow from user initiation to deliverable completion.

---

## Workflow Phases (9 Total)

```
INTAKE → SIZING → PLANNING → ARCHITECTURE → DEVELOPMENT → TESTING → DEPLOYMENT → OPERATIONS → HANDOFF
```

| Phase | Description | Key Activities |
|-------|-------------|----------------|
| **INTAKE** | Initial project requirements gathering | Chat with PM Agent, describe project |
| **SIZING** | Effort estimation and scope definition | T-shirt sizing, resource allocation |
| **PLANNING** | Detailed project planning | Sprint planning, milestone definition |
| **ARCHITECTURE** | System design and tech decisions | Solution Architect designs system |
| **DEVELOPMENT** | Code generation by Dev Agents | Multi-agent parallel development |
| **TESTING** | QA Agent generates and runs tests | Three-strike feedback loop |
| **DEPLOYMENT** | DevOps Agent creates CI/CD pipelines | Infrastructure as code |
| **OPERATIONS** | SRE Agent monitors and maintains | Observability, alerts |
| **HANDOFF** | Project delivery to customer | Documentation, knowledge transfer |

---

## Primary User Journeys

### Journey 1: New User Onboarding

**E2E Test:** `test_frontend.py`
**Personas:** New User, Business Owner

#### Steps

1. **Navigate to Registration**
   - User visits `http://localhost:3000/register`
   - Registration form displays

2. **Fill Registration Form**
   - Enter Tenant Name (company)
   - Enter Full Name
   - Enter Email Address
   - Enter Password (with confirmation)

3. **Submit Registration**
   - Click "Submit" button
   - Backend creates tenant and user
   - Redirect to Dashboard

4. **Explore Dashboard**
   - View welcome message
   - See "New Project" and "Import Existing" buttons
   - View empty project list

#### Success Criteria
- User redirected to `/dashboard` after registration
- Dashboard shows company name
- No console errors

---

### Journey 2: Greenfield Project (New Project from Scratch)

**E2E Test:** `test_grand_tour.py`
**Personas:** Product Owner, Developer

#### Steps

1. **Create New Project**
   - Click "New Project" button
   - Fill project name (e.g., "Enterprise Analytics Suite")
   - Enter project description
   - Select features (Full Stack, DevOps, etc.)
   - Click "Create Project"

2. **Intake Phase - Chat with PM Agent**
   - Workspace loads with INTAKE phase active
   - PM Agent greets user and asks clarifying questions
   - User responds with requirements
   - PM Agent confirms understanding

3. **Approve Specification**
   - User reviews generated spec in SPEC tab
   - User sends `/approve` command
   - System transitions to ARCHITECTURE phase

4. **Multi-Agent Execution**
   - Watch TASKS tab for DAG progress
   - Observe parallel agent activity
   - Monitor AGENTS tab for active workers

5. **Code Generation**
   - Dev Agents generate code
   - FILES tab populates with generated files
   - CODE tab shows file contents

6. **Testing Phase**
   - QA Agent generates test files
   - Tests run automatically
   - Three-strike rule: failures trigger regeneration

7. **Preview Result**
   - PREVIEW tab shows WebContainer
   - Live app runs in browser
   - User can interact with generated app

#### Success Criteria
- All phases complete (INTAKE → HANDOFF)
- Code files generated
- Preview renders application
- CI pipeline shows green

---

### Journey 3: Brownfield Import (Existing Codebase)

**E2E Test:** `test_brownfield_import.py`
**Personas:** Tech Lead, Migration Team

#### Steps

1. **Open Import Modal**
   - Click "Import Existing" button
   - Import modal appears

2. **Configure Import Source**
   - Enter project name
   - Enter description
   - Select source type: Git Repository
   - Enter repository URL (e.g., `https://github.com/expressjs/express.git`)

3. **Start Import**
   - Click "Start Import"
   - Progress indicators show:
     - Files indexed count
     - Chunks created count
     - Analysis progress

4. **Analysis Complete**
   - System analyzes codebase structure
   - Creates embeddings in ChromaDB
   - Generates initial understanding

5. **Enter Workspace**
   - Redirect to workspace
   - Codebase context available
   - Can extend existing functionality

#### Success Criteria
- Import progress shows files indexed
- Analysis completes without errors
- Workspace loads with codebase context
- Can query about existing code

---

### Journey 4: Code Preview (WebContainer)

**E2E Test:** `test_sandpack_preview.py`
**Personas:** Developer, QA Engineer

#### Steps

1. **Access Workspace**
   - Register user
   - Create or open project

2. **Verify Workspace Layout**
   - Left Panel: CHAT, SPEC, FILES, TASKS, AGENTS tabs
   - Right Panel: PREVIEW, CODE, TERMINAL, CI, EVENTS tabs
   - Phase badge visible (INTAKE, SIZING, etc.)

3. **Navigate Tabs**
   - Click PREVIEW tab
   - WebContainer boots
   - Click CODE tab to view source
   - Click TERMINAL tab for command line
   - Click CI tab for pipeline status

4. **Interact with Preview**
   - WebContainer loads React app
   - App renders in iframe
   - User can click buttons, fill forms
   - Real-time updates visible

#### Success Criteria
- All tabs render correctly
- PREVIEW shows WebContainer or loading state
- Tab navigation works smoothly
- No JavaScript errors in console

---

### Journey 5: Complexity Levels (App Difficulty Testing)

**E2E Test:** `test_complexity_levels.py`
**Personas:** Product Manager, Architect

#### Complexity Tiers

| Level | Type | Example Applications |
|-------|------|---------------------|
| 1-5 | Simple | Hello World, Counter, Calculator |
| 6-10 | Moderate | Todo List, Form Wizard, Chat App |
| 11-15 | Complex | E-commerce, Dashboard, Kanban Board |
| 16-20 | Advanced | Survey Builder, Analytics, CRM |
| E1-E10 | Enterprise | SSO Login, Admin Dashboard, API Gateway |

#### Steps

1. **Select Complexity Level**
   - Run test with prompt and level:
   ```bash
   python test_complexity_levels.py "Build a todo app" "todo"
   ```

2. **Project Setup**
   - Register user
   - Create project for complexity level

3. **Send Request**
   - Describe app requirements
   - Answer PM questions
   - Approve specification

4. **Wait for Generation**
   - Monitor FILES tab
   - Track generation time
   - Note file count

5. **Verify Preview**
   - Open PREVIEW tab
   - WebContainer boots
   - Verify app functionality
   - Test button clicks, form inputs

#### Success Criteria
- Files generated within timeout
- Preview loads and renders
- App matches requested functionality
- No dependency errors (except known: react-router-dom, jspdf)

---

### Journey 6: QA Feedback Loop (Three-Strike Rule)

**E2E Test:** `test_feedback_loop.py`
**Personas:** QA Lead, Developer

#### Three-Strike Rule

When tests fail, the system:
1. **Strike 1:** QA Agent regenerates failing tests
2. **Strike 2:** Dev Agent refactors code
3. **Strike 3:** Human escalation or skip

#### Steps

1. **Create Test Project**
   - Register user
   - Create counter app project

2. **Trigger Code Generation**
   - Describe requirements
   - Send `/approve` command
   - Wait for code generation

3. **Monitor Testing Phase**
   - Observe TESTING phase
   - Watch for test execution
   - Check CI tab for results

4. **Verify Feedback Loop**
   - If tests fail, observe retry
   - Check worker logs for regeneration
   - Monitor phase transitions

5. **Final State**
   - Workflow reaches DEPLOYMENT or COMPLETE
   - Tests pass (or max retries reached)
   - Preview shows working app

#### Success Criteria
- Testing phase triggered
- Feedback loop executes on failures
- Workflow progresses past TESTING
- CI shows final test results

---

### Journey 7: Full Workflow E2E (API + Frontend)

**E2E Test:** `test_workflow_e2e.py`
**Personas:** DevOps Engineer, SRE

#### Steps

1. **Health Check**
   - `GET /health` returns healthy status
   - Streaming enabled

2. **Create Project via API**
   - `POST /api/pods` with project name
   - Receive workflow_id and run_id

3. **Get Project Status**
   - `GET /api/pods/{workflow_id}`
   - Check phase and chat_history

4. **Send Message via API**
   - `POST /api/pods/{workflow_id}/message`
   - Message: "Build a hello world React app"

5. **Wait for PM Response**
   - Poll status endpoint
   - Verify PM: prefix in chat_history

6. **Frontend UI Test**
   - Open frontend in Playwright
   - Verify UI elements present
   - Test input interaction

7. **Approve and Verify**
   - Send `/approve` via API
   - Wait for implementation
   - Check code_files generated

#### Success Criteria
- API endpoints respond correctly
- Workflow created and progresses
- PM Agent responds to messages
- Code files generated after approval

---

### Journey 8: Interactive Demo (Headed Browser)

**E2E Test:** `test_headed_browser.py`
**Personas:** Demo Team, Sales Engineer

#### Steps

1. **Launch Visible Browser**
   - Browser window opens (headless=False)
   - Slow motion mode (500ms delay)
   - Large viewport (1400x900)

2. **Walk Through Registration**
   - Watch form fill animation
   - See submit button click
   - Observe redirect

3. **Create Project Visually**
   - See project creation flow
   - Watch PM Agent interaction
   - Observe approval flow

4. **Watch Code Generation**
   - Monitor FILES tab updates
   - See file count increase
   - Watch phase transitions

5. **Interact with Preview**
   - See WebContainer boot
   - Watch app render
   - Click buttons in generated app
   - Verify counter increments/decrements

6. **Capture Screenshot**
   - Final screenshot saved
   - Browser stays open for inspection

#### Success Criteria
- Visible browser interaction
- All steps observable
- Counter app functional
- Screenshot captured

---

## Test File Reference

| # | Journey | E2E Test File | Key Validations |
|---|---------|---------------|-----------------|
| 1 | New User Onboarding | `test_frontend.py` | Registration, dashboard, UI elements |
| 2 | Greenfield Project | `test_grand_tour.py` | Full workflow, multi-agent, preview |
| 3 | Brownfield Import | `test_brownfield_import.py` | Import modal, file indexing, analysis |
| 4 | Code Preview | `test_sandpack_preview.py` | WebContainer, tabs, chat input |
| 5 | Complexity Levels | `test_complexity_levels.py` | Variable prompts, generation time |
| 6 | QA Feedback Loop | `test_feedback_loop.py` | Testing phase, retry mechanism |
| 7 | Full Workflow E2E | `test_workflow_e2e.py` | API + frontend integration |
| 8 | Interactive Demo | `test_headed_browser.py` | Visible browser, interaction |

---

## Running the Tests

```bash
# Navigate to e2e directory
cd /Users/satish/qlp-projects/QLFactory/sovereign-firm/e2e

# Activate virtual environment
source venv/bin/activate

# Run individual tests
python test_frontend.py
python test_grand_tour.py
python test_brownfield_import.py
python test_sandpack_preview.py
python test_workflow_e2e.py
python test_feedback_loop.py
python test_complexity_levels.py "Build a todo app" "todo"
python test_headed_browser.py

# Run with headed mode (visible browser)
python test_grand_tour.py --headed
```

---

## Known Limitations

### Dependency Issues in WebContainer
- `react-router-dom`: Multi-page routing fails
- `jspdf`: PDF generation unavailable
- `jspdf-autotable`: PDF table generation unavailable

### Timing Considerations
- WebContainer boot: ~30-90 seconds
- Code generation: varies by complexity (2s - 60s)
- Test execution: depends on test count

---

## Appendix: Workflow State Machine

```
                    ┌─────────────┐
                    │   INTAKE    │
                    └──────┬──────┘
                           │ User describes project
                           ▼
                    ┌─────────────┐
                    │   SIZING    │
                    └──────┬──────┘
                           │ Effort estimation
                           ▼
                    ┌─────────────┐
                    │  PLANNING   │
                    └──────┬──────┘
                           │ User approves spec
                           ▼
                    ┌─────────────┐
                    │ ARCHITECTURE│
                    └──────┬──────┘
                           │ Solution design
                           ▼
                    ┌─────────────┐
                    │ DEVELOPMENT │
                    └──────┬──────┘
                           │ Multi-agent coding
                           ▼
                    ┌─────────────┐
                    │   TESTING   │◄──────┐
                    └──────┬──────┘       │
                           │              │ 3-strike retry
                           ├──────────────┘
                           │ Tests pass
                           ▼
                    ┌─────────────┐
                    │ DEPLOYMENT  │
                    └──────┬──────┘
                           │ CI/CD setup
                           ▼
                    ┌─────────────┐
                    │ OPERATIONS  │
                    └──────┬──────┘
                           │ Monitoring
                           ▼
                    ┌─────────────┐
                    │   HANDOFF   │
                    └─────────────┘
```
