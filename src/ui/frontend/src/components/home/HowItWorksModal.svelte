<script lang="ts">
	import Modal from "../shared/Modal.svelte";

	export let show: boolean;
	export let onClose: () => void;
</script>

<Modal {show} title="How it works" {onClose}>
	<div class="help-content">
		<section>
			<h3>Adding Games</h3>
			<p>
				Go to the <strong>Run</strong> page, select your game
				executable (and launcher if applicable), configure your
				settings, and click <strong>LAUNCH GAME</strong>.
			</p>
			<p>
				After the first run, the game will automatically appear here
				in <strong>Quick Launch</strong>.
			</p>
		</section>

		<section>
			<h3>Quick Launch</h3>
			<p>
				Click on any game card in this section to start it
				immediately with its saved configuration.
			</p>
		</section>

		<section>
			<h3>Managing Sessions</h3>
			<p>
				Active game sessions are displayed at the top. You can
				terminate them if they become unresponsive.
			</p>
		</section>

		<section>
			<h3>CLI Usage</h3>
			<p>
				You can also launch games directly from your terminal or add
				them to your desktop entries:
			</p>
			<code class="help-code">light-launcher /path/to/game.exe</code>
		</section>
	</div>
</Modal>

<style lang="scss">
	.help-content {
		display: flex;
		flex-direction: column;
		gap: 24px;
		color: var(--text-main, #eee);

		section {
			h3 {
				margin: 0 0 8px 0;
				font-size: 1.1rem;
				color: var(--accent-primary, #fff);
			}

			p {
				margin: 0;
				line-height: 1.6;
				font-size: 0.95rem;
				color: var(--text-dim, #aaa);

				strong {
					color: var(--text-main, #eee);
				}
			}

			.help-code {
				display: block;
				background: rgba(0, 0, 0, 0.3);
				padding: 12px;
				border-radius: 8px;
				font-family: monospace;
				font-size: 0.85rem;
				color: var(--accent-primary, #fff);
				margin-top: 10px;
				border: 1px solid rgba(255, 255, 255, 0.05);
			}

			& + section {
				padding-top: 16px;
				border-top: 1px solid rgba(255, 255, 255, 0.05);
			}
		}
	}
</style>
