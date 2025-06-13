---
trigger: glob
globs: src/ui/**/*.{svelte,ts}
---

Use TailwindCSS 4.x utility classes for styling; only author additional CSS when Tailwind cannot express the requirement (document the reason in a comment).

Custom CSS lives in src/ui/styles/archon.css; keep it ≤ 200 LoC per phase.

Run pnpm run lint:css (Stylelint Tailwind preset) locally; no errors.