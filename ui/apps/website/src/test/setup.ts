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

// jsdom does not implement window.scrollTo; provide a no-op shim.
Object.defineProperty(window, "scrollTo", {
  writable: true,
  configurable: true,
  value: () => {},
});
