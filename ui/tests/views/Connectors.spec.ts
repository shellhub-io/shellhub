import { VueWrapper, flushPromises } from "@vue/test-utils";
import { Router } from "vue-router";
import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import createCleanRouter from "@tests/utils/router";
import Connectors from "@/views/Connectors.vue";
import useConnectorStore from "@/store/modules/connectors";
import { createAxiosError } from "@tests/utils/axiosError";

vi.mock("@/store/api/connectors");

describe("Connectors View", () => {
  let wrapper: VueWrapper<InstanceType<typeof Connectors>>;
  let router: Router;

  const mountWrapper = async (mockError?: Error) => {
    router = createCleanRouter();
    await router.push({ name: "Connectors" });
    await router.isReady();

    wrapper = mountComponent(Connectors, {
      global: { plugins: [router] },
      piniaOptions: { stubActions: !mockError },
    });

    const connectorStore = useConnectorStore();
    if (mockError) vi.mocked(connectorStore.fetchConnectorList).mockRejectedValueOnce(mockError);
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when page loads", () => {
    beforeEach(() => mountWrapper());

    it("renders the page header", () => {
      expect(wrapper.find('[data-test="device-title"]').text()).toContain("Docker Connectors");
      expect(wrapper.text()).toContain("Docker Integration");
    });

    it("displays the containers navigation button", () => {
      const containersBtn = wrapper.find('[data-test="connector-add-btn"]');
      expect(containersBtn.exists()).toBe(true);
      expect(containersBtn.text()).toContain("Containers");
    });

    it("displays the connector table", () => {
      expect(wrapper.find('[data-test="connector-table-component"]').exists()).toBe(true);
    });
  });

  describe("navigation", () => {
    beforeEach(() => mountWrapper());

    it("navigates to containers when button is clicked", async () => {
      const pushSpy = vi.spyOn(router, "push");
      const containersBtn = wrapper.find('[data-test="connector-add-btn"]');

      await containersBtn.trigger("click");

      expect(pushSpy).toHaveBeenCalledWith("/containers");
    });
  });

  describe("error handling", () => {
    it("displays error message when 403 error occurs", async () => {
      await mountWrapper(createAxiosError(403, "Forbidden"));

      const connectorAdd = wrapper.findComponent({ name: "ConnectorAdd" });
      await connectorAdd.vm.$emit("update");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("You do not have permission to access the connectors.");
    });

    it("displays error message when generic error occurs", async () => {
      await mountWrapper(createAxiosError(500, "Server Error"));

      const connectorAdd = wrapper.findComponent({ name: "ConnectorAdd" });
      await connectorAdd.vm.$emit("update");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Error loading the connectors.");
    });
  });
});
