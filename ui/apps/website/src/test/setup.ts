import "@testing-library/jest-dom/vitest";
import { afterEach } from "vitest";
import { cleanup } from "@testing-library/react";

afterEach(cleanup);

// jsdom does not implement IntersectionObserver; provide a no-op shim so
// components that use it do not throw in tests.
class IntersectionObserverShim {
  observe() {}
  unobserve() {}
  disconnect() {}
}

Object.defineProperty(window, "IntersectionObserver", {
  writable: true,
  configurable: true,
  value: IntersectionObserverShim,
});

Object.defineProperty(window, "scrollTo", {
  writable: true,
  configurable: true,
  value: () => {},
});

global.ResizeObserver = class {
  observe() {}
  unobserve() {}
  disconnect() {}
};

Element.prototype.scrollIntoView = function () {};
