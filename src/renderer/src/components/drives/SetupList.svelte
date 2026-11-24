<script lang="ts">
	import type { SetupEntry } from "@shared";
	import BrowseInput from "@components/shared/BrowseInput.svelte";
	import Dropdown from "@components/shared/Dropdown.svelte";
	import IconButton from "@components/shared/IconButton.svelte";
	import { lettersAvailableTo, letterOptions } from "@lib/driveSetup";

	let { entries = $bindable([]) }: { entries: SetupEntry[] } = $props();

	function setLetter(index: number, letter: string) {
		entries[index].letter = letter;
	}

	function removeRow(index: number) {
		entries.splice(index, 1);
	}
</script>

<ul class="setup-list">
	{#each entries as entry, index (index)}
		<li class="setup-row">
			<div class="row-controls">
				<div class="letter-field">
					<Dropdown
						options={letterOptions(lettersAvailableTo(entries, entry.letter))}
						value={entry.letter}
						onChange={(letter) => setLetter(index, letter)}
					/>
				</div>
				<div class="path-field">
					<BrowseInput bind:value={entries[index].path} />
				</div>
				<IconButton icon="delete" label="Remove drive" tone="danger" onclick={() => removeRow(index)} />
			</div>
			{#if entry.letter === "c:"}
				<p class="warning">
					C: replaces the prefix's own C: drive. Remove this row to go back to Wine's default.
				</p>
			{/if}
		</li>
	{/each}
</ul>

<style lang="scss">
	.setup-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.setup-row {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 16px 18px;
		background: var(--bg-input);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-md);
	}

	.row-controls {
		display: flex;
		align-items: flex-end;
		gap: 12px;
	}

	.letter-field {
		flex: 0 0 130px;
	}

	.path-field {
		flex: 1;
		min-width: 0;
	}

	.warning {
		margin: 0;
		color: var(--warning);
		font-size: 0.85rem;
		font-weight: 600;
	}
</style>
