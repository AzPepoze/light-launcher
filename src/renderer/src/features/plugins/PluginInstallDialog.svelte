<script lang="ts">
	import { onMount } from "svelte";
	import { PickFolder } from "@lib/api";
	import Button from "@components/shared/Button.svelte";
	import Modal from "@components/shared/Modal.svelte";
	import { notifications } from "@stores/notificationStore";
	import { errorMessage } from "../../lib/features/rpc";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import type { PluginInfo, PluginPreview } from "../../lib/features/types";
	import PluginConsent from "./PluginConsent.svelte";
	import QueueProgress from "./QueueProgress.svelte";
	import {
		LINK_FOLDER_HINT,
		LINK_FOLDER_LABEL,
		PERMISSION_HINT,
		linkedSummary,
		pluginProgress
	} from "./pluginCopy";

	let {
		initialSource = "",
		initialSha256 = "",
		queue = null,
		onInstalled,
		onClose,
	}: {
		initialSource?: string;
		/** SHA256 the package must match, from a community index entry. Empty skips the check. */
		initialSha256?: string;
		/** Set when the plugin is one of several queued. */
		queue?: { position: number; total: number } | null;
		onInstalled: (plugin: PluginInfo) => void;
		onClose: () => void;
	} = $props();

	const title = $derived(queue ? pluginProgress(queue.position, queue.total) : "Install plugin");
	let source = $state("");
	let expectedSha256 = $state("");
	let preview = $state<PluginPreview | null>(null);
	let busy = $state(false);

	/** The pinned digest applies only while the box still holds the link it came with. */
	const pinnedSha256 = $derived(source.trim() === initialSource.trim() ? expectedSha256 : "");
	/** Linking a folder is offered only for a manual install, not a queued or community one. */
	const showLinkFolder = $derived(!queue && !initialSource);

	onMount(() => {
		if (!initialSource) return;
		source = initialSource;
		expectedSha256 = initialSha256;
		void check();
	});

	async function check() {
		const link = source.trim();
		if (!link) return;
		busy = true;
		preview = null;
		try {
			preview = await pluginsApi.previewPermissions(link, pinnedSha256);
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

	/** Picks a plugin folder and runs it in dev mode. A linked folder has no consent step. */
	async function linkFolder() {
		busy = true;
		try {
			const path = await PickFolder();
			if (!path) return;
			const plugin = await pluginsApi.devLink(path);
			notifications.success(linkedSummary(plugin.name));
			onInstalled(plugin);
		} catch (err) {
			notifications.error(errorMessage(err));
		} finally {
			busy = false;
		}
	}

	async function install(shown: PluginPreview) {
		busy = true;
		try {
			onInstalled(await pluginsApi.installOrUpdate(shown.source, !shown.signed, shown.consent));
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

	{#if showLinkFolder}
		<div class="link">
			<p class="hint">{LINK_FOLDER_HINT}</p>
			<Button size="sm" icon="folder_open" disabled={busy} onclick={linkFolder}>
				{LINK_FOLDER_LABEL}
			</Button>
		</div>
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
			font-size: 0.9375rem;
			color: var(--text-muted);
		}

		.input {
			width: 100%;
		}
	}

	.link {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		flex-wrap: wrap;
		margin-top: 16px;

		.hint {
			margin: 0;
			font-size: 0.875rem;
			color: var(--text-muted);
		}
	}
</style>
