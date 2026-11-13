<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import Icon from "@components/shared/Icon.svelte";

	export let onLaunch: () => void;
</script>

<div class="launch-btn">
	<Button
		variant="primary"
		style="width: 100%; height: 60px; padding: 0 32px; gap: 12px; font-size: 1.15rem; letter-spacing: 2px;"
		onclick={onLaunch}
	>
		<Icon name="rocket_launch" size={28} class="run-icon" />
		<span>LAUNCH GAME</span>
	</Button>
</div>

<style lang="scss">
	.launch-btn {
		width: 100%;

		:global(.button) {
			box-shadow: 0 4px 14px var(--accent-glow);
		}

		&:hover :global(.button) {
			box-shadow: 0 8px 24px var(--accent-glow);
		}

		:global(.run-icon) {
			transition: transform var(--transition-spring);
		}

		&:hover :global(.run-icon) {
			transform: rotate(-15deg);
		}
	}
</style>
