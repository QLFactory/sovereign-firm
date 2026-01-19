# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Sovereign Firm is an AI-powered software consultancy platform where AI agents act as digital employees. The platform dynamically spawns specialized agents (PM, Architect, Developer, QA, DevOps, SRE) to deliver complete software projects through Temporal workflows.

## Repository Structure

The project lives in `sovereign-firm/` and consists of:
- **Go Backend** (`cmd/`, `pkg/`): Orchestrator API server and Temporal worker
- **Next.js Frontend** (`frontend/`): React-based console UI
- **Skills** (`skills/`): YAML-defined agent capabilities (react-developer, go-developer, etc.)

## Development Commands

### Prerequisites
```bash
# Start infrastructure (Temporal, PostgreSQL, ChromaDB)
docker-compose up -d

# Ensure ollama is running locally for LLM (dev mode)
ollama serve
```

### Go Backend
```bash
# Run orchestrator (API server on :8080)
go run cmd/orchestrator/main.go

# Run worker (connects to Temporal on :7233)
go run cmd/worker/main.go

# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./pkg/agent/...

# Run single test
go test -run TestAgentPool ./pkg/agent/...

# Build binaries
go build -o orchestrator ./cmd/orchestrator
go build -o worker ./cmd/worker
```

### Frontend
```bash
cd frontend

# Development
npm run dev

# Tests
npm run test          # Run once
npm run test:watch    # Watch mode

# Lint
npm run lint

# Build
npm run build
```

### E2E Tests (Python/Playwright)
```bash
cd e2e
python -m venv venv
source venv/bin/activate
pip install pytest playwright
playwright install

# Run E2E tests
pytest test_frontend.py
pytest test_grand_tour.py
```

## Architecture

### Core Services
- **Orchestrator** (`cmd/orchestrator/`): Chi-based REST API exposing `/api/projects`, `/api/auth`, and WebSocket streaming at `/api/projects/{id}/stream`
- **Worker** (`cmd/worker/`): Temporal worker registering all activities (PM, Architect, DevAgent, QA, DevOps, SRE, etc.)

### Key Packages
- `pkg/sovereign/llm/`: LLM client abstraction (Ollama, Azure OpenAI)
- `pkg/agent/`: Agent pool, skill injection, SOC (Structured Output Contract)
- `pkg/firm/workflows/`: Temporal workflows (ConsultancyWorkflow, ProjectLifecycle)
- `pkg/firm/activities/`: Temporal activities for each agent role
- `pkg/streaming/`: WebSocket hub for real-time updates
- `pkg/treesitter/`: Language-agnostic code parsing
- `pkg/sandbox/`: Code execution isolation (container, Firecracker)

### Data Flow
1. Frontend creates project via POST `/api/projects`
2. Orchestrator starts Temporal workflow
3. Worker executes activities (PM chat → Architect design → Dev code → QA tests)
4. Real-time updates streamed via WebSocket
5. Artifacts stored in MinIO, state persisted in Postgres

### Environment Variables
Key variables (see `.env.example`):
- `TEMPORAL_HOST`: Temporal server address (default: localhost:7233)
- `DATABASE_URL`: PostgreSQL connection string
- `LLM_PROVIDER`: `ollama` or `azure`
- `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_API_KEY`: For Azure OpenAI
- `CHROMA_URL`: ChromaDB for vector storage (default: http://localhost:8000)

## Testing

### Go Unit Tests
Tests are co-located with source files (`*_test.go`). Key test files:
- `pkg/agent/agent_test.go`: Agent pool tests
- `pkg/firm/workflows/consultancy_test.go`: Workflow logic tests
- `pkg/treesitter/parser_test.go`: Code parsing tests

### Frontend Tests
Uses Vitest with React Testing Library. Config in `frontend/vitest.config.ts`.

### E2E Tests
Python tests in `e2e/` using Playwright for browser automation.

### BDD Tests
Cucumber-style tests in `tests/bdd/` with step definitions.

## Docker

```bash
# Full stack (backend + frontend + infrastructure)
docker-compose up -d

# Production build
docker-compose -f docker-compose.prod.yaml up -d
```

## Key APIs

### REST Endpoints
- `POST /api/auth/register`, `/api/auth/login`: Authentication
- `GET/POST /api/projects`: List/create projects
- `GET /api/projects/{id}/state`: Get workflow state
- `POST /api/projects/{id}/message`: Send message to workflow
- `POST /api/projects/import`: Brownfield project import

### WebSocket
- `GET /api/projects/{id}/stream`: Real-time workflow updates

### Legacy API (backward compatibility)
- `POST /api/pods`: Create workflow
- `GET /api/pods/{id}`: Get status
- `POST /api/pods/{id}/message`: Send message
