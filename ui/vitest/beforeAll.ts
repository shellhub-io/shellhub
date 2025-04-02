import MockAdapter from "axios-mock-adapter";
import { beforeAll } from "vitest";
import ResizeObserver from "resize-observer-polyfill";
import { systemApi } from "../src/api/http";

const mockedSystemInfo = {
  authentication: {
    local: false,
    saml: true,
  },
  endpoints: {
    api: "localhost:80",
    ssh: "localhost:22",
  },
  setup: true,
  version: "v0.18.3",
};

const systemApiMock = new MockAdapter(systemApi.getAxios());
systemApiMock.onGet("http://localhost:3000/info").reply(200, mockedSystemInfo);

global.CSS = { supports: () => false } as never;
global.ResizeObserver = ResizeObserver;

beforeAll(() => {
  global.CSS = {
    supports: () => false,
    escape: (str: string) => str,
  } as never;
});
