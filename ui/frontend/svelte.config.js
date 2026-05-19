import sveltePreprocess from "svelte-preprocess";

export default {
	preprocess: sveltePreprocess(),
	onwarn: (warning, handler) => {
		if (warning.code && (warning.code.startsWith("a11y_") || warning.code.startsWith("a11y-"))) {
			return;
		}
		handler(warning);
	},
	compilerOptions: {
		warningFilter: (warning) => {
			if (warning.code && (warning.code.startsWith("a11y_") || warning.code.startsWith("a11y-"))) {
				return false;
			}
			return true;
		}
	}
};
