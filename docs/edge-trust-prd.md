# zcid "Edge Trust" (Feral Compute + Cryptographic Guillotine)
**Status:** Planned
**Date:** April 2026

## 1. Executive Summary
**zcid Edge Trust** transforms zcid from a standard CI/CD tool into a universally deployable, zero-trust release orchestrator. By combining a lightweight, Kubernetes-independent deployment agent (**Feral Compute**) with an immutable, mathematically verifiable release ledger (**Cryptographic Guillotine**), zcid allows 1-person platform teams to securely deploy to any infrastructure (bare metal, IoT, edge) with SOC2-level auditability.

## 2. Problem Statement
1. **The Infrastructure Trap:** Modern CI/CD assumes Kubernetes. Deploying to edge devices (e.g., retail store servers, Raspberry Pis) or legacy bare-metal requires hacky SSH scripts or heavy orchestrators.
2. **The Compliance Nightmare:** Manual deployments and edge updates lack cryptographic proof. Auditors cannot easily verify that the exact binary running on a remote server passed security scans and integration tests.

## 3. Target Audience
* **Solo Platform Engineers:** Need to manage diverse infrastructure without the overhead of maintaining K8s clusters.
* **Regulated Industries (Fintech/Healthcare):** Require absolute, cryptographic proof of what is running in production.
* **Edge/IoT Operators:** Need to deploy updates to hundreds of low-resource devices securely.

---

## 4. Technical Architecture

### 4.1 High-Level System Topology
We are introducing a new standalone binary (`zcid-agent`) that operates on a pull-based model to easily bypass inbound firewall restrictions on edge devices.

```text
[ Edge Device / Bare Metal ]          [ zcid Control Plane ]
+--------------------------+          +--------------------+
|                          |  HTTPS   |                    |
|       zcid-agent         | <======> |    zcid-server     |
|  (Lightweight Go daemon) |  (Pull)  |    (Go / Gin)      |
|                          |          |                    |
+-------+---------+--------+          +--+-------+-------+-+
        |         |                      |       |       |
    [Execute]  [Verify]                  v       v       v
        |         |                  [Redis] [Postgres] [MinIO]
   (Local App) (Ed25519)             (State) (Ledger) (Artifacts)
```

### 4.2 Database Schema Additions (PostgreSQL)
* **`agents`**: Tracks physical/virtual nodes (`id`, `name`, `token_hash`, `environment_id`, `status`, `last_seen_at`).
* **`artifact_signatures`**: Stores cryptographic proof (`id`, `artifact_id`, `signature`, `public_key`, `signed_at`).
* **`deployment_ledger`**: Immutable audit log (`id`, `deployment_id`, `agent_id`, `artifact_signature_id`, `verification_status`, `executed_at`).

### 4.3 Cryptographic Security Model (The Guillotine)
1. **Key Generation:** On first boot, `zcid-server` generates an Ed25519 keypair. The private key is encrypted at rest using `ZCID_ENCRYPTION_KEY`.
2. **Signing:** When a pipeline completes, the server hashes the artifact (SHA-256) and signs the hash using the private key.
3. **Verification:** The `zcid-agent` downloads the artifact, hashes it, fetches the public key/signature, and runs `ed25519.Verify()`. If it fails, the binary is deleted and the deployment is guillotined.

---

## 5. Implementation Roadmap (Epics & Stories)

### Epic 1: Cryptographic Foundation (Backend Core)
* **Story 1.1: Key Management:** Generate Ed25519 keypair on boot, encrypt private key with `ZCID_ENCRYPTION_KEY`.
* **Story 1.2: Artifact Signing:** Hash and sign build artifacts upon pipeline completion; store in `artifact_signatures`.

### Epic 2: The Feral Agent (New Go Binary)
* **Story 2.1: Agent Scaffolding & Auth:** Create `zcid-agent` CLI with token-based authentication.
* **Story 2.2: The Guillotine Verification:** Implement download, SHA-256 hashing, and Ed25519 signature verification.
* **Story 2.3: Execution & Telemetry:** Execute verified binary and stream logs/status back to server.

### Epic 3: Control Plane & Ledger (Backend APIs)
* **Story 3.1: Agent Management APIs:** Implement `/api/v1/agents/register`, `/tasks`, `/status` and `agents` table.
* **Story 3.2: The Immutable Ledger:** Record verification results in `deployment_ledger`.

### Epic 4: Edge Fleet Dashboard (React Frontend)
* **Story 4.1: Fleet Management UI:** Build "Edge Nodes" tab to view connected agents and heartbeat status.
* **Story 4.2: Cryptographic Audit UI:** Update Deployment details to show signature and verification status.
