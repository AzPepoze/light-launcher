<script lang="ts">
	import type { RawSettings, SettingsField, SettingsSection } from "../../lib/features/settingsForm";
	import FormField from "./FormField.svelte";

	let {
		sections,
		raw = $bindable(),
		errors,
	}: {
		sections: SettingsSection[];
		raw: RawSettings;
		errors: Record<string, string>;
	} = $props();

	function hasError(section: SettingsSection): boolean {
		return section.fields.some((field) => Boolean(errors[field.key]));
	}
</script>

{#snippet fieldList(fields: SettingsField[])}
	{#each fields as field (field.key)}
		<FormField {field} bind:raw error={errors[field.key]} />
	{/each}
{/snippet}

<div class="form">
	{#each sections as section}
		{#if section.advanced}
			<details class="advanced" open={hasError(section)}>
				<summary>{section.title}</summary>
				<div class="fields">
					{@render fieldList(section.fields)}
				</div>
			</details>
		{:else}
			<div class="group">
				{#if section.title}
					<h4>{section.title}</h4>
				{/if}
				{@render fieldList(section.fields)}
			</div>
		{/if}
	{/each}
</div>

<style lang="scss">
	.form,
	.group,
	.fields {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	.group h4 {
		margin: 0;
		font-size: 0.75rem;
		font-weight: 800;
		letter-spacing: 1px;
		text-transform: uppercase;
		color: var(--text-muted);
	}

	.advanced {
		summary {
			cursor: pointer;
			font-size: 0.75rem;
			font-weight: 800;
			letter-spacing: 1px;
			text-transform: uppercase;
			color: var(--text-muted);
		}

		.fields {
			margin-top: 14px;
		}
	}
</style>
