# Sovereign Firm v2.0 - Implementation Plan

**Version:** 1.0  
**Created:** 2026-01-15  
**Last Updated:** 2026-01-15  
**Status:** 🟡 In Progress

---

## Progress Summary

| Phase | Status | Progress | Start Date | End Date |
|-------|--------|----------|------------|----------|
| Phase 0: Foundation | 🟡 In Progress | 83% | 2026-01-15 | - |
| Phase 1: Agent Core | 🔴 Not Started | 0% | - | - |
| Phase 2: Execution Layer | 🔴 Not Started | 0% | - | - |
| Phase 3: Intelligence Layer | 🔴 Not Started | 0% | - | - |
| Phase 4: Multi-Agent | 🔴 Not Started | 0% | - | - |
| Phase 5: Enterprise Features | 🔴 Not Started | 0% | - | - |
| Phase 6: Production Hardening | 🔴 Not Started | 0% | - | - |

**Legend:** 🔴 Not Started | 🟡 In Progress | 🟢 Complete | 🔵 Blocked

---

## Phase 0: Foundation

**Goal:** Replace Sandpack with WebContainers, establish streaming infrastructure

**Duration:** 1-2 weeks

### Entry Criteria
- [x] Architecture document complete (ARCHITECTURE.md)
- [x] Current codebase builds and runs
- [x] Development environment set up (Docker, Temporal, ChromaDB)

### Tasks

| ID | Task | Status | Owner | Notes |
|----|------|--------|-------|-------|
| 0.1 | Replace Sandpack with WebContainers SDK | 🟢 | AI | Completed - PodConsole.tsx rewritten |
| 0.2 | Implement streaming WebSocket endpoint | 🟢 | AI | Completed - pkg/streaming + orchestrator |
| 0.3 | Create chunked generation protocol | 🟢 | AI | Completed - protocol.go defines all event types |
| 0.4 | Update PodConsole for streaming | 🟢 | AI | Completed - useStreaming hook + event handling |
| 0.5 | Add basic error boundary | 🟢 | AI | Completed - ErrorBoundary.tsx |
| 0.6 | Write integration tests | 🔴 | - | E2E test suite |

### Testing Requirements

| Test Type | Description | Status |
|-----------|-------------|--------|
| Unit | WebContainers integration unit tests | 🔴 |
| Unit | WebSocket endpoint tests | 🔴 |
| Integration | Frontend ↔ Backend streaming | 🔴 |
| E2E | Generate code → See in preview | 🔴 |
| Performance | Streaming latency < 500ms | 🔴 |

### Exit Criteria
- [ ] WebContainers boots and runs code in browser
- [ ] Streaming WebSocket delivers incremental updates
- [ ] User sees first output within 5 seconds
- [ ] All tests passing
- [ ] Code reviewed and merged

---

## Phase 1: Agent Core

**Goal:** Implement dynamic agent lifecycle, skills, and structured output contracts

**Duration:** 2-3 weeks

### Entry Criteria
- [ ] Phase 0 complete
- [ ] Streaming infrastructure working

### Tasks

| ID | Task | Status | Owner | Notes |
|----|------|--------|-------|-------|
| 1.1 | Define AgentInstance struct | 🔴 | - | pkg/agent/agent.go |
| 1.2 | Implement agent lifecycle (spawn→terminate) | 🔴 | - | State machine |
| 1.3 | Create AgentPool manager | 🔴 | - | Pool with limits |
| 1.4 | Implement skill loading from YAML | 🔴 | - | skills/*.yaml |
| 1.5 | Create 5 core skills | 🔴 | - | react-dev, backend-dev, qa, pm, architect |
| 1.6 | Implement SOC (Structured Output Contracts) | 🔴 | - | JSON Schema validation |
| 1.7 | Add SOC retry logic | 🔴 | - | Retry on validation failure |
| 1.8 | Integrate MCP protocol for tools | 🔴 | - | Tool standardization |

### Testing Requirements

| Test Type | Description | Status |
|-----------|-------------|--------|
| Unit | AgentInstance lifecycle tests | 🔴 |
| Unit | AgentPool limits and cleanup | 🔴 |
| Unit | Skill YAML parsing | 🔴 |
| Unit | SOC validation (valid/invalid JSON) | 🔴 |
| Unit | SOC retry logic | 🔴 |
| Integration | Agent spawn → execute → terminate | 🔴 |
| Contract | All 5 skills produce valid output | 🔴 |

### Exit Criteria
- [ ] Agents can be spawned with specific skills
- [ ] Agent lifecycle managed correctly
- [ ] All LLM outputs validated against SOC
- [ ] Invalid outputs trigger retry with feedback
- [ ] All tests passing
- [ ] Code reviewed and merged

---

## Phase 2: Execution Layer

**Goal:** Multi-layer sandbox (WebContainers, gVisor, Firecracker) for code execution

**Duration:** 2-3 weeks

### Entry Criteria
- [ ] Phase 1 complete
- [ ] Agent core working

### Tasks

| ID | Task | Status | Owner | Notes |
|----|------|--------|-------|-------|
| 2.1 | Level 1: WebContainers integration | 🔴 | - | Browser sandbox |
| 2.2 | Level 2: Docker + gVisor setup | 🔴 | - | Container sandbox |
| 2.3 | Level 2: Resource limits (CPU, mem, time) | 🔴 | - | Prevent runaway |
| 2.4 | Level 2: Network restrictions | 🔴 | - | Allowlist npm |
| 2.5 | Level 3: Firecracker microVM (optional) | 🔴 | - | For untrusted code |
| 2.6 | Sandbox selection logic | 🔴 | - | Choose level by trust |
| 2.7 | Tree-sitter integration | 🔴 | - | Code analysis |
| 2.8 | Universal language detection | 🔴 | - | Auto-detect stack |

### Testing Requirements

| Test Type | Description | Status |
|-----------|-------------|--------|
| Unit | WebContainers file operations | 🔴 |
| Unit | gVisor container lifecycle | 🔴 |
| Unit | Resource limit enforcement | 🔴 |
| Unit | Tree-sitter parsing (5+ languages) | 🔴 |
| Security | Network isolation verified | 🔴 |
| Security | Filesystem isolation verified | 🔴 |
| Integration | Execute code in all 3 sandbox levels | 🔴 |

### Exit Criteria
- [ ] Code executes in WebContainers (browser)
- [ ] Code executes in gVisor container (server)
- [ ] Resource limits enforced
- [ ] Network restrictions working
- [ ] Tree-sitter parses 5+ languages
- [ ] All tests passing
- [ ] Security review complete

---

## Phase 3: Intelligence Layer

**Goal:** RAG system with ChromaDB, multi-level memory, codebase understanding

**Duration:** 2-3 weeks

### Entry Criteria
- [ ] Phase 2 complete
- [ ] Execution layer working

### Tasks

| ID | Task | Status | Owner | Notes |
|----|------|--------|-------|-------|
| 3.1 | Implement 3-level memory structure | 🔴 | - | Global, Client, Project |
| 3.2 | Codebase chunking strategy | 🔴 | - | Tree-sitter based |
| 3.3 | Embedding generation pipeline | 🔴 | - | Batch processing |
| 3.4 | Semantic search with ranking | 🔴 | - | Multi-level retrieval |
| 3.5 | Context injection into prompts | 🔴 | - | RAG pipeline |
| 3.6 | Project context accumulation | 🔴 | - | Learn from decisions |
| 3.7 | Brownfield analysis workflow | 🔴 | - | Index existing codebase |

### Testing Requirements

| Test Type | Description | Status |
|-----------|-------------|--------|
| Unit | Chunking produces valid segments | 🔴 |
| Unit | Embeddings have correct dimensions | 🔴 |
| Unit | Semantic search returns relevant results | 🔴 |
| Integration | RAG improves code generation | 🔴 |
| Integration | Brownfield analysis indexes correctly | 🔴 |
| Performance | Search latency < 100ms | 🔴 |
| Quality | Retrieved context is relevant (manual review) | 🔴 |

### Exit Criteria
- [ ] All 3 memory levels operational
- [ ] Semantic search returns relevant context
- [ ] Generated code uses retrieved context
- [ ] Brownfield codebases indexed successfully
- [ ] All tests passing
- [ ] Code reviewed and merged

---

## Phase 4: Multi-Agent Coordination

**Goal:** Multi-agent orchestration with DAG, file locking, Git branches

**Duration:** 3-4 weeks

### Entry Criteria
- [ ] Phase 3 complete
- [ ] Intelligence layer working

### Tasks

| ID | Task | Status | Owner | Notes |
|----|------|--------|-------|-------|
| 4.1 | Implement Meta-Agent (Engagement Manager) | 🔴 | - | Orchestrator |
| 4.2 | Task Dependency Graph (DAG) implementation | 🔴 | - | Scheduler |
| 4.3 | Parallel task execution | 🔴 | - | Same-layer parallelism |
| 4.4 | File-level locking | 🔴 | - | Prevent conflicts |
| 4.5 | Git branch per agent | 🔴 | - | Isolation |
| 4.6 | Branch merge orchestration | 🔴 | - | Meta-Agent merges |
| 4.7 | Agent CI pipeline | 🔴 | - | Per-agent validation |
| 4.8 | Self-correction loop | 🔴 | - | Retry on CI failure |
| 4.9 | Artifact context injection | 🔴 | - | Pass deps to agents |

### Testing Requirements

| Test Type | Description | Status |
|-----------|-------------|--------|
| Unit | DAG correctly orders tasks | 🔴 |
| Unit | DAG identifies parallel tasks | 🔴 |
| Unit | File locking prevents conflicts | 🔴 |
| Unit | Git branch operations work | 🔴 |
| Integration | Multi-agent builds complete app | 🔴 |
| Integration | CI failures trigger self-correction | 🔴 |
| Stress | 10 agents working concurrently | 🔴 |

### Exit Criteria
- [ ] Meta-Agent creates valid task DAGs
- [ ] Tasks execute in parallel where possible
- [ ] No file conflicts with locking
- [ ] Each agent commits to own branch
- [ ] Merges happen without conflicts
- [ ] CI validates each agent's work
- [ ] All tests passing
- [ ] Code reviewed and merged

---

## Phase 5: Enterprise Features

**Goal:** Multi-tenancy, security hardening, cost controls, HAP filtering

**Duration:** 3-4 weeks

### Entry Criteria
- [ ] Phase 4 complete
- [ ] Multi-agent working

### Tasks

| ID | Task | Status | Owner | Notes |
|----|------|--------|-------|-------|
| 5.1 | Multi-tenant data isolation | 🔴 | - | Per-tenant namespaces |
| 5.2 | Authentication/Authorization | 🔴 | - | OAuth/OIDC |
| 5.3 | HAP filtering integration | 🔴 | - | Content safety |
| 5.4 | PII detection and redaction | 🔴 | - | Privacy |
| 5.5 | Secret scanning | 🔴 | - | No keys in code |
| 5.6 | Cost tracking per tenant | 🔴 | - | Token accounting |
| 5.7 | Budget limits and alerts | 🔴 | - | Prevent overruns |
| 5.8 | Audit logging | 🔴 | - | All actions logged |
| 5.9 | Model routing by task type | 🔴 | - | Cost optimization |

### Testing Requirements

| Test Type | Description | Status |
|-----------|-------------|--------|
| Unit | Tenant isolation verified | 🔴 |
| Unit | HAP filter catches offensive content | 🔴 |
| Unit | PII detector finds sensitive data | 🔴 |
| Unit | Secret scanner catches API keys | 🔴 |
| Unit | Cost tracking accurate | 🔴 |
| Security | Cross-tenant access blocked | 🔴 |
| Security | Penetration testing | 🔴 |
| Compliance | Audit logs complete | 🔴 |

### Exit Criteria
- [ ] Tenants fully isolated
- [ ] No HAP content in outputs
- [ ] PII automatically redacted
- [ ] Secrets never in generated code
- [ ] Cost tracking accurate to 1%
- [ ] Budget limits enforced
- [ ] Security audit passed
- [ ] All tests passing

---

## Phase 6: Production Hardening

**Goal:** Production readiness, observability, disaster recovery

**Duration:** 2-3 weeks

### Entry Criteria
- [ ] Phase 5 complete
- [ ] Enterprise features working

### Tasks

| ID | Task | Status | Owner | Notes |
|----|------|--------|-------|-------|
| 6.1 | OpenTelemetry integration | 🔴 | - | Traces, metrics |
| 6.2 | Grafana dashboards | 🔴 | - | Monitoring |
| 6.3 | AlertManager rules | 🔴 | - | SLO/SLA |
| 6.4 | Database backup/restore | 🔴 | - | DR |
| 6.5 | Temporal backup/restore | 🔴 | - | Workflow state |
| 6.6 | Rate limiting | 🔴 | - | API protection |
| 6.7 | Load testing | 🔴 | - | Capacity planning |
| 6.8 | Chaos engineering tests | 🔴 | - | Resilience |
| 6.9 | Runbooks and documentation | 🔴 | - | Operations |
| 6.10 | Production deployment scripts | 🔴 | - | K8s manifests |

### Testing Requirements

| Test Type | Description | Status |
|-----------|-------------|--------|
| Unit | Metrics exported correctly | 🔴 |
| Integration | Traces span across services | 🔴 |
| Load | Handle 100 concurrent projects | 🔴 |
| Chaos | Survives pod termination | 🔴 |
| Chaos | Survives database failover | 🔴 |
| DR | Restore from backup works | 🔴 |
| Performance | P99 latency < 2s | 🔴 |

### Exit Criteria
- [ ] Full observability (traces, metrics, logs)
- [ ] Dashboards show system health
- [ ] Alerts fire on SLO breach
- [ ] Backup/restore tested
- [ ] Load test passes (100 concurrent)
- [ ] Chaos tests pass
- [ ] Runbooks complete
- [ ] Production deployment successful

---

## Overall Definition of Done

A phase is complete when:

1. ✅ All tasks marked complete
2. ✅ All tests passing (unit, integration, e2e)
3. ✅ Code reviewed by at least 1 person
4. ✅ Documentation updated
5. ✅ No critical/high bugs open
6. ✅ Exit criteria met
7. ✅ Merged to main branch
8. ✅ Deployed to staging environment

---

## Risk Register

| ID | Risk | Impact | Probability | Mitigation | Status |
|----|------|--------|-------------|------------|--------|
| R1 | WebContainers SDK limitations | High | Medium | Fallback to server containers | 🟢 Mitigated |
| R2 | LLM output inconsistency | High | High | SOC with retry logic | 🟡 Monitoring |
| R3 | Multi-agent conflicts | Medium | Medium | File locking, branches | 🟡 Planned |
| R4 | Cost overruns | Medium | Medium | Budget limits, alerts | 🟡 Planned |
| R5 | Security vulnerabilities | High | Low | Sandbox, scanning | 🟡 Planned |
| R6 | Performance bottlenecks | Medium | Medium | Load testing, optimization | 🟡 Planned |

---

## Session Handoff Notes

**Current Session:** 2026-01-15 (Updated 17:20 UTC)

### What Was Done
- Created comprehensive ARCHITECTURE.md (~2,200+ lines)
- Added sections for:
  - Agent Role Catalog (20+ specialist roles)
  - Task Dependency Graph (DAG)
  - Structured Output Contracts (SOC)
  - Industry Patterns (MCP, A2A, Plan-and-Execute)
  - HAP Filtering / Content Safety
  - Multi-layer Sandbox Architecture
- Pushed all changes to GitHub
- **Completed Phase 0: Foundation (83%)**
  - ✅ Task 0.1: Replaced Sandpack with WebContainers SDK
  - ✅ Task 0.2: Implemented streaming WebSocket endpoint (pkg/streaming)
  - ✅ Task 0.3: Created chunked generation protocol (18 event types)
  - ✅ Task 0.4: Updated PodConsole with useStreaming hook
  - ✅ Task 0.5: Added ErrorBoundary component
  - 🔴 Task 0.6: Integration tests (remaining)
- **E2E Validation Completed:**
  - Two-column layout renders correctly
  - PM Agent chat working
  - WebContainers boots and installs deps
  - WebSocket hub logs connections
  - Known issue: "Insufficient resources" in some browsers

### Next Steps
1. Complete Phase 0: Task 0.6 (Integration tests)
2. Begin Phase 1: Agent Core
   - AgentInstance struct + lifecycle
   - AgentPool manager
   - Skill loading from YAML
   - SOC (Structured Output Contracts)

### Blockers
- None critical
- Minor: WebSocket browser resource limits (non-blocking)

### Notes for Next Session
- Consider simplifying WebSocket reconnection logic
- Add integration tests for WebSocket flow
- Start with AgentInstance struct in Phase 1

---

## Change Log

| Date | Author | Changes |
|------|--------|---------|
| 2026-01-15 | AI Architect | Initial plan created |
| 2026-01-15 | AI Architect | Task 0.1 complete: WebContainers replaces Sandpack |
| 2026-01-15 | AI Architect | Tasks 0.2-0.5 complete: Streaming + Error Boundary |
| 2026-01-15 | AI Architect | E2E validation completed, minor issues noted |

