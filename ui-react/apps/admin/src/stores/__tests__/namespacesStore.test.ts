import { describe, it, expect, vi, beforeEach } from "vitest";
import { useNamespacesStore } from "../namespacesStore";
import { useAuthStore } from "../authStore";
import { useConnectivityStore } from "../connectivityStore";

vi.mock("../../api/namespaces", () => ({
  getNamespaces: vi.fn(),
  getNamespace: vi.fn(),
  getNamespaceToken: vi.fn(),
  createNamespace: vi.fn(),
  updateNamespace: vi.fn(),
  deleteNamespace: vi.fn(),
  leaveNamespace: vi.fn(),
}));

import {
  getNamespaces,
  getNamespace,
  getNamespaceToken,
  createNamespace as createNamespaceApi,
  deleteNamespace as deleteNamespaceApi,
  leaveNamespace as leaveNamespaceApi,
} from "../../api/namespaces";

const mockedGetNamespaces = vi.mocked(getNamespaces);
const mockedGetNamespace = vi.mocked(getNamespace);
const mockedGetNamespaceToken = vi.mocked(getNamespaceToken);
const mockedCreateNamespace = vi.mocked(createNamespaceApi);
const mockedDeleteNamespace = vi.mocked(deleteNamespaceApi);
const mockedLeaveNamespace = vi.mocked(leaveNamespaceApi);

beforeEach(() => {
  useNamespacesStore.setState({
    namespaces: [],
    currentNamespace: null,
    loading: false,
    loaded: false,
    error: null,
  });

  useAuthStore.setState({
    token: "old-token",
    tenant: "old-tenant",
    role: "owner",
  });

  useConnectivityStore.getState().markUp();

  // Stub window.location
  Object.defineProperty(window, "location", {
    writable: true,
    value: { href: "", reload: vi.fn(), replace: vi.fn() },
  });

  vi.clearAllMocks();
});

describe("namespacesStore", () => {
  describe("fetch", () => {
    it("loads namespaces list", async () => {
      const ns = [{ tenant_id: "t1", name: "ns1" }];
      mockedGetNamespaces.mockResolvedValue(ns as never);

      await useNamespacesStore.getState().fetch();

      const state = useNamespacesStore.getState();
      expect(state.namespaces).toEqual(ns);
      expect(state.loaded).toBe(true);
      expect(state.loading).toBe(false);
    });

    it("sets error on failure when API is reachable", async () => {
      mockedGetNamespaces.mockRejectedValue(new Error("fail"));

      await useNamespacesStore.getState().fetch();

      const state = useNamespacesStore.getState();
      expect(state.error).toBe("Failed to load namespaces");
      expect(state.loaded).toBe(true);
    });

    it("sets API-down error when connectivity is lost", async () => {
      useConnectivityStore.getState().markDown();
      mockedGetNamespaces.mockRejectedValue(new Error("network"));

      await useNamespacesStore.getState().fetch();

      const state = useNamespacesStore.getState();
      expect(state.error).toBe("Unable to reach the API");
      expect(state.loaded).toBe(false);
    });
  });

  describe("switchNamespace", () => {
    it("gets new token and updates auth session", async () => {
      mockedGetNamespaceToken.mockResolvedValue({
        token: "new-token",
        role: "administrator",
      } as never);

      await useNamespacesStore.getState().switchNamespace("new-tenant");

      expect(mockedGetNamespaceToken).toHaveBeenCalledWith("new-tenant");

      const auth = useAuthStore.getState();
      expect(auth.token).toBe("new-token");
      expect(auth.tenant).toBe("new-tenant");
      expect(auth.role).toBe("administrator");

      expect(window.location.reload).toHaveBeenCalled();
    });
  });

  describe("createNamespace", () => {
    it("creates namespace, switches token, and reloads", async () => {
      mockedCreateNamespace.mockResolvedValue({
        tenant_id: "created-tenant",
        name: "new-ns",
      } as never);
      mockedGetNamespaceToken.mockResolvedValue({
        token: "created-token",
        role: "owner",
      } as never);

      await useNamespacesStore.getState().createNamespace("new-ns");

      expect(mockedCreateNamespace).toHaveBeenCalledWith("new-ns");
      expect(mockedGetNamespaceToken).toHaveBeenCalledWith("created-tenant");

      const auth = useAuthStore.getState();
      expect(auth.token).toBe("created-token");
      expect(auth.tenant).toBe("created-tenant");

      expect(window.location.reload).toHaveBeenCalled();
    });

    it("sets error on failure", async () => {
      mockedCreateNamespace.mockRejectedValue(new Error("duplicate"));

      await expect(
        useNamespacesStore.getState().createNamespace("dup"),
      ).rejects.toThrow("duplicate");

      const state = useNamespacesStore.getState();
      expect(state.error).toBe("duplicate");
      expect(state.loading).toBe(false);
    });
  });

  describe("deleteNamespace", () => {
    it("deletes namespace, logs out, and redirects", async () => {
      mockedDeleteNamespace.mockResolvedValue(undefined as never);

      await useNamespacesStore.getState().deleteNamespace("t1");

      expect(mockedDeleteNamespace).toHaveBeenCalledWith("t1");
      expect(useAuthStore.getState().token).toBeNull();
      expect(window.location.replace).toHaveBeenCalledWith("/v2/ui/login");
    });
  });

  describe("leaveNamespace", () => {
    it("leaves namespace, logs out, and redirects", async () => {
      mockedLeaveNamespace.mockResolvedValue(undefined as never);

      await useNamespacesStore.getState().leaveNamespace("t1");

      expect(mockedLeaveNamespace).toHaveBeenCalledWith("t1");
      expect(useAuthStore.getState().token).toBeNull();
      expect(window.location.replace).toHaveBeenCalledWith("/v2/ui/login");
    });
  });

  describe("fetchCurrent", () => {
    it("sets current namespace", async () => {
      const ns = { tenant_id: "t1", name: "ns1" };
      mockedGetNamespace.mockResolvedValue(ns as never);

      await useNamespacesStore.getState().fetchCurrent("t1");

      expect(useNamespacesStore.getState().currentNamespace).toEqual(ns);
    });

    it("clears current namespace on error", async () => {
      useNamespacesStore.setState({
        currentNamespace: { tenant_id: "old" } as never,
      });
      mockedGetNamespace.mockRejectedValue(new Error("not found"));

      await useNamespacesStore.getState().fetchCurrent("bad");

      expect(useNamespacesStore.getState().currentNamespace).toBeNull();
    });
  });
});
