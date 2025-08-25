import { createPinia, setActivePinia } from "pinia";
import { describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { webEndpointsApi } from "@/api/http";
import useWebEndpointsStore from "@/store/modules/web_endpoints";

const initialWebEndpoints = {
  data: [
    { uid: "abc123", address: "localhost", port: 8080, expires_in: 60000 },
    { uid: "def456", address: "127.0.0.1", port: 8081, expires_in: 60000 },
  ],
  headers: {
    "x-total-count": "2",
  },
};

describe("WebEndpoints store", () => {
  const mockWebEndpointsApi = new MockAdapter(webEndpointsApi.getAxios());
  setActivePinia(createPinia());
  const webEndpointsStore = useWebEndpointsStore();

  it("Returns web endpoints default variables", () => {
    expect(webEndpointsStore.webEndpoints).toEqual([]);
    expect(webEndpointsStore.webEndpointCount).toBe(0);
    expect(webEndpointsStore.showWebEndpoints).toBe(false);
  });

  it("Fetches web endpoints with default params and updates state", async () => {
    mockWebEndpointsApi
      .onGet("http://localhost:3000/api/web-endpoints?page=1&per_page=10")
      .reply(200, initialWebEndpoints.data, initialWebEndpoints.headers);

    await webEndpointsStore.fetchWebEndpointsList();

    expect(webEndpointsStore.webEndpoints).toEqual(initialWebEndpoints.data);
    expect(webEndpointsStore.webEndpointCount).toBe(2);
    expect(webEndpointsStore.showWebEndpoints).toBe(true);
  });

  it("Creates a new web endpoint", async () => {
    const payload = { uid: "abc123", host: "localhost", port: 8080, ttl: -1 };
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(200);
    const storeSpy = vi.spyOn(webEndpointsStore, "createWebEndpoint");
    await webEndpointsStore.createWebEndpoint(payload);

    expect(storeSpy).not.toThrow();
  });

  it("Deletes a web endpoint", async () => {
    const address = "localhost";
    mockWebEndpointsApi.onDelete(`http://localhost:3000/api/web-endpoints/${address}`).reply(200);
    const storeSpy = vi.spyOn(webEndpointsStore, "deleteWebEndpoint");

    await webEndpointsStore.deleteWebEndpoint(address);

    expect(storeSpy).not.toThrow();
  });
});
