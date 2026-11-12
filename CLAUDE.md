# Light Launcher: rules for Claude

## Reuse existing components

- Before writing any Svelte component or helper, look for one that exists in `src/renderer/src/components/shared/` and the feature folder.
- Use the shared component (for example `Dropdown`, `Modal`, `AppMenu`, `SlideButton`, `RangeSlider`, `NumericInput`, `BrowseInput`, `MenuTrigger`, `PageHeader`) instead of raw HTML or a copy.
- If no component fits, say so in the report and explain the gap. Do not add a near-duplicate.
- Workers get the same rule from `.wk-rules.md`.
