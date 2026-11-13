<script lang="ts">
	/* Material Symbols Rounded icon. Decorative only: put the meaning in a label. */
	let {
		name,
		size = 20,
		filled = false,
		class: className = ""
	}: { name: string; size?: number; filled?: boolean; class?: string } = $props();
</script>

<span
	class="material-symbols-rounded {className}"
	aria-hidden="true"
	style:font-size="{size}px"
	style:font-variation-settings="'FILL' {filled ? 1 : 0}, 'wght' 400"
>{name}</span>

<style lang="scss">
	.material-symbols-rounded {
		display: inline-block;
		flex-shrink: 0;
		line-height: 1;
		user-select: none;
	}
</style>
