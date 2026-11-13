import "material-icons/iconfont/material-icons.css";
import "material-symbols/rounded.css";
import "./style.scss";
import { mount } from "svelte";
import App from "./App.svelte";
import { installAutoscroll } from "@lib/autoscroll";
import { signalReady } from "@lib/transport";

const app = mount(App, {
	target: document.getElementById("app") as HTMLElement
});

installAutoscroll();

signalReady();

export default app;
