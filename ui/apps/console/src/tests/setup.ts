import "@testing-library/jest-dom/vitest";
import { afterAll, afterEach, beforeAll, vi } from "vitest";
import { cleanup } from "@testing-library/react";
import { server } from "./msw";

beforeAll(async () => {
  await import("@/api/fetchInterceptors");
  server.listen({ onUnhandledFrame: "error" });
});
afterEach(() => {
  server.resetHandlers();
  cleanup();
});
afterAll(() => server.close());

process.env.TZ = "UTC";

vi.mock("@/env");
vi.mock("@/hooks/useFocusTrap");
vi.mock("@xterm/xterm", () => ({ Terminal: buildXtermFake() }));
vi.mock("@xterm/addon-fit", () => ({ FitAddon: buildXtermFake() }));
vi.mock("@xterm/addon-web-links", () => ({ WebLinksAddon: buildXtermFake() }));
vi.mock("@xterm/addon-webgl", () => ({ WebglAddon: buildXtermFake() }));

function buildXtermFake() {
  return class {
    cols = 80;
    rows = 24;
    options = {};
    constructor() {
      return new Proxy(this, {
        get: (target, key) =>
          key in target
            ? target[key as keyof typeof target]
            : () => ({ dispose: () => {} }),
      });
    }
  };
}

HTMLDialogElement.prototype.showModal = function (this: HTMLDialogElement) {
  this.setAttribute("open", "");
};

HTMLDialogElement.prototype.close = function (this: HTMLDialogElement) {
  this.removeAttribute("open");
};

Element.prototype.scrollIntoView = function () {};

global.ResizeObserver = class {
  observe() {}
  unobserve() {}
  disconnect() {}
};

global.IntersectionObserver = class {
  root = null;
  rootMargin = "";
  scrollMargin = "";
  thresholds = [];
  observe() {}
  unobserve() {}
  disconnect() {}
  takeRecords() {
    return [];
  }
};
