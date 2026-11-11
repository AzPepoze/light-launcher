<script lang="ts">
	import {
		canResetToDefault,
		defaultInput,
		pickMethodFor,
		type FieldInput,
		type PickMethod,
		type RawSettings,
		type SettingsField
	} from "../../lib/features/settingsForm";
	import ListEditor from "./ListEditor.svelte";

	let {
		field,
		raw = $bindable(),
		error,
	}: {
		field: SettingsField;
		raw: RawSettings;
		error?: string;
	} = $props();

	let inputId = $derived(`setting-${field.key}`);
	let hintId = $derived(`${inputId}-hint`);
	let current = $derived(raw[field.key]);
	let pickMethod = $derived(pickMethodFor(field, window.llShell));
	let resettable = $derived(canResetToDefault(field, raw));

	function textOf(input: FieldInput | undefined): string {
		return typeof input === "string" ? input : "";
	}

	function flagOf(input: FieldInput | undefined): boolean {
		return input === true;
	}

	function itemsOf(input: FieldInput | undefined): string[] {
		return Array.isArray(input) ? input : [];
	}

	function resetToDefault() {
		const fallback = defaultInput(field);
		if (fallback !== undefined) raw[field.key] = fallback;
	}

	async function browse(method: PickMethod) {
		const picked = await window.llShell?.[method]?.();
		if (picked) raw[field.key] = picked;
	}
</script>

<div class="field">
	{#if field.kind === "boolean"}
		<label class="switch">
			<input
				type="checkbox"
				role="switch"
				checked={flagOf(current)}
				aria-describedby={hintId}
				onchange={(event) => (raw[field.key] = event.currentTarget.checked)}
			/>
			<span class="track" aria-hidden="true"></span>
			<span>{field.label}{field.required ? " *" : ""}</span>
		</label>
	{:else}
		<label for={inputId}>{field.label}{field.required ? " *" : ""}</label>

		{#if field.kind === "select"}
			<select
				id={inputId}
				class="input"
				value={textOf(current)}
				aria-describedby={hintId}
				onchange={(event) => (raw[field.key] = event.currentTarget.value)}
			>
				{#if !field.required}<option value="">Not set</option>{/if}
				{#each field.options as option (option)}
					<option value={String(option)}>{option}</option>
				{/each}
			</select>
		{:else if field.kind === "number"}
			<input
				id={inputId}
				class="input"
				type="number"
				step={field.integer ? 1 : "any"}
				min={field.min}
				max={field.max}
				value={textOf(current)}
				aria-describedby={hintId}
				aria-invalid={Boolean(error)}
				oninput={(event) => (raw[field.key] = event.currentTarget.value)}
			/>
		{:else if field.kind === "path"}
			<div class="path">
				<input
					id={inputId}
					class="input"
					type="text"
					value={textOf(current)}
					aria-describedby={hintId}
					aria-invalid={Boolean(error)}
					oninput={(event) => (raw[field.key] = event.currentTarget.value)}
				/>
				{#if pickMethod}
					<button class="btn sm" type="button" onclick={() => browse(pickMethod)}>Browse</button>
				{/if}
			</div>
		{:else if field.kind === "list"}
			<ListEditor {field} items={itemsOf(current)} onchange={(next) => (raw[field.key] = next)} />
		{:else}
			<input
				id={inputId}
				class="input"
				type="text"
				value={textOf(current)}
				aria-describedby={hintId}
				aria-invalid={Boolean(error)}
				oninput={(event) => (raw[field.key] = event.currentTarget.value)}
			/>
		{/if}
	{/if}

	{#if field.description}
		<p class="hint" id={hintId}>{field.description}</p>
	{/if}
	{#if field.examples.length > 0}
		<p class="hint">Examples: {field.examples.join(", ")}</p>
	{/if}
	{#if error}
		<p class="error" role="alert">{error}</p>
	{/if}
	{#if resettable}
		<button class="reset" type="button" onclick={resetToDefault}>Reset to default</button>
	{/if}
</div>

<style lang="scss">
	.field {
		display: flex;
		flex-direction: column;
		gap: 6px;

		label {
			font-size: 0.85rem;
			font-weight: 700;
		}

		p {
			margin: 0;
		}
	}

	.switch {
		display: flex;
		align-items: center;
		gap: 10px;
		cursor: pointer;

		input {
			position: absolute;
			opacity: 0;
			pointer-events: none;
		}

		.track {
			position: relative;
			width: 38px;
			height: 22px;
			border-radius: 11px;
			background: var(--bg-input);
			border: 2px solid var(--glass-border);
			transition: background var(--transition-fast);
			flex-shrink: 0;

			&::after {
				content: "";
				position: absolute;
				top: 2px;
				left: 2px;
				width: 14px;
				height: 14px;
				border-radius: 50%;
				background: var(--text-muted);
				transition: transform var(--transition-fast);
			}
		}

		input:checked + .track {
			background: var(--accent-primary);
			border-color: var(--accent-primary);

			&::after {
				transform: translateX(16px);
				background: var(--bg-elevated);
			}
		}

		input:focus-visible + .track {
			box-shadow: 0 0 0 4px var(--accent-glow);
		}
	}

	.path {
		display: flex;
		gap: 8px;

		.input {
			flex: 1;
			min-width: 0;
		}
	}

	.hint {
		font-size: 0.75rem;
		color: var(--text-muted);
	}

	.error {
		font-size: 0.75rem;
		color: var(--danger);
	}

	.reset {
		align-self: flex-start;
		padding: 0;
		border: none;
		background: none;
		font-size: 0.75rem;
		font-weight: 700;
		color: var(--accent-primary);
		cursor: pointer;
	}
</style>
