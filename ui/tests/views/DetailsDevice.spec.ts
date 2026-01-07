import { VueWrapper, flushPromises } from "@vue/test-utils";
import { Router } from "vue-router";
import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import createCleanRouter from "@tests/utils/router";
import DetailsDevice from "@/views/DetailsDevice.vue";
import { IDevice } from "@/interfaces/IDevice";
import { envVariables } from "@/envVariables";
import { formatFullDateTime } from "@/utils/date";
import { mockDevice } from "@tests/views/mocks";
import { createAxiosError } from "@tests/utils/axiosError";
import useDevicesStore from "@/store/modules/devices";

vi.mock("@/store/api/devices");

vi.mock("@/envVariables", () => ({
  envVariables: {
    hasWebEndpoints: false,
    isEnterprise: true,
  },
}));

describe("Details Device View", () => {
  let wrapper: VueWrapper<InstanceType<typeof DetailsDevice>>;
  let router: Router;

  const device: IDevice = mockDevice;

  const mountWrapper = async ({
    deviceId = "123456",
    initialDevice = device,
    mockError,
  }: {
    deviceId?: string;
    initialDevice?: Partial<IDevice>;
    mockError?: Error;
  } = {}) => {
    localStorage.setItem("tenant", "fake-tenant-data");

    router = createCleanRouter();
    await router.push({ name: "DeviceDetails", params: { identifier: deviceId } });
    await router.isReady();

    wrapper = mountComponent(DetailsDevice, {
      global: { plugins: [router] },
      piniaOptions: {
        ...(mockError ? {} : { initialState: { devices: { device: initialDevice } } }),
        stubActions: !mockError,
      },
    });

    const devicesStore = useDevicesStore();
    if (mockError) vi.mocked(devicesStore.fetchDevice).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when device loads successfully", () => {
    beforeEach(() => mountWrapper());

    it("renders all device detail fields with correct values", () => {
      const uidField = wrapper.find('[data-test="device-uid-field"]');
      const macField = wrapper.find('[data-test="device-mac-field"]');
      const prettyNameField = wrapper.find('[data-test="device-pretty-name-field"]');
      const versionField = wrapper.find('[data-test="device-version-field"]');
      const tagsField = wrapper.find('[data-test="device-tags-field"]');
      const lastSeenField = wrapper.find('[data-test="device-last-seen-field"]');

      expect(uidField.text()).toContain(device.uid);
      expect(macField.text()).toContain(device.identity.mac);
      expect(prettyNameField.text()).toContain(device.info.pretty_name);
      expect(versionField.text()).toContain(device.info.version);
      expect(tagsField.text()).toContain(device.tags[0].name);
      expect(lastSeenField.text()).toContain(formatFullDateTime(device.last_seen));
    });

    it("displays the device name in the header", () => {
      expect(wrapper.text()).toContain(device.name);
    });

    it("displays the connect button for accepted devices", () => {
      expect(wrapper.find('[data-test="connect-btn"]').exists()).toBe(true);
    });

    it("displays device action menu items", () => {
      const actionsList = wrapper.findComponent({ name: "VList" });
      expect(actionsList.find('[data-test="device-rename-component"]').exists()).toBe(true);
      expect(actionsList.find('[data-test="open-tags-btn"]').exists()).toBe(true);
      expect(actionsList.find('[data-test="device-delete-item"]').exists()).toBe(true);
    });
  });

  describe("when device has different statuses", () => {
    it.each([
      ["pending", false],
      ["rejected", false],
      ["accepted", true],
    ] as const)(
      "displays connect button: %s -> %s",
      async (status, shouldShowButton) => {
        await mountWrapper({ initialDevice: { ...device, status } });
        expect(wrapper.find('[data-test="connect-btn"]').exists()).toBe(shouldShowButton);
      },
    );
  });

  describe("when device has no tags", () => {
    beforeEach(() => mountWrapper({ initialDevice: { ...device, tags: [] } }));

    it("still renders the device details", () => {
      expect(wrapper.find('[data-test="device-uid-field"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="device-tags-field"]').exists()).toBe(false);
    });
  });

  describe("when device fails to load", () => {
    beforeEach(() => mountWrapper({ deviceId: "inexistent-device", mockError: createAxiosError(404, "Device not found") }));

    it("shows error message when device does not load", () => {
      expect(wrapper.text()).toContain("Something is wrong, try again !");
    });

    it("does not render device detail fields", () => {
      expect(wrapper.find('[data-test="device-uid-field"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="connect-btn"]').exists()).toBe(false);
    });

    it("displays error snackbar notification", () => {
      expect(mockSnackbar.showError).toHaveBeenCalledWith("There was an error loading the device details.");
    });
  });

  describe("when web endpoints feature is enabled", () => {
    const findActionsList = () => wrapper.findComponent({ name: "VList" });

    it("displays create web endpoint button when feature is enabled", async () => {
      vi.mocked(envVariables).hasWebEndpoints = true;
      await mountWrapper();
      expect(findActionsList().find('[data-test="create-web-endpoint-btn"]').exists()).toBe(true);
    });

    it("does not display create web endpoint button when feature is disabled", async () => {
      vi.mocked(envVariables).hasWebEndpoints = false;
      await mountWrapper();
      expect(findActionsList().find('[data-test="create-web-endpoint-btn"]').exists()).toBe(false);
    });
  });
});
