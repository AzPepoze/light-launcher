import { describe, expect, it } from "vitest";
import { autoscrollSpeed, findScrollableAncestor, isAutoscrollBlocked } from "./autoscroll";

interface FakeStyle {
	overflowX: string;
	overflowY: string;
}

interface FakeDocument {
	scrollingElement: FakeElement | null;
	defaultView: {
		getComputedStyle(element: FakeElement): FakeStyle;
	};
}

class FakeElement {
	tagName = "DIV";
	parentElement: FakeElement | null = null;
	clientHeight = 100;
	scrollHeight = 100;
	clientWidth = 100;
	scrollWidth = 100;
	overflowX = "visible";
	overflowY = "visible";
	attributes: Record<string, string> = {};
	ownerDocument: FakeDocument;

	constructor(document: FakeDocument) {
		this.ownerDocument = document;
	}

	getAttribute(name: string): string | null {
		return this.attributes[name] ?? null;
	}

	append(child: FakeElement): this {
		child.parentElement = this;
		return this;
	}
}

function createDocument(): FakeDocument {
	return {
		scrollingElement: null,
		defaultView: {
			getComputedStyle: (element) => ({
				overflowX: element.overflowX,
				overflowY: element.overflowY
			})
		}
	};
}

function createElement(document: FakeDocument, init: Partial<FakeElement> = {}): FakeElement {
	const element = new FakeElement(document);
	return Object.assign(element, init);
}

function asElement(element: FakeElement): HTMLElement {
	return element as unknown as HTMLElement;
}

describe("autoscrollSpeed", () => {
	it("stays still inside the 12 px dead zone", () => {
		expect(autoscrollSpeed(0)).toBe(0);
		expect(autoscrollSpeed(12)).toBe(0);
		expect(autoscrollSpeed(-12)).toBe(0);
	});

	it("grows with the distance beyond the dead zone", () => {
		const near = autoscrollSpeed(20);
		const far = autoscrollSpeed(80);
		expect(near).toBeGreaterThan(0);
		expect(far).toBeGreaterThan(near);
	});

	it("scrolls the other way for negative offsets", () => {
		expect(autoscrollSpeed(-40)).toBe(-autoscrollSpeed(40));
	});

	it("caps the speed", () => {
		expect(autoscrollSpeed(100_000)).toBe(autoscrollSpeed(1_000_000));
	});

	it("returns 0 for non-finite offsets", () => {
		expect(autoscrollSpeed(Number.NaN)).toBe(0);
		expect(autoscrollSpeed(Number.POSITIVE_INFINITY)).toBe(0);
		expect(autoscrollSpeed(Number.NEGATIVE_INFINITY)).toBe(0);
	});
});

describe("findScrollableAncestor", () => {
	it("returns the nearest vertically scrollable ancestor", () => {
		const document = createDocument();
		const outer = createElement(document, { scrollHeight: 400, overflowY: "auto" });
		const middle = createElement(document);
		const inner = createElement(document);
		outer.append(middle);
		middle.append(inner);

		expect(findScrollableAncestor(asElement(inner))).toBe(asElement(outer));
	});

	it("returns the nearest horizontally scrollable ancestor", () => {
		const document = createDocument();
		const row = createElement(document, { scrollWidth: 600, overflowX: "auto" });
		const child = createElement(document);
		row.append(child);

		expect(findScrollableAncestor(asElement(child))).toBe(asElement(row));
	});

	it("accepts a scrollable element as the start", () => {
		const document = createDocument();
		const scroller = createElement(document, { scrollHeight: 400, overflowY: "scroll" });

		expect(findScrollableAncestor(asElement(scroller))).toBe(asElement(scroller));
	});

	it("skips overflow that only overflows without scrolling", () => {
		const document = createDocument();
		const visible = createElement(document, { scrollHeight: 400 });
		const hidden = createElement(document, { scrollHeight: 400, overflowY: "hidden" });
		const child = createElement(document);
		hidden.append(child);

		expect(findScrollableAncestor(asElement(child))).toBeNull();
		expect(findScrollableAncestor(asElement(visible))).toBeNull();
	});

	it("falls back to the page scroller", () => {
		const document = createDocument();
		const root = createElement(document, { scrollHeight: 2000 });
		document.scrollingElement = root;
		const child = createElement(document);

		expect(findScrollableAncestor(asElement(child))).toBe(asElement(root));
	});

	it("returns null when nothing can scroll", () => {
		expect(findScrollableAncestor(null)).toBeNull();
	});
});

describe("isAutoscrollBlocked", () => {
	it("allows plain elements", () => {
		const document = createDocument();
		expect(isAutoscrollBlocked(asElement(createElement(document)))).toBe(false);
	});

	it("blocks form fields", () => {
		const document = createDocument();
		expect(isAutoscrollBlocked(asElement(createElement(document, { tagName: "INPUT" })))).toBe(
			true
		);
		expect(isAutoscrollBlocked(asElement(createElement(document, { tagName: "TEXTAREA" })))).toBe(
			true
		);
	});

	it("blocks links with an href", () => {
		const document = createDocument();
		const link = createElement(document, { tagName: "A", attributes: { href: "/games" } });
		const child = createElement(document);
		link.append(child);

		expect(isAutoscrollBlocked(asElement(link))).toBe(true);
		expect(isAutoscrollBlocked(asElement(child))).toBe(true);
	});

	it("allows anchors without an href", () => {
		const document = createDocument();
		expect(isAutoscrollBlocked(asElement(createElement(document, { tagName: "A" })))).toBe(false);
	});

	it("blocks editable regions and their children", () => {
		const document = createDocument();
		const editable = createElement(document, {
			attributes: { contenteditable: "plaintext-only" }
		});
		const child = createElement(document);
		editable.append(child);

		expect(isAutoscrollBlocked(asElement(editable))).toBe(true);
		expect(isAutoscrollBlocked(asElement(child))).toBe(true);
	});

	it("allows contenteditable=false", () => {
		const document = createDocument();
		const element = createElement(document, { attributes: { contenteditable: "false" } });

		expect(isAutoscrollBlocked(asElement(element))).toBe(false);
	});
});
