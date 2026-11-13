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
	import Button from "@components/shared/Button.svelte";
	import Dropdown, { asOptions, type DropdownOption } from "@components/shared/Dropdown.svelte";
	import Switch from "@components/shared/Switch.svelte";
	import NumberField from "@components/shared/NumberField.svelte";
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
	let labelText = $derived(`${field.label}${field.required ? " *" : ""}`);
	let current = $derived(raw[field.key]);
	let pickMethod = $derived(pickMethodFor(field, window.llShell));
	let resettable = $derived(canResetToDefault(field, raw));
	let selectItems = $derived(selectOptionsOf(field));

	function selectOptionsOf(settings: SettingsField): DropdownOption[] {
		if (settings.kind !== "select") return [];
		const items = asOptions(settings.options.map(String));
		return settings.required ? items : [{ value: "", label: "Not set" }, ...items];
	}

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
		<Switch
			checked={flagOf(current)}
			aria-describedby={hintId}
			onChange={(checked) => (raw[field.key] = checked)}
		>
			<span>{field.label}{field.required ? " *" : ""}</span>
		</Switch>
	{:else}
		{#if field.kind !== "number"}
			<label for={inputId}>{labelText}</label>
		{/if}

		{#if field.kind === "select"}
			<Dropdown
				id={inputId}
				aria-describedby={hintId}
				options={selectItems}
				value={textOf(current)}
				onChange={(value) => (raw[field.key] = value)}
			/>
		{:else if field.kind === "number"}
			<NumberField
				id={inputId}
				label={labelText}
				value={textOf(current)}
				min={field.min}
				max={field.max}
				step={field.integer ? 1 : "any"}
				invalid={Boolean(error)}
				describedBy={hintId}
				oninput={(value) => (raw[field.key] = value)}
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
					<Button size="sm" onclick={() => browse(pickMethod)}>Browse</Button>
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
		<Button variant="text" size="sm" class="reset" onclick={resetToDefault}>Reset to default</Button>
	{/if}
</div>

<style lang="scss">
	.field {
		display: flex;
		flex-direction: column;
		gap: 6px;

		label {
			font-size: 0.9375rem;
			font-weight: 600;
		}

		p {
			margin: 0;
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
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.error {
		font-size: 0.875rem;
		color: var(--danger);
	}

	:global(.field .reset) {
		align-self: flex-start;
	}
</style>
