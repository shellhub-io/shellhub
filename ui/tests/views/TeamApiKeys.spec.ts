import { VueWrapper, flushPromises } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import TeamApiKeys from "@/views/TeamApiKeys.vue";
import useApiKeysStore from "@/store/modules/api_keys";
import { createAxiosError } from "@tests/utils/axiosError";

vi.mock("@/store/api/api_keys");

describe("Team Api Keys View", () => {
  let wrapper: VueWrapper<InstanceType<typeof TeamApiKeys>>;
  let apiKeysStore: ReturnType<typeof useApiKeysStore>;

  const mockApiKeys = [
    {
      name: "fake-api-key",
      tenant_id: "00000000-0000-4000-0000-000000000000",
      role: "administrator" as const,
      created_by: "xxxxxxxx",
      created_at: "",
      updated_at: "",
      expires_in: 1753815353,
    },
  ];

  const mountWrapper = async (hasKeys = true, mockError?: Error) => {
    const initialState = {
      apiKeys: {
        apiKeys: hasKeys ? mockApiKeys : [],
        apiKeysCount: hasKeys ? 1 : 0,
      },
    };

    apiKeysStore = useApiKeysStore();
    if (mockError) {
      vi.mocked(apiKeysStore.fetchApiKeys).mockRejectedValueOnce(mockError);
    }

    wrapper = mountComponent(TeamApiKeys, { piniaOptions: { initialState, stubActions: !mockError } });

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when API keys exist", () => {
    beforeEach(async () => { await mountWrapper(); });

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="api-key-title"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("API Keys");
      expect(pageHeader.text()).toContain("Team Management");
    });

    it("displays the generate API key button in header", () => {
      const generateBtn = wrapper.findComponent({ name: "ApiKeyGenerate" });
      expect(generateBtn.exists()).toBe(true);
    });

    it("displays the API key list", () => {
      expect(wrapper.find('[data-test="api-key-list"]').exists()).toBe(true);
    });

    it("does not show the no items message", () => {
      expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
    });

    it("does not show loading spinner", () => {
      const loadingSpinner = wrapper.findComponent({ name: "VProgressCircular" });
      expect(loadingSpinner.exists()).toBe(false);
    });
  });

  describe("when no API keys exist", () => {
    beforeEach(async () => { await mountWrapper(false); });

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="api-key-title"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("API Keys");
    });

    it("does not display the generate API key button in header", () => {
      const pageHeader = wrapper.find('[data-test="api-key-title"]');
      const generateBtn = pageHeader.findComponent({ name: "ApiKeyGenerate" });
      expect(generateBtn.exists()).toBe(false);
    });

    it("does not display the API key list", () => {
      expect(wrapper.find('[data-test="api-key-list"]').exists()).toBe(false);
    });

    it("shows the no items message", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      expect(noItemsMessage.exists()).toBe(true);
      expect(noItemsMessage.text()).toContain("API Keys");
      expect(noItemsMessage.text()).toContain("authenticate and integrate external applications");
    });

    it("displays generate API key button in no items message", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      const generateBtn = noItemsMessage.findComponent({ name: "ApiKeyGenerate" });
      expect(generateBtn.exists()).toBe(true);
    });

    it("does not show loading spinner", () => {
      const loadingSpinner = wrapper.findComponent({ name: "VProgressCircular" });
      expect(loadingSpinner.exists()).toBe(false);
    });
  });

  describe("when loading API keys fails", () => {
    beforeEach(() => mountWrapper(false, createAxiosError(500, "Internal Server Error")));

    it("displays error snackbar notification", () => {
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load API keys.");
    });

    it("still renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="api-key-title"]');
      expect(pageHeader.exists()).toBe(true);
    });

    it("shows the no items message when error occurs", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      expect(noItemsMessage.exists()).toBe(true);
    });

    it("hides loading spinner after error", () => {
      const loadingSpinner = wrapper.findComponent({ name: "VProgressCircular" });
      expect(loadingSpinner.exists()).toBe(false);
    });
  });

  describe("loading state", () => {
    it("shows loading spinner initially", async () => {
      vi.mocked(apiKeysStore.fetchApiKeys).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(), 100)),
      );

      wrapper = mountComponent(TeamApiKeys, {
        piniaOptions: {
          initialState: { apiKeys: { apiKeys: [], apiKeysCount: 0 } },
          stubActions: false,
        },
      });

      // Before promises resolve, loading should be true
      const loadingSpinner = wrapper.findComponent({ name: "VProgressCircular" });
      expect(loadingSpinner.exists()).toBe(true);

      await flushPromises();
    });
  });
});
