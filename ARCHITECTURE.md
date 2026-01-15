# Sovereign Firm v2.0 - Architecture Design Document (ADD)

**Version:** 2.0  
**Date:** 2026-01-15  
**Status:** Draft  
**Authors:** Technical Architecture Team

---

## Executive Summary

Sovereign Firm is an **Enterprise-Grade Agentic AI Consultancy Platform** that provides automated software development services through dynamically spawned, skill-equipped AI agents. Unlike simple code generators, Sovereign Firm operates as a full software consultancy with:

- **Dynamic Agent Teams** - Agents are spawned on-demand with specific skills for each project
- **Full SDLC Coverage** - From requirements through production deployment
- **Brownfield & Greenfield** - Works on new projects and existing codebases
- **Real-Time Collaboration** - Streaming output, checkpoints, human-in-the-loop
- **Enterprise Security** - Isolation, audit trails, cost controls
- **Universal Language Support** - Not just React—works with Go, Python, Rust, Java, and more
- **Tree-sitter Powered** - Language-agnostic code parsing and surgical edits
- **Multi-Layer Sandboxing** - Browser, container, and microVM isolation

---

## Key Differentiators

| Feature | Bolt.new / v0 | GitHub Copilot | Sovereign Firm |
|---------|---------------|----------------|----------------|
| Multi-agent | ❌ | ❌ | ✅ Dynamic teams |
| Brownfield | ❌ | Partial | ✅ Full repo analysis |
| Language Support | JS/TS only | Many | ✅ Universal (Tree-sitter) |
| Validation | Basic | None | ✅ Critic + Ensemble |
| Deployment | ❌ | ❌ | ✅ K8s, Vercel, etc. |
| Enterprise Security | ❌ | ❌ | ✅ Multi-layer sandbox |

---

## Industry Patterns & Standards (State of the Art - January 2026)

This section documents the agentic AI patterns, protocols, and frameworks available as of January 2026, and how Sovereign Firm aligns with or adopts them.

### Agent Design Pattern Evolution

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    AGENT PATTERN EVOLUTION                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  2023-2024: ReAct (Reasoning + Acting)                                     │
│     └─ Agent reasons → acts → observes → repeats                           │
│     └─ Simple loop, good for single tasks                                  │
│                                                                             │
│  2025: Plan-and-Execute ("2026 Gold Standard")                             │
│     └─ High-reasoning model creates comprehensive plan                     │
│     └─ Specialized "Sniper" models execute individual steps               │
│     └─ 5x faster execution, 3x cost reduction                              │
│     └─ ✅ WE USE THIS: Meta-Agent plans → Specialist agents execute       │
│                                                                             │
│  2025-2026: Multi-Agent Orchestration ("Microservices for AI")             │
│     └─ Teams of specialized agents, not one general-purpose agent         │
│     └─ DAG-based task dependencies for ordering                           │
│     └─ Bounded autonomy with human checkpoints                            │
│     └─ ✅ WE USE THIS: Role catalog, TDG, human-in-the-loop               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Standardized Protocols

| Protocol | Provider | Purpose | Our Status |
|----------|----------|---------|------------|
| **MCP (Model Context Protocol)** | Anthropic (Nov 2024) | LLM ↔ Tool integration | ✅ **Adopting** - Standard for tool calls |
| **A2A (Agent-to-Agent)** | Google (Apr 2025) | Agent ↔ Agent communication | 🔄 Potential - Agent Cards for discovery |

**MCP Integration:**
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    MCP (Model Context Protocol)                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Standard for connecting LLMs to external tools and data:                  │
│                                                                             │
│  ┌─────────────────┐      JSON-RPC 2.0      ┌─────────────────┐            │
│  │   MCP Client    │◄─────────────────────►│   MCP Server    │            │
│  │  (LLM Host)     │                        │  (Tool/Data)    │            │
│  └─────────────────┘                        └─────────────────┘            │
│                                                                             │
│  Benefits for Sovereign Firm:                                              │
│  - Standardized tool definitions (no custom integrations)                  │
│  - Ecosystem of pre-built MCP servers                                      │
│  - Interoperability with other AI platforms                                │
│  - Donated to Agentic AI Foundation (Dec 2025) - industry standard        │
│                                                                             │
│  We'll implement MCP servers for our tools:                                │
│  - file_read, file_write, npm_run, git commands                           │
│  - terraform_plan, kubectl_apply, docker_build                            │
│  - semantic_search (RAG), code_analyze (Tree-sitter)                      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Multi-Agent Framework Comparison

| Framework | Architecture | Strengths | Weaknesses | Our Approach |
|-----------|-------------|-----------|------------|--------------|
| **LangGraph** | Graph-based | Production-ready v1.0, stateful checkpoints | Steeper learning curve | Inspired our TDG design |
| **CrewAI** | Role-based | Intuitive roles, 5.76x faster | Overkill for simple tasks | Inspired our Role Catalog |
| **AutoGen** | Conversational | Native code execution | Verbose setup | Inspired human-in-the-loop |
| **Temporal** | Durable execution | Fault-tolerant, auto-retry, state persistence | Requires workflow thinking | ✅ **We use Temporal** |

**Why Temporal (not LangGraph/CrewAI directly):**
```
┌─────────────────────────────────────────────────────────────────────────────┐
│  TEMPORAL AS AGENT ORCHESTRATOR                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Used by: OpenAI Codex, Replit                                             │
│                                                                             │
│  Why Temporal for AI agents:                                               │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  1. Durable Execution                                               │    │
│  │     - Survives crashes, restarts, network failures                 │    │
│  │     - Workflow resumes exactly where it left off                   │    │
│  │                                                                     │    │
│  │  2. Automatic Retry                                                 │    │
│  │     - LLM calls fail? Retry with exponential backoff               │    │
│  │     - Tool execution times out? Retry automatically                │    │
│  │                                                                     │    │
│  │  3. State Persistence                                               │    │
│  │     - Agent context persisted for days/weeks                       │    │
│  │     - Perfect for long-running projects                            │    │
│  │                                                                     │    │
│  │  4. Human-in-the-Loop                                               │    │
│  │     - Workflow pauses via Signals                                  │    │
│  │     - Awaits human approval, then resumes                          │    │
│  │                                                                     │    │
│  │  5. Event Sourcing                                                  │    │
│  │     - Full history of every decision and LLM call                  │    │
│  │     - Replay for debugging, audit trails                           │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Competitive Landscape (Code Generation Agents)

| Product | Architecture | Key Innovation | Our Differentiation |
|---------|-------------|----------------|---------------------|
| **Cursor** | MoE + semantic indexing | Speculative edits, codebase embeddings | We're a full consultancy, not just an IDE |
| **Devin** | Autonomous AI engineer | Long-term planning, full dev environment | We have multi-agent teams with specialist roles |
| **Windsurf** | Cascade agent | Deep context, automatic memories | We have enterprise features (security, GitOps) |
| **GitHub Copilot** | Inline completion | Ubiquitous, IDE integration | We do full projects, not just snippets |

**Our Unique Position:**
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    SOVEREIGN FIRM DIFFERENTIATION                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Cursor/Windsurf = Developer TOOLS (augment individual developer)          │
│  Devin = Single AI ENGINEER (1 autonomous agent)                           │
│  Sovereign Firm = AI CONSULTANCY (full team of specialist agents)         │
│                                                                             │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  We provide what others don't:                                      │    │
│  │                                                                     │    │
│  │  ✅ Multi-agent teams (not just one agent)                         │    │
│  │  ✅ Specialist roles (Cloud Architect, SRE, Security)              │    │
│  │  ✅ Enterprise features (multi-tenancy, audit, compliance)         │    │
│  │  ✅ Brownfield support (work on existing codebases)                │    │
│  │  ✅ Full SDLC (requirements → deployment, not just code)           │    │
│  │  ✅ Human-in-the-loop (checkpoints, approvals)                     │    │
│  │  ✅ Cost controls (budget limits, model selection)                 │    │
│  │  ✅ GitOps integration (PRs, branch protection)                    │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Standards We Adopt

| Standard | What | Implementation |
|----------|------|----------------|
| **MCP** | Tool integration | MCP servers for all tools |
| **JSON Schema** | Output contracts | SOC with JSON Schema validation |
| **OpenTelemetry** | Observability | Traces for every agent action |
| **GitOps** | Deployment | ArgoCD/Flux integration |
| **SLSA** | Supply chain security | Signed artifacts, SBOM |

---

## 1. Vision & Objectives

### 1.1 Vision Statement
"To be the AI-powered software consultancy that enterprises trust for mission-critical development—operating with the reliability, security, and transparency of a top-tier human consultancy."

### 1.2 Key Objectives

| Objective | Success Metric |
|-----------|---------------|
| Dynamic Agent Teams | Spawn 1-10 agents per project based on complexity |
| Brownfield Support | Successfully modify 80%+ of imported codebases |
| Streaming Generation | User sees first output within 5 seconds |
| Production Deployment | One-click deploy to K8s/Vercel |
| Enterprise Security | SOC 2 compliant isolation and audit |
| Cost Efficiency | 50% cheaper than equivalent human consultancy |

---

## 2. Architecture Overview

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      WEB CONSOLE (Next.js)                          │   │
│  │  - Project Dashboard        - War Room View                         │   │
│  │  - Live Preview             - Agent Activity Feed                   │   │
│  │  - Approval Gates           - Cost Tracking                         │   │
│  │  - Repo Import              - Streaming Code View                   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────────────┤
│                            API GATEWAY LAYER                                │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     ORCHESTRATOR (Go)                                │   │
│  │  - REST API                 - WebSocket Streaming                   │   │
│  │  - Authentication           - Rate Limiting                         │   │
│  │  - Tenant Routing           - Request Validation                    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────────────┤
│                          ORCHESTRATION LAYER                                │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      TEMPORAL (Workflow Engine)                      │   │
│  │  - Project Lifecycle Workflow                                       │   │
│  │  - Agent Coordination Workflow                                      │   │
│  │  - Approval Gate Workflow                                           │   │
│  │  - Deployment Pipeline Workflow                                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────────────┤
│                             AGENT LAYER                                     │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      ENGAGEMENT MANAGER                              │   │
│  │  (Meta-Agent: Plans teams, routes work, monitors progress)          │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│  ┌─────────────────────────────────┴─────────────────────────────────┐     │
│  │                         AGENT POOL                                 │     │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │     │
│  │  │   Agent 1   │ │   Agent 2   │ │   Agent 3   │ │   Agent N   │ │     │
│  │  │  FE Dev     │ │  BE Dev     │ │  QA Eng     │ │  DevOps     │ │     │
│  │  │  [Skills]   │ │  [Skills]   │ │  [Skills]   │ │  [Skills]   │ │     │
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │     │
│  │                                                                    │     │
│  │  Lifecycle: SPAWN → ONBOARD → ACTIVE → OFFBOARD → TERMINATE      │     │
│  └────────────────────────────────────────────────────────────────────┘     │
├─────────────────────────────────────────────────────────────────────────────┤
│                           EXECUTION LAYER                                   │
│  ┌────────────────────────────┐    ┌────────────────────────────────┐      │
│  │      PREVIEW SANDBOX       │    │     PRODUCTION PIPELINE        │      │
│  │                            │    │                                │      │
│  │  WebContainers (Browser)   │    │  Docker/Kubernetes             │      │
│  │  - Instant preview         │    │  - Real npm test               │      │
│  │  - Hot reload              │◄──►│  - Real build                  │      │
│  │  - Interactive             │    │  - Security scanning           │      │
│  │                            │    │  - Deploy to cloud             │      │
│  └────────────────────────────┘    └────────────────────────────────┘      │
├─────────────────────────────────────────────────────────────────────────────┤
│                          INTEGRATION LAYER                                  │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐              │
│  │   GIT PROVIDER  │ │  DEPLOY TARGETS │ │  NOTIFICATIONS  │              │
│  │  GitHub/GitLab  │ │  Vercel/K8s/AWS │ │  Slack/Email    │              │
│  └─────────────────┘ └─────────────────┘ └─────────────────┘              │
├─────────────────────────────────────────────────────────────────────────────┤
│                          PLATFORM SERVICES                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │
│  │   LLM    │ │  Vector  │ │ Secrets  │ │ Observ.  │ │  Billing │        │
│  │  Router  │ │   DB     │ │  Vault   │ │  Stack   │ │  Meter   │        │
│  │(Multi)   │ │ (Chroma) │ │ (Vault)  │ │  (OTEL)  │ │          │        │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘        │
├─────────────────────────────────────────────────────────────────────────────┤
│                         DATA PERSISTENCE                                    │
│  ┌──────────────────────┐    ┌──────────────────────┐                      │
│  │     PostgreSQL       │    │      Redis           │                      │
│  │  - Temporal state    │    │  - Session cache     │                      │
│  │  - Project metadata  │    │  - Pub/Sub messages  │                      │
│  │  - Audit logs        │    │  - Rate limiting     │                      │
│  └──────────────────────┘    └──────────────────────┘                      │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Agent System Design

### 3.1 Agent Lifecycle

```
    ┌──────────────────────────────────────────────────────────────────┐
    │                      AGENT LIFECYCLE                             │
    └──────────────────────────────────────────────────────────────────┘
    
    ┌─────────┐    ┌───────────┐    ┌────────┐    ┌───────────┐    ┌────────────┐
    │  SPAWN  │───►│ ONBOARDING│───►│ ACTIVE │───►│OFFBOARDING│───►│ TERMINATED │
    └─────────┘    └───────────┘    └────────┘    └───────────┘    └────────────┘
         │              │               │               │                │
         │              │               │               │                │
    - Load role    - Inject skills  - Execute tasks  - Archive work    - Release
      config       - Set context    - Communicate    - Submit finals     resources
    - Create ID    - Build prompt   - Report status  - Handoff          - Log metrics
    - Register     - Validate       - Stream output    context          
```

### 3.2 Agent Instance Structure

```go
type AgentInstance struct {
    // Identity
    ID            string          `json:"id"`
    Name          string          `json:"name"`        // "Alice" - humanized name
    Role          string          `json:"role"`        // "senior-frontend-developer"
    
    // Configuration
    Skills        []Skill         `json:"skills"`      // Injected capabilities
    SystemPrompt  string          `json:"-"`           // Role + skills + context
    Model         string          `json:"model"`       // Which LLM to use
    
    // Project Context
    ProjectID     string          `json:"project_id"`
    Context       *AgentContext   `json:"context"`     // What this agent knows
    
    // Lifecycle
    Status        AgentStatus     `json:"status"`
    CreatedAt     time.Time       `json:"created_at"`
    TerminatedAt  *time.Time      `json:"terminated_at,omitempty"`
    
    // Work
    CurrentTask   *Task           `json:"current_task,omitempty"`
    Artifacts     []Artifact      `json:"artifacts"`   // Files produced
    
    // Communication
    Inbox         chan Message    `json:"-"`           // Receive messages
    Outbox        chan Message    `json:"-"`           // Send messages
    
    // Metrics
    TokensUsed    int64           `json:"tokens_used"`
    TasksComplete int             `json:"tasks_complete"`
}
```

### 3.3 Skill System

```yaml
# Example: skills/react-developer/skill.yaml
name: react-developer
version: 1.0.0
description: Build React components following best practices

capabilities:
  - functional-components
  - hooks
  - typescript
  - testing

tools:
  - file_write
  - file_read
  - npm_run
  - semantic_search

prompts:
  system: |
    You are a Senior React Developer with expertise in:
    - React 18+ with hooks and functional components
    - TypeScript for type safety
    - Modern CSS (Tailwind, CSS Modules, or styled-components)
    - Testing with React Testing Library
    
    You follow these conventions:
    - Use PascalCase for component names
    - Use camelCase for functions and variables
    - One component per file
    - Colocate tests with components
    
  examples:
    - description: "Create a button component"
      input: "Create a reusable button component with variants"
      output: |
        // Button.tsx
        interface ButtonProps {
          variant: 'primary' | 'secondary' | 'danger';
          children: React.ReactNode;
          onClick?: () => void;
        }
        
        export function Button({ variant, children, onClick }: ButtonProps) {
          const baseStyles = "px-4 py-2 rounded font-medium";
          const variantStyles = {
            primary: "bg-blue-500 text-white hover:bg-blue-600",
            secondary: "bg-gray-200 text-gray-800 hover:bg-gray-300",
            danger: "bg-red-500 text-white hover:bg-red-600",
          };
          
          return (
            <button 
              className={`${baseStyles} ${variantStyles[variant]}`}
              onClick={onClick}
            >
              {children}
            </button>
          );
        }
```

### 3.4 Inter-Agent Communication

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      AGENT MESSAGE BUS                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Message Types:                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │  QUESTION    - Agent asks another agent for information              │ │
│  │  ANSWER      - Response to a question                                 │ │
│  │  ARTIFACT    - Agent shares a deliverable (file, schema, etc.)       │ │
│  │  STATUS      - Agent updates its status                               │ │
│  │  REQUEST     - Agent asks another to do something                     │ │
│  │  REVIEW      - Agent asks for review of work                          │ │
│  │  BROADCAST   - Message to all agents                                  │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  Example Flow:                                                              │
│  ┌─────────────┐                              ┌─────────────┐              │
│  │  FE Agent   │                              │  BE Agent   │              │
│  └──────┬──────┘                              └──────┬──────┘              │
│         │                                            │                      │
│         │  QUESTION: "What's the user API schema?"   │                      │
│         │───────────────────────────────────────────►│                      │
│         │                                            │                      │
│         │  ANSWER: { id, name, email, avatar }       │                      │
│         │◄───────────────────────────────────────────│                      │
│         │                                            │                      │
│         │  ARTIFACT: UserCard.tsx (uses schema)      │                      │
│         │───────────────────────────────────────────►│                      │
│         │                                            │                      │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.5 Intelligence Layer (Vector DB, RAG, Embeddings)

The "brain" of the platform—how agents learn and remember:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      INTELLIGENCE LAYER                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  THREE LEVELS OF MEMORY:                                                    │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  LEVEL 1: GLOBAL KNOWLEDGE (Platform-wide)                          │   │
│  │                                                                      │   │
│  │  What: Best practices, design patterns, framework docs              │   │
│  │  Source: Open-source repos, official docs, Stack Overflow           │   │
│  │  Access: All agents on all projects                                 │   │
│  │  Examples: "React patterns", "Go idioms", "K8s configs"             │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  LEVEL 2: CLIENT KNOWLEDGE (Per-tenant)                             │   │
│  │                                                                      │   │
│  │  What: Client's codebase, conventions, past projects               │   │
│  │  Source: Imported repos, completed projects, feedback               │   │
│  │  Access: Only agents working on this client's projects              │   │
│  │  Examples: "How does Acme Corp structure their components?"         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  LEVEL 3: PROJECT CONTEXT (This task)                               │   │
│  │                                                                      │   │
│  │  What: Current project specs, decisions made, files generated      │   │
│  │  Source: This workflow execution                                    │   │
│  │  Access: Only agents working on this specific project               │   │
│  │  Examples: "What did the Architect decide?", "Files created?"       │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**RAG Pipeline:**

```
Agent Task: "Implement user authentication"
       │
       ▼
1. EMBED QUERY → "user authentication" → [0.23, -0.45, ...]
       │
       ▼
2. PARALLEL SEARCH across all levels:
   - Global:  "JWT auth Go pattern" (score: 0.92)
   - Client:  "Existing auth code" (score: 0.88)
   - Project: "Architect: use NextAuth" (score: 0.95)
       │
       ▼
3. RANK & SELECT (Top-K with diversity)
       │
       ▼
4. INJECT INTO AGENT PROMPT as context
       │
       ▼
5. GENERATE with context → Consistent, informed output
```

**Implementation:**

```go
type IntelligenceLayer struct {
    GlobalStore   *ChromaDB  // Platform-wide patterns
    ClientStores  map[string]*ChromaDB  // Per-tenant
    ProjectStore  *ChromaDB  // Current project
    Embedder      llm.Client
}

func (il *IntelligenceLayer) EnrichPrompt(agentID string, task string) (string, error) {
    // Embed the task
    embedding, _ := il.Embedder.Embed(task)
    
    // Search all levels in parallel
    var results []Document
    
    // Global patterns
    global, _ := il.GlobalStore.SimilaritySearch(embedding, 3)
    results = append(results, global...)
    
    // Client-specific patterns
    client, _ := il.ClientStores[agentID].SimilaritySearch(embedding, 3)
    results = append(results, client...)
    
    // Project context
    project, _ := il.ProjectStore.SimilaritySearch(embedding, 5)
    results = append(results, project...)
    
    // Rank by score and diversity
    selected := il.RankAndDedupe(results, 8)
    
    // Build context string
    return il.FormatAsContext(selected), nil
}
```

### 3.6 Multi-Agent Coordination

How multiple agents work without stepping on each other:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    MULTI-AGENT COORDINATION                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. TASK ASSIGNMENT (DAG-based)                                            │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  Meta-Agent creates Directed Acyclic Graph:                        │ │
│     │                                                                     │ │
│     │       [Database Schema]                                             │ │
│     │             │                                                       │ │
│     │    ┌───────┴───────┐                                               │ │
│     │    ▼               ▼                                               │ │
│     │ [API Routes]   [Type Defs]                                         │ │
│     │    │               │                                               │ │
│     │    └───────┬───────┘                                               │ │
│     │            ▼                                                       │ │
│     │      [Frontend]────► [Tests]                                       │ │
│     │                                                                     │ │
│     │  Dependencies respected: Frontend waits for API Routes             │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  2. FILE-LEVEL LOCKING                                                      │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  Before editing, agent must acquire lock:                          │ │
│     │                                                                     │ │
│     │  FE Agent: "I want to edit Header.tsx"                             │ │
│     │  Lock Manager: "Granted. Lock expires in 60s."                     │ │
│     │                                                                     │ │
│     │  BE Agent: "I want to edit Header.tsx"                             │ │
│     │  Lock Manager: "DENIED. FE Agent has lock."                        │ │
│     │                                                                     │ │
│     │  Lock types: EXCLUSIVE (write) | SHARED (read)                     │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  3. EACH AGENT = FEATURE BRANCH                                            │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  main                                                               │ │
│     │    │                                                                │ │
│     │    ├── fe-agent/header-component                                   │ │
│     │    ├── be-agent/user-api                                           │ │
│     │    └── qa-agent/user-tests                                         │ │
│     │                                                                     │ │
│     │  Each agent commits to their branch                                │ │
│     │  Meta-Agent merges when dependencies complete                      │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  4. SHARED WORKSPACE STATE                                                  │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  All agents see same virtual filesystem:                           │ │
│     │                                                                     │ │
│     │  Workspace {                                                        │ │
│     │    files: Map<path, content>                                       │ │
│     │    locks: Map<path, agentId>                                       │ │
│     │    branches: Map<agentId, branchName>                              │ │
│     │  }                                                                  │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.7 Agent CI Pipeline (Continuous Integration)

Every agent gets instant feedback like a human developer:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       AGENT CI PIPELINE                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Every agent commit triggers CI:                                           │
│                                                                             │
│  Agent saves file                                                          │
│       │                                                                     │
│       ▼                                                                     │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  STAGE 1: INSTANT CHECKS (< 5 seconds)                             │    │
│  │  ✓ Syntax valid (Tree-sitter parse)                                │    │
│  │  ✓ Types valid (TypeScript/Go compile)                             │    │
│  │  ✓ Lint pass (ESLint/golint)                                       │    │
│  │                                                                     │    │
│  │  If FAIL → Agent gets feedback → Self-correction                   │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│       │                                                                     │
│       ▼ (if passed)                                                        │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  STAGE 2: UNIT TESTS (< 30 seconds)                                │    │
│  │  Run affected tests only (test dependency graph)                   │    │
│  │                                                                     │    │
│  │  If FAIL → Agent sees error → Self-correction loop                 │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│       │                                                                     │
│       ▼ (if passed)                                                        │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  STAGE 3: INTEGRATION (< 2 minutes)                                │    │
│  │  - Build entire project                                            │    │
│  │  - Run full test suite                                             │    │
│  │  - Check for breaking changes                                      │    │
│  │                                                                     │    │
│  │  If FAIL → Notify agents → Rollback → Collaborate to fix          │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│       │                                                                     │
│       ▼ (if passed)                                                        │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  STAGE 4: MERGE TO MAIN                                            │    │
│  │  Auto-merge if: tests pass + no conflicts + critic score OK       │    │
│  │  Otherwise: Human review required                                  │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Self-Correction Loop:**

```go
func (agent *Agent) ExecuteWithFeedback(task Task) (*Result, error) {
    maxRetries := 3
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        // Generate code
        code := agent.Generate(task)
        
        // Commit to agent's branch
        agent.Workspace.Commit(code)
        
        // Trigger CI
        ciResult := agent.CI.Run()
        
        if ciResult.Passed {
            return &Result{Code: code, Attempt: attempt}, nil
        }
        
        // FEEDBACK LOOP: Agent sees what failed
        agent.Context.AddFeedback(Feedback{
            Attempt:    attempt,
            Error:      ciResult.Error,
            FailedTest: ciResult.FailedTests,
        })
        
        // Agent uses feedback to self-correct
    }
    
    return nil, ErrMaxRetriesExceeded  // Escalate to human
}
```

### 3.8 Embedded Git (Version Control)

Git is embedded from the start for every project—safety, rollback, and history:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       EMBEDDED GIT LAYER                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  WHY EMBEDDED GIT:                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  - Rollback to any previous state instantly                        │    │
│  │  - Compare what changed between versions                           │    │
│  │  - Multiple agents work on branches (isolation)                    │    │
│  │  - Full history for debugging and audit                            │    │
│  │  - Export to GitHub/GitLab when ready                              │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  REPO STRUCTURE (Per Project):                                             │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  main (protected, always working)                                  │    │
│  │    │                                                                │    │
│  │    ├── architect/v1                                                 │    │
│  │    │   └── "Initial architecture"                                  │    │
│  │    │                                                                │    │
│  │    ├── fe-agent/header                                             │    │
│  │    │   ├── "Created Header"                                        │    │
│  │    │   ├── "Fixed type error" (self-correction)                    │    │
│  │    │   └── "Added tests"                                           │    │
│  │    │                                                                │    │
│  │    ├── be-agent/api                                                │    │
│  │    │   └── "Implemented /api/users"                                │    │
│  │    │                                                                │    │
│  │    └── qa-agent/tests                                              │    │
│  │        └── "Integration tests"                                     │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  ROLLBACK OPTIONS:                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  1. Soft rollback  - Revert last commit, keep history             │    │
│  │  2. Hard rollback  - Reset to specific commit                     │    │
│  │  3. Branch reset   - Abandon agent's branch, restart              │    │
│  │  4. Full rollback  - Go back to checkpoint                        │    │
│  │                                                                     │    │
│  │  All rollbacks are versioned (can undo the undo!)                 │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  EXPORT TO EXTERNAL GIT:                                                    │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  When project ready:                                               │    │
│  │  1. Push to GitHub/GitLab                                          │    │
│  │  2. Create Pull Request                                            │    │
│  │  3. Full history preserved                                         │    │
│  │  4. External CI/CD triggers                                        │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.9 Structured Output Contracts (SOC)

LLMs are non-deterministic. Without strict output contracts, we get chaos:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                  STRUCTURED OUTPUT CONTRACTS (SOC)                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  PROBLEM: Without contracts, same prompt yields different formats          │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  Request: "Generate a React component"                             │    │
│  │                                                                     │    │
│  │  Response 1: { "App.js": "..." }                                   │    │
│  │  Response 2: { "files": [{"name": "App.js", ...}] }               │    │
│  │  Response 3: ```jsx\nexport function App()...```                  │    │
│  │                                                                     │    │
│  │  Result: Parser breaks. Pipeline fails. CHAOS.                     │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  SOLUTION: JSON Schema enforcement at EVERY LLM call                       │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  1. DEFINE CONTRACT (JSON Schema)                                  │    │
│  │  2. INJECT SCHEMA INTO PROMPT                                      │    │
│  │  3. VALIDATE OUTPUT AGAINST SCHEMA                                 │    │
│  │  4. RETRY WITH ERROR FEEDBACK (if invalid)                         │    │
│  │  5. ESCALATE after max retries                                     │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Contract Catalog:**

| Contract | Used By | Schema |
|----------|---------|--------|
| `CodeBundle` | DevAgent.GenerateCode | `{ "/src/file.tsx": "content" }` |
| `TestBundle` | QAAgent.GenerateTests | `{ "/src/file.test.tsx": "content" }` |
| `ArchitectureDecision` | ArchitectAgent.Design | `{ summary, stack, tasks[] }` |
| `CriticReview` | CriticAgent.Review | `{ verdict, score, issues[] }` |
| `ChatResponse` | PMAgent.Chat | `{ message, action, spec? }` |
| `TargetedEdit` | DevAgent.Edit | `{ file, startLine, endLine, content }` |

**Example Contract (JSON Schema):**

```json
{
  "name": "CodeBundle",
  "description": "Map of filename to code content",
  "schema": {
    "type": "object",
    "additionalProperties": { "type": "string" },
    "propertyNames": {
      "pattern": "^/src/.*\\.(js|jsx|ts|tsx|css|json)$"
    }
  }
}
```

**Contract Enforcement in Go:**

```go
type StructuredExecutor struct {
    LLM        llm.Client
    Contract   *Contract
    MaxRetries int
}

func (e *StructuredExecutor) Execute(ctx context.Context, prompt string) (json.RawMessage, error) {
    fullPrompt := prompt + e.Contract.GeneratePromptSuffix()
    
    for attempt := 1; attempt <= e.MaxRetries; attempt++ {
        response, _ := e.LLM.Generate(ctx, fullPrompt)
        cleaned := cleanJSONResponse(response)
        
        result, _ := e.Contract.Validate(cleaned)
        if result.Valid {
            return json.RawMessage(cleaned), nil  // SUCCESS
        }
        
        // RETRY: Add validation errors to prompt
        fullPrompt = fmt.Sprintf("%s\n\nPREVIOUS FAILED: %s\nFix and retry.",
            fullPrompt, result.Errors)
    }
    
    return nil, ErrMaxRetriesExceeded
}
```

**Additional Guardrails:**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    DETERMINISM GUARDRAILS                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. TEMPERATURE CONTROL                                                     │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  Code generation     │ 0.2   │ Deterministic                       │ │
│     │  Architecture        │ 0.5   │ Creative but structured             │ │
│     │  Chat                │ 0.8   │ Natural                             │ │
│     │  Review/Critic       │ 0.1   │ Objective                           │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  2. RESPONSE FORMAT (OpenAI/Azure API)                                      │
│     response_format: { "type": "json_object" }                             │
│                                                                             │
│  3. FEW-SHOT EXAMPLES                                                       │
│     Include 2-3 examples showing exact expected format                     │
│                                                                             │
│  4. OUTPUT SANITIZATION                                                     │
│     - Strip markdown code blocks                                           │
│     - Remove leading/trailing prose                                        │
│     - Fix common JSON errors (trailing commas)                             │
│                                                                             │
│  5. CONTRACT VERSIONING                                                     │
│     contracts/v1/code_bundle.json                                          │
│     contracts/v2/code_bundle.json (breaking change)                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.10 Task Dependency Graph (TDG)

Tasks have natural dependencies—you can't build a component that uses a type before the type exists:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    TASK DEPENDENCY GRAPH (TDG)                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  WHY: Ensure meaningful code generation                                    │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  ❌ Without DAG:                                                    │    │
│  │     - Frontend uses User type that doesn't exist                   │    │
│  │     - API calls database function not yet written                  │    │
│  │     - Tests import components that haven't been generated          │    │
│  │                                                                     │    │
│  │  ✅ With DAG:                                                       │    │
│  │     - Types/interfaces generated first                              │    │
│  │     - Implementations use those types                              │    │
│  │     - Consumers wait for what they consume                          │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  DEPENDENCY LAYERS (Bottom-up):                                            │
│                                                                             │
│  LAYER 0: Foundation ───────────────────────────────────────────────────   │
│     [package.json] [tsconfig] [types/*.ts] [prisma/schema]                 │
│                               │                                             │
│                               ▼                                             │
│  LAYER 1: Services ─────────────────────────────────────────────────────   │
│     [lib/db.ts] [lib/auth.ts] [repositories/*.ts]                          │
│                               │                                             │
│                               ▼                                             │
│  LAYER 2: API ──────────────────────────────────────────────────────────   │
│     [api/tasks/route.ts] [api/users/route.ts] [api/auth/route.ts]         │
│                               │                                             │
│                               ▼                                             │
│  LAYER 3: Components ───────────────────────────────────────────────────   │
│     [TaskCard.tsx] [TaskList.tsx] [UserAvatar.tsx]                        │
│                               │                                             │
│                               ▼                                             │
│  LAYER 4: Pages ────────────────────────────────────────────────────────   │
│     [dashboard/page.tsx] [tasks/page.tsx]                                  │
│                               │                                             │
│                               ▼                                             │
│  LAYER 5: Tests ────────────────────────────────────────────────────────   │
│     [*.test.tsx] [*.test.ts]                                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Parallel Execution Within Layers:**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Time ──────────────────────────────────────────────────────►              │
│                                                                             │
│  T1: [package.json] [tsconfig] [types/Task.ts] [types/User.ts] ← PARALLEL │
│       └──────────────────────────┬───────────────────────────┘              │
│                                  ▼                                          │
│  T2: [prisma/schema] [lib/db.ts] [lib/auth.ts]                 ← PARALLEL │
│       └──────────────────────────┬───────────────────────────┘              │
│                                  ▼                                          │
│  T3: [api/tasks] [api/users]                                   ← PARALLEL │
│       └──────────────────────────┬───────────────────────────┘              │
│                                  ▼                                          │
│  T4: [TaskCard] [UserCard]                                     ← PARALLEL │
│       └──────────────────────────┬───────────────────────────┘              │
│                                  ▼                                          │
│  T5: [TaskList] (depends on TaskCard)                          ← SEQUENTIAL│
│                                                                             │
│  Result: Maximum parallelism while respecting dependencies                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Task Definition:**

```go
type Task struct {
    ID           string        `json:"id"`
    Name         string        `json:"name"`
    Type         TaskType      `json:"type"`          // config, schema, service, api, component, page, test
    
    // Dependencies
    DependsOn    []string      `json:"depends_on"`    // Task IDs that must complete first
    Produces     []string      `json:"produces"`      // Files this task creates
    Requires     []string      `json:"requires"`      // Files this task needs to exist
    
    // Execution
    Status       TaskStatus    `json:"status"`        // pending, ready, running, complete, failed
    AssignedTo   string        `json:"assigned_to"`   // Agent ID
}

type TaskStatus string
const (
    TaskStatusPending   TaskStatus = "pending"   // Waiting for dependencies
    TaskStatusReady     TaskStatus = "ready"     // All deps complete, can start
    TaskStatusRunning   TaskStatus = "running"   // Agent working
    TaskStatusComplete  TaskStatus = "complete"  // Done
    TaskStatusBlocked   TaskStatus = "blocked"   // Dependency failed
)
```

**DAG Scheduler:**

```go
type DAGScheduler struct {
    Tasks     map[string]*Task
    Completed map[string]bool
}

// GetReadyTasks returns all tasks whose dependencies are satisfied
func (s *DAGScheduler) GetReadyTasks() []*Task {
    var ready []*Task
    
    for _, task := range s.Tasks {
        if task.Status != TaskStatusPending {
            continue
        }
        
        allDepsComplete := true
        for _, depID := range task.DependsOn {
            if !s.Completed[depID] {
                allDepsComplete = false
                break
            }
        }
        
        if allDepsComplete {
            task.Status = TaskStatusReady
            ready = append(ready, task)
        }
    }
    
    return ready  // All returned tasks can run in PARALLEL
}
```

**Artifact Context Injection:**

When an agent starts a task, it receives the outputs of its dependencies:

```go
// Before TaskList agent starts:
context := AgentContext{
    Task: "Create TaskList component",
    AvailableArtifacts: map[string]string{
        "src/types/task.ts":           "export interface Task { ... }",
        "src/components/TaskCard.tsx": "export function TaskCard({ task }: { task: Task }) { ... }",
    },
}

// Agent knows EXACTLY what exists - no guessing, no hallucinating imports
```

**Example DAG (YAML):**

```yaml
tasks:
  - id: type-task
    name: Define Task interface
    type: schema
    depends_on: []
    produces: [src/types/task.ts]
    
  - id: type-user
    name: Define User interface  
    type: schema
    depends_on: []
    produces: [src/types/user.ts]
    
  - id: api-tasks
    name: Task API routes
    type: api
    depends_on: [type-task]          # Must wait for Task type
    requires: [src/types/task.ts]
    produces: [src/app/api/tasks/route.ts]
    
  - id: comp-task-card
    name: TaskCard component
    type: component
    depends_on: [type-task]          # Must wait for Task type
    requires: [src/types/task.ts]
    produces: [src/components/TaskCard.tsx]
    
  - id: comp-task-list
    name: TaskList component
    type: component
    depends_on: [comp-task-card, api-tasks]  # Must wait for BOTH
    requires: [src/components/TaskCard.tsx, src/types/task.ts]
    produces: [src/components/TaskList.tsx]
    
  - id: test-task-card
    name: TaskCard tests
    type: test
    depends_on: [comp-task-card]     # Must wait for component
    requires: [src/components/TaskCard.tsx]
    produces: [src/components/TaskCard.test.tsx]
```

### 3.11 Agent Role Catalog

A comprehensive roster of specialist agents, organized by domain:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         AGENT ROLE CATALOG                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  MANAGEMENT & PLANNING                                                      │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  🎯 Engagement Manager   - Team planning, DAG creation, progress   │    │
│  │  📋 Product Manager      - Requirements, user stories, priorities  │    │
│  │  🏛️ Solution Architect   - System design, tech stack, API contracts│    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  DEVELOPMENT                                                                │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  ⚛️ Frontend Developer   - React/Vue, CSS, state management       │    │
│  │  🔧 Backend Developer    - API routes, business logic, auth       │    │
│  │  💾 Database Engineer    - Schema design, migrations, queries     │    │
│  │  📱 Mobile Developer     - React Native, Flutter, native          │    │
│  │  🔌 API Developer        - REST/GraphQL, OpenAPI, SDKs            │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  INFRASTRUCTURE & CLOUD                                                     │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  ☁️ Cloud Architect      - AWS/GCP/Azure, multi-region, cost      │    │
│  │  🚀 DevOps Engineer      - CI/CD, GitHub Actions, Docker          │    │
│  │  ☸️ Kubernetes Engineer   - K8s manifests, Helm, service mesh     │    │
│  │  🏗️ Infrastructure Eng   - Terraform, Pulumi, CloudFormation     │    │
│  │  🔔 SRE                   - Monitoring, alerting, SLO/SLA         │    │
│  │  🌐 Platform Engineer    - Developer platforms, self-service      │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  QUALITY & SECURITY                                                         │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  🧪 QA Engineer          - Unit/integration/E2E tests             │    │
│  │  🔒 Security Engineer    - Threat modeling, SAST/DAST, compliance │    │
│  │  ⚡ Performance Engineer - Load testing, profiling, caching       │    │
│  │  📊 Code Critic          - Code review, best practices, scoring   │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  SPECIALISTS                                                                │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  🤖 ML/AI Engineer       - Model integration, RAG, embeddings     │    │
│  │  📊 Data Engineer        - ETL, warehousing, analytics            │    │
│  │  🔄 Integration Spec     - Third-party APIs, webhooks, OAuth      │    │
│  │  📝 Technical Writer     - API docs, README, architecture docs    │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Team Composition by Project Type:**

| Project Type | Team Size | Key Roles |
|--------------|-----------|-----------|
| **Simple Web App** | 3-4 | Engagement Mgr, Frontend, Backend, QA |
| **Standard SaaS** | 6-8 | + Architect, Database, DevOps, Security |
| **Enterprise Platform** | 10-15 | + Cloud Arch, K8s, Infra, SRE, Tech Writer |
| **AI/ML Project** | 5-7 | + ML/AI Engineer, Data Engineer |

**Example: Cloud Architect Skill**

```yaml
# skills/cloud-architect/skill.yaml
name: cloud-architect
version: 1.0.0
description: Design cloud infrastructure

capabilities:
  - aws
  - gcp
  - azure
  - multi-region
  - cost-optimization
  - compliance

tools:
  - terraform_plan
  - cost_calculator
  - architecture_diagram
  - compliance_check

prompts:
  system: |
    You are a Senior Cloud Architect with expertise in:
    - AWS, GCP, and Azure cloud platforms
    - Multi-region and disaster recovery
    - Cost optimization and FinOps
    - Compliance (SOC2, HIPAA, GDPR)
    
    Design scalable, secure, cost-effective architectures.
```

**Example: SRE Skill**

```yaml
# skills/sre/skill.yaml
name: sre
version: 1.0.0
description: Site Reliability Engineering

capabilities:
  - monitoring
  - alerting
  - slo-sla
  - incident-response
  - chaos-engineering

tools:
  - prometheus_config
  - grafana_dashboard
  - alertmanager_rules
  - runbook_generator

prompts:
  system: |
    You are a Senior SRE with expertise in:
    - Observability (metrics, logs, traces)
    - Prometheus, Grafana, AlertManager
    - SLO/SLA definition and tracking
    - Incident response and runbooks
    
    Ensure systems are reliable, observable, recoverable.
```

**Example: Kubernetes Engineer Skill**

```yaml
# skills/kubernetes-engineer/skill.yaml
name: kubernetes-engineer
version: 1.0.0
description: Kubernetes deployments

capabilities:
  - k8s-manifests
  - helm-charts
  - ingress
  - service-mesh
  - autoscaling

tools:
  - kubectl_apply
  - helm_install
  - manifest_validate
  - security_scan

prompts:
  system: |
    You are a Senior Kubernetes Engineer with expertise in:
    - K8s Deployments, Services, ConfigMaps
    - Helm charts
    - Ingress (nginx, traefik)
    - Service mesh (Istio)
    
    Generate production-ready K8s configs with:
    - Resource limits/requests
    - Health checks
    - Security contexts
```

---

## 4. Project Modes

### 4.1 Greenfield Mode (New Projects)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         GREENFIELD WORKFLOW                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. DISCOVERY                                                               │
│     User describes what they want                                          │
│     PM Agent clarifies requirements                                        │
│     Output: Project Specification                                          │
│                                                                             │
│  2. PLANNING                                                                │
│     Architect Agent designs system                                         │
│     Meta-Agent assembles team                                              │
│     Output: Architecture, Team composition, Task breakdown                 │
│                                                                             │
│  3. GENERATION (Chunked/Streaming)                                         │
│     └─ Step 1: Project structure      → User sees skeleton                │
│     └─ Step 2: Core components        → Preview updates                   │
│     └─ [CHECKPOINT: Review structure]                                      │
│     └─ Step 3: Feature implementation → Preview updates iteratively       │
│     └─ Step 4: Tests                  → Tests appear and run              │
│     └─ [CHECKPOINT: Review implementation]                                 │
│                                                                             │
│  4. VERIFICATION                                                            │
│     All tests pass                                                         │
│     Security scan clean                                                    │
│     Build succeeds                                                         │
│                                                                             │
│  5. DEPLOYMENT                                                              │
│     Create Git repo with all code                                          │
│     Open PR for review                                                     │
│     Deploy to staging                                                      │
│     [HUMAN APPROVAL]                                                       │
│     Deploy to production                                                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.2 Brownfield Mode (Existing Projects)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         BROWNFIELD WORKFLOW                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. INGESTION                                                               │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  Input: GitHub URL / ZIP upload / Git clone URL                    │ │
│     │                                                                     │ │
│     │  Actions:                                                           │ │
│     │  - Clone repository                                                 │ │
│     │  - Parse all files                                                  │ │
│     │  - Detect: framework, language, patterns, dependencies              │ │
│     │  - Index into ChromaDB for semantic search                          │ │
│     │  - Generate architecture diagram                                    │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  2. ANALYSIS                                                                │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  Architect Agent produces:                                          │ │
│     │                                                                     │ │
│     │  📊 Codebase Report:                                                │ │
│     │     - Framework: Next.js 14 (App Router)                           │ │
│     │     - Language: TypeScript 5.2                                     │ │
│     │     - Styling: Tailwind CSS                                        │ │
│     │     - State: Zustand + React Query                                 │ │
│     │     - Database: Prisma + PostgreSQL                                │ │
│     │     - Auth: NextAuth.js                                            │ │
│     │                                                                     │ │
│     │  🎨 Conventions:                                                    │ │
│     │     - Component naming: PascalCase                                 │ │
│     │     - File structure: Feature-based folders                        │ │
│     │     - Testing: Vitest + Testing Library                            │ │
│     │                                                                     │ │
│     │  ⚠️ Technical Debt:                                                 │ │
│     │     - 3 deprecated dependencies                                    │ │
│     │     - Low test coverage (42%)                                      │ │
│     │     - Inconsistent error handling                                  │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  3. TASK UNDERSTANDING                                                      │
│     User: "Add a shopping cart feature"                                    │
│                                                                             │
│     Agents identify:                                                       │
│     - Where cart UI belongs (follows existing patterns)                   │
│     - Where cart API belongs (follows existing API structure)             │
│     - Database schema changes needed                                       │
│     - Which existing code to modify vs create new                         │
│                                                                             │
│  4. TARGETED MODIFICATION                                                   │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  Output: Git diff (not full replacement)                           │ │
│     │                                                                     │ │
│     │  + src/features/cart/CartDrawer.tsx      (new)                     │ │
│     │  + src/features/cart/CartItem.tsx        (new)                     │ │
│     │  + src/features/cart/useCart.ts          (new)                     │ │
│     │  M src/app/layout.tsx                    (modified: add provider)  │ │
│     │  + src/app/api/cart/route.ts             (new)                     │ │
│     │  M prisma/schema.prisma                  (modified: add Cart)      │ │
│     │  + src/features/cart/__tests__/Cart.test.tsx (new)                 │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  5. VERIFICATION                                                            │
│     - Existing tests still pass                                            │
│     - New tests pass                                                       │
│     - No breaking changes detected                                         │
│                                                                             │
│  6. INTEGRATION                                                             │
│     - Create feature branch                                                │
│     - Commit changes                                                       │
│     - Open PR with detailed description                                    │
│     - Request human review                                                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 5. Streaming/Chunked Generation

### 5.1 Generation Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      STREAMING GENERATION                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Traditional (Fire & Forget):                                              │
│  ┌───────────┐                                          ┌───────────┐      │
│  │  Request  │ ─────────[60 seconds]──────────────────► │  Result   │      │
│  └───────────┘                                          └───────────┘      │
│                      User waits blindly                                     │
│                                                                             │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│  Streaming (Real-time Feedback):                                           │
│                                                                             │
│  Request                                                                    │
│     │                                                                       │
│     ▼                                                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Step 1: Generate package.json                                      │   │
│  │          ──► Stream to UI (500ms)                                   │   │
│  │          ──► User sees: "📦 Setting up project..."                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│     │                                                                       │
│     ▼                                                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Step 2: Generate App.tsx (line by line)                            │   │
│  │          ──► Stream each chunk to UI                                │   │
│  │          ──► User sees code appearing                               │   │
│  │          ──► Preview hot-reloads                                    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│     │                                                                       │
│     ▼                                                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  [CHECKPOINT]                                                        │   │
│  │  "Here's the basic structure. Continue?"                            │   │
│  │  [Continue] [Modify] [Stop]                                         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│     │                                                                       │
│     ▼ (User clicks Continue)                                               │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Step 3: Generate components...                                      │   │
│  │          ──► Continue streaming                                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.2 Streaming Protocol

```typescript
// WebSocket message types for streaming

interface StreamMessage {
  type: 'file_start' | 'file_chunk' | 'file_complete' | 
        'checkpoint' | 'agent_status' | 'error';
  payload: any;
  timestamp: number;
}

// File content streaming
interface FileStartMessage {
  type: 'file_start';
  payload: {
    path: string;           // "/src/components/Button.tsx"
    agentId: string;        // Which agent is generating
    description: string;    // "Creating button component"
  };
}

interface FileChunkMessage {
  type: 'file_chunk';
  payload: {
    path: string;
    content: string;        // Chunk of code
    isComplete: boolean;
  };
}

// Checkpoint for user intervention
interface CheckpointMessage {
  type: 'checkpoint';
  payload: {
    id: string;
    title: string;
    description: string;
    options: string[];      // ["Continue", "Modify", "Stop"]
    filesComplete: string[];
    preview: string;        // Screenshot URL
  };
}
```

---

## 6. Verification & Critic System

### 6.1 Validation Philosophy

Nothing is "done" without passing multiple validation layers. We use a defense-in-depth approach:

1. **Automated Checks** - Fast, zero cost, catches obvious issues
2. **Critic Agents** - LLM-based review, catches quality issues
3. **Ensemble Validation** - Multiple agents for critical decisions
4. **Self-Correction** - Automatic retry on failure
5. **Human Gate** - Final approval for production

### 6.2 Validation Pipeline

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    VALIDATION LAYERS                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  LAYER 1: AUTOMATED CHECKS (Fast, No LLM Cost)                             │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  ✓ Syntax validation    - Does it parse?                          │    │
│  │  ✓ Type checking        - TypeScript/Go compile?                  │    │
│  │  ✓ Lint pass            - ESLint/golint clean?                    │    │
│  │  ✓ Tests pass           - npm test / go test                      │    │
│  │  ✓ Build succeeds       - npm run build                           │    │
│  │  ✓ Security scan        - No vulnerabilities?                     │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  LAYER 2: CRITIC AGENT (LLM-based Review)                                  │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  Specialized agent that ONLY reviews, never generates:            │    │
│  │  - Code quality assessment                                        │    │
│  │  - Requirements match verification                                │    │
│  │  - Architecture review                                            │    │
│  │  - Security analysis                                              │    │
│  │  Output: PASS/FAIL with score and detailed issues                 │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  LAYER 3: ENSEMBLE VALIDATION (For Critical Decisions)                     │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  Multiple agents independently solve the same problem:            │    │
│  │  Agent A (GPT-4) ─┐                                               │    │
│  │  Agent B (Claude) ─┼──► Arbiter Agent compares and picks best     │    │
│  │  Agent C (GPT-4)  ─┘    or synthesizes hybrid solution            │    │
│  │                                                                    │    │
│  │  Used for: Architecture, security-critical, complex algorithms    │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  LAYER 4: SELF-CORRECTION LOOP                                             │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  Generate ──► Validate ──► Pass? ──Yes──► Done                    │    │
│  │     ▲                        │ No                                  │    │
│  │     │                        ▼                                     │    │
│  │     │                   Retries < 3? ──No──► Escalate to Human    │    │
│  │     │                        │ Yes                                 │    │
│  │     └────────────────────────┘                                     │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 6.3 Critic Agent Definition

```yaml
# skills/code-critic/skill.yaml
name: code-critic
version: 1.0.0
description: Reviews and critiques generated code

personality:
  role: "Senior Code Reviewer"
  temperament: "Critical but constructive"

rules:
  - NEVER generate or fix code, only critique
  - Always provide specific line references
  - Rate severity: BLOCKER | MAJOR | MINOR | NITPICK
  - Must pass/fail with clear reasoning

output_format:
  verdict: "PASS | FAIL"
  score: "0-100"
  issues:
    - severity: "BLOCKER|MAJOR|MINOR|NITPICK"
      file: "path/to/file"
      line: 42
      issue: "Description"
      suggestion: "How to fix"
  praise: ["What was done well"]
  summary: "Overall assessment"
```

### 6.4 Ensemble Pattern

```go
// Multiple agents solve independently, arbiter picks best
type EnsembleConfig struct {
    Problem        string
    Agents         []AgentConfig  // Different models/temperatures
    Arbiter        AgentConfig    // Who decides
    RequiredAgree  int            // Quorum
}

// Used for:
// - Architecture decisions
// - Security-critical code
// - Complex algorithms
// - When confidence is low
```

### 6.5 Validation Modes by Complexity

| Task Complexity | Validation Mode | Time | Cost |
|-----------------|-----------------|------|------|
| Simple (1 file) | Automated only | 10s | $0 |
| Standard (feature) | Automated + Critic | 30s | $0.05 |
| Complex (module) | Automated + Critic + Human | 60s | $0.10 |
| Critical (security) | Ensemble + All above | 90s | $0.25 |

---

## 7. Security Model


### 6.1 Isolation Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SECURITY LAYERS                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  TENANT ISOLATION:                                                          │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Each client gets:                                                   │   │
│  │  - Separate Kubernetes namespace                                    │   │
│  │  - Dedicated database schema                                        │   │
│  │  - Isolated ChromaDB collection                                     │   │
│  │  - Network policies preventing cross-tenant access                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  EXECUTION ISOLATION:                                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Each project execution runs in:                                     │   │
│  │  - Ephemeral container (destroyed after use)                        │   │
│  │  - No network access by default (allowlist for npm registry)        │   │
│  │  - Read-only filesystem except workspace                            │   │
│  │  - Resource limits (CPU, memory, time)                              │   │
│  │  - No access to host system                                         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  SECRET MANAGEMENT:                                                         │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  - All secrets stored in HashiCorp Vault                            │   │
│  │  - Agents never see raw secrets (injected at runtime)               │   │
│  │  - Audit log for all secret access                                  │   │
│  │  - Automatic rotation                                               │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  CODE SCANNING:                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Before deployment:                                                  │   │
│  │  - SAST (Static Application Security Testing)                       │   │
│  │  - Dependency vulnerability scan                                    │   │
│  │  - Secret detection (prevent accidental commits)                    │   │
│  │  - License compliance check                                         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 7.2 Content Safety & HAP Filtering

HAP (Hate speech, Abuse, Profanity) filtering is critical for enterprise AI systems:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    HAP FILTERING ARCHITECTURE                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  WHAT IS HAP?                                                               │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  H - Hate speech: Expressions targeting groups (race, religion)   │    │
│  │  A - Abusive language: Bullying, demeaning, hurtful content       │    │
│  │  P - Profanity: Expletives, insults, explicit language            │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  WHY WE NEED IT:                                                           │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  1. Prompt injection with offensive content                        │    │
│  │  2. LLMs generating inappropriate comments in code                 │    │
│  │  3. Brownfield codebases with existing offensive content          │    │
│  │  4. Multi-agent communication amplifying harmful content          │    │
│  │  5. Enterprise compliance requirements                             │    │
│  │  6. Regulated industry standards (finance, healthcare)            │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  HAP FILTERING PIPELINE:                                                   │
│                                                                             │
│  ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐              │
│  │  User   │────►│ HAP     │────►│  LLM    │────►│ HAP     │────► Output │
│  │  Input  │     │ Filter  │     │         │     │ Filter  │              │
│  └─────────┘     │ (IN)    │     └─────────┘     │ (OUT)   │              │
│                  └─────────┘                     └─────────┘              │
│                       │                               │                    │
│                       ▼                               ▼                    │
│          Block/Flag/Redact input          Replace with safe message      │
│                                                                             │
│  IMPLEMENTATION:                                                           │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  1. Classification Model                                           │    │
│  │     - Use IBM granite-guardian or similar HAP detector            │    │
│  │     - Score 0-1 for each sentence (1 = likely HAP)                │    │
│  │     - Configurable threshold per tenant (default 0.5)             │    │
│  │                                                                     │    │
│  │  2. Action Based on Score                                          │    │
│  │     - Score < 0.3: Pass through                                    │    │
│  │     - Score 0.3-0.7: Flag for review, continue                    │    │
│  │     - Score > 0.7: Block/Replace/Escalate                         │    │
│  │                                                                     │    │
│  │  3. Offensive Span Identification                                  │    │
│  │     - Don't just flag, identify EXACTLY which words are HAP       │    │
│  │     - Surgical redaction instead of full-message blocking         │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  ADDITIONAL CONTENT GUARDRAILS:                                            │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  - PII Detection: Block/redact personal identifiable info         │    │
│  │  - Secret Detection: Prevent API keys in generated code           │    │
│  │  - Copyright Detection: Flag potentially copyrighted content      │    │
│  │  - Bias Detection: Identify potentially biased outputs            │    │
│  │  - Topic Restriction: Block specific topics per tenant config     │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Integration with SOC (Structured Output Contracts):**

```go
type ContentSafetyMiddleware struct {
    HAPFilter     *HAPClassifier
    PIIDetector   *PIIDetector
    SecretScanner *SecretScanner
    Threshold     float64
}

func (m *ContentSafetyMiddleware) FilterInput(ctx context.Context, input string) (string, error) {
    // 1. HAP Filter
    hapScore, spans := m.HAPFilter.Analyze(input)
    if hapScore > m.Threshold {
        return "", ErrHAPContentDetected{Spans: spans}
    }
    
    // 2. PII Redaction
    input = m.PIIDetector.Redact(input)
    
    return input, nil
}

func (m *ContentSafetyMiddleware) FilterOutput(ctx context.Context, output string) (string, error) {
    // 1. HAP Filter
    hapScore, spans := m.HAPFilter.Analyze(output)
    if hapScore > m.Threshold {
        output = m.HAPFilter.RedactSpans(output, spans)
        // Or replace entirely:
        // return "I cannot generate content with harmful language.", nil
    }
    
    // 2. Secret Detection
    if secrets := m.SecretScanner.Scan(output); len(secrets) > 0 {
        output = m.SecretScanner.Redact(output, secrets)
        log.Warn("Secrets detected and redacted from output")
    }
    
    return output, nil
}
```

---

## 8. Cost Management

### 7.1 Token Economics

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         COST CONTROL                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  MODEL ROUTING:                                                             │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │                                                                     │    │
│  │  Task Type           → Model         → Cost/1K tokens              │    │
│  │  ─────────────────────────────────────────────────────────────      │    │
│  │  Chat/Clarification  → GPT-3.5       → $0.002                      │    │
│  │  Simple code         → Claude Haiku  → $0.00025                    │    │
│  │  Complex code        → GPT-4         → $0.06                       │    │
│  │  Architecture        → Claude Opus   → $0.015                      │    │
│  │  Code review         → o1            → $0.015                      │    │
│  │  Embeddings          → ada-002       → $0.0001                     │    │
│  │                                                                     │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  BUDGET CONTROLS:                                                           │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  Per-project limits:                                                │    │
│  │  - Soft limit: Alert at 80% budget                                 │    │
│  │  - Hard limit: Pause and require approval                          │    │
│  │                                                                     │    │
│  │  Per-agent limits:                                                  │    │
│  │  - Max tokens per task                                             │    │
│  │  - Max retries before escalation                                   │    │
│  │                                                                     │    │
│  │  Real-time tracking:                                                │    │
│  │  - Dashboard showing spend                                         │    │
│  │  - Alerts via Slack/email                                          │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 8. Universal Language Support

### 8.1 Supported Stacks

Sovereign Firm is **language-agnostic**, supporting multiple technology stacks from day one:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    SUPPORTED TECHNOLOGY STACKS                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  FRONTEND         │  BACKEND          │  MOBILE          │  INFRASTRUCTURE │
│  ────────────────  │  ────────────────  │  ─────────────── │  ─────────────  │
│  React / Next.js   │  Go               │  React Native    │  Terraform      │
│  Vue / Nuxt        │  Node.js / Express│  Flutter         │  Pulumi         │
│  Svelte            │  Python / FastAPI │  Swift           │  K8s YAML       │
│  Angular           │  Rust / Axum      │  Kotlin          │  Docker         │
│  Astro             │  Java / Spring    │                  │  Helm           │
│                    │  .NET / C#        │                  │  CloudFormation │
│                    │  Ruby on Rails    │                  │                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 8.2 Stack Detection (Brownfield)

Automatic detection of project technology:

| Marker File | Detected Stack |
|-------------|---------------|
| `package.json` | Node.js ecosystem |
| `go.mod` | Go |
| `Cargo.toml` | Rust |
| `pom.xml` / `build.gradle` | Java (Maven/Gradle) |
| `pyproject.toml` / `requirements.txt` | Python |
| `*.csproj` / `*.sln` | .NET / C# |
| `Gemfile` | Ruby |
| `pubspec.yaml` | Flutter/Dart |

### 8.3 Language-Specific Executors

Each language has its own executor configuration:

```yaml
# executors/go.yaml
name: go-executor
image: golang:1.22-alpine
commands:
  build: go build ./...
  test: go test -v ./...
  lint: golangci-lint run
  run: go run main.go
  
# executors/python.yaml
name: python-executor
image: python:3.12-slim
commands:
  install: pip install -r requirements.txt
  test: pytest -v
  lint: ruff check .
  run: python main.py
```

---

## 9. Tree-sitter Integration

### 9.1 What is Tree-sitter?

Tree-sitter is a fast, incremental parsing library that provides consistent AST (Abstract Syntax Tree) access across 100+ programming languages. This is **essential** for language-agnostic code understanding.

### 9.2 Use Cases in Sovereign Firm

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    TREE-SITTER USE CASES                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. CODEBASE UNDERSTANDING (Brownfield Import)                             │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  Parse entire codebase to extract:                                 │ │
│     │  - Classes, functions, methods                                     │ │
│     │  - Import/export relationships                                     │ │
│     │  - Type definitions                                                │ │
│     │  - Call graphs                                                     │ │
│     │  - Dependencies between files                                      │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  2. TARGETED EDITS (Surgical Modifications)                                │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  Instead of LLM rewriting entire files:                            │ │
│     │  - Find exact AST node to modify (function, class, etc.)           │ │
│     │  - Extract just that node's code                                   │ │
│     │  - LLM modifies only that snippet                                  │ │
│     │  - Insert back at exact position                                   │ │
│     │  Result: Minimal diffs, preserve formatting, safer edits           │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  3. SYMBOL EXTRACTION (For Memory/RAG)                                     │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  Index codebase by symbols:                                        │ │
│     │  - Function: createUser at lib/users.ts:42                         │ │
│     │  - Class: UserService at services/user.go:15                       │ │
│     │  - Type: UserDTO at types/user.ts:8                                │ │
│     │                                                                     │ │
│     │  Enables semantic queries:                                          │ │
│     │  - "Find all usages of X"                                          │ │
│     │  - "What functions call Y?"                                        │ │
│     │  - "Show me all API endpoints"                                     │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  4. PATTERN DETECTION (Convention Analysis)                                │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  Tree-sitter queries to detect patterns:                           │ │
│     │  - "Find all React components" (TSX)                               │ │
│     │  - "Find all functions with error handling" (Go)                   │ │
│     │  - "Find all API endpoints" (Express/FastAPI)                      │ │
│     │  - "Find all database queries" (Prisma/SQLAlchemy)                 │ │
│     │                                                                     │ │
│     │  Result: Understand codebase conventions for consistent generation │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  5. SYNTAX VALIDATION (Pre-commit Check)                                   │
│     ┌────────────────────────────────────────────────────────────────────┐ │
│     │  Before committing generated code:                                 │ │
│     │  - Parse to ensure valid syntax (any language)                     │ │
│     │  - Compare AST structure before/after edit                         │ │
│     │  - Detect unintended structural changes                            │ │
│     │  - Fast (no need to run full compiler)                             │ │
│     └────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 9.3 Implementation

```go
import (
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/typescript"
    "github.com/smacker/go-tree-sitter/golang"
    "github.com/smacker/go-tree-sitter/python"
)

// Parse any file based on extension
func (p *Parser) Parse(filename string, code []byte) (*sitter.Tree, error) {
    lang := p.detectLanguage(filename)
    parser := sitter.NewParser()
    parser.SetLanguage(lang)
    return parser.Parse(nil, code), nil
}

// Find all functions in a file
func (p *Parser) FindFunctions(tree *sitter.Tree, lang *sitter.Language) []Function {
    query, _ := sitter.NewQuery([]byte(`
        (function_declaration
            name: (identifier) @name
        ) @function
    `), lang)
    
    cursor := sitter.NewQueryCursor()
    cursor.Exec(query, tree.RootNode())
    
    var functions []Function
    for {
        match, ok := cursor.NextMatch()
        if !ok { break }
        functions = append(functions, extractFunction(match))
    }
    return functions
}

// Targeted edit: Replace just one function body
func (p *Parser) ReplaceFunction(code []byte, funcName string, newBody string) []byte {
    tree, _ := p.Parse("file.ts", code)
    funcNode := p.findFunctionByName(tree, funcName)
    
    // Surgical replacement at exact byte positions
    result := make([]byte, 0, len(code)+len(newBody))
    result = append(result, code[:funcNode.StartByte()]...)
    result = append(result, []byte(newBody)...)
    result = append(result, code[funcNode.EndByte():]...)
    
    return result
}
```

---

## 10. Multi-Layer Sandbox Architecture

### 10.1 Isolation Philosophy

All code execution happens in isolated sandboxes. We use **three isolation levels** based on trust and requirements:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    SANDBOX ISOLATION LEVELS                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  LEVEL 1: BROWSER SANDBOX (WebContainers)                                  │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  For: Instant preview, light testing (JS/TS only)                  │    │
│  │  Runs in: User's browser tab                                       │    │
│  │  Isolation: Browser security model                                 │    │
│  │  Pros: Instant feedback, no server cost, hot reload                │    │
│  │  Cons: Browser-only languages, limited resources                   │    │
│  │  Trust Level: User's own code                                      │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  LEVEL 2: CONTAINER SANDBOX (Docker + gVisor)                              │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  For: Full builds, tests, multi-language execution                 │    │
│  │  Runs on: Our infrastructure                                       │    │
│  │  Isolation:                                                         │    │
│  │    - gVisor runtime (kernel-level sandboxing)                      │    │
│  │    - Network namespace (no external access by default)             │    │
│  │    - Read-only root filesystem                                     │    │
│  │    - Resource limits (CPU, memory, time, disk)                     │    │
│  │    - No privileged capabilities                                    │    │
│  │    - Seccomp profiles                                              │    │
│  │  Supports: ALL languages (Go, Python, Rust, Java, etc.)            │    │
│  │  Trust Level: Platform-generated code                              │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  LEVEL 3: EPHEMERAL MICROVM (Firecracker)                                  │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  For: Untrusted code, client uploads, security testing             │    │
│  │  Runs on: Dedicated microVM that's destroyed after use             │    │
│  │  Isolation: Full VM isolation (Firecracker microVM)                │    │
│  │  Boot time: ~125ms                                                 │    │
│  │  Use case: When client uploads their own repo                      │    │
│  │  Trust Level: Untrusted external code                              │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 10.2 Sandbox Selection Matrix

| Scenario | Sandbox Level | Why |
|----------|---------------|-----|
| Preview React component | Browser (WebContainers) | Instant, user sees changes live |
| Run npm test | Browser or Container | Browser for speed, Container for accuracy |
| Build Go backend | Container (gVisor) | Needs real Go toolchain |
| Run full integration tests | Container (gVisor) | Needs database, multiple processes |
| Client uploads unknown repo | MicroVM (Firecracker) | Untrusted code, maximum isolation |
| Deploy to staging | Container (gVisor) | Needs kubectl, cloud credentials |

### 10.3 Container Security Configuration

```yaml
# sandbox/gvisor-config.yaml
apiVersion: v1
kind: Pod
metadata:
  name: code-sandbox
  annotations:
    # Use gVisor runtime
    io.kubernetes.cri.untrusted-executor: "gvisor"
spec:
  runtimeClassName: gvisor
  
  # Security context
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 1000
    seccompProfile:
      type: RuntimeDefault
    
  containers:
    - name: sandbox
      image: sovereign/executor:latest
      
      # Resource limits
      resources:
        limits:
          cpu: "2"
          memory: "4Gi"
          ephemeral-storage: "10Gi"
        requests:
          cpu: "0.5"
          memory: "1Gi"
      
      # Security settings
      securityContext:
        allowPrivilegeEscalation: false
        readOnlyRootFilesystem: true
        capabilities:
          drop: ["ALL"]
      
      # Execution timeout
      activeDeadlineSeconds: 600  # 10 minutes max
      
      # Volume mounts
      volumeMounts:
        - name: workspace
          mountPath: /workspace
        - name: tmp
          mountPath: /tmp
          
  # Network isolation
  dnsPolicy: None
  hostNetwork: false
```

---

## 11. Technology Stack


| Layer | Technology | Rationale |
|-------|------------|-----------|
| **Frontend** | Next.js 14 | App Router, Server Components, great DX |
| **API** | Go | Performance, type safety, great for infrastructure |
| **Orchestration** | Temporal | Durable workflows, built-in retries, state management |
| **Preview** | WebContainers | Real Node.js in browser, npm works |
| **Execution** | Kubernetes | Container orchestration, scaling |
| **Vector DB** | ChromaDB | Simple, good for RAG |
| **Cache** | Redis | Sessions, pub/sub, rate limiting |
| **Database** | PostgreSQL | Reliable, Temporal compatible |
| **Secrets** | HashiCorp Vault | Enterprise-grade secret management |
| **Observability** | OpenTelemetry + Grafana | Distributed tracing, metrics |
| **LLM** | Multi-provider | Azure OpenAI, Anthropic, local Ollama |
| **Git** | GitHub/GitLab API | Version control, PRs |
| **Deploy** | Vercel/K8s | Flexible deployment targets |

---

## 9. Implementation Phases

### Phase 0: Foundation (Week 1)
- [ ] Replace Sandpack with WebContainers
- [ ] Implement streaming WebSocket for real-time updates
- [ ] Basic chunked generation (file-by-file)

### Phase 1: Agent Core (Week 2)
- [ ] Agent instance struct and lifecycle
- [ ] Agent pool with spawn/terminate
- [ ] Skill loading from YAML
- [ ] Meta-agent for team planning

### Phase 2: Greenfield Flow (Week 3)
- [ ] Complete streaming generation
- [ ] Checkpoint system
- [ ] Preview hot-reload integration
- [ ] Basic test execution

### Phase 3: Brownfield Support (Week 4)
- [ ] Repository ingestion
- [ ] Codebase analysis agent
- [ ] Pattern detection
- [ ] Targeted modification (diff-based)

### Phase 4: GitOps (Week 5)
- [ ] GitHub integration
- [ ] Branch per project
- [ ] PR creation
- [ ] Diff visualization

### Phase 5: Security (Week 6)
- [ ] Container isolation
- [ ] Secret management
- [ ] Code scanning
- [ ] Network policies

### Phase 6: Human-in-the-Loop (Week 7)
- [ ] Approval gates workflow
- [ ] Slack/email notifications
- [ ] Review interface
- [ ] Rollback capability

### Phase 7: Enterprise Features (Week 8)
- [ ] Multi-tenancy
- [ ] Cost tracking
- [ ] Usage quotas
- [ ] Audit logging

### Phase 8: Production Deployment (Week 9-10)
- [ ] Kubernetes deployment
- [ ] Vercel integration
- [ ] CI/CD pipeline
- [ ] Monitoring & alerting

---

## 10. Success Metrics

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Time to first preview | < 5 seconds | Timestamp from request to first file rendered |
| Agent spawn time | < 2 seconds | Time from spawn request to agent active |
| Brownfield success rate | > 80% | % of imported repos successfully modified |
| Test pass rate | > 90% | % of generated code that passes tests |
| Deployment success | > 95% | % of approved deploys that succeed |
| Token efficiency | < $0.10/KLOC | Cost per 1000 lines of generated code |
| User satisfaction | > 4.5/5 | Post-project survey |

---

## 11. Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| LLM generates insecure code | High | Medium | Automated security scanning before deploy |
| Cost overruns | Medium | High | Hard budget limits, model routing |
| Agent loops forever | Medium | Medium | Timeout policies, max iterations |
| Brownfield breaks existing code | High | Medium | Run existing tests, diff review |
| WebContainers browser compat | Medium | Low | Fallback to server-side execution |

---

## Appendix A: Directory Structure

```
sovereign-firm/
├── cmd/
│   ├── orchestrator/           # API Gateway
│   └── worker/                 # Temporal Worker
├── pkg/
│   ├── agents/                 # NEW: Agent system
│   │   ├── pool.go            # Agent pool management
│   │   ├── instance.go        # Agent instance
│   │   ├── lifecycle.go       # Spawn/terminate
│   │   ├── communication.go   # Inter-agent messaging
│   │   └── meta/              # Meta-agent
│   ├── skills/                 # NEW: Skill system
│   │   ├── loader.go          # Load skills from YAML
│   │   ├── registry.go        # Skill registry
│   │   └── builtin/           # Built-in skills
│   ├── execution/              # NEW: Execution layer
│   │   ├── preview/           # WebContainers integration
│   │   ├── sandbox/           # Docker execution
│   │   └── deploy/            # Deployment targets
│   ├── brownfield/             # NEW: Brownfield support
│   │   ├── ingestion.go       # Import repositories
│   │   ├── analysis.go        # Codebase analysis
│   │   └── modification.go    # Targeted edits
│   ├── gitops/                 # NEW: Git integration
│   │   ├── github.go
│   │   └── gitlab.go
│   ├── sovereign/              # Core SDK (existing)
│   │   ├── llm/               # LLM clients
│   │   └── memory/            # Vector store
│   └── firm/                   # Business logic (existing, refactor)
│       └── workflows/          # Temporal workflows
├── skills/                     # Skill definitions (YAML)
│   ├── react-developer/
│   ├── go-developer/
│   ├── devops-engineer/
│   └── qa-engineer/
├── frontend/                   # Next.js UI
└── deploy/                     # Deployment configs
```

---

**Document Status:** Ready for review  
**Next Steps:** Approve and begin Phase 0 implementation
