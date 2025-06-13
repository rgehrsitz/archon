## Archon: MVP Scope Definition & Prioritization

Archon’s first public release (v1.0) must deliver a **compelling, stable, and focused set of capabilities** that validate the product vision, prove the plugin ecosystem, and delight early adopters—while avoiding scope creep. This document defines what is **in** and **out** of scope for the MVP, establishes priority tiers, and clarifies the initial industry personas we explicitly target.

---

### 1. Guiding Principles for MVP
1. **Core Value First** – Users must be able to **capture**, **version**, and **compare** physical‑asset configurations with zero‑friction setup.
2. **Plugin Foundation, Not Abundance** – Ship only the minimal plugin set required to prove extensibility; defer niche integrations.
3. **Single‑User Excellence** – Nail offline/local workflows before layering deep collaboration features.
4. **Polish Over Proliferation** – Fewer features finished to a high standard beat many half‑baked ones.

---

### 2. MVP Must‑Haves (🚀 Tier 0)
| Area | Capability | Notes |
|------|------------|-------|
| **Hierarchy Editor** | Create / edit / delete components; drag‑and‑drop re‑parenting | Single‑file storage default |
| **Snapshot System** | Manual snapshot creation (commit + tag); auto‑snapshot on large structural changes | Tag conflict handling, snapshot history UI |
| **Semantic Diff** | Compare any two snapshots; tree view with added/removed/modified highlights | Summary + per‑component details (lazy‑load) |
| **Storage** | Single‑file project layout (`components.json`, `attachments/`, `archon.json`) | Multi‑file mode migration tool deferred |
| **Attachments** | Add/remove files; Git LFS onboarding dialogue | Basic metadata; no thumbnail previews yet |
| **Plugin Runtime** | Front‑end JS/TS sandbox with permissions + PluginContext APIs | Signing optional in dev mode; verification CLI shipped |
| **Plugin Manager UI** | Install/enable/disable plugins from local file | Trust dialog for permissions |
| **Minimal Plugin Bundle** | See §4 | Bundled & enabled by default |
| **Testing & Integrity** | JSON Schema validation; unit tests for diff, snapshot, storage; CI workflow | Golden diff fixtures |
| **Basic UX Polish** | Keyboard navigation, focus outlines, dark/light theme with WCAG AA contrast | i18n hooks present but only English strings |

❗ Everything in Tier 0 must be **implemented, documented, and tested** before v1 ships.

---

### 3. Tier 1 Enhancements (post‑MVP, target v1.x)
| Feature | Rationale |
|---------|-----------|
| **Auto‑Snapshot interval & retention UI** | Quality‑of‑life once core snapshot works |
| **CSV/Excel Export Plugin** | Complements importer; low complexity |
| **Rack Visualizer Plugin** | Proves visualization extension point |
| **Basic Validation/Linter Plugin** | Enforce required properties; demo of Validation category |
| **Git Remote Push/Pull Buttons** | Enables manual sync without full branching |
| **Credentials stored in keyring** | Security hardening once remotes introduced |

---

### 4. Minimal Plugin Bundle (Ship with MVP)
| Plugin | Category | Purpose |
|--------|----------|---------|
| **CSV Importer** | Import | Jump‑starts data entry from existing spreadsheets |
| **JSON Exporter** | Export | Allows users to extract raw config for external tools |
| **Hello‑World Example** | Template | Boilerplate + step-by-step docs for third‑party authors; accelerates ecosystem adoption |

These three plugins are enough to **prove the API**, exercise Import/Export hooks, and give developers a clear starting example. The Hello‑World plugin doubles as onboarding scaffold with documentation for plugin developers.

---

### 5. Features Explicitly **Out of Scope for MVP**
- **Branching UI** (backend support present; UI toggle later)
- **Semantic Merge Driver & Conflict UI** (use Git textual merge + manual resolve for now)
- **Live Sync Plugins** (SNMP, OPC‑UA)—deliver after plugin ecosystem validated
- **Advanced Visualization Plugins** (network topology graph, floor plan)
- **Full i18n translations** beyond English; RTL support
- **High‑contrast / reduced‑motion toggles**
- **Local multi‑user accounts & role management**
- **Cloud sync backend / web collaboration**
- **Mobile or tablet UI**

---

### 6. Target Industries & Personas for v1
| Persona | Sector | Why v1 Focus? |
|---------|--------|---------------|
| **Lab Technician** | Scientific & R&D labs | Small teams, immediate need for repeatable setups; spreadsheets today |
| **Test Engineer** | Industrial / Automotive test rigs | Gains from snapshot/diff; CSV import fits existing data |
| **Museum Technician** | Exhibit maintenance teams | Straightforward hierarchies; attachments for manuals/photos |
| **Small Data‑Center Operator** | SMB IT | Simplified inventory; interested in diff & JSON export |

These personas share **simple hierarchies** and rely heavily on **manual documentation today**, making Archon’s MVP feature set immediately valuable.

---

### 7. Success Criteria for MVP (Exit Criteria)
1. **User Scenario Completion**: Each target persona can install Archon, create a project, import via CSV, attach a photo/PDF, take snapshots, compare diffs, and restore a snapshot—*without using the command line*.
2. **Plugin Proof**: Third‑party developer can clone Hello‑World plugin, modify, sign (optional), load into Archon, and execute with expected result.
3. **Stability**: Zero blocker or critical bugs after two‑week beta with internal testers; **qualitative feedback**: ≥ 80% of beta users report “satisfied” or “very satisfied” with core workflows.
4. **Performance**: Load project with 2 000 components in ≤ 3 seconds; diff two such snapshots in ≤ 1 second.
5. **Documentation**: Quick‑start guide, plugin authoring guide, and API reference published.
6. **Beta Adoption Metrics**: Onboard at least X beta users and gather feedback showing Archon addresses key pain points (quantitative and qualitative surveys).

---

### 8. Implementation Timeline (high‑level)
| Phase | Length | Milestones |
|-------|--------|------------|
| **Foundation** | 4 weeks | Storage, snapshot, diff, canonical JSON, unit tests; track **technical debt** items for deferred features |
| **Plugin Runtime** | 2 weeks | Sandbox, PluginContext, CSV importer dev |
| **UI Core** | 3 weeks | Hierarchy editor, snapshot history, diff viewer |
| **Polish & Docs** | 2 weeks | Accessibility hooks, example plugin, docs; plan post-MVP backlog |
| **Beta + Bugfix** | 3 weeks | Internal testing, performance tuning, qualitative survey |

---

### 9. Risk & Mitigation
| Risk | Impact | Mitigation |
|------|--------|-----------|
| **Scope Creep** | Delays & instability | Strict Tier 0 list; any new request requires triage vs. Tier 0 |
| **Plugin Runtime Complexity** | Security/performance issues | Frontend‑only JS/TS plugins initially; backend plugins deferred |
| **Performance on Large Hierarchies** | Poor UX | Early profiling; golden performance tests |
| **Diff Edge‑Cases** | Incorrect comparisons | Golden files and fuzz tests on diff engine |
| **Technical Debt Accumulation** | Slows future velocity | Document deferred features; allocate capacity in post-MVP sprints; maintain modular architecture |

---

*End of MVP Scope Definition & Prioritization (Incorporated qualitative feedback, technical debt tracking, and clarified Hello‑World purpose)*

