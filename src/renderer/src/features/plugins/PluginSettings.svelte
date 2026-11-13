<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";
	import { notifications } from "@stores/notificationStore";
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
	const SAVED_COPY = "Settings saved";

	let fields = $state<SettingsField[]>([]);
	let sections = $state<SettingsSection[]>([]);
	/** The form state at load or at the last save. Revert and the unsaved-changes check compare with it. */
	let baseline = $state<RawSettings>({});
	let raw = $state<RawSettings>({});
	let errors = $state<Record<string, string>>({});
	let loading = $state(true);
	let loadFailed = $state(false);
	let saving = $state(false);
	let loadId = 0;

	let dirty = $derived(changedKeys(fields, raw, baseline).length > 0);

	async function load(id: string) {
		const current = ++loadId;
		loading = true;
		loadFailed = false;
		errors = {};
		try {
			const response = await pluginsApi.settingsGet(id);
			if (current !== loadId) return;
			fields = buildSettingsFields(response.schema);
			sections = buildSettingsSections(fields);
			baseline = rawSettings(fields, response.values);
			raw = cloneRaw(baseline);
		} catch {
			if (current === loadId) loadFailed = true;
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
		if (Object.keys(check.errors).length > 0) {
			notifications.error(FIX_FIELDS);
			return;
		}
		const savedState = cloneRaw(raw);
		saving = true;
		try {
			await pluginsApi.settingsSet(pluginId, check.values);
			baseline = savedState;
			notifications.success(SAVED_COPY);
		} catch (err) {
			const problems = err instanceof InvokeError ? err.fields : [];
			errors = serverErrors(problems);
			notifications.error(problems.length > 0 ? FIX_FIELDS : errorMessage(err));
		} finally {
			saving = false;
		}
	}

	function revert() {
		raw = cloneRaw(baseline);
		errors = {};
	}
</script>

<section aria-labelledby="settings-title">
	<h3 id="settings-title">Settings</h3>

	{#if loading}
		<LoadingState label="Loading settings..." />
	{:else if loadFailed}
		<EmptyState icon="error" title="Settings could not be loaded">
			<Button size="sm" onclick={() => load(pluginId)}>Retry</Button>
		</EmptyState>
	{:else if fields.length === 0}
		<EmptyState icon="tune" title="This plugin has no settings." />
	{:else}
		<SchemaForm {sections} bind:raw {errors} />
		<div class="footer">
			<Button size="sm" disabled={!dirty || saving} onclick={revert}>Revert</Button>
			<Button variant="primary" size="sm" disabled={!dirty || saving} onclick={save}>
				Save settings
			</Button>
			{#if dirty}
				<span class="unsaved"><span class="dot" aria-hidden="true"></span>Unsaved changes</span>
			{/if}
		</div>
	{/if}
</section>

<style lang="scss">
	section {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	h3 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.footer {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 12px;
		font-size: 0.9375rem;
	}

	.unsaved {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		color: var(--text-main);
	}

	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--accent-primary);
	}
</style>
