import "@testing-library/jest-dom/vitest";
import { afterEach, vi } from "vitest";
import { cleanup } from "@testing-library/react";

// Anything that formats a date renders it in the machine's timezone, so a test
// asserting on a rendered date passes or fails depending on where it runs. Pin
// it here rather than in the npm script, so running vitest directly gets the
// same answer CI does.
process.env.TZ = "UTC";

vi.mock("@/env");
vi.mock("@/hooks/useFocusTrap");

afterEach(cleanup);

HTMLDialogElement.prototype.showModal = function (this: HTMLDialogElement) {
  this.setAttribute("open", "");
};

HTMLDialogElement.prototype.close = function (this: HTMLDialogElement) {
  this.removeAttribute("open");
};

// jsdom does not implement scrollIntoView; components that scroll active items
// into view (e.g. the command palette) call it inside effects.
Element.prototype.scrollIntoView = function () {};

// jsdom does not implement ResizeObserver; Floating UI's autoUpdate needs it.
global.ResizeObserver = class {
  observe() {}
  unobserve() {}
  disconnect() {}
};
