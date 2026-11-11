package main

import "encoding/json"

// bannerScript injects a small non-interactive banner above the page. It never blocks
// clicks or changes the page itself, and it is not added to child frames.
func bannerScript(message string) string {
	if message == "" {
		return ""
	}
	text, err := json.Marshal(message)
	if err != nil {
		return ""
	}
	return `(function () {
	if (window.top !== window) return;
	window.addEventListener("DOMContentLoaded", function () {
		if (!document.body) return;
		var banner = document.createElement("div");
		banner.textContent = ` + string(text) + `;
		banner.style.cssText = "position:fixed;top:0;left:0;right:0;z-index:2147483647;" +
			"padding:6px 12px;background:#1f2937;color:#fff;font:13px system-ui,sans-serif;" +
			"pointer-events:none;text-align:center;";
		document.body.appendChild(banner);
	});
})();`
}
