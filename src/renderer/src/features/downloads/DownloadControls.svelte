<script lang="ts">
	import NumberField from "@components/shared/NumberField.svelte";
	import type { DownloadsState } from "./downloadsState.svelte";

	let { downloads }: { downloads: DownloadsState } = $props();

	let url = $state("");
	let limitInput = $state("0");

	function submitUrl(event: SubmitEvent) {
		event.preventDefault();
		const link = url.trim();
		if (!link) return;
		downloads.enqueue(link);
		url = "";
	}

	function applyLimit(event: SubmitEvent) {
		event.preventDefault();
		const kbps = Math.max(0, Math.floor(Number(limitInput)) || 0);
		downloads.setSpeedLimit(kbps);
	}
</script>

<div class="controls">
	<form class="form" onsubmit={submitUrl}>
		<label class="field grow">
			<span class="label">Add a direct link</span>
			<input class="input" type="url" placeholder="https://..." bind:value={url} />
		</label>
		<button class="btn sm primary" type="submit" disabled={!url.trim()}>Add</button>
	</form>

	<form class="form" onsubmit={applyLimit}>
		<div class="speed-limit">
			<NumberField
				id="speed-limit"
				label="Speed limit (KB/s, 0 = none)"
				value={limitInput}
				min={0}
				step={64}
				oninput={(value) => (limitInput = value)}
			/>
		</div>
		<button class="btn sm" type="submit">Apply</button>
	</form>
</div>

<style lang="scss">
	.controls {
		display: flex;
		flex-wrap: wrap;
		gap: 14px 20px;
		align-items: flex-end;
		margin-bottom: 20px;
	}

	.form {
		display: flex;
		align-items: flex-end;
		gap: 8px;
		margin: 0;
		min-width: 0;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;

		&.grow {
			flex: 1 1 260px;
		}
	}

	.label {
		font-size: 0.75rem;
		color: var(--text-muted);
	}

	.speed-limit {
		width: 150px;
	}
</style>
