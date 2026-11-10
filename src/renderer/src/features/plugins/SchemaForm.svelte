<script lang="ts">
	import type { RawValues, SchemaField } from "./schemaForm";

	let {
		fields,
		raw = $bindable(),
		errors,
	}: {
		fields: SchemaField[];
		raw: RawValues;
		errors: Record<string, string>;
	} = $props();

	function setText(key: string, event: Event) {
		raw[key] = (event.currentTarget as HTMLInputElement | HTMLSelectElement).value;
	}

	function setChecked(key: string, event: Event) {
		raw[key] = (event.currentTarget as HTMLInputElement).checked;
	}

	function textValue(key: string): string {
		const value = raw[key];
		return typeof value === "string" ? value : "";
	}
</script>

<div class="form">
	{#each fields as field (field.key)}
		{@const errorId = `${field.key}-error`}
		<div class="field">
			{#if field.kind === "boolean"}
				<label class="check">
					<input
						type="checkbox"
						checked={raw[field.key] === true}
						onchange={(event) => setChecked(field.key, event)}
						aria-describedby={errorId}
					/>
					<span>{field.label}</span>
				</label>
			{:else}
				<label for={field.key}>{field.label}{field.required ? " *" : ""}</label>
				{#if field.kind === "select"}
					<select
						id={field.key}
						class="input"
						value={textValue(field.key)}
						onchange={(event) => setText(field.key, event)}
						aria-describedby={errorId}
					>
						{#each field.options as option (option)}
							<option value={String(option)}>{option}</option>
						{/each}
					</select>
				{:else if field.kind === "number"}
					<input
						id={field.key}
						class="input"
						type="number"
						step={field.integer ? 1 : "any"}
						min={field.min}
						max={field.max}
						value={textValue(field.key)}
						oninput={(event) => setText(field.key, event)}
						aria-describedby={errorId}
						aria-invalid={Boolean(errors[field.key])}
					/>
				{:else}
					<input
						id={field.key}
						class="input"
						type="text"
						value={textValue(field.key)}
						oninput={(event) => setText(field.key, event)}
						aria-describedby={errorId}
						aria-invalid={Boolean(errors[field.key])}
					/>
				{/if}
			{/if}
			{#if field.description}
				<p class="hint">{field.description}</p>
			{/if}
			{#if errors[field.key]}
				<p class="error" id={errorId} role="alert">{errors[field.key]}</p>
			{/if}
		</div>
	{/each}
</div>

<style lang="scss">
	.form {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

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

	.check {
		display: flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;
	}

	.hint {
		font-size: 0.75rem;
		color: var(--text-muted);
	}

	.error {
		font-size: 0.75rem;
		color: var(--danger);
	}
</style>
