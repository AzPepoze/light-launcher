# Reuse existing components

- Before writing any Svelte component or helper, look for one that exists in `src/renderer/src/components/shared/` and the feature folder.
- Use the shared component (for example `Dropdown`, `Modal`, `AppMenu`, `SlideButton`, `RangeSlider`, `NumericInput`, `BrowseInput`, `MenuTrigger`, `PageHeader`) instead of raw HTML or a copy.
- UI kit in `components/shared/`:
     - `Icon`: Material Symbols Rounded icon, use instead of `material-icons` spans.
     - `Button`: buttons with `variant` and `size`, spinner while `loading`.
     - `IconButton`: icon-only button with a required `label`.
     - `Notice`: tinted message box for info, warning, error or success.
     - `EmptyState`: centered icon, title, hint and action.
     - `LoadingState`: small spinner with a label.
     - `Badge`: small pill with a tone.
     - `Section`: card with a title, help line and actions.
     - `SettingRow`: title and help on the left, control on the right.
     - `ChoiceRow`: full-width selectable row for filter choices, with a tint when selected and an optional count.
     - `ProgressBar`: bar with a value, or indeterminate when `value` is null.
     - `TransferStages`: Download, Extract, Install steps for a transfer.
     - `TransferMeta`: speed, time left, size and percent of a transfer.
     - `Sparkline`: small line chart of recent samples.
- Transient messages (a check failed, a download was added) go through the notification store (`@stores/notificationStore`), never as stacked Notice blocks; Notice is for page-blocking states only.
- Pages must not start network work on open unless the person asked for it.
- If no component fits, say so in the report and explain the gap. Do not add a near-duplicate.

## Keep the gap list current

- `docs/known-gaps.md` is the single list of what is missing, unproven or planned.
- Add a row the moment the owner asks for something or an agent reports a gap. Tick it when it is done.
- Re-read it after every batch of agents and before every final report.
