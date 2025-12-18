# Project Specification: Guardian-Log

**Subject:** Self-Hosted LLM-Powered DNS Anomaly Explainer  
**Target Stack:** Go (Backend), TypeScript/React (Frontend), AdGuard Home (Integration)  
**Author:** Software Architect (via Gemini)

---

## 1. Project Overview
Guardian-Log is a middleware service that acts as a reasoning layer for home network security. It monitors **AdGuard Home** DNS query logs to identify "first-seen" or anomalous traffic patterns per device. It then leverages an LLM (Claude, Gemini, ChatGPT, or local Ollama) to generate human-readable explanations of what the traffic represents.

### Core Objectives
* **Contextual Awareness:** Move beyond binary "Allow/Block" lists.
* **Self-Hosted Privacy:** Keep the data ingestion and baseline logic local.
* **Provider Agnosticism:** Support multiple LLM backends (Cloud and Local).

---

## 2. System Architecture

### Components
1.  **Ingestion Engine (Go):** Polls the AdGuard Home `/control/querylog` API.
2.  **State Manager (Go/BoltDB):** Maintains a persistent "Baseline" of `(DeviceID + Domain)` pairs.
3.  **Analysis Orchestrator (Go):** Coordinates WHOIS enrichment and LLM requests.
4.  **Security Dashboard (React/Vite):** Visualizes anomalies and manages whitelist/blacklist actions.

---

## 3. Technical Requirements

### R1: Data Ingestion & Baselines (The "Memory")
* **Polling:** The service must poll the AdGuard Home API on a configurable interval (default 5-10s).
* **Deduplication:** Only process unique queries.
* **Persistence:** Use a lightweight KV store (BoltDB or SQLite). Store mappings of `Client_ID -> List[Domains]`.
* **Anomaly Definition:** Any query where the Domain has never been requested by that specific Client_ID in the past.

### R2: Enrichment & LLM Integration (The "Brain")
* **Enrichment:** For any flagged anomaly, perform a background `whois` lookup or use a public RDAP API to gather:
    * Registrar (e.g., NameCheap, GoDaddy).
    * Country of Registration.
    * Creation Date.
* **LLM Interface:** Implement a strategy pattern to support:
    * OpenAI (GPT-4o)
    * Anthropic (Claude 3.5 Sonnet)
    * Google (Gemini 1.5 Pro)
    * Ollama (Local Llama3/Mistral)
* **Prompting:** The LLM must be strictly instructed to return a valid JSON object.

### R3: API & Frontend (The "Interface")
* **REST API:** Go-based API to serve "Pending Anomalies" and "Baseline Stats."
* **UI Components:**
    * **Activity Feed:** Cards showing the Device, Domain, LLM Explanation, and Risk Score.
    * **Action Buttons:** * `Approve`: Add to baseline.
        * `Block`: Call AdGuard Home API to block the domain for that client.
    * **Settings:** Configure API keys and AdGuard credentials.

---

## 4. LLM Prompt Specification
The developer must use the following system prompt (or similar) to ensure consistency:

> **System Prompt:**
> You are a Network Security Analyst. Analyze the following DNS query from a home network.
> Context: Device: [Name], Domain: [Domain], WHOIS: [Data].
> 
> Respond ONLY in JSON format:
> {
>   "classification": "Safe" | "Suspicious" | "Malicious",
>   "explanation": "1-2 sentence description of what this domain is and why the device is calling it.",
>   "risk_score": 1-10,
>   "suggested_action": "Allow" | "Investigate" | "Block"
> }

---

## 5. Development Milestones

| Milestone | Deliverables | Checkpoint |
| :--- | :--- | :--- |
| **M1: Core Ingestor** | AdGuard API integration + BoltDB storage logic. | Log "First Seen" events to the terminal in real-time. |
| **M2: LLM Service** | Multi-provider client + WHOIS enrichment. | Receive a structured JSON explanation from a chosen LLM. |
| **M3: Web Dashboard** | React UI with anomaly feed and Approve/Block actions. | Fully "Approve" an anomaly via UI and see it persist in the DB. |
| **M4: Dockerization** | `docker-compose.yml` for easy deployment. | Run the entire stack with one command. |

---

## 6. Environment Variables
```env
AGH_URL=[http://192.168.1.2:8080](http://192.168.1.2:8080)
AGH_USER=admin
AGH_PASS=password
LLM_PROVIDER=ollama # or openai, anthropic,