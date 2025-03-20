import { beforeAll } from "vitest";
import ResizeObserver from "resize-observer-polyfill";

global.CSS = { supports: () => false } as never;
global.ResizeObserver = ResizeObserver;

beforeAll(() => {
  global.CSS = {
    supports: () => false,
    escape: (str: string) => str,
  } as never;
});
