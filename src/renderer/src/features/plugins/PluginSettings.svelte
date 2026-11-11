<script lang="ts">
	import { errorMessage } from "../../lib/features/rpc";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import {
		buildSettingsFields,
		buildSettingsSections,
		changedKeys,
		cloneRaw,
		collectSettings,
		rawSettings,
		serverErrors,
		type RawSettings,
		type SettingsField,
		type SettingsSection
	} from "../../lib/features/settingsForm";
	import { InvokeError } from "../../lib/transport";
	import SchemaForm from "./SchemaForm.svelte";

	let { pluginId }: { pluginId: string } = $props();

	const FIX_FIELDS = "Fix the highlighted fields.";

	let fields = $state<SettingsField[]>([]);
	let sections = $state<SettingsSection[]>([]);
	/** The form state at load or at the last save. Revert and the unsaved-changes check compare with it. */
	let baseline = $state<RawSettings>({});
	let raw = $state<RawSettings>({});
	let errors = $state<Record<string, string>>({});
	let loading = $state(true);
	let saving = $state(false);
	let message = $state("");
	let loadId = 0;

	let dirty = $derived(changedKeys(fields, raw, baseline).length > 0);

	async function load(id: string) {
		const current = ++loadId;
		loading = true;
		message = "";
		errors = {};
		try {
			const response = await pluginsApi.settingsGet(id);
			if (current !== loadId) return;
			fields = buildSettingsFields(response.schema);
			sections = buildSettingsSections(fields);
			baseline = rawSettings(fields, response.values);
			raw = cloneRaw(baseline);
		} catch (err) {
			if (current === loadId) message = errorMessage(err);
		} finally {
			if (current === loadId) loading = false;
		}
	}

	$effect(() => {
		void load(pluginId);
	});

	async function save() {
		const check = collectSettings(fields, raw);
		errors = check.errors;
		message = "";
		if (Object.keys(check.errors).length > 0) {
			message = FIX_FIELDS;
			return;
		}
		const saved = cloneRaw(raw);
		saving = true;
		try {
			await pluginsApi.settingsSet(pluginId, check.values);
			baseline = saved;
			message = "Saved.";
		} catch (err) {
			const problems = err instanceof InvokeError ? err.fields : [];
			errors = serverErrors(problems);
			message = problems.length > 0 ? FIX_FIELDS : errorMessage(err);
		} finally {
			saving = false;
		}
	}

	function revert() {
		raw = cloneRaw(baseline);
		errors = {};
		message = "";
	}
</script>

<section aria-labelledby="settings-title">
	<h3 id="settings-title">Settings</h3>

	{#if loading}
		<p class="muted">Loading settings...</p>
	{:else if fields.length === 0}
		<p class="muted" role={message ? "alert" : undefined}>
			{message || "This plugin has no settings."}
		</p>
	{:else}
		<SchemaForm {sections} bind:raw {errors} />
		<div class="footer">
			<button class="btn sm" type="button" disabled={!dirty || saving} onclick={revert}>
				Revert
			</button>
			<button class="btn sm primary" type="button" disabled={!dirty || saving} onclick={save}>
				Save settings
			</button>
			{#if dirty}<span class="muted">Unsaved changes</span>{/if}
			{#if message}<span class="muted" role="status">{message}</span>{/if}
		</div>
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
		flex-wrap: wrap;
		align-items: center;
		gap: 12px;
	}
</style>
