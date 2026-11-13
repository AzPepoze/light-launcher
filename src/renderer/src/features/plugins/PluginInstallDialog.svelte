<script lang="ts">
	import { onMount } from "svelte";
	import Button from "@components/shared/Button.svelte";
	import Modal from "@components/shared/Modal.svelte";
	import { notifications } from "@stores/notificationStore";
	import { errorMessage } from "../../lib/features/rpc";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import type { PluginInfo, PluginPreview } from "../../lib/features/types";
	import PluginConsent from "./PluginConsent.svelte";
	import QueueProgress from "./QueueProgress.svelte";
	import { PERMISSION_HINT, pluginProgress } from "./pluginCopy";

	let {
		initialSource = "",
		queue = null,
		onInstalled,
		onClose,
	}: {
		initialSource?: string;
		/** Set when the plugin is one of several queued. */
		queue?: { position: number; total: number } | null;
		onInstalled: (plugin: PluginInfo) => void;
		onClose: () => void;
	} = $props();

	const title = $derived(queue ? pluginProgress(queue.position, queue.total) : "Install plugin");
	let source = $state("");
	let preview = $state<PluginPreview | null>(null);
	let busy = $state(false);

	onMount(() => {
		if (!initialSource) return;
		source = initialSource;
		void check();
	});

	async function check() {
		const link = source.trim();
		if (!link) return;
		busy = true;
		preview = null;
		try {
			preview = await pluginsApi.previewPermissions(link);
		} catch (err) {
			notifications.error(errorMessage(err));
		} finally {
			busy = false;
		}
	}

	/** In a queue, Cancel skips this plugin. Alone, it goes back to the link form. */
	function cancelReview() {
		if (queue) onClose();
		else preview = null;
	}

	async function install(shown: PluginPreview) {
		busy = true;
		try {
			onInstalled(await pluginsApi.install(shown.source, !shown.signed, shown.consent));
		} catch (err) {
			notifications.error(errorMessage(err));
		} finally {
			busy = false;
		}
	}
</script>

<Modal show={true} {title} showDone={false} {onClose}>
	{#if queue}
		<QueueProgress position={queue.position} total={queue.total} />
	{/if}

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
		<Button type="submit" size="sm" disabled={busy || !source.trim()}>Check</Button>
	</form>

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
			font-size: 0.9375rem;
			color: var(--text-muted);
		}

		.input {
			width: 100%;
		}
	}
</style>
