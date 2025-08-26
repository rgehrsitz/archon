# Archon v1 — UI/UX Screen Blueprint (Wails + Svelte 5)

> This blueprint translates the product spec and ADRs into a concrete, desktop‑class interface for Archon. It prioritizes clarity, speed at 10k–50k nodes, and approachable Git‑backed workflows (Snapshot/History/Sync). It is written to hand directly to engineers and designers.

---

## 0) Guiding Principles

* **Hierarchy first.** The hierarchy is the product. Editing nodes and seeing context must be effortless.
* **Snapshots, not branches.** Keep the in‑app model linear and approachable: *Snapshot • History • Sync*.
* **Semantic over textual.** Diffs/merges speak the domain language: moves, renames, property deltas.
* **Fast everywhere.** Virtualize large lists, prefetch, and keep main interactions <50ms when possible.
* **Safe extensibility.** Plugins are sandboxed with explicit consent and clear capabilities.
* **Desktop‑class feel.** Menus, context menus, drag and drop, keyboard power, resizable panes.

---

## 1) Design System & Tech (Svelte 5)

* **Styling:** Tailwind v4 with CSS variables (`@theme`) and light/dark modes.
* **Components:** shadcn‑svelte (dialogs, command palette, resizable, menus) + Bits/Melt headless for custom trees.
* **Icons:** Lucide.
* **Virtualization:** TanStack Virtual (trees/lists). Fallbacks available.
* **Tokens:** Semantic color/space/typography tokens; density set to compact‑by‑default with a “cozy” option.

---

## 2) Navigation Model

* **Primary Shell:** A three‑pane *Hierarchy Workbench* with a universal **top command bar** and **status bar**.
* **Global Surfaces:**

  * **Command Palette** (`⌘K / Ctrl+K`): actions, nodes, snapshots, plugins.
  * **Left App Sidebar:** Project switcher, Recent projects, global sections.
  * **Right Panels Dock:** Inspector, Attachments, History, Plugins, Logs.
* **Secondary Screens:** Project Dashboard, History/Snapshots, Diff/Merge, Import Wizard, Plugin Manager, Settings.

---

## 3) Screen Catalog (14 screens)

### S1. Welcome & Project Switcher

* **When:** App launch / `⌘O`.
* **Layout:** Left column (Recent projects), right column (Open/Clone/Create cards) with drag‑and‑drop for folders/URLs.
* **Details:** Shows last opened, Git status badges (ahead/behind), theme toggle, quick tips. Supports “Open Recent” context menu.

### S2. Create / Open Project Wizard

* **Steps:** Mode → Location → (optional) Git Remote + LFS → Initialize.
* **Affordances:** Detects if Git/LFS available; offers to enable LFS at first attachment use. Shows policy notes for proxy/secrets if detected.

### S3. **Project Dashboard** (Overview)

* **Purpose:** Snapshot of health and activity.
* **Cards:** Recent Snapshots, Pending Sync, Recent Changes (last N commits), Index Health, Plugins activity, Quick Actions (Import CSV, New Node, Create Snapshot).
* **Empty state:** Friendly guidance with links into Workbench.

### S4. **Hierarchy Workbench** (Primary)

* **Layout:**

  * **A. Miller Columns** (3–5 adaptive columns) for path‑based navigation with **virtualized** rows and lazy children loading.
  * **B. Details Inspector** (right dock): Name, Description, Properties editor (typed inputs), Attachments, Quick Links, Validation.
  * **C. Secondary dock tabs:** History, Attachments browser, Plugin Panels, Logs.
  * **D. Top command bar:** Breadcrumbs, omnibox search (`/`), view modes, snapshot button, sync menu.
  * **E. Status bar:** Project path, index state, Git remote, background tasks.
* **Interactions:**

  * Click/keyboard to traverse columns, **multi‑select** with Shift/Ctrl, drag‑and‑drop reorder/move (with conflict guards).
  * Context menu per node (Rename, Duplicate, Move, Add… , Snapshot subtree, Create Jira issue, etc.).
  * Inline create (`Enter`), inline rename (`F2`), inline add property.
  * **Quick‑Edit Drawer** (optional) for batch edits without leaving context.
* **Assistive:** Inline validation (names unique among siblings), warnings if depth >64, property type hints.

### S5. Properties & Attachment Inspector (Dock Tab)

* **Fields:** Typed editors (string/number/bool/date/attachment). Attachment subpanel for preview, replace, verify hash, open external.
* **UX:** Dirty state badges; undo/redo per field; copy as JSON; keyboard nav across fields.
* **Affordances:** “Promote to Template” (later), “Copy path,” “Open in Finder/Explorer.”

### S6. Global Search & Filters (Top Command Bar)

* **Search box** with:

  * **FTS** across names/properties.
  * **Structured filters:** `key:value`, `type:int`, `has:attachment`, `path:/Lab A/*`.
  * **Saved Searches** with pinning.
* **Results:** Inline dropdown (top 8) + full Search panel (virtualized). Actions on result (Go, Peek, Add to selection).

### S7. Snapshots & History

* **View:** Linear timeline with Snapshot tags, authors, messages. Filter: “Snapshots only,” “All commits.”
* **Actions:** Create Snapshot (`⌘S`), tag rename (if allowed), notes/labels, restore as working set (read‑only preview vs checkout).

### S8. **Semantic Diff Viewer**

* **Compare:** Snapshot ↔ Snapshot, Snapshot ↔ Working.
* **Presentation:** Change list grouped by type (Renames, Moves, Property changes, Adds/Deletes). Clicking focuses the Workbench with change highlights.
* **Details panel:** Before/after values, path, user, timestamp. Toggle to text diff for power users.

### S9. **Merge Conflict Resolver**

* **Layout:** Left (Theirs), Right (Ours), Middle chooser per conflicting field. Visual widget for sibling reorder conflicts.
* **Batch tools:** “Prefer Ours/Theirs” per section, “Accept latest timestamp,” “Accept non‑null.”
* **Safety:** Dry‑run preview; final summary before commit.

### S10. Import Wizard (Plugin‑driven)

* **Steps:** Select plugin → Validate → Preview mapped tree → Choose merge target (new project or node) → Apply (as draft) → Review → Snapshot.
* **UX:** Row sampling preview, schema hints, per‑run network consent if plugin requests it, progress with cancel.

### S11. **Plugin Manager & Permissions**

* **Tabs:** Installed, Discover (later/gated), Permissions.
* **Installed:** Enable/disable, version, kinds (Importer/Panel/Validator/etc.), access badges (readRepo, writeRepo, net, attachments…).
* **Permissions:** Grant/revoke with expiration, scope display (e.g., `secrets:jira*`).
* **Panels:** Panel‑type plugins appear as dockable tabs within the Workbench.

### S12. Settings

* **Sections:** Appearance (theme, density), Git (remote, SSH/PAT helper), LFS, Index (rebuild), Proxy/Secrets policy visibility, Shortcuts, Logging (open folder), About.
* **Danger Zone:** Reset cache/index, clear secrets (with confirmations).

### S13. Attachments Manager

* **Grid:** Attachment file list with type, size, referenced‑by count, hash verification, “Open Containing Node,” LFS status.
* **Bulk:** Verify all, GC unused (dry‑run), rehash, export selection.

### S14. Notifications & Log Center

* **Drawer:** Background operations with progress (clone, import, index rebuild). Persist failures with actions (Retry, Open Logs, Copy error).

---

## 4) Key Flows (maps)

### Flow A — Edit → Snapshot → Sync

1. User edits properties (Inspector) → autosave journal.
2. Press **Snapshot** → name + description → commit+tag.
3. **Sync** menu shows ahead/behind → Pull (if needed) → Push.

### Flow B — Import CSV → Preview → Merge → Snapshot

1. Import Wizard → CSV plugin → validate + preview.
2. Choose target node → apply as draft (ephemeral branch/working set).
3. Review in Diff Viewer → Snapshot.

### Flow C — Conflict Resolution

1. Pull reveals conflict → open **Merge Resolver**.
2. Per‑field decisions → summary → commit.
3. Return to Workbench with change highlights.

---

## 5) Micro‑Interactions & Shortcuts

* **Global:** `⌘K` command palette, `/` focus search, `⌘N` new node, `⌘S` snapshot, `⌘⇧S` snapshot + push, `⌘Z/⇧⌘Z` undo/redo.
* **Tree:** Arrow keys navigate, `→` expand, `←` collapse, `Enter` open, `F2` rename, `Delete` delete.
* **Workbench:** `Alt` while dragging to **duplicate** node; `Ctrl` to **link as reference** (future DAG).
* **Inspector:** `Tab/Shift+Tab` field nav; `⌘C` copies selected property as `key=value`.

---

## 6) Performance Patterns

* Virtualize rows at every tree level; windowed rendering.
* Lazy load children on expand; prefetch adjacent ranges.
* Debounced index queries; progressive result append.
* Background index rebuild on clone/import; non‑blocking UI.

---

## 7) Reliability & Error UX

* Unified error envelope → friendly messages with “More details…”.
* Long operations: cancellable progress, automatic rollback on failure.
* Autosave dirty nodes; recovery prompt on restart.
* Snapshot create/push show toast + timeline insertion.

---

## 8) Accessibility & Internationalization

* Tree uses ARIA `treegrid` patterns; full keyboard control.
* Contrast‑safe tokens; focus outlines; prefers‑reduced‑motion respected.
* English‑first, string externalization scaffold; UTF‑8 content throughout.

---

## 9) Theming & Visual Language

### 9.1 Brand Accents (updated)

* **Primary accent (Cobalt):**

  * Hex: **#2A5ADA** (cobalt)
  * Usage: primary buttons, selection, links, focus rings, key icons.
  * Foreground on accent: **#FFFFFF** (ensures WCAG AA on buttons/links).
* **Palette pairing:** neutral grays + off‑whites for balance.

### 9.2 Density & Modes

* **Default density:** **compact** (desktop‑first). Toggle **cozy** via `data-density="cozy"` on `<html>`.
* **Themes:** light/dark via `data-theme="light|dark"` on `<html>`.

### 9.3 Token Mapping (CSS variables consumed by Tailwind utilities)

```css
/* app.css (excerpt) */
:root {
  /* Light mode: white bg, soft gray panels, cobalt accents */
  --bg:            255 255 255;     /* #FFFFFF */
  --surface:       246 248 250;     /* #F6F8FA panels */
  --fg:            15 18 20;        /* #0F1214 text */
  --muted-fg:      102 112 133;     /* #667085 subtle text */
  --border:        229 231 235;     /* #E5E7EB */

  --accent:        42 90 218;       /* #2A5ADA cobalt */
  --accent-foreground: 255 255 255; /* #FFFFFF */

  --success:       16 185 129;      /* #10B981 */
  --warn:          234 179 8;       /* #EAB308 */
  --error:         239 68 68;       /* #EF4444 */
}

:root[data-theme="dark"] {
  /* Dark mode: charcoal bg, medium gray surfaces, cobalt highlights */
  --bg:            18 20 23;        /* #121417 */
  --surface:       27 31 36;        /* #1B1F24 */
  --fg:            225 229 234;     /* #E1E5EA subtle light gray text */
  --muted-fg:      160 165 170;     /* #A0A5AA */
  --border:        44 49 56;        /* #2C3138 */

  --accent:        91 140 255;      /* #5B8CFF cobalt highlight */
  --accent-foreground: 15 18 20;    /* #0F1214 for chips/badges on accent */
}

html { font-size: 14px; }
html[data-density="cozy"] { font-size: 15px; }

/* Utility helpers (Tailwind v4 arbitrary color syntax) */
.btn-accent { background: rgb(var(--accent)); color: rgb(var(--accent-foreground)); }
.panel { background: rgb(var(--surface)); border-color: rgb(var(--border)); }
.text-muted { color: rgb(var(--muted-fg)); }
```

> Tailwind classes can reference these via `bg-[rgb(var(--accent))]`, `text-[rgb(var(--fg))]`, etc. Contrast meets or exceeds AA for large text/UI elements in both themes; primary button text on cobalt is white in light mode and deep charcoal in dark mode.

## 10) Component/Route Map (Svelte)


``` (Svelte)
src/routes/
+layout.svelte            // app shell, menus, status bar, command palette
/welcome/+page.svelte     // S1
/project/\[id]/dashboard   // S3
/project/\[id]/workbench   // S4–S6 docked panels
/project/\[id]/history     // S7
/project/\[id]/diff        // S8
/project/\[id]/merge       // S9
/project/\[id]/import      // S10
/settings                 // S12
/plugins                  // S11

src/lib/components/
workbench/MillerColumns.svelte
workbench/TreeColumn.svelte
workbench/InspectorPanel.svelte
workbench/AttachmentsPanel.svelte
workbench/HistoryPanel.svelte
diff/SemanticDiffViewer.svelte
merge/ConflictResolver.svelte
plugins/PermissionDialog.svelte
common/CommandBar.svelte
common/CommandPalette.svelte
common/ToastCenter.svelte
```

---

## 11) Data & Git Concepts → UI Mapping

- **UUIDv7 IDs:** stable across moves → enables Move/Rename semantics.
- **Sharded /nodes/<id>.json:** tiny diffs, partial saves → fast Undo/Redo and merges.
- **Snapshots:** commit+tag pairs → visible as “Snapshots” in timeline.
- **SQLite FTS index:** instant find; supports saved searches.
- **LFS Attachments:** transparently handled; visual LFS status in Attachment Manager.
- **Credential helpers:** OS‑native auth, inline prompts in Sync menu.

---

## 12) Empty States & Skeletons

- Welcome (no projects), Dashboard (no activity), Workbench (no selection), Diff (no changes), Plugins (none installed).
- Skeletons for tree columns and inspector on initial load.

---

## 13) Future (v2+) Considerations

- **Multi‑window** (Wails v3): open Diff/Merge in separate window; large monitor workflows.
- **DAG/cross‑links:** convert Tree to Graph explorer; breadcrumbs as paths, not strict parentage.
- **Background Sync** and snapshot policies; approvals via Validators.
- **Plugin Gallery** with signing/verifications and ratings.

---

### Appendix A — Visual Behaviors (quick notes)
- Column widths are resizable with snap points; double‑click to auto‑fit.
- Drag‑drop shows a *ghost* with target path; illegal drops shake + tooltip.
- History uses color accents for tags; tooltips show author+time+message.
- Diff highlights affected nodes in the columns (badges on ancestors).
- Merge screen constrains height of long property lists; sticky header with summary.

### Appendix B — Example Permission Copy (Plugin)
- **readRepo**: “Let this plugin read nodes and properties.”
- **writeRepo**: “Allow edits and commits (creates snapshots).”
- **attachments**: “Read and write files attached to nodes.”
- **net**: “Make network requests (scoped by policy).”
- **secrets:jira***: “Use your stored Jira credentials to pull/push issues.”

---

**Deliverable ready**: Brand accents and density defaults applied. Below, you’ll find low‑fidelity wireframes for the key screens and Svelte starter stubs (routes + core components) to accelerate implementation.



---

## 15) Low‑Fidelity Wireframes (key screens)

> These are ASCII layout specs that mirror the interactions; engineers can translate directly into Svelte components.

### 15.1 Workbench (S4 — Miller Columns + Inspector)
```

┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│ Command Bar:  breadcrumb / \[ / Search… ]  View: Miller ▾  Snapshot ⌘S  Sync ⌃⌘S             │
├───────────────┬───────────────────────────┬───────────────────────────┬───────────────────────┤
│ Column A      │ Column B                  │ Column C                  │ Inspector (dock)      │
│ \[Sites]       │ \[Lab A]                   │ \[Bench 3]                 │ ┌───────────────────┐ │
│  ▸ Lab A      │  ▸ Rooms                  │  ▸ Oscilloscope 1         │ │ Name              │ │
│  ▸ Lab B      │  ▸ Benches                │  ▸ Power Supply           │ │ Description       │ │
│  ▸ Storage    │  ▸ Storage                │  ▸ DUTs                   │ │ Properties (typed)│ │
│               │                           │                           │ │  - voltage\:int    │ │
│  \[virtualized rows… scroll]               │ \[virtualized rows…]       │ │  - serial\:string  │ │
│               │                           │                           │ └───────────────────┘ │
├───────────────┴───────────────────────────┴───────────────────────────┴───────────┬───────────┤
│ Dock Tabs:  Inspector • History • Attachments • Plugins • Logs                    │ Status    │
└───────────────────────────────────────────────────────────────────────────────────┴───────────┘

```

### 15.2 Semantic Diff Viewer (S8)
```

┌───────────────────────────────────────────────────────────────────────────────┐
│ Compare: \[Snapshot A] vs \[Working]    Group by: Type ▾   Filters: \[x] Moves  │
├───────────────────────────────────────────────────────────────────────────────┤
│ Changes                                                                      │
│  ▸ Renames (3)                                                                │
│    - Bench 2  →  Bench B                                                      │
│  ▸ Moves (2)                                                                  │
│    - Oscilloscope 1  /Lab A/Bench 3 → /Lab A/Bench 4                          │
│  ▸ Property changes (7)                                                       │
│    - Power Supply.max\_voltage: 30 → 60                                        │
│                                                                               │
│ Details (right): before/after values • author • timestamp • open in workbench │
└───────────────────────────────────────────────────────────────────────────────┘

```

### 15.3 Merge Conflict Resolver (S9)
```

┌────────────────────────────────────────────────────────────────────────────────────────────┐
│ Merge: Pull resulted in conflicts  |  Strategy: Prefer Ours ▾ |  Dry‑run ✔  |  Commit ▶    │
├────────────────────────────────────────────────────────────────────────────────────────────┤
│ Node: /Lab A/Bench 3/Power Supply                                                           │
│  Property: max\_voltage                                                                      │
│   Theirs: 48V   ◀──────────────┬──────────────▶   Ours: 60V                                  │
│                                 Accept Theirs  |  Accept Ours                                │
│  Sibling order: \[PSU, Osc1, DUT1]  vs  \[Osc1, PSU, DUT1]  →  ⟷ visual reorder widget         │
│ Summary pane (bottom): 3 props chosen Theirs, 1 Ours, 1 unresolved                           │
└────────────────────────────────────────────────────────────────────────────────────────────┘

```

### 15.4 Project Dashboard (S3)
```

┌──────────────────────────────────────────────────────────────────────────────┐
│ Project: Archon Demo   \[Open Workbench] \[Import CSV] \[Snapshot]               │
├──────────────┬──────────────┬─────────────────────┬───────────────────────────┤
│ Recent Snaps │ Pending Sync │ Recent Changes     │ Plugins Activity          │
│  • v1.2      │  ahead: 2    │  • 7 edits today   │  • Jira panel pulled 5    │
│  • v1.1      │  behind: 1   │  • 1 merge         │    issues                 │
├──────────────┴──────────────┴─────────────────────┴───────────────────────────┤
│ Index Health • Attachment LFS • Quick tips                                     │
└──────────────────────────────────────────────────────────────────────────────┘

```

### 15.5 Import Wizard (S10)
```

Step 1: Choose Plugin → Step 2: Validate → Step 3: Preview → Step 4: Target → Step 5: Apply → Review
┌──────────────────────────────────────────────────────────────────────────────┐
│ Plugin: CSV Importer  |  Network access: none                                │
│ Preview rows (virtualized)  |  column→property mapping editor                 │
│ \[Apply as Draft]  \[Cancel]                                                    │
└──────────────────────────────────────────────────────────────────────────────┘

```

### 15.6 Plugin Manager (S11)
```

┌──────────────────────────────────────────────────────────────────────────────┐
│ Installed • Permissions • (Discover—later)                                     │
├──────────────────────────────────────────────────────────────────────────────┤
│ CSV Importer (Importer)   v0.1   \[Enable]  Access: readRepo                   │
│ Jira Panel (Panel, Net)   v0.2   \[Disable] Access: readRepo, net, secrets\:jira│
│ \[Configure] \[Revoke]                                                             │
└──────────────────────────────────────────────────────────────────────────────┘

````

---

## 16) Svelte Starter Stubs (routes + components)

> Minimal, compile‑ready scaffolding with theme/density controls. Hook up real data as stores later.

### 16.1 `src/routes/+layout.svelte`
```svelte
<script lang="ts">
  import CommandBar from "$lib/components/common/CommandBar.svelte";
  let theme = (localStorage.getItem("theme") as "light"|"dark") || "light";
  let density = (localStorage.getItem("density") as "compact"|"cozy") || "compact";
  $: document.documentElement.dataset.theme = theme;
  $: document.documentElement.dataset.density = density === "cozy" ? "cozy" : "";
  function toggleTheme(){ theme = theme === "light" ? "dark" : "light"; localStorage.setItem("theme", theme); }
  function toggleDensity(){ density = density === "compact" ? "cozy" : "compact"; localStorage.setItem("density", density); }
</script>

<div class="min-h-screen bg-[rgb(var(--bg))] text-[rgb(var(--fg))]">
  <header class="sticky top-0 border-b border-black/10 dark:border-white/10">
    <CommandBar on:toggleTheme={toggleTheme} on:toggleDensity={toggleDensity} />
  </header>
  <main class="h-[calc(100vh-40px)]"> <slot /> </main>
  <footer class="h-8 border-t border-black/10 dark:border-white/10 flex items-center px-2 text-xs">
    <span class="opacity-70">Archon • Status: Ready</span>
  </footer>
</div>
````

### 16.2 `src/lib/components/common/CommandBar.svelte`

```svelte
<script lang="ts">
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();
</script>
<div class="h-10 px-2 flex items-center gap-2 bg-[rgb(var(--muted))]">
  <div class="font-medium">Archon</div>
  <div class="flex-1"></div>
  <button class="px-2 py-1 rounded bg-[rgb(var(--accent))] text-[rgb(var(--accent-foreground))]" on:click={() => dispatch("toggleTheme")}>Theme</button>
  <button class="px-2 py-1 rounded border" on:click={() => dispatch("toggleDensity")}>Density</button>
</div>
```

### 16.3 `src/routes/project/[id]/workbench/+page.svelte`

```svelte
<script lang="ts">
  import MillerColumns from "$lib/components/workbench/MillerColumns.svelte";
  import InspectorPanel from "$lib/components/workbench/InspectorPanel.svelte";
</script>
<div class="grid grid-cols-[1fr_1fr_1fr_360px] h-full">
  <MillerColumns class="col-span-3" />
  <InspectorPanel />
</div>
```

### 16.4 `src/lib/components/workbench/MillerColumns.svelte`

```svelte
<script lang="ts">
  let columns = ["Sites", "Lab A", "Bench 3"]; // placeholder
</script>
<div class="grid grid-cols-3 border-r h-full">
  {#each columns as name}
    <div class="border-r overflow-auto">
      <div class="sticky top-0 bg-[rgb(var(--muted))] px-2 py-1 text-sm font-medium">{name}</div>
      <ul class="text-sm">
        {#each Array.from({length: 200}) as _, i}
          <li class="px-2 py-1 hover:bg-black/5 dark:hover:bg-white/5 cursor-default">Row {i+1}</li>
        {/each}
      </ul>
    </div>
  {/each}
</div>
```

### 16.5 `src/lib/components/workbench/InspectorPanel.svelte`

```svelte
<script lang="ts">let name = "Oscilloscope 1";</script>
<aside class="h-full overflow-auto p-3">
  <h3 class="text-sm font-semibold mb-2">Inspector</h3>
  <label class="block text-xs opacity-70">Name</label>
  <input class="w-full border px-2 py-1 rounded mb-3" bind:value={name} />
  <label class="block text-xs opacity-70">Properties</label>
  <div class="space-y-2 text-sm">
    <div class="grid grid-cols-[160px_1fr] gap-2 items-center">
      <div>max_voltage</div>
      <input class="border px-2 py-1 rounded" value="60" />
    </div>
    <div class="grid grid-cols-[160px_1fr] gap-2 items-center">
      <div>serial</div>
      <input class="border px-2 py-1 rounded" value="ABC123" />
    </div>
  </div>
</aside>
```

### 16.6 `src/lib/components/diff/SemanticDiffViewer.svelte`

```svelte
<script lang="ts">
  export let changes: Array<{type:string, label:string}> = [
    {type:"rename", label:"Bench 2 → Bench B"},
    {type:"move", label:"Oscilloscope 1 /Bench 3 → /Bench 4"},
  ];
</script>
<div class="grid grid-cols-[360px_1fr] h-full">
  <aside class="border-r overflow-auto">
    <div class="px-2 py-1 text-sm font-semibold sticky top-0 bg-[rgb(var(--muted))]">Changes</div>
    <ul class="text-sm">
      {#each changes as c}
        <li class="px-2 py-1 hover:bg-black/5 dark:hover:bg-white/5">{c.type}: {c.label}</li>
      {/each}
    </ul>
  </aside>
  <section class="p-3 overflow-auto">
    <h3 class="text-sm font-semibold mb-2">Details</h3>
    <div class="text-sm opacity-70">Select a change to view before/after.</div>
  </section>
</div>
```

### 16.7 `src/lib/components/merge/ConflictResolver.svelte`

```svelte
<script lang="ts">
  let conflicts = [{ key: "max_voltage", ours: "60", theirs: "48" }];
</script>
<div class="space-y-3 p-3">
  <div class="text-sm font-semibold">Conflicts</div>
  {#each conflicts as c}
    <div class="rounded border p-2 text-sm">
      <div class="opacity-70">{c.key}</div>
      <div class="flex items-center gap-2 mt-1">
        <button class="px-2 py-1 border rounded">Accept Theirs ({c.theirs})</button>
        <button class="px-2 py-1 border rounded">Accept Ours ({c.ours})</button>
      </div>
    </div>
  {/each}
</div>
```

### 16.8 `src/app.css` (theme + density glue)

```css
@import "tailwindcss";
/* Variables defined in Section 9.3 */
body { color: rgb(var(--fg)); background: rgb(var(--bg)); }
.btn-accent { background: rgb(var(--accent)); color: rgb(var(--accent-foreground)); }
```

---

