<script lang="ts">
	import { ListPrefixDrives } from "@lib/api";
	import type { PrefixDrive } from "@shared";
	import Badge from "@components/shared/Badge.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";

	let { prefixPath }: { prefixPath: string } = $props();

	const KIND_LABELS: Record<string, string> = {
		system: "System",
		root: "Linux root",
		partition: "Disk",
		folder: "Folder"
	};

	let drives = $state<PrefixDrive[]>([]);
	let status = $state<"loading" | "ready" | "failed">("loading");
	let requestId = 0;

	$effect(() => {
		loadDrives(prefixPath);
	});

	/** The prefix folder may not exist yet, so a failed read is shown as a hint, not an error. */
	async function loadDrives(path: string) {
		const id = ++requestId;
		status = "loading";
		try {
			const list = await ListPrefixDrives(path);
			if (id !== requestId) return;
			drives = list;
			status = "ready";
		} catch {
			if (id !== requestId) return;
			drives = [];
			status = "failed";
		}
	}

	function kindLabel(kind: string): string {
		return KIND_LABELS[kind] ?? kind;
	}
</script>

{#if status === "loading"}
	<LoadingState label="Loading drives..." />
{:else if status === "failed"}
	<EmptyState icon="hard_drive" title="No drives yet" hint="Run a Wine tool in this prefix first." />
{:else if drives.length === 0}
	<EmptyState icon="hard_drive" title="No drives in this prefix" hint="Set up drives in Drive setup." />
{:else}
	<ul class="drive-list">
		{#each drives as drive (drive.letter)}
			<li class="drive-row">
				<span class="letter">{drive.letter.toUpperCase()}</span>
				<Badge tone="accent">{kindLabel(drive.kind)}</Badge>
				<span class="target" title={drive.target}>{drive.target}</span>
				{#if !drive.exists}
					<Badge tone="danger" icon="error">Missing</Badge>
				{/if}
			</li>
		{/each}
	</ul>
	<p class="hint">To change these, use Drive setup.</p>
{/if}

<style lang="scss">
	.drive-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.drive-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 16px;
		background: var(--bg-input);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-md);
	}

	.letter {
		min-width: 28px;
		color: var(--text-main);
		font-weight: 800;
	}

	.hint {
		margin: 8px 0 0;
		color: var(--text-muted);
		font-size: 0.85rem;
	}

	.target {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		color: var(--text-muted);
		font-size: 0.85rem;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
</style>
