// Defines window.llShell and sends outside links to the default browser. Runs before each page script.
(function () {
	const pending = new Map();
	let nextId = 1;

	function request(method) {
		return new Promise((resolve) => {
			const id = nextId++;
			pending.set(id, resolve);
			window.llRequestPick(method, id);
		});
	}

	window.llShell = {
		pickFolder: () => request("pickFolder"),
		pickFile: () => request("pickFile"),
		_resolve(id, value) {
			const resolve = pending.get(id);
			pending.delete(id);
			if (resolve) resolve(value);
		}
	};

	document.addEventListener("click", (event) => {
		const link = event.target.closest("a[href]");
		if (!link) return;
		const target = new URL(link.href, location.href);
		const isWeb = target.protocol === "http:" || target.protocol === "https:";
		if (isWeb && target.origin !== location.origin) {
			event.preventDefault();
			window.llOpenExternal(target.href);
		}
	});
})();
