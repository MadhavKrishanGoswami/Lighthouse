# Lighthouse Architecture

This document provides a detailed overview of the Lighthouse application's architecture, its components, and how they communicate.

---

## 1. Core Philosophy

Lighthouse is a lightweight, self-hosted solution for monitoring and automatically updating Docker containers across multiple hosts. Its design is guided by the following principles:

* **Simplicity**: Every component has a single, clear responsibility.
* **Efficiency**: Communication is built on gRPC, using bidirectional streaming for real-time, low-latency updates.
* **Decentralized Execution**: A central Orchestrator manages state and dispatches commands, while container operations are executed by agents on each host.
* **Clear Contracts**: Protocol Buffers (`.proto` files) define strict API contracts, ensuring type safety and clarity between services.
* **Source of Truth**: A PostgreSQL database stores all state data, ensuring consistency and reliability across the system.

---

## 2. High-Level Architecture

The Lighthouse system consists of four primary services, a database for persistence, and external dependencies like the Docker daemon and container registry.

### Architecture Diagram
<img width="3840" height="2900" alt="Flowchart Lighthouse Sept 10 2025" src="https://github.com/user-attachments/assets/f7ec1eca-a42d-4546-a853-0de76e6c3815" />

---

## 3. Component Deep Dive

### Orchestrator

The **Orchestrator** is the brain of Lighthouse. It does not directly perform container operations but coordinates all other services.

**Responsibilities:**

* **State Management**: Maintains host and container states in PostgreSQL, treating it as the source of truth.
* **Agent Communication**: Manages registration and heartbeat RPCs, maintaining bidirectional gRPC streams with Host Agents.
* **TUI Communication**: Streams host and container data to the TUI, while also accepting configuration commands.
* **Update Coordination**: Based on cron schedules, queries the database for containers under watch and tasks the Registry Monitor to check for updates.

---

### Host Agent

The **Host Agent** is a lightweight client running on every managed host.

**Responsibilities:**

* **Host Registration**: Registers itself with the Orchestrator by sending system information and container details.
* **Docker Interaction**: Translates commands from the Orchestrator into `docker pull`, `docker stop`, and `docker run`.
* **Persistent Connection**: Maintains a persistent gRPC stream with the Orchestrator for real-time commands and status updates.
* **Status Reporting**: Periodically sends heartbeats and update progress (e.g., `PULLING`, `STARTING`, `FAILED`) back to the Orchestrator.

---

### Registry Monitor

The **Registry Monitor** is a microservice dedicated to checking for new container image versions.

**Responsibilities:**

* **Update Detection**: Receives a list of container images from the Orchestrator and checks for new tags or digests.
* **Registry Communication**: Queries external registries like Docker Hub to find updates.
* **Reporting**: Returns a list of images with available updates to the Orchestrator.

---

### Lighthouse TUI

The **TUI (Text-based User Interface)** is the control panel where users interact with Lighthouse.

**Responsibilities:**

* **Real-time Visualization**: Receives streams of host and container data to display live status updates.
* **User Interaction**: Allows toggling of container watches and setting cron schedules.
* **Configuration**: Enables users to update system settings such as the frequency of update checks.

---

### PostgreSQL Database

The **PostgreSQL database** acts as the persistent storage layer and source of truth.

**Data Stored:**

* Host information (IP, MAC address, etc.)
* Container configurations
* Auto-update watch settings
* Update logs and system settings

**Why PostgreSQL?**

* Reliable, robust, and structured storage
* Ensures consistency and integrity of data across services

---

## 4. Communication Protocol (gRPC)

Lighthouse uses gRPC for all internal communication, enabling fast, typed, and stream-capable RPCs defined via `.proto` files.

### Services and Methods

#### **HostAgentService**

* `RegisterHost` (unary): Agent sends initial state to Orchestrator.
* `Heartbeat` (unary): Periodic alive signals with container status.
* `ConnectAgentStream` (bidirectional stream): Orchestrator sends commands, Agent streams update statuses.

#### **RegistryMonitorService**

* `CheckUpdates` (unary): Checks for new image versions.

#### **TUIService**

* `SendDatastream` (server-streaming): Orchestrator streams live data to TUI.
* `SetWatch` / `SetCronTime` (unary): TUI sends configuration commands.

---

## 5. Core Workflow: The Automated Update Cycle

The primary flow of Lighthouse is the automated detection and deployment of container updates.

### Sequence Diagram
<img width="3173" height="3840" alt="Lighthouse Sequence Diagram Sept 10 2025" src="https://github.com/user-attachments/assets/8f32e4cd-b99e-4336-afd4-7231b056265f" />


