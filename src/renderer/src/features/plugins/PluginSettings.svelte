<script lang="ts">
	import { onMount } from "svelte";
	import { errorMessage } from "../../lib/features/rpc";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import SchemaForm from "./SchemaForm.svelte";
	import {
		buildSchemaFields,
		collectValues,
		rawFormValues,
		type RawValues,
		type SchemaField,
	} from "./schemaForm";

	let { pluginId }: { pluginId: string } = $props();

	let fields = $state<SchemaField[]>([]);
	let raw = $state<RawValues>({});
	let errors = $state<Record<string, string>>({});
	let loading = $state(true);
	let saving = $state(false);
	let message = $state("");

	onMount(async () => {
		try {
			const response = await pluginsApi.settingsGet(pluginId);
			fields = buildSchemaFields(response.schema);
			raw = rawFormValues(fields, response.values);
		} catch (err) {
			message = errorMessage(err);
		} finally {
			loading = false;
		}
	});

	async function save() {
		const result = collectValues(fields, raw);
		errors = result.errors;
		message = "";
		if (Object.keys(result.errors).length > 0) return;
		saving = true;
		try {
			await pluginsApi.settingsSet(pluginId, result.values);
			message = "Saved.";
		} catch (err) {
			message = errorMessage(err);
		} finally {
			saving = false;
		}
	}
</script>

<section aria-labelledby="settings-title">
	<h3 id="settings-title">Settings</h3>

	{#if loading}
		<p class="muted">Loading settings...</p>
	{:else if fields.length === 0}
		<p class="muted">This plugin has no settings.</p>
	{:else}
		<SchemaForm {fields} bind:raw {errors} />
		<div class="footer">
			<button class="btn sm primary" disabled={saving} onclick={save}>Save settings</button>
			{#if message}<span class="muted" role="status">{message}</span>{/if}
		</div>
	{/if}

	{#if !loading && fields.length === 0 && message}
		<p class="muted" role="alert">{message}</p>
	{/if}
</section>

<style lang="scss">
	section {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	h3 {
		margin: 0;
		font-size: 0.8rem;
		font-weight: 800;
		letter-spacing: 1px;
		text-transform: uppercase;
		color: var(--text-muted);
	}

	.muted {
		margin: 0;
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.footer {
		display: flex;
		align-items: center;
		gap: 12px;
	}
</style>
