import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { namespacesApi } from "@/api/http";
import useConnectorStore from "@/store/modules/connectors";

describe("Connectors Pinia Store", () => {
  setActivePinia(createPinia());
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());
  let connectorStore: ReturnType<typeof useConnectorStore>;

  beforeEach(() => {
    connectorStore = useConnectorStore();
  });

  afterEach(() => {
    mockNamespacesApi.reset();
  });

  describe("initial state", () => {
    it("should have initial state values", () => {
      expect(connectorStore.connectors).toEqual([]);
      expect(connectorStore.connector).toEqual({});
      expect(connectorStore.connectorInfo).toEqual({});
      expect(connectorStore.connectorCount).toBe(0);
    });
  });

  describe("actions", () => {
    it("should fetch connector list successfully", async () => {
      const connectorsData = [
        { uid: "1", name: "Connector 1", enable: true, address: "127.0.0.1", port: 8080, secure: false },
        { uid: "2", name: "Connector 2", enable: false, address: "127.0.0.2", port: 8081, secure: true },
      ];

      mockNamespacesApi.onGet("http://localhost:3000/api/connector?page=1&per_page=10").reply(200, connectorsData, {
        "x-total-count": "2",
      });

      await connectorStore.fetchConnectorList({ page: 1, perPage: 10 });

      expect(connectorStore.connectors).toEqual(connectorsData);
      expect(connectorStore.connectorCount).toBe(2);
    });

    it("should handle empty connector list", async () => {
      mockNamespacesApi.onGet("http://localhost:3000/api/connector?page=1&per_page=10").reply(200, [], {
        "x-total-count": "0",
      });

      await connectorStore.fetchConnectorList({ page: 1, perPage: 10 });

      expect(connectorStore.connectors).toEqual([]);
      expect(connectorStore.connectorCount).toBe(0);
    });

    it("should fetch connector by ID", async () => {
      const connectorData = { uid: "1", name: "Connector 1", enable: true, address: "127.0.0.1", port: 8080, secure: false };

      mockNamespacesApi.onGet("http://localhost:3000/api/connector/1").reply(200, connectorData);

      await connectorStore.fetchConnectorById("1");

      expect(connectorStore.connector).toEqual(connectorData);
    });

    it("should get connector info", async () => {
      const infoData = { status: "connected", message: "Connection successful" };

      mockNamespacesApi.onGet("http://localhost:3000/api/connector/1/info").reply(200, infoData);

      await connectorStore.getConnectorInfo("1");

      expect(connectorStore.connectorInfo).toEqual(infoData);
    });

    it("should create connector", async () => {
      const createData = { name: "New Connector", enable: true, address: "127.0.0.3", port: 8082, secure: false };

      mockNamespacesApi.onPost("http://localhost:3000/api/connector").reply(201);

      await expect(connectorStore.createConnector(createData)).resolves.not.toThrow();
    });

    it("should update connector", async () => {
      const updateData = { uid: "1", name: "Updated Connector", enable: false, address: "127.0.0.1", port: 8080, secure: true };

      mockNamespacesApi.onPatch("http://localhost:3000/api/connector/1").reply(200);

      await expect(connectorStore.updateConnector(updateData)).resolves.not.toThrow();
    });

    it("should delete connector", async () => {
      mockNamespacesApi.onDelete("http://localhost:3000/api/connector/1").reply(200);

      await expect(connectorStore.deleteConnector("1")).resolves.not.toThrow();
    });

    it("should handle fetch connector list error", async () => {
      mockNamespacesApi.onGet("http://localhost:3000/api/connector?page=1&per_page=10").reply(500);

      await expect(connectorStore.fetchConnectorList({ page: 1, perPage: 10 })).rejects.toThrow();

      expect(connectorStore.connectors).toEqual([]);
      expect(connectorStore.connectorCount).toBe(0);
    });

    it("should handle fetch connector by ID error", async () => {
      mockNamespacesApi.onGet("http://localhost:3000/api/connector/1").reply(404);

      await expect(connectorStore.fetchConnectorById("1")).rejects.toThrow();
    });

    it("should handle get connector info error", async () => {
      mockNamespacesApi.onGet("http://localhost:3000/api/connector/1/info").reply(500);

      await expect(connectorStore.getConnectorInfo("1")).rejects.toThrow();
    });

    it("should handle create connector error", async () => {
      const createData = { name: "New Connector", enable: true, address: "127.0.0.3", port: 8082, secure: false };

      mockNamespacesApi.onPost("http://localhost:3000/api/connector").reply(400);

      await expect(connectorStore.createConnector(createData)).rejects.toThrow();
    });

    it("should handle update connector error", async () => {
      const updateData = { uid: "1", name: "Updated Connector", enable: false, address: "127.0.0.1", port: 8080, secure: true };

      mockNamespacesApi.onPatch("http://localhost:3000/api/connector/1").reply(400);

      await expect(connectorStore.updateConnector(updateData)).rejects.toThrow();
    });

    it("should handle delete connector error", async () => {
      mockNamespacesApi.onDelete("http://localhost:3000/api/connector/1").reply(500);

      await expect(connectorStore.deleteConnector("1")).rejects.toThrow();
    });
  });
});
