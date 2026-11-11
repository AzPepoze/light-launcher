<script lang="ts">
	import { onMount } from "svelte";
	import Modal from "@components/shared/Modal.svelte";
	import { errorMessage } from "../../lib/features/rpc";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import type { PluginInfo, PluginPreview } from "../../lib/features/types";
	import PluginConsent from "./PluginConsent.svelte";
	import { PERMISSION_HINT } from "./pluginCopy";

	let {
		initialSource = "",
		progress = "",
		onInstalled,
		onClose,
	}: {
		initialSource?: string;
		/** "Plugin 2 of 3" when the plugin is one of several queued. */
		progress?: string;
		onInstalled: (plugin: PluginInfo) => void;
		onClose: () => void;
	} = $props();

	const title = $derived(progress || "Install plugin");
	let source = $state("");

	onMount(() => {
		if (!initialSource) return;
		source = initialSource;
		void check();
	});
	let preview = $state<PluginPreview | null>(null);
	let busy = $state(false);
	let message = $state("");

	async function check() {
		const link = source.trim();
		if (!link) return;
		busy = true;
		message = "";
		preview = null;
		try {
			preview = await pluginsApi.previewPermissions(link);
		} catch (err) {
			preview = null;
			message = errorMessage(err);
		} finally {
			busy = false;
		}
	}

	/** In a queue, Cancel skips this plugin. Alone, it goes back to the link form. */
	function cancelReview() {
		if (progress) onClose();
		else preview = null;
	}

	async function install(shown: PluginPreview) {
		busy = true;
		message = "";
		try {
			onInstalled(await pluginsApi.install(shown.source, !shown.signed, shown.consent));
		} catch (err) {
			message = errorMessage(err);
		} finally {
			busy = false;
		}
	}
</script>

<Modal show={true} {title} showDone={false} {onClose}>
	<form
		class="row"
		onsubmit={(event) => {
			event.preventDefault();
			check();
		}}
	>
		<label class="field">
			<span>Plugin link or file</span>
			<input class="input" type="text" bind:value={source} placeholder="https://... or a file path" />
		</label>
		<button class="btn sm" type="submit" disabled={busy || !source.trim()}>Check</button>
	</form>

	{#if message}
		<p class="error" role="alert">{message}</p>
	{/if}

	{#if preview}
		{@const shown = preview}
		<PluginConsent
			preview={shown}
			hint={PERMISSION_HINT}
			lines={shown.permissions}
			emptyText="No extra permissions."
			confirmLabel="Install plugin"
			{busy}
			onConfirm={() => install(shown)}
			onCancel={cancelReview}
		/>
	{/if}
</Modal>

<style lang="scss">
	.row {
		display: flex;
		align-items: flex-end;
		gap: 8px;
		margin: 0;

		.field {
			flex: 1;
			min-width: 0;
			display: flex;
			flex-direction: column;
			gap: 6px;
			font-size: 0.85rem;
			color: var(--text-muted);
		}

		.input {
			width: 100%;
		}
	}

	.error {
		margin: 12px 0 0;
		font-size: 0.85rem;
		color: var(--danger);
	}
</style>
