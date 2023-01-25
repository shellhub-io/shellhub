import { beforeAll } from "vitest";
import ResizeObserver from "resize-observer-polyfill";

(global as any).CSS = { supports: () => false };
global.ResizeObserver = ResizeObserver;

beforeAll(() => {
  global.CSS = {
    supports: (str: string) => false,
    escape: (str: string) => str,
  };
});
