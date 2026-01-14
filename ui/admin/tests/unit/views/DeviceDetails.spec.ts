import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import { formatFullDateTime } from "@/utils/date";
import useDevicesStore from "@admin/store/modules/devices";
import DeviceDetails from "@admin/views/DeviceDetails.vue";
import { mockDevice } from "../mocks";

vi.mock("@admin/store/api/devices");

describe("DeviceDetails", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceDetails>>;

  const mountWrapper = async (mockError?: Error) => {
    const router = createCleanAdminRouter();
    await router.push({ name: "deviceDetails", params: { id: mockDevice.uid } });
    await router.isReady();

    wrapper = mountComponent(DeviceDetails, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: { adminDevices: mockError ? {} : { device: mockDevice } },
        stubActions: !mockError,
      },
    });

    const devicesStore = useDevicesStore();
    if (mockError) vi.mocked(devicesStore.fetchDeviceById).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when device loads successfully", () => {
    beforeEach(() => mountWrapper());

    it("displays the device name in the card title", () => {
      expect(wrapper.find(".text-h6").text()).toBe(mockDevice.name);
    });

    it("shows online status icon", () => {
      const icon = wrapper.find('[data-test="online-icon"]');
      expect(icon.exists()).toBe(true);
    });

    it("shows status chip with correct value", () => {
      const statusChip = wrapper.find('[data-test="device-status-chip"]');
      expect(statusChip.exists()).toBe(true);
      expect(statusChip.text()).toBe(mockDevice.status);
    });

    it("displays device uid", () => {
      const uidField = wrapper.find('[data-test="device-uid-field"]');
      expect(uidField.text()).toContain("UID:");
      expect(uidField.text()).toContain(mockDevice.uid);
    });

    it("displays mac address", () => {
      const macField = wrapper.find('[data-test="device-mac-field"]');
      expect(macField.text()).toContain("MAC:");
      expect(macField.text()).toContain(mockDevice.identity.mac);
    });

    it("displays operating system", () => {
      const osField = wrapper.find('[data-test="device-pretty-name-field"]');
      expect(osField.text()).toContain("Operating System:");
      expect(osField.text()).toContain(mockDevice.info.pretty_name);
    });

    it("displays agent version", () => {
      const versionField = wrapper.find('[data-test="device-version-field"]');
      expect(versionField.text()).toContain("Agent Version:");
      expect(versionField.text()).toContain(mockDevice.info.version);
    });

    it("displays architecture", () => {
      const archField = wrapper.find('[data-test="device-architecture-field"]');
      expect(archField.text()).toContain("Architecture:");
      expect(archField.text()).toContain(mockDevice.info.arch);
    });

    it("displays platform", () => {
      const platformField = wrapper.find('[data-test="device-platform-field"]');
      expect(platformField.text()).toContain("Platform:");
      expect(platformField.text()).toContain(mockDevice.info.platform);
    });

    it("displays namespace with link", () => {
      const namespaceField = wrapper.find('[data-test="device-namespace-field"]');
      expect(namespaceField.text()).toContain("Namespace:");
      const link = namespaceField.find("a");
      expect(link.exists()).toBe(true);
      expect(link.text()).toBe(mockDevice.namespace);
    });

    it("displays tenant id", () => {
      const tenantField = wrapper.find('[data-test="device-tenant-id-field"]');
      expect(tenantField.text()).toContain("Tenant ID:");
      expect(tenantField.text()).toContain(mockDevice.tenant_id);
    });

    it("displays remote address", () => {
      const remoteAddrField = wrapper.find('[data-test="device-remote-addr-field"]');
      expect(remoteAddrField.text()).toContain("Remote Address:");
      expect(remoteAddrField.text()).toContain(mockDevice.remote_addr);
    });

    it("displays created at date", () => {
      const createdAtField = wrapper.find('[data-test="device-created-at-field"]');
      expect(createdAtField.text()).toContain("Created At:");
      expect(createdAtField.text()).toContain(formatFullDateTime(mockDevice.created_at));
    });

    it("displays last seen date", () => {
      const lastSeenField = wrapper.find('[data-test="device-last-seen-field"]');
      expect(lastSeenField.text()).toContain("Last Seen:");
      expect(lastSeenField.text()).toContain(formatFullDateTime(mockDevice.last_seen));
    });

    it("displays public key", () => {
      const publicKeyField = wrapper.find('[data-test="device-public-key-field"]');
      expect(publicKeyField.text()).toContain("Public Key:");
      expect(publicKeyField.text()).toContain(mockDevice.public_key);
    });
  });

  describe("when device fails to load", () => {
    it("shows error snackbar", () => mountWrapper(createAxiosError(404, "Not Found")).then(() => {
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to get device details.");
    }));
  });
});
