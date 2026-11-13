<script lang="ts">
	let { label = "Loading..." }: { label?: string } = $props();
</script>

<div class="loading-state" role="status">
	<span class="ring" aria-hidden="true"></span>
	<span class="label">{label}</span>
</div>

<style lang="scss">
	.loading-state {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		padding: 48px 24px;
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.ring {
		width: 18px;
		height: 18px;
		border: 2px solid var(--glass-border);
		border-top-color: var(--accent-primary);
		border-radius: 50%;
		animation: ll-spin 1s linear infinite;
	}
</style>
