import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { mockApiKeys } from "@tests/mocks/apiKey";
import * as hasPermissionModule from "@/utils/permission";
import ApiKeyList from "@/components/Team/ApiKeys/ApiKeyList.vue";
import useApiKeysStore from "@/store/modules/api_keys";
import moment from "moment";
import handleError from "@/utils/handleError";

vi.mock("@/utils/permission");

describe("ApiKeyList", () => {
  let wrapper: VueWrapper<InstanceType<typeof ApiKeyList>>;
  let apiKeysStore: ReturnType<typeof useApiKeysStore>;

  const mountWrapper = ({
    apiKeys = mockApiKeys,
    canDeleteApiKey = true,
  } = {}) => {
    vi.mocked(hasPermissionModule.default).mockReturnValue(canDeleteApiKey);

    wrapper = mountComponent(ApiKeyList, {
      piniaOptions: {
        initialState: {
          apiKeys: { apiKeys, apiKeysCount: apiKeys.length },
        },
      },
    });

    apiKeysStore = useApiKeysStore();
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe("Component rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders data table", () => {
      const table = wrapper.find('[data-test="api-key-list"]');
      expect(table.exists()).toBe(true);
    });

    it("displays API key names", () => {
      const firstKeyName = wrapper.find('[data-test="key-name"]');
      expect(firstKeyName.exists()).toBe(true);
    });

    it("displays API key roles", () => {
      const keyRole = wrapper.find('[data-test="key-name"]');
      expect(keyRole.exists()).toBe(true);
    });

    it("displays expiration dates", () => {
      const expiryDate = wrapper.find('[data-test="key-expiry-date"]');
      expect(expiryDate.exists()).toBe(true);
    });

    it("displays action menu for each key", () => {
      const menu = wrapper.find('[data-test="menu-key-component"]');
      expect(menu.exists()).toBe(true);
    });
  });

  describe("Data fetching", () => {
    it("calls fetchApiKeys with correct params", async () => {
      mountWrapper();
      await wrapper.findComponent({ name: "DataTable" }).vm.$emit("update:page", 2);
      await flushPromises();

      expect(apiKeysStore.fetchApiKeys).toHaveBeenCalledWith({
        page: 2,
        perPage: 10,
        sortField: "name",
        sortOrder: "asc",
      });
    });

    it("calls fetchApiKeys when items per page changes", async () => {
      mountWrapper();
      vi.clearAllMocks();

      await wrapper.findComponent({ name: "DataTable" }).vm.$emit("update:itemsPerPage", 20);
      await flushPromises();

      expect(apiKeysStore.fetchApiKeys).toHaveBeenCalledWith({
        page: 1,
        perPage: 20,
        sortField: "name",
        sortOrder: "asc",
      });
    });
  });

  describe("Sorting", () => {
    beforeEach(() => mountWrapper());

    it("calls fetchApiKeys when sorting by field", async () => {
      vi.clearAllMocks();

      await wrapper.findComponent({ name: "DataTable" }).vm.$emit("update:sort", "expires_in");
      await flushPromises();

      expect(apiKeysStore.fetchApiKeys).toHaveBeenCalledWith({
        page: 1,
        perPage: 10,
        sortField: "expires_in",
        sortOrder: "desc",
      });
    });

    it("toggles sort order on second click", async () => {
      vi.clearAllMocks();

      await wrapper.findComponent({ name: "DataTable" }).vm.$emit("update:sort", "name");
      await flushPromises();

      expect(apiKeysStore.fetchApiKeys).toHaveBeenCalledWith({
        page: 1,
        perPage: 10,
        sortField: "name",
        sortOrder: "desc",
      });

      await wrapper.findComponent({ name: "DataTable" }).vm.$emit("update:sort", "name");
      await flushPromises();

      expect(apiKeysStore.fetchApiKeys).toHaveBeenLastCalledWith({
        page: 1,
        perPage: 10,
        sortField: "name",
        sortOrder: "asc",
      });
    });
  });

  describe("Expiration formatting", () => {
    it("displays 'Never' for never-expiring keys", () => {
      const neverExpiringKey = {
        ...mockApiKeys[0],
        expires_in: -1,
      };
      mountWrapper({ apiKeys: [neverExpiringKey] });

      const expiryDate = wrapper.find('[data-test="key-expiry-date"]');
      expect(expiryDate.text()).toBe("Never");
    });

    it("displays future expiration date correctly", () => {
      const futureUnixTime = moment().add(30, "days").unix();
      const futureKey = {
        ...mockApiKeys[0],
        expires_in: futureUnixTime,
      };
      mountWrapper({ apiKeys: [futureKey] });

      const expiryDate = wrapper.find('[data-test="key-expiry-date"]');
      const expected = `Expires on ${moment.unix(futureUnixTime).format("MMM D YYYY")}.`;
      expect(expiryDate.text()).toBe(expected);
    });

    it("displays past expiration date correctly", () => {
      const pastUnixTime = moment().subtract(10, "days").unix();
      const expiredKey = {
        ...mockApiKeys[0],
        expires_in: pastUnixTime,
      };
      mountWrapper({ apiKeys: [expiredKey] });

      const expiryDate = wrapper.find('[data-test="key-expiry-date"]');
      const expected = `Expired on ${moment.unix(pastUnixTime).format("MMM D YYYY")}.`;
      expect(expiryDate.text()).toBe(expected);
    });
  });

  describe("Key expiration status", () => {
    it("does not apply warning class for never-expiring keys", () => {
      const neverExpiringKey = {
        ...mockApiKeys[0],
        expires_in: -1,
      };
      mountWrapper({ apiKeys: [neverExpiringKey] });

      const expiryDate = wrapper.find('[data-test="key-expiry-date"]');
      expect(expiryDate.classes()).not.toContain("text-warning");
    });

    it("does not apply warning class for future expiration", () => {
      const futureUnixTime = moment().add(30, "days").unix();
      const futureKey = {
        ...mockApiKeys[0],
        expires_in: futureUnixTime,
      };
      mountWrapper({ apiKeys: [futureKey] });

      const expiryDate = wrapper.find('[data-test="key-expiry-date"]');
      expect(expiryDate.classes()).not.toContain("text-warning");
    });

    it("applies warning class for expired keys", () => {
      const pastUnixTime = moment().subtract(10, "days").unix();
      const expiredKey = {
        ...mockApiKeys[0],
        expires_in: pastUnixTime,
      };
      mountWrapper({ apiKeys: [expiredKey] });

      const expiryDate = wrapper.find('[data-test="key-expiry-date"]');
      expect(expiryDate.classes()).toContain("text-warning");
    });

    it("shows alert icon for expired keys", () => {
      const pastUnixTime = moment().subtract(10, "days").unix();
      const expiredKey = {
        ...mockApiKeys[0],
        expires_in: pastUnixTime,
      };
      mountWrapper({ apiKeys: [expiredKey] });

      const icon = wrapper.find("tbody tr td i.mdi-clock-alert-outline");
      expect(icon.exists()).toBe(true);
    });

    it("shows key icon for non-expired keys", () => {
      const futureUnixTime = moment().add(30, "days").unix();
      const futureKey = {
        ...mockApiKeys[0],
        expires_in: futureUnixTime,
      };
      mountWrapper({ apiKeys: [futureKey] });

      const icon = wrapper.find("tbody tr td i.mdi-key-outline");
      expect(icon.exists()).toBe(true);
    });
  });

  describe("Error handling", () => {
    it("shows error snackbar on 403 status", async () => {
      const error = createAxiosError(403, "Forbidden");

      mountWrapper();
      vi.mocked(apiKeysStore.fetchApiKeys).mockRejectedValueOnce(error);
      await flushPromises();

      await wrapper.findComponent({ name: "DataTable" }).vm.$emit("update:page", 2);
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("You are not authorized to view this API key.");
      expect(handleError).not.toHaveBeenCalled();
    });

    it("shows generic error snackbar for other errors", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      mountWrapper();
      vi.mocked(apiKeysStore.fetchApiKeys).mockRejectedValueOnce(error);
      await flushPromises();

      await wrapper.findComponent({ name: "DataTable" }).vm.$emit("update:page", 2);
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load API keys.");
      expect(handleError).toHaveBeenCalled();
    });
  });

  describe("Refresh functionality", () => {
    beforeEach(() => mountWrapper());

    it("calls fetchApiKeys when refresh is called", async () => {
      await wrapper.vm.refresh();
      await flushPromises();

      expect(apiKeysStore.fetchApiKeys).toHaveBeenCalled();
    });
  });

  describe("Permission-based actions", () => {
    it("enables edit and delete when user has permission", () => {
      mountWrapper({ canDeleteApiKey: true });

      expect(wrapper.findAllComponents({ name: "ApiKeyEdit" }).length).toBeGreaterThan(0);
      expect(wrapper.findAllComponents({ name: "ApiKeyDelete" }).length).toBeGreaterThan(0);
    });
  });
});
