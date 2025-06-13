## Archon: Collaboration & Multi-User Design

Collaboration and multi-user support are strategic for Archon’s growth. While initial releases focus on single-user workflows, designing with collaboration in mind prevents costly rework. This section outlines:
1. Branching vs. Linear-Only Commit Flows
2. Conflict Detection & Resolution Strategies
3. Multi-User Tagging, Authorship Tracking, Change Attribution
4. Integration with Git Hosting Providers (GitHub, GitLab, etc.)
5. Permissions, Access Control, and Roles (MVP focus on remote model)
6. Collaboration UX Considerations
7. Implementation Guidance & Testing

---

### 1. Branching vs. Linear-Only Commit Flows

#### 1.1 Hybrid Model: Linear Default, Branch-Aware Backend
- **Backend Preparedness**: The Git Abstraction Layer in Go must include functions for creating, listing, checking out, and merging branches from day one, even if UI initially hides them. This ensures future branch support without major rewrites.
- **Linear-Only UI Initially**:
  - Default user experience shows a single chronological history of snapshots (commits), labeled clearly as “Snapshots”. Non-technical users see a simple timeline.
  - Underlying Git supports branching; advanced users may create branches externally via Git CLI, and Archon can detect and allow merging via UI.
- **Opt-In Branching Mode**:
  - Provide a setting or advanced toggle (“Enable Branching Mode”) to expose branch-related UI when users are ready for parallel workflows.
  - Until enabled, hide branching elements to avoid confusion for casual users.
- **Draft & Restore Workflow**:
  - Encourage manual or auto-snapshots before experimenting. Users can restore a previous snapshot to branch externally or via UI when branching becomes available.

#### 1.2 Branch Creation & Switching (Advanced/Opt-In)
- **Create Branch from Snapshot**:
  - Dialog: user selects a snapshot, enters branch name. Underlying Git: `git branch <name> <snapshot-ref>`.
  - UI feedback: confirm branch created, offer to switch to it.
- **Switch Branch**:
  - When branching mode enabled, allow user to choose active branch from a dropdown. On switch, load the tip state of that branch.
  - Warn about unsaved changes before switching.
- **Branch Listing**:
  - Show list of branches in a simple list or minimal graph, with current branch highlighted.

#### 1.3 Merge (Advanced/Opt-In)
- **Merge Initiation**:
  - User selects “Merge branch” action, chooses source and target branches. UI clearly states source and target names.
- **Merge Preview**:
  - Before merge, perform semantic pre-merge: auto-merge non-conflicting edits, identify conflicting components/properties.
  - Show summary: e.g., “Auto-merged X components; Y conflicts need resolution.”
- **Merge Execution**:
  - Launch conflict resolution UI (see Section 2) if conflicts exist; otherwise auto-commit with descriptive message.
- **Post-Merge**:
  - Validate merged state against schema. If invalid, prompt user to correct before final commit.
  - Commit merged snapshot with message like “Merge branch 'feature-xyz' into 'main'” plus summary.

---

### 2. Conflict Detection & Resolution Strategies

#### 2.1 Semantic Conflict Detection
- **Pre-Merge Semantic Merge**:
  - Use JSON-aware merge: parse component hierarchies to auto-merge non-overlapping edits. For example, if users edited different properties of the same component, merge automatically.
  - Identify true conflicts where the same property or hierarchy position was modified differently.
- **Contextual Conflict Metadata**:
  - For each conflict, gather context: local change details (“Your change: set firmware_version to 1.3.0”), remote change details (“Incoming change from Alice Smith: set firmware_version to 1.2.5”). Include author name/email and snapshot tags.

#### 2.2 Conflict Resolution UI
- **Enhanced Three-Pane Interface**:
  - **Left Pane**: Incoming/Remote version, annotated with author and timestamp.
  - **Right Pane**: Local version, annotated with author (or “Your change”) and timestamp.
  - **Center Pane**: Merged result, updated live as user selects resolutions.
- **Property-Level Context**:
  - Show human-centric descriptions: e.g., “Your change: set ‘firmware_version’ to '1.3.0'”, “Remote (Alice Smith, June 5, 2025): set ‘firmware_version’ to '1.2.5'.”
  - For moved components, indicate original and new parent contexts.
- **Conflict Navigation & Batch Actions**:
  - Sidebar lists all conflicts; clicking navigates to detailed resolution pane.
  - Option to apply consistent resolution patterns (e.g., “Prefer remote for property X across all conflicts”).
- **Validation & Commit**:
  - After resolutions, validate merged JSON. If issues arise, highlight problematic components.
  - On commit, include conflict summary in commit message.

#### 2.3 Textual Fallback & Markers
- If semantic merge cannot fully handle complex cases, fall back to showing raw JSON conflict markers in a text view as last resort, but prioritize semantic resolution UI.

#### 2.4 Merge from External Branches
- Detect external branches created via CLI and allow merging via same semantic process in UI.

---

### 3. Multi-User Tagging, Authorship Tracking, Change Attribution

#### 3.1 Per-Component Blame & History
- **Last-Change Metadata**:
  - On each snapshot commit, record for each component/property the modifying user and timestamp (from Git author). Store in component metadata: `last_modified_by`, `last_modified_at`, and the snapshot tag or commit hash.
- **Blame UI**:
  - In component detail panel, display a small history icon next to each property. Clicking opens a popover listing chronological changes: author, timestamp, snapshot/tag, previous value → new value.
- **Component History View**:
  - Allow user to view full history of a component: list of snapshots where it changed, with details.
- **Snapshot Labels & Annotations**:
  - Support user-defined labels or annotations on snapshots (e.g., “QA Review”, “Production Release”). Store in snapshot metadata.
- **Contributors Summary**:
  - On merge commits, list contributors merged (from Git commit metadata).

#### 3.2 Audit Reports
- **Change Reports**:
  - Generate a report summarizing changes between two snapshots for selected components, including who changed what and when.
- **User Activity Logs**:
  - Aggregate per-user changes across project: e.g., “Alice modified 5 components on June 5, 2025.” Useful for auditing.

---

### 4. Integration with Git Hosting Providers

#### 4.1 Remote Setup & Basic Sync (MVP)
- **Remote Configuration UI**:
  - Simple dialog to add/edit remote repository URL. Use secure credential retrieval from OS keyring.
- **Push/Pull Buttons**:
  - Clear “Push” and “Pull” operations in UI. On Pull, detect remote changes and prompt user to merge if needed.
- **Conflict Flow**:
  - After Pull, if changes exist, run semantic merge pre-check; if conflicts, enter Conflict Resolution UI.

#### 4.2 Pull Request Workflows (Future)
- **Open Pull Request Action**:
  - After branch creation, offer “Open Pull Request” which opens hosting provider in browser pre-filled with branch and base.
- **Review Status Display**:
  - Optionally show PR build/CI status badges in Archon UI when user toggles advanced collaboration mode.

#### 4.3 CI/CD Triggers (Future)
- **Snapshot Tags & CI**:
  - When pushing snapshot tags, external CI workflows can be triggered via hosting provider hooks. Archon can optionally display summary of CI results.

#### 4.4 Permissions and Access Control via Hosting
- **Leverage Hosting Provider**:
  - Rely on Git hosting for push/pull permissions; Archon UI respects errors returned by remote when lacking permissions.
- **Local Roles Deferred**:
  - For MVP, skip implementing local user roles; rely entirely on remote repository access control.

---

### 5. Permissions, Access Control, and Roles

#### 5.1 MVP Focus on Remote Model
- **Skip Local Account System Initially**:
  - Rely on Git hosting provider for collaboration permissions (push/pull rights). Simplifies implementation.
- **Future Local Roles**:
  - Optionally in advanced deployments, implement local user accounts and roles (admin, contributor, viewer) with encrypted storage, but defer until demand arises.

---

### 6. Collaboration UX Considerations

#### 6.1 Simplified History & Branch Visualization
- **Linear View Default**:
  - Display snapshots in chronological order with labels. Hide branching elements unless advanced mode enabled.
- **Branch-Aware View (Opt-In)**:
  - Show simplified branch graph or list of branches. Keep UI minimal: avoid complex graphs; focus on functional clarity.

#### 6.2 Guided Merge & Pull Workflows
- **Incoming Changes Notification**:
  - When remote changes exist, show prompt “New changes available from remote. Pull now?” with likely impact (e.g., number of snapshots ahead).
- **Merge Preview**:
  - Before merging, display summary: auto-merged items and conflicts needing resolution.

#### 6.3 Blame & History UI
- **Inline History Icons**:
  - Small icon next to each property in detail pane opens change history popover.
- **Filterable Snapshot List**:
  - Allow filtering snapshots by author, date, or label to navigate collaborative history.

#### 6.4 Notifications & Integrations
- **In-App Notifications**:
  - Notify user when operations complete (push success) or require action (pull conflicts).
- **External Notifications (Future)**:
  - Plugins can send notifications (e.g., Slack) when snapshots or merges occur.

---

### 7. Implementation Guidance & Testing

#### 7.1 Git Abstraction Layer
- **Branch Support from Day One**:
  - Implement functions: CreateBranch(name, fromRef), ListBranches(), CheckoutBranch(name), MergeBranch(source, target), DeleteBranch(name). Ensure semantic merge logic integrates here.
- **Testing**:
  - Unit tests for branch operations using ephemeral repos. Test operations both via CLI execution and Go Git library if used.

#### 7.2 Conflict Resolution Testing
- **Semantic Merge Tests**:
  - Create test scenarios with divergent edits: different property edits auto-merge, conflicting edits prompt resolution. Use golden fixtures for expected merged outputs.
- **Contextual Metadata**:
  - Test that conflict UI receives correct metadata (author, timestamp, previous values).

#### 7.3 Authorship & Audit Testing
- **Simulate Multiple Git Authors**:
  - In tests, configure commits with different author.name/email and verify per-component metadata updates and blame UI data.
- **Audit Log Entries**:
  - Verify logging of merge events with contributor lists.

#### 7.4 Git Hosting Integration Testing
- **Mock Hosting APIs**:
  - Use stubs for GitHub/GitLab API calls when testing PR creation. Verify correct URLs and parameters.
- **Credential Flows**:
  - Test push/pull with valid and invalid credentials; ensure UI displays clear error messages.

#### 7.5 UX & Usability Testing
- **User Studies**:
  - Conduct usability sessions with non-technical and technical personas to validate simplicity of linear flows and clarity of opt-in branching.
- **E2E Automated Tests**:
  - Automate new project workflows: set remote, push initial snapshot, pull changes, trigger conflict scenarios, resolve via UI.

#### 7.6 Security in Collaboration
- **Credential Security**:
  - Test that credentials used for remote operations are retrieved from keyring and not exposed.
- **Permission Enforcement**:
  - Verify that Archon handles remote errors (e.g., insufficient permissions) gracefully and informs user.

---

*End of Collaboration & Multi-User Design Deep Dive (Refined with Hybrid Model, Contextual Merge UI, Blame Icon, MVP Remote Focus)*

