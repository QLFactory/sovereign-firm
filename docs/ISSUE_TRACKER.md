# Sovereign Firm - Issue Tracker

## Overview

This document tracks all identified issues from the deep codebase analysis. Issues are prioritized by functionality and robustness first, with security deferred until before go-live.

**Last Updated**: 2026-01-19
**Analysis Scope**: Full codebase trace of backend (Go) and frontend (Next.js)

---

## Priority Legend

| Priority | Description | Timeline |
|----------|-------------|----------|
| **P0** | Critical - Causes crashes, deadlocks, or data loss | Week 1-2 |
| **P1** | High - Data integrity or workflow correctness | Week 3 |
| **P2** | Medium - Frontend stability and UX | Week 4 |
| **P3** | Security - Address before go-live | Pre-launch |

---

## P0 - Critical Reliability Issues

### ISS-001: Deadlock in Message Checker
- **File**: `pkg/agent/executor.go`
- **Lines**: 188-219
- **Description**: `checkForNewMessages()` holds mutex lock while receiving from channel. If channel blocks, entire executor deadlocks.
- **Impact**: Complete workflow stall, unrecoverable without restart
- **Fix**: Release lock before channel receive, use select with timeout
- **Status**: [x] **FIXED** - 2026-01-19
- **Resolution**: Refactored to use non-blocking `select` with `default` case. Lock acquired only briefly per message, never during channel operations.

### ISS-002: Goroutine Leak - Agent Context
- **File**: `pkg/agent/agent.go`
- **Line**: 157
- **Description**: Agent context uses `context.Background()` instead of pool context. Agent goroutines survive pool shutdown.
- **Impact**: Memory leak, zombie goroutines accumulate over time
- **Fix**: Pass pool context to agent, propagate cancellation
- **Status**: [x] **FIXED** - 2026-01-19
- **Resolution**: `NewAgentInstance` now takes `parentCtx` parameter. Pool passes its context, so agent goroutines are cancelled on pool shutdown.

### ISS-003: Goroutine Leak - Outbox Forwarder
- **File**: `pkg/agent/pool.go`
- **Lines**: 116-134
- **Description**: Outbox forwarder goroutine started but never tracked or stopped on pool shutdown.
- **Impact**: Leaked goroutines, potential panic on closed channel
- **Fix**: Track goroutine with WaitGroup, cancel on shutdown
- **Status**: [x] **FIXED** - 2026-01-19
- **Resolution**: Outbox forwarder now uses `select` with `p.ctx.Done()` to exit on pool shutdown. Also handles channel close gracefully.

### ISS-004: Silent Message Loss - Agent Outbox
- **File**: `pkg/agent/agent.go`
- **Lines**: 281-306
- **Description**: Non-blocking send to outbox channel. If buffer full, message silently dropped.
- **Impact**: Inter-agent messages lost without notification
- **Fix**: Use blocking send with timeout, log/retry on failure
- **Status**: [x] **FIXED** - 2026-01-19
- **Resolution**: `SendMessage` now uses 100ms timeout, logs warning on drop, returns bool to indicate success/failure.

### ISS-005: Silent Message Loss - Pool Inbox
- **File**: `pkg/agent/pool.go`
- **Lines**: 242-253
- **Description**: Non-blocking send to agent inbox. If buffer full, message silently dropped.
- **Impact**: Workflow signals lost, agents miss instructions
- **Fix**: Use blocking send with timeout, return error to caller
- **Status**: [x] **FIXED** - 2026-01-19
- **Resolution**: New `deliverToAgent` helper uses 100ms timeout. Logs warning on drop. Also logs when message sent to unknown agent.

### ISS-006: No Panic Recovery in Message Processor
- **File**: `pkg/agent/agent.go`
- **Lines**: 308-331
- **Description**: `processMessages()` goroutine has no defer/recover. Panic crashes entire agent.
- **Impact**: Single bad message can crash agent and stall workflow
- **Fix**: Add defer with recover, log panic and continue
- **Status**: [x] **FIXED** - 2026-01-19
- **Resolution**: New `safeHandleMessage` wrapper with defer/recover. Panics are logged and agent continues processing.

### ISS-007: Activity Errors Continue Workflow
- **File**: `pkg/firm/workflows/consultancy.go`
- **Lines**: 336, 416, 461
- **Description**: Activity errors are logged but workflow continues with corrupted state instead of failing fast.
- **Impact**: Cascading failures, invalid deliverables
- **Fix**: Return error from workflow on critical activity failure
- **Status**: [ ] Open

### ISS-008: Test Failures Don't Block Deployment
- **File**: `pkg/firm/workflows/consultancy.go`
- **Line**: 931
- **Description**: QA test failures are logged but deployment proceeds anyway.
- **Impact**: Broken code gets deployed
- **Fix**: Check test results, block deployment on failure
- **Status**: [ ] Open

---

## P1 - Data Integrity Issues

### ISS-009: Incomplete DAG State
- **File**: `pkg/firm/workflows/consultancy.go`
- **Lines**: 251-266
- **Description**: DAG only defines 7 tasks but workflow has 11 phases. Missing: QA, SRE, BROWNFIELD, COMPLETE.
- **Impact**: Progress tracking incorrect, UI shows wrong state
- **Fix**: Add all phases to DAG definition
- **Status**: [ ] Open

### ISS-010: Non-Deterministic Timestamp
- **File**: `pkg/firm/workflows/consultancy.go`
- **Line**: 1184
- **Description**: Uses `time.Now()` instead of `workflow.Now(ctx)`. Breaks Temporal's deterministic replay.
- **Impact**: Workflow replay fails, state corruption on recovery
- **Fix**: Replace with `workflow.Now(ctx)`
- **Status**: [ ] Open

### ISS-011: File Collisions in AllCodeFiles
- **File**: `pkg/firm/workflows/consultancy.go`
- **Lines**: 659-730
- **Description**: Files from all agents merged into single map without namespacing. Same filename from different agents overwrites.
- **Impact**: Code lost, wrong files used in build
- **Fix**: Namespace by agent ID or use list instead of map
- **Status**: [ ] Open

### ISS-012: Brownfield Results Not Passed to Architect
- **File**: `pkg/firm/workflows/consultancy.go`
- **Lines**: ~400-450
- **Description**: Brownfield analysis results stored but not passed to Architect activity input.
- **Impact**: Architect designs without knowledge of existing codebase
- **Fix**: Include brownfield results in ArchitectInput struct
- **Status**: [ ] Open

### ISS-013: Fake Token Counting
- **File**: `pkg/agent/executor.go`
- **Lines**: 62-64
- **Description**: Token count estimated as `len(resp.Response) / 4`. Not actual token count.
- **Impact**: Inaccurate billing, context window miscalculation
- **Fix**: Use actual token count from LLM response
- **Status**: [ ] Open

### ISS-014: Unbounded Conversation History
- **File**: `pkg/agent/executor.go`
- **Lines**: 49, 78-79
- **Description**: Messages appended to conversation history without limit. Never truncated.
- **Impact**: Memory growth, eventual OOM, context window overflow
- **Fix**: Implement sliding window or summarization
- **Status**: [ ] Open

---

## P2 - Frontend Issues

### ISS-015: Polling Dependency Array Bug
- **File**: `frontend/app/components/PodConsole.tsx`
- **Line**: 473
- **Description**: `selectedFile` in useEffect dependency array causes polling to reset on every file selection.
- **Impact**: UX stuttering, lost updates during file browsing
- **Fix**: Remove `selectedFile` from dependency array or use ref
- **Status**: [ ] Open

### ISS-016: Memory Leak - File Buffers
- **File**: `frontend/app/components/PodConsole.tsx`
- **Lines**: 179-189
- **Description**: `fileBuffersRef` accumulates file contents but never cleared on project change.
- **Impact**: Browser memory grows, eventual crash on long sessions
- **Fix**: Clear buffers on project change or component unmount
- **Status**: [ ] Open

### ISS-017: Files Only Merged, Never Replaced
- **File**: `frontend/app/components/PodConsole.tsx`
- **Lines**: 453-456
- **Description**: File updates merged into existing map. Deleted files persist.
- **Impact**: Stale files shown, incorrect file tree
- **Fix**: Replace file map on full refresh, merge only on incremental
- **Status**: [ ] Open

### ISS-018: Silent Disconnect Handling
- **File**: `frontend/app/hooks/useStreaming.ts`
- **Lines**: 329-331
- **Description**: `onDisconnect` callback defined but does nothing. No retry logic.
- **Impact**: WebSocket disconnect silently fails, user sees stale data
- **Fix**: Implement reconnection with exponential backoff
- **Status**: [ ] Open

### ISS-019: Token Refresh Race Condition
- **File**: `frontend/app/lib/api/client.ts`
- **Lines**: 88-106
- **Description**: Multiple concurrent 401 responses each trigger token refresh. Race condition on refresh token usage.
- **Impact**: Auth failures, user logged out unexpectedly
- **Fix**: Add mutex/queue for refresh, deduplicate concurrent refreshes
- **Status**: [ ] Open

### ISS-020: Multi-Tab localStorage Sync
- **File**: `frontend/app/lib/api/client.ts`
- **Lines**: 43-44
- **Description**: Token stored in localStorage but no cross-tab sync. One tab refreshes, others have stale token.
- **Impact**: Multi-tab usage causes auth errors
- **Fix**: Use storage event listener or BroadcastChannel
- **Status**: [ ] Open

---

## P3 - Security Issues (Pre-Launch)

### ISS-021: Path Traversal in MCP Tools
- **File**: `pkg/mcp/builtin.go`
- **Line**: 614
- **Description**: `resolvePath()` accepts absolute paths and `..` traversal. Can read/write outside workspace.
- **Impact**: Arbitrary file access on host system
- **Fix**: Validate paths are within workspace, reject absolute paths and `..`
- **Status**: [ ] Open

### ISS-022: No File Size Limits
- **File**: `pkg/mcp/builtin.go`
- **Lines**: 81, 117
- **Description**: `file_read` and `file_write` have no size bounds. Can read/write arbitrarily large files.
- **Impact**: DoS via memory exhaustion
- **Fix**: Add configurable max file size limit
- **Status**: [ ] Open

### ISS-023: Prompt Injection - Dev Activity
- **File**: `pkg/firm/activities/dev.go`
- **Line**: 94
- **Description**: User specification concatenated directly into LLM prompt without sanitization.
- **Impact**: User can inject instructions to LLM
- **Fix**: Escape or structure user input in prompt
- **Status**: [ ] Open

### ISS-024: Prompt Injection - Architect Activity
- **File**: `pkg/firm/activities/architect.go`
- **Line**: 283
- **Description**: User input embedded in architect prompt without sanitization.
- **Impact**: User can manipulate architecture decisions
- **Fix**: Escape or structure user input in prompt
- **Status**: [ ] Open

### ISS-025: Prompt Injection - QA Activity
- **File**: `pkg/firm/activities/qa.go`
- **Line**: 79
- **Description**: User spec passed directly to QA prompt.
- **Impact**: User can bypass test generation
- **Fix**: Escape or structure user input in prompt
- **Status**: [ ] Open

### ISS-026: Network Access in Test Sandbox
- **File**: `pkg/firm/activities/testrunner.go`
- **Line**: 103
- **Description**: `config.NetworkEnabled = true` gives test sandbox full network access.
- **Impact**: Tests can exfiltrate data or attack internal services
- **Fix**: Disable network by default, whitelist specific endpoints if needed
- **Status**: [ ] Open

### ISS-027: Azure API Key in Environment
- **File**: `docker-compose.yaml`
- **Lines**: 69-72
- **Description**: Azure OpenAI API key passed via environment variable from `.env`.
- **Impact**: Key visible in container inspection, process listing
- **Fix**: Use Docker secrets or external secret manager
- **Status**: [ ] Open

---

## Implementation Plan

### Week 1-2: Reliability (P0)
1. Fix deadlock in executor.go (ISS-001)
2. Fix goroutine leaks (ISS-002, ISS-003)
3. Add message overflow handling (ISS-004, ISS-005)
4. Add panic recovery (ISS-006)
5. Fix activity error handling (ISS-007)
6. Block deployment on test failure (ISS-008)

### Week 3: Data Integrity (P1)
1. Complete DAG tasks (ISS-009)
2. Fix timestamp determinism (ISS-010)
3. Namespace files (ISS-011)
4. Pass brownfield to architect (ISS-012)
5. Implement proper token counting (ISS-013)
6. Add conversation history limits (ISS-014)

### Week 4: Frontend (P2)
1. Fix polling dependency array (ISS-015)
2. Clear file buffers (ISS-016)
3. Fix file replacement logic (ISS-017)
4. Implement reconnection logic (ISS-018)
5. Fix token refresh race (ISS-019)
6. Add multi-tab sync (ISS-020)

### Pre-Launch: Security (P3)
1. Add path validation (ISS-021)
2. Add file size limits (ISS-022)
3. Sanitize LLM prompts (ISS-023, ISS-024, ISS-025)
4. Disable test network (ISS-026)
5. Move to secret manager (ISS-027)

---

## Progress Summary

| Priority | Total | Open | In Progress | Done |
|----------|-------|------|-------------|------|
| P0 | 8 | 2 | 0 | 6 |
| P1 | 6 | 6 | 0 | 0 |
| P2 | 6 | 6 | 0 | 0 |
| P3 | 7 | 7 | 0 | 0 |
| **Total** | **27** | **21** | **0** | **6** |

---

## Notes

- All line numbers are approximate and may shift with code changes
- Security issues (P3) are intentionally deferred per product decision
- Each issue should be converted to GitHub issue when ready to work on it
- Test coverage should be added alongside each fix
