// Defines window.llShell. Each call posts JSON to the native shell, which answers with _resolve.
(function () {
	const pending = new Map();
	let nextId = 1;

	function request(method) {
		return new Promise((resolve) => {
			const id = nextId++;
			pending.set(id, resolve);
			window.webkit.messageHandlers.llshell.postMessage(JSON.stringify({ id, method }));
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
})();
