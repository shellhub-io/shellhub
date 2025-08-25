import { describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { containersApi } from "@/api/http";
import useContainersStore from "@/store/modules/containers";
import { IContainer } from "@/interfaces/IContainer";

describe("Containers Pinia Store", () => {
  setActivePinia(createPinia());
  const mockContainers = new MockAdapter(containersApi.getAxios());
  const containersStore = useContainersStore();

  describe("initial state", () => {
    it("should have initial state values", () => {
      expect(containersStore.containers).toEqual([]);
      expect(containersStore.container).toEqual({});
      expect(containersStore.containerCount).toBe(0);
      expect(containersStore.showContainers).toBe(false);
    });
  });

  describe("actions", () => {
    it("should fetch container list successfully", async () => {
      const containersData = [
        { uid: "a582b47a42d", name: "Container 1" },
        { uid: "a582b47a42e", name: "Container 2" },
      ];

      mockContainers.onGet("http://localhost:3000/api/containers?page=1&per_page=10&status=accepted").reply(200, containersData, {
        "x-total-count": "2",
      });

      await containersStore.fetchContainerList({ page: 1, perPage: 10, status: "accepted" });

      expect(containersStore.containers).toEqual(containersData);
      expect(containersStore.containerCount).toBe(2);
    });

    it("should handle empty container list", async () => {
      mockContainers.onGet("http://localhost:3000/api/containers?page=1&per_page=10&status=accepted").reply(200, [], {
        "x-total-count": "0",
      });

      await containersStore.fetchContainerList({ page: 1, perPage: 10, status: "accepted" });

      expect(containersStore.containers).toEqual([]);
      expect(containersStore.containerCount).toBe(0);
    });

    it("should fetch container by ID", async () => {
      const containerData = { uid: "a582b47a42d", name: "Container 1" };

      mockContainers.onGet("http://localhost:3000/api/containers/a582b47a42d").reply(200, containerData);

      await containersStore.getContainer("a582b47a42d");

      expect(containersStore.container).toEqual(containerData);
    });

    it("should remove container", async () => {
      mockContainers.onDelete("http://localhost:3000/api/containers/a582b47a42d").reply(200);

      await expect(containersStore.removeContainer("a582b47a42d")).resolves.not.toThrow();
    });

    it("should rename container", async () => {
      const renameData = { uid: "a582b47a42d", name: { name: "Updated Container 1" } };

      // Set initial container state
      containersStore.container = { uid: "a582b47a42d", name: "Container 1" } as IContainer;

      mockContainers.onPut("http://localhost:3000/api/containers/a582b47a42d").reply(200);

      await containersStore.renameContainer(renameData);

      expect(containersStore.container.name).toBe("Updated Container 1");
    });

    it("should accept container", async () => {
      mockContainers.onPatch("http://localhost:3000/api/containers/a582b47a42d/accept").reply(200);

      await expect(containersStore.acceptContainer("a582b47a42d")).resolves.not.toThrow();
    });

    it("should reject container", async () => {
      mockContainers.onPatch("http://localhost:3000/api/containers/a582b47a42d/reject").reply(200);

      await expect(containersStore.rejectContainer("a582b47a42d")).resolves.not.toThrow();
    });

    it("should handle fetch container list error", async () => {
      mockContainers.onGet("http://localhost:3000/api/containers?page=1&per_page=10&status=accepted").reply(500);

      await expect(containersStore.fetchContainerList({ page: 1, perPage: 10, status: "accepted" })).rejects.toThrow();

      expect(containersStore.containers).toEqual([]);
      expect(containersStore.containerCount).toBe(0);
    });

    it("should handle fetch container by ID error", async () => {
      mockContainers.onGet("http://localhost:3000/api/containers/a582b47a42d").reply(404);

      await expect(containersStore.getContainer("a582b47a42d")).rejects.toThrow();

      expect(containersStore.container).toEqual({});
    });

    it("should handle remove container error", async () => {
      mockContainers.onDelete("http://localhost:3000/api/containers/a582b47a42d").reply(500);

      await expect(containersStore.removeContainer("a582b47a42d")).rejects.toThrow();
    });

    it("should handle rename container error", async () => {
      const renameData = { uid: "a582b47a42d", name: { name: "Updated Container 1" } };

      mockContainers.onPut("http://localhost:3000/api/containers/a582b47a42d").reply(400);

      await expect(containersStore.renameContainer(renameData)).rejects.toThrow();
    });

    it("should handle accept container error", async () => {
      mockContainers.onPatch("http://localhost:3000/api/containers/a582b47a42d/accept").reply(500);

      await expect(containersStore.acceptContainer("a582b47a42d")).rejects.toThrow();
    });

    it("should handle reject container error", async () => {
      mockContainers.onPatch("http://localhost:3000/api/containers/a582b47a42d/reject").reply(500);

      await expect(containersStore.rejectContainer("a582b47a42d")).rejects.toThrow();
    });
  });
});
