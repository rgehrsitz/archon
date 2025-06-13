---
trigger: glob
globs: src/ui/**/*.svelte
---

Strictly use Svelte 5.x syntax (checked via svelte-check --compiler-version 5).  No svelte:component fallback blocks, on:mount legacy helpers, or $: oldReactiveStyle.

Components must export named props via export let …—no implicit $$props unless required.

Prefer slot composition over nested component imports when only markup is reused.