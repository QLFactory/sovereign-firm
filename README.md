# Sovereign AI Consultancy ("The Firm")

A self-contained, air-gapped capable platform where AI Agents act as digital employees.

## Architecture
This project follows a custom "Sovereign" architecture, avoiding heavy agent frameworks in favor of a clean, typed Go SDK.

### Directory Map

```text
/sovereign-firm
├── /cmd
│   ├── /orchestrator    # Main API Server - Runs on :8080
│   └── /worker          # Temporal Worker  - Connects to :7233
├── /pkg
│   ├── /sovereign       # CUSTOM SDK (The Brain)
│   ├── /firm            # BUSINESS LOGIC (The Office)
├── /frontend            # Next.js Console - Runs on :3000
```

## Getting Started

### 1. Boot the Infrastructure
We use Docker Compose to spin up Temporal, PostgreSQL, and ChromaDB.
```bash
docker-compose up -d
```
*Note: Ensure you have `ollama serve` running locally on your host machine.*

### 2. Start the Backend
```bash
# Terminal 1: The Agency API
go run cmd/orchestrator/main.go

# Terminal 2: The Agents (Worker)
go run cmd/worker/main.go
```

### 3. Start the Frontend
```bash
# Terminal 3: The Office
cd frontend
npm run dev
```

Visit `http://localhost:3000` to access the Sovereign Console.
