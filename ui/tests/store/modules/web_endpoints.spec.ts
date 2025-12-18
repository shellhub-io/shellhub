import { createPinia, setActivePinia } from "pinia";
import { beforeEach, afterEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { webEndpointsApi } from "@/api/http";
import useWebEndpointsStore from "@/store/modules/web_endpoints";
import { IWebEndpoint } from "@/interfaces/IWebEndpoints";
import { IDevice } from "@/interfaces/IDevice";

const mockWebEndpointBase: IWebEndpoint = {
  address: "localhost:8080",
  port: 8080,
  expires_in: "2026-03-31T23:59:59Z",
  full_address: "http://localhost:8080",
  host: "localhost",
  device_uid: "device-123",
  device: {
    uid: "device-123",
    name: "Test Device",
  } as IDevice,
};

describe("WebEndpoints Store", () => {
  let mockWebEndpointsApi: MockAdapter;
  let store: ReturnType<typeof useWebEndpointsStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    mockWebEndpointsApi = new MockAdapter(webEndpointsApi.getAxios());
    store = useWebEndpointsStore();
  });

  afterEach(() => { mockWebEndpointsApi.reset(); });

  describe("Initial State", () => {
    it("should have correct default values", () => {
      expect(store.webEndpoints).toEqual([]);
      expect(store.webEndpointCount).toBe(0);
      expect(store.showWebEndpoints).toBe(false);
    });
  });

  describe("fetchWebEndpointsList", () => {
    const fetchListUrl = "http://localhost:3000/api/web-endpoints?page=1&per_page=10";

    it("should fetch web endpoints list successfully with default params", async () => {
      const mockWebEndpoints = [
        mockWebEndpointBase,
        { ...mockWebEndpointBase, uid: "def456", address: "127.0.0.1", port: 8081 },
      ];

      mockWebEndpointsApi
        .onGet(fetchListUrl)
        .reply(200, mockWebEndpoints, {
          "x-total-count": "2",
        });

      await store.fetchWebEndpointsList();

      expect(store.webEndpoints).toEqual(mockWebEndpoints);
      expect(store.webEndpointCount).toBe(2);
      expect(store.showWebEndpoints).toBe(true);
    });

    it("should handle empty web endpoints list", async () => {
      mockWebEndpointsApi
        .onGet(fetchListUrl)
        .reply(200, [], {
          "x-total-count": "0",
        });

      await store.fetchWebEndpointsList();

      expect(store.webEndpoints).toEqual([]);
      expect(store.webEndpointCount).toBe(0);
      expect(store.showWebEndpoints).toBe(false);
    });

    it("should reset state when request fails", async () => {
      mockWebEndpointsApi
        .onGet(fetchListUrl)
        .reply(403, { message: "Insufficient permissions" });

      await expect(store.fetchWebEndpointsList()).rejects.toBeAxiosErrorWithStatus(403);

      expect(store.webEndpoints).toEqual([]);
      expect(store.webEndpointCount).toBe(0);
      expect(store.showWebEndpoints).toBe(false);
    });

    it("should reset state when network error occurs", async () => {
      mockWebEndpointsApi
        .onGet(fetchListUrl)
        .networkError();

      await expect(store.fetchWebEndpointsList()).rejects.toThrow();

      expect(store.webEndpoints).toEqual([]);
      expect(store.webEndpointCount).toBe(0);
      expect(store.showWebEndpoints).toBe(false);
    });
  });

  describe("createWebEndpoint", () => {
    const createUrl = "http://localhost:3000/api/web-endpoints";

    it("should create web endpoint successfully", async () => {
      const payload = { uid: "abc123", host: "localhost", port: 8080, ttl: -1 };

      mockWebEndpointsApi.onPost(createUrl).reply(200);

      await expect(store.createWebEndpoint(payload)).resolves.not.toThrow();
    });

    it("should handle validation error when creating web endpoint", async () => {
      const payload = { uid: "abc123", host: "localhost", port: 8080, ttl: -1 };

      mockWebEndpointsApi
        .onPost(createUrl)
        .reply(400, { message: "Invalid web endpoint data" });

      await expect(store.createWebEndpoint(payload)).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle network error when creating web endpoint", async () => {
      const payload = { uid: "abc123", host: "localhost", port: 8080, ttl: -1 };

      mockWebEndpointsApi
        .onPost(createUrl)
        .networkError();

      await expect(store.createWebEndpoint(payload)).rejects.toThrow();
    });
  });

  describe("deleteWebEndpoint", () => {
    it("should delete web endpoint successfully", async () => {
      const address = "localhost";

      mockWebEndpointsApi.onDelete(`http://localhost:3000/api/web-endpoints/${address}`).reply(200);

      await expect(store.deleteWebEndpoint(address)).resolves.not.toThrow();
    });

    it("should handle not found error when deleting web endpoint", async () => {
      const address = "localhost";

      mockWebEndpointsApi
        .onDelete(`http://localhost:3000/api/web-endpoints/${address}`)
        .reply(404, { message: "Web endpoint not found" });

      await expect(store.deleteWebEndpoint(address)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle network error when deleting web endpoint", async () => {
      const address = "localhost";

      mockWebEndpointsApi
        .onDelete(`http://localhost:3000/api/web-endpoints/${address}`)
        .networkError();

      await expect(store.deleteWebEndpoint(address)).rejects.toThrow();
    });
  });
});
