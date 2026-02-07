# AGENT.md

This document defines **rules, constraints, and intentions** for any developer or AI agent
working on this repository.

The goal is to keep the project **coherent, evolvable, and aligned with its original purpose**.

---

## 🎯 Project Intent

This project is a **training-oriented railway control center simulator**, not a game.

Key intentions:

- Training use over entertainment
- Deterministic, explainable simulation
- Multiple dispatchers operating on a shared state
- Safety and correctness over visual fidelity

Any change must respect these principles.

---

## 🧠 Architectural Principles (Must Follow)

### Layered Architecture (DDD-inspired)

Strict separation of concerns:

- **Domain**
  - Pure business rules and invariants
  - No HTTP, JSON, DB, time.Now, random, or framework dependencies
- **Application (UseCase)**
  - Orchestrates domain behavior
  - Returns DTOs, not domain entities
- **Infrastructure**
  - Implements repositories and external services
- **Presentation**
  - HTTP / WebSocket / UI adapters
  - Must not touch domain objects directly
- **DI**
  - Dependency wiring only
  - No business logic

Violating layer boundaries is considered a bug.

---

## 🚫 Forbidden Actions

The following are explicitly **not allowed** without redesign discussion:

- Introducing Event Sourcing or CQRS prematurely
- Returning domain entities from UseCases
- Adding persistence logic to domain or application layers
- Letting repositories generate business IDs
- Making HTTP handlers contain business rules
- Introducing game mechanics (scores, randomness, fun-first logic)

---

## 🚆 Simulation Model Rules

### Line & Stations

- Line shape is **linear**
- Stations are represented as **block boundaries**
- No "station entity" with position/state
- Blocks are the only occupiable units (fixed block system)

Station -- Block -- Station -- Block -- Station


### Trains

- Trains always exist **on a block**
- Trains never "exist at a station"
- Train position is expressed as:
  - current block
  - progress (0.0–1.0)
  - direction
- Speed is constant (for MVP)

### Station Logic

- Station events occur **only when a train reaches a block boundary**
- Departure permission is checked **only when entering the next block**
- Waiting at stations is represented by clamping progress to the boundary

---

## ⏱️ Time Handling

- Simulation advances only via explicit `Tick(dt)`
- No implicit time progression
- No direct `time.Now()` usage in domain
- If time is needed in UseCases, inject a Clock abstraction

---

## 🔐 Invariants (Must Not Be Broken)

- One block may be occupied by **at most one train**
- A train may not enter an occupied block
- A train may not depart a station without permission
- Terminal stations stop trains permanently (unless redesigned)

Any code that bypasses these checks is invalid.

---

## 🧩 Repository Rules

- Repository interfaces belong to the **domain layer**
- Repositories must not:
  - Generate IDs
  - Apply business rules
  - Mutate domain state implicitly
- `Get / Create / Save` semantics must be respected

---

## 🧪 Testing Philosophy

- Domain logic must be testable without HTTP or infrastructure
- UseCases should be testable with in-memory repositories
- Simulation behavior should be deterministic under fixed inputs

---

## 🤖 AI Agent Guidelines

When acting as an AI agent on this repository:

- Prefer **clarity over cleverness**
- Do not introduce abstractions unless there is a concrete need
- Ask before restructuring core domain concepts
- Assume training correctness > performance > convenience

If unsure, **do less, not more**.

---

## 🧭 Evolution Strategy

This project is expected to grow in this order:

1. Time-based simulation (`Tick`)
2. WebSocket state synchronization
3. Logging / replay
4. Persistence
5. Line branching / points

Skipping ahead in this list is discouraged.

---

## 📌 Status

- MVP architecture is stable
- Changes should be incremental and justified

---

## 📝 Final Note

This codebase is designed to be **understandable first, extensible second**.

If a change makes the system harder to reason about,
it is probably the wrong change.
