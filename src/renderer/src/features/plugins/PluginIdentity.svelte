<script lang="ts">
	import { identityHue, identityInitial } from "../../lib/features/identityHue";

	let { id, name, large = false }: { id: string; name: string; large?: boolean } = $props();

	const hue = $derived(identityHue(id));
	const initial = $derived(identityInitial(name));
</script>

<span class="identity" class:large style:--identity-hue={hue} aria-hidden="true">{initial}</span>

<style lang="scss">
	.identity {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		flex-shrink: 0;
		border-radius: var(--radius-sm);
		background: hsl(var(--identity-hue) 62% 80%);
		color: #111;
		font-size: 1rem;
		font-weight: 800;

		&.large {
			width: 52px;
			height: 52px;
			font-size: 1.4rem;
		}
	}
</style>
