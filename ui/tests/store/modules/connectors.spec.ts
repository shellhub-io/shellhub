import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { namespacesApi } from "@/api/http";
import useConnectorStore from "@/store/modules/connectors";
import { IConnector, IConnectorPayload } from "@/interfaces/IConnector";
import { buildUrl } from "../../utils/url";

const mockConnectorBase: IConnector = {
  uid: "connector-123",
  tenant_id: "tenant-456",
  address: "127.0.0.1",
  port: 8080,
  status: {
    state: "connected",
    message: "Connection successful",
  },
  enable: true,
  secure: false,
};

const mockConnectorPayloadBase: IConnectorPayload = {
  uid: "connector-123",
  enable: true,
  secure: false,
  address: "127.0.0.1",
  port: 8080,
};

describe("Connectors Store", () => {
  let mockNamespacesApi: MockAdapter;
  let store: ReturnType<typeof useConnectorStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());
    store = useConnectorStore();
  });

  afterEach(() => { mockNamespacesApi.reset(); });

  describe("Initial State", () => {
    it("should have correct default values", () => {
      expect(store.connectors).toEqual([]);
      expect(store.connector).toEqual({});
      expect(store.connectorInfo).toEqual({});
      expect(store.connectorCount).toBe(0);
    });
  });

  describe("fetchConnectorList", () => {
    const baseUrl = "http://localhost:3000/api/connector";

    it("should fetch connector list successfully with pagination", async () => {
      const mockConnectors = [
        mockConnectorBase,
        { ...mockConnectorBase, uid: "connector-456", address: "127.0.0.2", port: 8081, enable: false, secure: true },
      ];

      mockNamespacesApi.onGet(buildUrl(baseUrl, { page: "1", per_page: "10" })).reply(200, mockConnectors, {
        "x-total-count": "2",
      });

      await store.fetchConnectorList({ page: 1, perPage: 10 });

      expect(store.connectors).toEqual(mockConnectors);
      expect(store.connectorCount).toBe(2);
    });

    it("should handle empty connector list", async () => {
      mockNamespacesApi.onGet(buildUrl(baseUrl, { page: "1", per_page: "10" })).reply(200, [], {
        "x-total-count": "0",
      });

      await store.fetchConnectorList({ page: 1, perPage: 10 });

      expect(store.connectors).toEqual([]);
      expect(store.connectorCount).toBe(0);
    });

    it("should reset state when request fails with permission error", async () => {
      mockNamespacesApi.onGet(buildUrl(baseUrl, { page: "1", per_page: "10" })).reply(403, { message: "Insufficient permissions" });

      await expect(store.fetchConnectorList({ page: 1, perPage: 10 })).rejects.toBeAxiosErrorWithStatus(403);

      expect(store.connectors).toEqual([]);
      expect(store.connectorCount).toBe(0);
    });

    it("should reset state when network error occurs", async () => {
      mockNamespacesApi.onGet(buildUrl(baseUrl, { page: "1", per_page: "10" })).networkError();

      await expect(store.fetchConnectorList({ page: 1, perPage: 10 })).rejects.toThrow();

      expect(store.connectors).toEqual([]);
      expect(store.connectorCount).toBe(0);
    });
  });

  describe("fetchConnectorById", () => {
    const generateFetchByIdUrl = (id: string) => `http://localhost:3000/api/connector/${id}`;

    it("should fetch connector by ID successfully", async () => {
      mockNamespacesApi.onGet(generateFetchByIdUrl("connector-123")).reply(200, mockConnectorBase);

      await store.fetchConnectorById("connector-123");

      expect(store.connector).toEqual(mockConnectorBase);
    });

    it("should handle not found error when fetching connector", async () => {
      mockNamespacesApi.onGet(generateFetchByIdUrl("connector-123")).reply(404, { message: "Connector not found" });

      await expect(store.fetchConnectorById("connector-123")).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw error when network error occurs", async () => {
      mockNamespacesApi.onGet(generateFetchByIdUrl("connector-123")).networkError();

      await expect(store.fetchConnectorById("connector-123")).rejects.toThrow();
    });
  });

  describe("getConnectorInfo", () => {
    const generateInfoUrl = (id: string) => `http://localhost:3000/api/connector/${id}/info`;

    it("should get connector info successfully", async () => {
      const mockInfo = { status: "connected", message: "Connection successful" };

      mockNamespacesApi.onGet(generateInfoUrl("connector-123")).reply(200, mockInfo);

      await store.getConnectorInfo("connector-123");

      expect(store.connectorInfo).toEqual(mockInfo);
    });

    it("should handle permission error when getting connector info", async () => {
      mockNamespacesApi.onGet(generateInfoUrl("connector-123")).reply(403, { message: "Forbidden" });

      await expect(store.getConnectorInfo("connector-123")).rejects.toBeAxiosErrorWithStatus(403);
    });

    it("should throw error when network error occurs", async () => {
      mockNamespacesApi.onGet(generateInfoUrl("connector-123")).networkError();

      await expect(store.getConnectorInfo("connector-123")).rejects.toThrow();
    });
  });

  describe("createConnector", () => {
    const createUrl = "http://localhost:3000/api/connector";

    it("should create connector successfully", async () => {
      const { uid: _uid, ...createData } = mockConnectorPayloadBase;

      mockNamespacesApi.onPost(createUrl).reply(201);

      await expect(store.createConnector(createData)).resolves.not.toThrow();
    });

    it("should handle validation errors when creating connector", async () => {
      const { uid: _uid, ...createData } = mockConnectorPayloadBase;

      mockNamespacesApi.onPost(createUrl).reply(400, { message: "Invalid request data" });

      await expect(store.createConnector(createData)).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should throw error when network error occurs", async () => {
      const { uid: _uid, ...createData } = mockConnectorPayloadBase;

      mockNamespacesApi.onPost(createUrl).networkError();

      await expect(store.createConnector(createData)).rejects.toThrow();
    });
  });

  describe("updateConnector", () => {
    const generateUpdateUrl = (id: string) => `http://localhost:3000/api/connector/${id}`;

    it("should update connector successfully", async () => {
      mockNamespacesApi.onPatch(generateUpdateUrl("connector-123")).reply(200);

      await expect(store.updateConnector(mockConnectorPayloadBase)).resolves.not.toThrow();
    });

    it("should handle not found error when updating connector", async () => {
      mockNamespacesApi.onPatch(generateUpdateUrl("connector-123")).reply(404, { message: "Connector not found" });

      await expect(store.updateConnector(mockConnectorPayloadBase)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw error when network error occurs", async () => {
      mockNamespacesApi.onPatch(generateUpdateUrl("connector-123")).networkError();

      await expect(store.updateConnector(mockConnectorPayloadBase)).rejects.toThrow();
    });
  });

  describe("deleteConnector", () => {
    const generateDeleteUrl = (id: string) => `http://localhost:3000/api/connector/${id}`;

    it("should delete connector successfully", async () => {
      mockNamespacesApi.onDelete(generateDeleteUrl("connector-123")).reply(200);

      await expect(store.deleteConnector("connector-123")).resolves.not.toThrow();
    });

    it("should handle permission error when deleting connector", async () => {
      mockNamespacesApi.onDelete(generateDeleteUrl("connector-123")).reply(403, { message: "Insufficient permissions" });

      await expect(store.deleteConnector("connector-123")).rejects.toBeAxiosErrorWithStatus(403);
    });

    it("should throw error when network error occurs", async () => {
      mockNamespacesApi.onDelete(generateDeleteUrl("connector-123")).networkError();

      await expect(store.deleteConnector("connector-123")).rejects.toThrow();
    });
  });
});
