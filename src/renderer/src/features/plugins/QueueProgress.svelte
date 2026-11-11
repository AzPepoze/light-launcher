<script lang="ts">
	import { pluginProgress } from "./pluginCopy";

	let { position, total }: { position: number; total: number } = $props();

	const percent = $derived((position / total) * 100);
</script>

<div
	class="track"
	role="progressbar"
	aria-label="Install queue"
	aria-valuemin={1}
	aria-valuemax={total}
	aria-valuenow={position}
	aria-valuetext={pluginProgress(position, total)}
>
	<div class="fill" style:width="{percent}%"></div>
</div>

<style lang="scss">
	.track {
		height: 4px;
		border-radius: var(--radius-pill);
		background: var(--bg-elevated);
		overflow: hidden;
		margin-bottom: 16px;
	}

	.fill {
		height: 100%;
		border-radius: var(--radius-pill);
		background: var(--accent-primary);
		transition: width var(--motion-base) var(--ease-out);
	}
</style>
