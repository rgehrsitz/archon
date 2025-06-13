## Archon: Internationalization (i18n) & Accessibility (a11y) Readiness

To ensure Archon is inclusive, global-ready, and accessible, we build hooks and guidelines from day one—even if full support rolls out in later versions. This document outlines strategies and best practices for:
1. Prioritization: Must-Haves vs. Nice-to-Haves for v1
2. String Extraction & Localization Framework
3. Date/Time, Number, and Unit Formatting
4. Locale Detection, Overrides, and OS Integration
5. UI/UX Considerations for Right-to-Left (RTL) and Multilingual Interfaces
6. Accessibility Foundations: Semantic HTML, ARIA, Keyboard, Screen Readers
7. Color Contrast, Focus Indicators, High Contrast & Reduced Motion
8. Automated & Manual Testing for i18n & a11y
9. Implementation Guidance in Svelte/Wails Context
10. Documentation & Developer Guidelines

---

### 1. Prioritization: Must-Haves vs. Nice-to-Haves for v1

- **Must-Haves**:
  - Externalize UI strings and basic localization support for English (default) and one additional language if resources permit.
  - Use Intl APIs for date/time/number formatting in current locale.
  - Semantic HTML structure and basic ARIA roles for core components (buttons, dialogs, tree view).
  - Keyboard navigation for main flows and visible focus indicators.
  - Color contrast meeting WCAG AA for default theme.
  - CI checks for missing translation keys and basic a11y issues via Axe.
- **Nice-to-Haves (deferred or optional toggles)**:
  - Full translations for multiple locales beyond English.
  - Granular user settings for date/time format variations (MM/DD vs DD/MM), numeric separators, unit preferences.
  - RTL layout support and mirrored iconography.
  - High Contrast Mode and Reduced Motion respecting OS preferences.
  - Pseudo-localization and extensive pluralization testing for many languages.
  - Accessible interactive diagrams with text summaries.
  - Manual user testing sessions with screen readers in early alpha stages.

---

### 2. String Extraction & Localization Framework

#### 2.1 Externalize All UI Strings
- **Resource Files**: Store user-facing text in external files (`locales/en.json`, etc.). JSON-based format is straightforward for Svelte.
- **Key-Based Lookup**: Use descriptive keys (e.g., `snapshot.createButton`) consistently.
- **Avoid Concatenation**: Use full template strings with placeholders.
- **Prioritization**: For v1, extract strings from main UI; leave less-used or developer-only messages for later.

#### 2.2 Translation & Interpolation
- **Placeholders & Interpolation**: Named placeholders ("Snapshot {name} created at {time}").
- **Pluralization**: Use Intl.PluralRules for dynamic counts. For v1, focus on common plural patterns in English; plan extension for other locales later.
- **Fallback Language**: Always fallback to English if translation missing.

#### 2.3 Build & Runtime Integration
- **Extraction Tooling**: Provide simple CLI (`npm run i18n:extract`) to gather keys from Svelte files.
- **Runtime Loading**: Detect locale or use user override; load locale file at startup. Bundle default language; allow placing additional locale JSON in `locales/`.
- **Logging Missing Keys**: Log warnings in dev mode but not disrupt users at runtime.

#### 2.4 Developer Workflow
- **Translation Template**: Generate `locales/en.template.json` for translators.
- **Validation in CI**: Check locale files include all keys from template.
- **Manual Testing**: Encourage reviewing UI with pseudo-localized strings for layout issues.

---

### 3. Date/Time, Number, and Unit Formatting

#### 3.1 Use Intl APIs
- **Date/Time**: Format dates/times via `Intl.DateTimeFormat`, converting stored ISO 8601 UTC to user locale/timezone.
- **Numbers**: Use `Intl.NumberFormat` for separators.
- **Units & Currency**: Use appropriate Intl options; allow user preference for currency code if cost metadata appears.

#### 3.2 UI Components & Inputs
- **Date Pickers**: Use or build components respecting locale conventions (first day of week, format). For v1, default to English patterns; design flexible enough for later locale injection.
- **Number Inputs**: Accept locale-specific separators; normalize internally. For v1, document limitations and plan improvements.
- **Relative Times**: Use `Intl.RelativeTimeFormat` for status messages ("5 minutes ago").

#### 3.3 Storage vs Display & User Preferences
- **Canonical Storage**: Always store date/time in ISO UTC; convert only on display.
- **User Overrides**: In settings, allow overriding timezone or date format patterns (Nice-to-Have). For v1, respect system locale/timezone automatically.

---

### 4. Locale Detection, Overrides, and OS Integration

#### 4.1 Detect System Locale & Timezone
- **Wails/Go Backend**: On startup, read OS locale/timezone to set defaults.
- **User Override UI**: For v1, simple dropdown to select language from available locales; advanced format overrides deferred.

#### 4.2 Fallback Strategy
- **Missing Translations**: Fallback to English; log missing keys for developers.
- **RTL Detection**: For RTL locales, plan layout adjustments; initial v1 may defer full RTL but include hooks (e.g., setting `dir` attribute).

#### 4.3 OS-Level Accessibility Settings
- **High Contrast & Reduced Motion**: Detect OS settings if possible via Wails/WebView. For v1, design CSS to support high contrast themes and minimal animations; actual toggles may be deferred.
- **Theme Respect**: Follow OS light/dark mode; ensure contrast in both.

---

### 5. UI/UX Considerations for RTL and Multilingual Interfaces

#### 5.1 Bidirectional Layout Support (Hooked)
- **CSS Direction**: Structure components with logical CSS properties; set `dir` dynamically when RTL support enabled in future.
- **Iconography**: Use transform-friendly icons or mirrored assets; design baseline so flipping is feasible.
- **Layout Flexibility**: Ensure UI elements can expand for longer text; avoid fixed widths.

#### 5.2 Font & Character Support
- **Unicode Coverage**: Bundle or reference fonts covering broad scripts. For v1, test basic Latin and plan adding additional fonts later.
- **Input Handling**: Ensure text fields handle IME composition correctly; test with non-Latin input.
- **Sorting & Collation**: Use `Intl.Collator` when sorting names; v1 may default to locale-aware behavior based on system locale.

---

### 6. Accessibility Foundations

#### 6.1 Semantic HTML & ARIA
- **Markup**: Use semantic elements in Svelte for buttons, nav, main, sections.
- **ARIA Roles**: For custom widgets (tree, dialogs), follow WAI-ARIA Authoring Practices. Ensure `role`, `aria-expanded`, `aria-label` etc.
- **Labels**: Associate labels with inputs; include alt text or aria-label for icons.

#### 6.2 Keyboard Navigation & Focus
- **Tab Order**: Define logical tab flow; use `tabindex` for custom elements.
- **Focus Indicators**: Visible outlines or styles; do not remove without replacement.
- **Shortcuts**: Provide discoverable keyboard shortcuts; ensure no conflict with screen readers or OS.
- **Tree Navigation**: Implement arrow key support for hierarchy tree using ARIA tree patterns.

#### 6.3 Screen Reader Support & Live Regions
- **Announcements**: Use `aria-live` for status updates (e.g., "Snapshot created").
- **Dialog Focus Management**: Trap focus in modals; announce title; restore focus.
- **Complex Visuals**: Provide text alternatives or summaries for diff views or diagrams; v1 include basic summaries.
- **Manual Testing**: Regularly test with NVDA, VoiceOver to catch issues automated tools miss.

#### 6.4 Color Contrast & Visual Aids
- **Contrast Ratios**: Ensure WCAG AA standards in default theme; test dark/light.
- **Avoid Color-Only**: Use icons or text alongside color cues.
- **High Contrast Mode**: CSS should support toggling to high-contrast (Nice-to-Have), respecting OS if detected.
- **Reduced Motion**: Minimize animations; allow user or OS preference to disable non-essential motion.

---

### 7. Automated & Manual Testing for i18n & a11y

#### 7.1 i18n Testing
- **Lint Missing Keys**: Static checks for translation calls without keys.
- **CI Locale Validation**: Load each available locale to catch errors.
- **Pseudo-Localization**: Use pseudo-locales in dev to reveal layout issues.
- **Pluralization**: Automated tests for plural forms for English and key target locales.
- **Performance Balance**: For desktop app, bundle core locales; lazy-load additional locales if size grows significantly.

#### 7.2 Accessibility Testing
- **Automated Tools**: Integrate Axe in Playwright or Jest for core UI flows.
- **Keyboard Navigation Tests**: Automated scripts to traverse UI via keyboard only.
- **Color Contrast Checks**: Automated contrast analysis on components.
- **Manual User Testing**: Schedule periodic sessions with assistive technology users to validate real-world usability.

---

### 8. Implementation Guidance in Svelte/Wails Context

#### 8.1 i18n Libraries & Integration
- **Svelte i18n**: Use a lightweight library (e.g., `svelte-i18n`) supporting reactive translations, interpolation, pluralization.
- **Setup & Extraction**: Initialize locale store, use translation functions in components. Provide CLI extraction script and integrate into CI.
- **Runtime Locale Change**: Support changing language at runtime; components re-render accordingly.

#### 8.2 Date/Number Formatting Utilities
- **Wrapper Module**: Centralize Intl usage; reactively update on locale change.
- **User Overrides Hook**: Design API for future format overrides.

#### 8.3 RTL & Theming
- **Logical CSS**: Tailwind logical utilities and dynamic `dir` attribute.
- **OS Theme Integration**: Detect OS dark/light mode via Wails and apply matching theme.
- **High Contrast & Reduced Motion**: Hook into OS preferences if available; design CSS variables to support toggling.

#### 8.4 Accessible Components in Svelte
- **Tree View**: Follow WAI-ARIA for tree widget; use `bind:this` for focus management.
- **Dialogs & Modals**: Use ARIA roles, trap focus, manage announcements.
- **Buttons/Controls**: Ensure keyboard focusable, `aria-label` for icon-only buttons.
- **Live Regions**: Use Svelte reactivity to update `aria-live` regions for dynamic messages.

#### 8.5 Testing Setup
- **Playwright Tests**: Launch Archon in test mode, simulate locale changes, keyboard nav, run Axe checks.
- **Pseudo-Localization Mode**: Debug mode to wrap strings for layout testing.

---

### 9. Documentation & Developer Guidelines

- **i18n Guide**: How to add keys, run extraction, add locale files, handle pluralization and context.
- **a11y Checklist**: Semantic markup, ARIA attributes, keyboard navigation, contrast, screen reader testing.
- **User Settings Specs**: Outline future settings: language selection, date/time format overrides, high contrast toggle, reduced motion toggle.
- **Testing Recipes**: Steps for automated and manual testing of i18n/a11y.
- **Translator Instructions**: Context notes for translators, handling placeholders and plural forms.

---

*End of Internationalization & Accessibility Readiness Deep Dive (Refined with Prioritization, OS-Level Settings, User Overrides, Manual Testing Emphasis, Performance Considerations)*

