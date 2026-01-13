# 🧠 BotMatrix AI, RAG & MCP Core Guide

> **Version**: 2.5
> **Status**: Core architecture implemented
> [🌐 English](AI_RAG_MCP_GUIDE.md) | [简体中文](../zh-CN/core/AI_RAG_MCP_GUIDE.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

This guide details BotMatrix's AI capabilities, RAG (Retrieval-Augmented Generation) architecture, the MCP (Model Context Protocol) layer, and the **Digital Employee** system.

---

## 1. MCP Layer (Model Context Protocol)

MCP is the "driver interface" for BotMatrix, decoupling capability providers from model consumers.
- **Resources**: Static/dynamic data access (logs, DB reports).
- **Tools**: Action execution (send messages, call APIs).
- **Prompts**: Pre-defined templates for specific personas or tasks.

### 1.1 Storage & Vector Integration
- **Persistence**: Cognitive memory is stored in PostgreSQL via `CognitiveMemoryService`.
- **pgvector**: Every memory/knowledge fragment is vectorized (e.g., using BGE-M3) and stored for semantic search.

---

## 2. RAG 2.0 (Retrieval-Augmented Generation)

RAG allows bots to "bootstrap" their knowledge from system docs and external databases.
- **Hybrid Search**: Combines vector (semantic) and full-text (keyword) search.
- **Query Refinement**: Automatically optimizes user queries before retrieval.

---

## 3. Digital Employee System

**Digital Employees** are anthropomorphic AI agents with IDs, roles, skills, and KPI tracking.
- **Identity**: Linked to `IdentityGORM` for departmental and enterprise context.
- **Collaboration (Agent Mesh)**: Supports synchronous consultation and asynchronous task delegation across nodes or enterprises.

### 3.1 KPI Framework
Performance is automatically calculated based on:
- **Success Rate** (40%)
- **Efficiency** (30%)
- **Autonomy** (20%)
- **Cost/Token Usage** (10%)

---

## 4. Security & Privacy (Privacy Bastion)

- **PII Masking**: Automatically strips phone numbers and names before sending data to LLMs.
- **Audit Trails**: `AIAgentTrace` logs every tool call, parameter, and result for transparency.
