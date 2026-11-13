<script lang="ts">
	import { platform } from "@stores/platformStore";
	import { notifications } from "@stores/notificationStore";
	import Badge from "@components/shared/Badge.svelte";
	import Button from "@components/shared/Button.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import IconButton from "@components/shared/IconButton.svelte";
	import { downloadsApi } from "../../lib/features/downloadsApi";
	import { bytes } from "../../lib/features/format";
	import { isFileForOs } from "../../lib/features/platform";
	import type { GameFile } from "../../lib/features/types";
	import { PREPARING_TEXT, queueGameFile } from "./gameFileQueue";
	import { fileIcon, groupFiles, kindBadge, visibleHost } from "./fileGroups";

	let { gameId, files }: { gameId: number; files: GameFile[] } = $props();

	let preparingId = $state<number | null>(null);
	/** Platform groups the person closed. Every group starts open. */
	let closedGroups = $state<Record<string, boolean>>({});

	const groups = $derived(groupFiles(files, $platform.os));

	async function queue(file: GameFile) {
		const text = await queueGameFile(gameId, file, enqueueFile, (id) => {
			preparingId = id;
		});
		if (text) notifications.info(text);
	}

	function enqueueFile(targetGameId: number, fileId: number): Promise<number> {
		return downloadsApi.enqueueFile(targetGameId, fileId);
	}

	function toggleGroup(platformName: string) {
		closedGroups = { ...closedGroups, [platformName]: !closedGroups[platformName] };
	}
</script>

<section>
	<h3 id="files-title">Files</h3>

	{#if files.length === 0}
		<p class="muted">No files found for this game.</p>
	{:else}
		<div class="groups" aria-labelledby="files-title">
			{#each groups as group (group.platform)}
				{@const open = !closedGroups[group.platform]}
				<div class="group">
					<Button variant="text" size="sm" icon={open ? "expand_less" : "expand_more"} onclick={() => toggleGroup(group.platform)} aria-expanded={open}>
						<Icon name={group.icon} size={16} />
						<span>{group.label}</span>
						<span class="count">{group.files.length}</span>
					</Button>

					{#if open}
						<ul class="files">
							{#each group.files as file (file.id)}
								{@const name = file.filename ?? file.label}
								{@const host = visibleHost(file.host)}
								{@const badge = kindBadge(file.kind)}
								<li>
									<span class="type"><Icon name={fileIcon(file)} size={20} /></span>
									<div class="info">
										<p class="label" title={name}>{name}</p>
										<p class="meta">
											{#if file.size}<span>{bytes(file.size)}</span>{/if}
											{#if host}<span>{host}</span>{/if}
											{#if badge}<Badge>{badge}</Badge>{/if}
											{#if !isFileForOs(file.platform, $platform.os)}<Badge>Other OS</Badge>{/if}
											{#if file.requiresAuth}<span>Needs sign-in</span>{/if}
										</p>
									</div>
									<IconButton
										icon={preparingId === file.id ? "hourglass_top" : "download"}
										label={preparingId === file.id ? PREPARING_TEXT : `Download ${name}`}
										disabled={preparingId !== null}
										onclick={() => queue(file)}
									/>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</section>

<style lang="scss">
	section {
		display: flex;
		flex-direction: column;
		gap: 10px;
		min-width: 0;

		h3 {
			margin: 0;
			font-size: 1.25rem;
			font-weight: 600;
		}

		p {
			margin: 0;
		}
	}

	.groups {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.group {
		display: flex;
		flex-direction: column;
		gap: 6px;
		align-items: stretch;
	}

	.count {
		color: var(--text-muted);
	}

	.files {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 6px;

		li {
			display: flex;
			align-items: center;
			gap: 10px;
			padding: 8px 10px;
			border-radius: var(--radius-md);
			background: var(--bg-surface);
			min-width: 0;
		}
	}

	.type {
		display: inline-flex;
		flex-shrink: 0;
		color: var(--text-muted);
	}

	.info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.label {
		font-weight: 600;
		font-size: 0.95rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 4px 8px;
		color: var(--text-muted);
		font-size: 0.8rem;
	}

	.muted {
		color: var(--text-muted);
		font-size: 0.875rem;
	}
</style>
