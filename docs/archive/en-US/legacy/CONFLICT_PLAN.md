# Conflict & Improvement Plan

> [🌐 English](CONFLICT_PLAN.md) | [简体中文](../zh-CN/CONFLICT_PLAN.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

This document records outdated documentation, logic conflicts, and future improvement directions within the project.

## 🚩 Identified Conflicts

### 1. Message Routing Logic
- **Conflict**: Some routing priorities described in `ROUTING_RULES.md` are not fully consistent with the actual implementation in `src/BotNexus/handlers.go`.
- **Status**: More complex wildcard matching logic has been introduced in the code, but the documentation has not been updated yet.
- **Plan**: Synchronize the documentation and add detailed explanations for regex routing and wildcard priorities.

### 2. Redis Dependency
- **Conflict**: Some old documents still mention using local files for session storage, while `REDIS_UPGRADE.md` explicitly states the switch to Redis.
- **Status**: Some plugins may still attempt to read/write local JSON files.
- **Plan**: Comprehensively clean up local storage descriptions in old documents and unify them to the Redis access pattern.

### 3. Multi-platform Compatibility
- **Conflict**: `SERVER_MANUAL.md` is mainly focused on the Python side, while implementation details for the Go side (e.g., OneBot adapter layer) are not fully described.
- **Status**: Developers lack a unified interface specification when adapting to new platforms.
- **Plan**: Establish a general `PLATFORM_ADAPTER_GUIDE.md`.

---

## 📅 Improvement Roadmap

### Phase 1: Documentation Standardization (Completed)
- [x] Establish a documentation center index.
- [x] Supplement system architecture diagrams and API references.
- [x] Eliminate isolated documents and establish hierarchical links.

### Phase 2: Content Completion (In Progress)
- [ ] Supplement **[Advanced Plugin Development Tutorial]**: How to handle multimedia messages.
- [ ] Write **[Troubleshooting Manual]**.
- [ ] Perfect **[Detailed Database Schema Description]**.

### Phase 3: Automated Synchronization
- [ ] Introduce `swagger` or similar tools to automatically generate API documentation from code comments.
- [ ] Add documentation format validation to the CI/CD workflow.

---

*If you find any documentation errors, please submit an Issue or Pull Request.*
