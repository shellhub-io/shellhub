import "@testing-library/jest-dom/vitest";
import { afterEach, vi } from "vitest";
import { cleanup } from "@testing-library/react";

vi.mock("@/env");

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
