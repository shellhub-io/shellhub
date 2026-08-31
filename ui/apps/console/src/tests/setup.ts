import "@testing-library/jest-dom/vitest";
import { afterEach, vi } from "vitest";
import type { Mock } from "vitest";
import { cleanup } from "@testing-library/react";

globalThis.mockSdkGen = <T extends Record<string, Mock>>(mocks: T): T => {
  vi.doMock("@/client/sdk.gen", async (importOriginal) => ({
    ...(await (importOriginal as () => Promise<Record<string, unknown>>)()),
    ...mocks,
  }));
  return mocks;
};

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

Element.prototype.scrollIntoView = function () {};

global.ResizeObserver = class {
  observe() {}
  unobserve() {}
  disconnect() {}
};
