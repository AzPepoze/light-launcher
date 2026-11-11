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
	const SAVED_COPY = "Saved";

	let fields = $state<SettingsField[]>([]);
	let sections = $state<SettingsSection[]>([]);
	/** The form state at load or at the last save. Revert and the unsaved-changes check compare with it. */
	let baseline = $state<RawSettings>({});
	let raw = $state<RawSettings>({});
	let errors = $state<Record<string, string>>({});
	let loading = $state(true);
	let saving = $state(false);
	let message = $state("");
	/** True after a save, until the form changes again. Drives the "Saved" fade. */
	let saved = $state(false);
	let loadId = 0;

	let dirty = $derived(changedKeys(fields, raw, baseline).length > 0);

	$effect(() => {
		if (dirty) saved = false;
	});

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
		saved = false;
		if (Object.keys(check.errors).length > 0) {
			message = FIX_FIELDS;
			return;
		}
		const savedState = cloneRaw(raw);
		saving = true;
		try {
			await pluginsApi.settingsSet(pluginId, check.values);
			baseline = savedState;
			saved = true;
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
		saved = false;
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
			{#if dirty}
				<span class="unsaved"><span class="dot" aria-hidden="true"></span>Unsaved changes</span>
			{/if}
			<span class="saved" class:shown={saved} role="status">{saved ? SAVED_COPY : ""}</span>
		</div>
		{#if message}
			<p class="error" role="alert">{message}</p>
		{/if}
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

	.muted {
		margin: 0;
		font-size: 0.9375rem;
		color: var(--text-muted);
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

	.saved {
		color: var(--success);
		font-weight: 600;
		opacity: 0;

		&.shown {
			animation: saved-flash 2.4s var(--ease-out) both;
		}
	}

	.error {
		margin: 0;
		font-size: 0.9375rem;
		color: var(--danger);
	}

	/* Reduced motion keeps the confirmation visible instead of fading it out. */
	@media (prefers-reduced-motion: reduce) {
		.saved {
			opacity: 1;

			&.shown {
				animation: none;
			}
		}
	}

	@keyframes saved-flash {
		0% {
			opacity: 0;
		}
		12%,
		70% {
			opacity: 1;
		}
		100% {
			opacity: 0;
		}
	}
</style>
