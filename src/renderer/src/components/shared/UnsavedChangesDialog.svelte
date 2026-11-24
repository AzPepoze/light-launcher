<script lang="ts">
	import Button from "./Button.svelte";
	import Modal from "./Modal.svelte";
	import { leavePrompt, type LeaveChoice } from "@stores/leaveGuard";

	function answer(choice: LeaveChoice) {
		$leavePrompt?.answer(choice);
	}
</script>

<Modal show={$leavePrompt !== null} title="Unsaved changes" showDone={false} onClose={() => answer("stay")}>
	<p class="message">Save your changes before you leave?</p>

	<div slot="footer" class="actions">
		<Button onclick={() => answer("stay")}>Stay</Button>
		<Button variant="secondary" onclick={() => answer("discard")}>Discard</Button>
		<Button variant="primary" onclick={() => answer("save")}>Save and leave</Button>
	</div>
</Modal>

<style lang="scss">
	.message {
		margin: 0;
		font-size: 0.95rem;
		color: var(--text-muted);
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: 8px;
	}
</style>
