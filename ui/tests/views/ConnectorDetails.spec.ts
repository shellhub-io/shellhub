import { VueWrapper, flushPromises } from "@vue/test-utils";
import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import createCleanRouter from "@tests/utils/router";
import ConnectorDetails from "@/views/ConnectorDetails.vue";
import useConnectorStore from "@/store/modules/connectors";
import { createAxiosError } from "@tests/utils/axiosError";

vi.mock("@/store/api/connectors");
vi.mock("@/utils/permission", () => ({
  default: vi.fn(() => true),
}));

describe("ConnectorDetails View", () => {
  let wrapper: VueWrapper<InstanceType<typeof ConnectorDetails>>;
  let connectorStore: ReturnType<typeof useConnectorStore>;

  const connectorId = "connector-123";

  const createMockConnector = (overrides = {}) => ({
    uid: connectorId,
    tenant_id: "tenant-123",
    address: "localhost",
    port: 2375,
    status: {
      state: "running",
      message: "",
    },
    secure: true,
    enable: true,
    ...overrides,
  });

  const mockConnector = createMockConnector();

  const mountWrapper = async (mockError?: Error) => {
    const router = createCleanRouter();
    await router.push({ name: "ConnectorDetails", params: { id: connectorId } });
    await router.isReady();

    wrapper = mountComponent(ConnectorDetails, {
      global: { plugins: [router] },
      piniaOptions: { stubActions: !mockError },
    });

    connectorStore = useConnectorStore();
    connectorStore.connector = mockConnector;

    if (mockError) {
      vi.mocked(connectorStore.fetchConnectorById).mockRejectedValueOnce(mockError);
      vi.mocked(connectorStore.getConnectorInfo).mockRejectedValueOnce(mockError);
    }

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when page loads", () => {
    beforeEach(() => mountWrapper());

    it("renders the page title", () => {
      expect(wrapper.text()).toContain("Connector Details");
    });

    it("calls fetchConnectorById and getConnectorInfo on mount", () => {
      expect(connectorStore.fetchConnectorById).toHaveBeenCalledWith(connectorId);
      expect(connectorStore.getConnectorInfo).toHaveBeenCalledWith(connectorId);
    });
  });

  describe("when connector data loads successfully", () => {
    beforeEach(() => mountWrapper());

    it("displays the connector address", () => {
      expect(wrapper.find('[data-test="sshid-chip text-overline"]').text()).toContain("localhost:2375");
    });

    it("displays the connector as secure", () => {
      const chip = wrapper.find('[data-test="sshid-chip text-overline"]');
      expect(chip.classes()).toContain("text-success");
    });

    it("displays the enable/disable switch", () => {
      const switchElement = wrapper.find('[data-test="connector-enable-switch"]');
      expect(switchElement.exists()).toBe(true);
    });
  });

  describe("when toggling connector state", () => {
    beforeEach(() => mountWrapper());

    it("updates the connector when switch is toggled", async () => {
      const switchElement = wrapper.find('[data-test="connector-enable-switch"]');
      await switchElement.trigger("click");
      await flushPromises();

      expect(connectorStore.updateConnector).toHaveBeenCalledWith(
        expect.objectContaining({
          uid: connectorId,
          enable: false,
        }),
      );
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("The connector has been updated.");
    });
  });

  describe("when fetching connector fails", () => {
    beforeEach(() => mountWrapper(createAxiosError(404, "Not Found")));

    it("displays error message", () => {
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Error loading the connector.");
    });

    it("shows error state when connector uid is missing", async () => {
      connectorStore.connector = createMockConnector({ uid: "" });
      await flushPromises();

      expect(wrapper.text()).toContain("Something went wrong, try again!");
    });
  });
});
