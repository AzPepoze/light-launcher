<script lang="ts">
	export let isSelected: boolean = false;
	export let inline: boolean = false;
</script>

<div class={inline ? "selection-checkbox" : "selection-overlay"}>
	<div class="checkbox" class:checked={isSelected}>
		{#if isSelected}
			<span class="material-icons" style="font-size: 16px;">check</span>
		{/if}
	</div>
</div>

<style lang="scss">
	.selection-overlay {
		position: absolute;
		top: 10px;
		right: 10px;
		z-index: 10;
	}

	.selection-checkbox {
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.checkbox {
		width: 22px;
		height: 22px;
		border-radius: var(--radius-sm);
		border: 2px solid rgba(255, 255, 255, 0.4);
		background: rgba(0, 0, 0, 0.4);
		backdrop-filter: blur(4px);
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all var(--transition-fast);

		&.checked {
			background: var(--accent-primary) !important;
			border-color: var(--accent-primary) !important;
			color: var(--bg-base);
		}
	}
</style>
