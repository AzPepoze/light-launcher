// Defines window.llShell. Each call sends a console message that the Go shell answers with _resolve.
(function () {
	const prefix = "llshell:";
	const pending = new Map();
	let nextId = 1;

	function request(method) {
		return new Promise((resolve) => {
			const id = nextId++;
			pending.set(id, resolve);
			console.debug(prefix + JSON.stringify({ id, method }));
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
