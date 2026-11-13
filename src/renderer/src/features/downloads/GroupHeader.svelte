<script lang="ts">
	import Badge from "@components/shared/Badge.svelte";
	import Button from "@components/shared/Button.svelte";
	import IconButton from "@components/shared/IconButton.svelte";
	import { type RowGroup, groupTitle } from "./downloadGroups";

	let {
		group,
		count,
		open = false,
		onToggle,
		onClear
	}: {
		group: RowGroup;
		count: number;
		/** Only the finished group can be opened and closed. */
		open?: boolean;
		onToggle?: () => void;
		onClear?: () => void;
	} = $props();

	const title = $derived(groupTitle(group));
</script>

<li class="group-header">
	<h2>
		{#if group === "finished"}
			<Button variant="text" size="sm" icon={open ? "expand_less" : "expand_more"} onclick={onToggle}>
				{title}
			</Button>
		{:else}
			{title}
		{/if}
	</h2>
	<Badge>{count}</Badge>
	{#if group === "finished" && onClear}
		<span class="clear">
			<IconButton icon="clear_all" label="Clear finished list" onclick={onClear} />
		</span>
	{/if}
</li>

<style lang="scss">
	.group-header {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-top: 12px;
		list-style: none;

		h2 {
			margin: 0;
			color: var(--text-main);
			font-size: 0.95rem;
			font-weight: 700;
		}
	}

	.clear {
		margin-left: auto;
	}
</style>
