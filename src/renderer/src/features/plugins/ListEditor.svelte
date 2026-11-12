<script lang="ts">
	import Dropdown, { asOptions } from "@components/shared/Dropdown.svelte";
	import {
		addItem,
		moveItem,
		removeItem,
		unusedOptions,
		type SettingsField
	} from "../../lib/features/settingsForm";

	let {
		field,
		items,
		onchange,
	}: {
		field: SettingsField;
		items: string[];
		onchange: (next: string[]) => void;
	} = $props();

	let draft = $state("");
	let chosen = $state("");
	let choices = $derived(unusedOptions(field, items));
	let choiceItems = $derived(asOptions(choices));
	let usesOptions = $derived(field.options.length > 0);
	let pick = $derived(choices.includes(chosen) ? chosen : (choices[0] ?? ""));

	function addDraft() {
		onchange(addItem(items, usesOptions ? pick : draft));
		draft = "";
	}

	function onDraftKey(event: KeyboardEvent) {
		if (event.key === "Enter") {
			event.preventDefault();
			addDraft();
		}
	}
</script>

<div class="list">
	{#each items as item, index (index)}
		<div class="row">
			<span class="name">{item}</span>
			<button
				class="btn sm"
				type="button"
				aria-label="Move {item} up"
				disabled={index === 0}
				onclick={() => onchange(moveItem(items, index, -1))}>Up</button
			>
			<button
				class="btn sm"
				type="button"
				aria-label="Move {item} down"
				disabled={index === items.length - 1}
				onclick={() => onchange(moveItem(items, index, 1))}>Down</button
			>
			<button
				class="btn sm"
				type="button"
				aria-label="Remove {item}"
				onclick={() => onchange(removeItem(items, index))}>Remove</button
			>
		</div>
	{/each}

	{#if usesOptions}
		{#if choices.length > 0}
			<div class="add">
				<div class="choose">
					<Dropdown
						aria-label="Item to add"
						options={choiceItems}
						value={pick}
						onChange={(value) => (chosen = value)}
					/>
				</div>
				<button class="btn sm primary" type="button" onclick={addDraft}>Add</button>
			</div>
		{/if}
	{:else}
		<div class="add">
			<input
				class="input"
				type="text"
				aria-label="New item"
				bind:value={draft}
				onkeydown={onDraftKey}
			/>
			<button class="btn sm primary" type="button" onclick={addDraft}>Add</button>
		</div>
	{/if}
</div>

<style lang="scss">
	.list {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.row,
	.add {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.name {
		flex: 1;
		min-width: 0;
		font-weight: 700;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.add .input,
	.choose {
		flex: 1;
		min-width: 0;
	}
</style>
