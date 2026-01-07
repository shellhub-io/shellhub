import { VueWrapper, flushPromises } from "@vue/test-utils";
import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import WebEndpoints from "@/views/WebEndpoints.vue";
import useWebEndpointsStore from "@/store/modules/web_endpoints";
import { createAxiosError } from "@tests/utils/axiosError";
import createCleanRouter from "@tests/utils/router";

vi.mock("@/store/api/web_endpoints");
vi.mock("@/store/api/devices", () => ({
  fetchDevices: () => Promise.resolve({ data: [], headers: { "x-total-count": "0" } }),
}));
vi.mock("@/utils/permission", () => ({
  default: vi.fn(() => true),
}));

describe("WebEndpoints View", () => {
  let wrapper: VueWrapper<InstanceType<typeof WebEndpoints>>;
  let webEndpointsStore: ReturnType<typeof useWebEndpointsStore>;
  const router = createCleanRouter();

  const mountWrapper = async (mockError?: Error) => {
    wrapper = mountComponent(WebEndpoints, { global: { plugins: [router] }, piniaOptions: { stubActions: !mockError } });

    webEndpointsStore = useWebEndpointsStore();

    if (mockError) vi.spyOn(webEndpointsStore, "fetchWebEndpointsList").mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when page loads", () => {
    beforeEach(() => mountWrapper());

    it("renders the page header", () => {
      expect(wrapper.text()).toContain("Web Endpoints");
      expect(wrapper.text()).toContain("Web Access");
    });

    it("displays the create web endpoint button", () => {
      const createBtn = wrapper.find('[data-test="tunnel-create-dialog-btn"]');
      expect(createBtn.exists()).toBe(true);
      expect(createBtn.text()).toContain("Create Web Endpoint");
    });

    it("calls fetchWebEndpointsList on mount", () => {
      expect(webEndpointsStore.fetchWebEndpointsList).toHaveBeenCalledWith();
    });
  });

  describe("when web endpoints exist", () => {
    beforeEach(async () => {
      await mountWrapper();
      webEndpointsStore.showWebEndpoints = true;
      await flushPromises();
    });

    it("displays the search field", () => {
      const searchField = wrapper.find('[data-test="search-text"]');
      expect(searchField.exists()).toBe(true);
    });

    it("displays the web endpoints table", () => {
      expect(wrapper.find('[data-test="web-endpoints-table-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
    });

    it("filters web endpoints on search", async () => {
      const searchField = wrapper.find('[data-test="search-text"] input');
      await searchField.setValue("localhost");
      await searchField.trigger("keyup");
      await flushPromises();

      expect(webEndpointsStore.fetchWebEndpointsList).toHaveBeenCalledWith(
        expect.objectContaining({
          filter: expect.any(String),
        }),
      );
    });
  });

  describe("when no web endpoints exist", () => {
    beforeEach(async () => {
      await mountWrapper();
      webEndpointsStore.showWebEndpoints = false;
      await flushPromises();
    });

    it("displays the no items message", () => {
      expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="web-endpoints-table-component"]').exists()).toBe(false);
    });

    it("does not display the search field", () => {
      expect(wrapper.find('[data-test="search-text"]').exists()).toBe(false);
    });

    it("displays the create button in no items message", () => {
      const noItemsSection = wrapper.find('[data-test="no-items-message-component"]');
      const createBtn = noItemsSection.find('[data-test="tunnel-create-dialog-btn"]');
      expect(createBtn.exists()).toBe(true);
    });
  });

  describe("when fetching web endpoints fails", () => {
    beforeEach(() => mountWrapper(createAxiosError(500, "Internal Server Error")));

    it("displays error message", () => {
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load web endpoints.");
    });
  });

  describe("web endpoint creation", () => {
    beforeEach(() => mountWrapper());

    it("opens create dialog when button is clicked", async () => {
      const createBtn = wrapper.find('[data-test="tunnel-create-dialog-btn"]');
      await createBtn.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "WebEndpointCreate" });
      expect(dialog.props("modelValue")).toBe(true);
    });

    it("calls searchWebEndpoints when create dialog emits update", async () => {
      const createBtn = wrapper.find('[data-test="tunnel-create-dialog-btn"]');

      await createBtn.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "WebEndpointCreate" });
      dialog.vm.$emit("update");
      await flushPromises();

      expect(webEndpointsStore.fetchWebEndpointsList).toHaveBeenCalled();
    });
  });
});
