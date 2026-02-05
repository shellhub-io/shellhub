import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import { mockDeviceOnlineWithTags, mockDeviceOfflineNoTags, mockDevicePending } from "@tests/mocks/device";
import DeviceTable from "@/components/Tables/DeviceTable.vue";
import { IDevice, IDeviceMethods } from "@/interfaces/IDevice";
import * as hasPermissionModule from "@/utils/permission";
import { createCleanRouter } from "@tests/utils/router";
import handleError from "@/utils/handleError";

describe("DeviceTable", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceTable>>;
  let storeMethods: IDeviceMethods;

  const mockDevices: IDevice[] = [mockDeviceOnlineWithTags, mockDeviceOfflineNoTags];
  const mockPendingDevices: IDevice[] = [mockDevicePending];

  const createMockStoreMethods = (devices: IDevice[] = mockDevices): IDeviceMethods => ({
    fetchDevices: vi.fn().mockResolvedValue(undefined),
    getList: vi.fn(() => devices),
    getCount: vi.fn(() => devices.length),
    getFilter: vi.fn(() => undefined),
  });

  const mountWrapper = ({
    mockStoreMethods = undefined as IDeviceMethods | undefined,
    status = "accepted" as "accepted" | "pending",
    header = "primary" as "primary" | "secondary",
    variant = "device" as "device" | "container",
    hasTagUpdatePermission = true,
    hasDeviceRemovePermission = true,
  } = {}) => {
    vi.spyOn(hasPermissionModule, "default").mockImplementation((permission: string) => {
      if (permission === "tag:update") return hasTagUpdatePermission;
      if (permission === "device:remove") return hasDeviceRemovePermission;
      return true;
    });

    storeMethods = mockStoreMethods || createMockStoreMethods();

    wrapper = mountComponent(DeviceTable, {
      props: {
        storeMethods,
        status,
        header,
        variant,
      },
      global: { plugins: [createCleanRouter()] },
    });
  };

  beforeEach(() => {
    mountWrapper();
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Component rendering", () => {
    it("Renders DataTable component", () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      expect(dataTable.exists()).toBe(true);
    });

    it("Passes correct props to DataTable", () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      expect(dataTable.props("items")).toEqual(mockDevices);
      expect(dataTable.props("totalCount")).toBe(2);
      expect(dataTable.props("itemsPerPageOptions")).toEqual([10, 20, 50, 100]);
      expect(dataTable.props("tableName")).toBe("devices");
    });

    it("Renders with container variant", () => {
      wrapper.unmount();
      mountWrapper({ variant: "container" });

      const dataTable = wrapper.findComponent({ name: "DataTable" });
      expect(dataTable.exists()).toBe(true);
    });
  });

  describe("Headers configuration", () => {
    it("Shows primary headers for accepted devices", () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      const headers = dataTable.props("headers");

      expect(headers).toHaveLength(6);
      expect(headers[0].text).toBe("Online");
      expect(headers[1].text).toBe("Hostname");
      expect(headers[2].text).toBe("Operating System");
      expect(headers[3].text).toBe("SSHID");
      expect(headers[4].text).toBe("Tags");
      expect(headers[5].text).toBe("Actions");
    });

    it("Shows secondary headers for pending devices", () => {
      wrapper.unmount();
      mountWrapper({ header: "secondary" });

      const dataTable = wrapper.findComponent({ name: "DataTable" });
      const headers = dataTable.props("headers");

      expect(headers).toHaveLength(4);
      expect(headers[0].text).toBe("Hostname");
      expect(headers[1].text).toBe("Operating System");
      expect(headers[2].text).toBe("Request Time");
      expect(headers[3].text).toBe("Actions");
    });

    it("Shows Image header for container variant", () => {
      wrapper.unmount();
      mountWrapper({ variant: "container" });

      const dataTable = wrapper.findComponent({ name: "DataTable" });
      const headers = dataTable.props("headers");

      const osHeader = headers.find((h: { value: string }) => h.value === "operating_system");
      expect(osHeader?.text).toBe("Image");
    });
  });

  describe("Device fetching", () => {
    it("Fetches devices on mount", async () => {
      await flushPromises();

      expect(storeMethods.fetchDevices).toHaveBeenCalledWith({
        perPage: 10,
        page: 1,
        status: "accepted",
        sortField: undefined,
        sortOrder: undefined,
      });
    });

    it("Shows error when fetch fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      wrapper.unmount();
      mountWrapper();
      vi.mocked(storeMethods.fetchDevices).mockRejectedValueOnce(error);
      await flushPromises();
      await wrapper.findComponent({ name: "DeviceDelete" }).vm.$emit("update");
      expect(handleError).toHaveBeenCalled();
    });
  });

  describe("Accepted devices row rendering", () => {
    it("Renders TerminalConnectButton for each device", () => {
      const buttons = wrapper.findAllComponents({ name: "TerminalConnectButton" });
      expect(buttons).toHaveLength(2);
      expect(buttons[0].props("deviceUid")).toBe("device-1");
      expect(buttons[0].props("deviceName")).toBe("device-one");
      expect(buttons[0].props("online")).toBe(true);
    });

    it("Renders DeviceIcon for each device", () => {
      const icons = wrapper.findAllComponents({ name: "DeviceIcon" });
      expect(icons).toHaveLength(2);
      expect(icons[0].props("icon")).toBe("ubuntu");
    });

    it("Shows device tags", () => {
      const tagChips = wrapper.findAll('[data-test="tag-chip"]');
      expect(tagChips).toHaveLength(2);
      expect(tagChips[0].text()).toBe("production");
      expect(tagChips[1].text()).toBe("web-server");
    });

    it("Shows 'No tags' when device has no tags", () => {
      const rows = wrapper.findAll("tr");
      const secondRow = rows[2]; // Index 2 because the headers occupy the first row
      expect(secondRow.text()).toContain("No tags");
    });

    it("Renders SSHID chip for each device", () => {
      const sshidChips = wrapper.findAll('[data-test="sshid-chip"]');
      expect(sshidChips).toHaveLength(2);
      expect(sshidChips[0].text()).toContain("user-ns.device-one@");
    });

    it("Shows SSHID help button", () => {
      const helpButtons = wrapper.findAll('[data-test="sshid-help-btn"]');
      expect(helpButtons).toHaveLength(2);
    });

    it("Opens terminal helper when clicking SSHID help button", async () => {
      const helpButton = wrapper.find('[data-test="sshid-help-btn"]');
      await helpButton.trigger("click");

      const helperContent = new DOMWrapper(document.body).find('[data-test="sshid-helper-component"] .v-overlay__content');
      expect(helperContent.exists()).toBe(true);
      expect(helperContent.attributes("style")).not.toContain("display: none");
    });

    it("Renders actions menu button for each device", () => {
      const menuButtons = wrapper.findAll('[data-test="open-actions-menu-btn"]');
      expect(menuButtons).toHaveLength(2);
    });

    it("Shows device details link", () => {
      wrapper?.unmount();
      mountWrapper({ status: "pending" });

      const detailsLinks = wrapper.findAllComponents({ name: "RouterLink" });
      expect(detailsLinks.length).toBeGreaterThan(0);
    });

    it("Renders TagFormUpdate component", () => {
      const tagForms = wrapper.findAllComponents({ name: "TagFormUpdate" });
      expect(tagForms).toHaveLength(2);
      expect(tagForms[0].props("deviceUid")).toBe("device-1");
      expect(tagForms[0].props("tagsList")).toEqual(mockDevices[0].tags);
    });

    it("Renders DeviceRename component", () => {
      const renameComponents = wrapper.findAllComponents({ name: "DeviceRename" });
      expect(renameComponents).toHaveLength(2);
      expect(renameComponents[0].props("uid")).toBe("device-1");
      expect(renameComponents[0].props("name")).toBe("device-one");
    });

    it("Renders DeviceDelete component", () => {
      const deleteComponents = wrapper.findAllComponents({ name: "DeviceDelete" });
      expect(deleteComponents).toHaveLength(2);
      expect(deleteComponents[0].props("uid")).toBe("device-1");
      expect(deleteComponents[0].props("variant")).toBe("device");
    });

    it("Disables TagFormUpdate when user lacks permission", () => {
      wrapper.unmount();
      mountWrapper({ hasTagUpdatePermission: false });

      const tagForms = wrapper.findAllComponents({ name: "TagFormUpdate" });
      expect(tagForms[0].props("hasAuthorization")).toBe(false);
    });

    it("Disables DeviceDelete when user lacks permission", () => {
      wrapper.unmount();
      mountWrapper({ hasDeviceRemovePermission: false });

      const deleteComponents = wrapper.findAllComponents({ name: "DeviceDelete" });
      expect(deleteComponents[0].props("hasAuthorization")).toBe(false);
    });
  });

  describe("Pending devices row rendering", () => {
    it("Renders different structure for pending devices", () => {
      const mockStoreMethods = createMockStoreMethods(mockPendingDevices);
      wrapper.unmount();
      mountWrapper({ status: "pending", header: "secondary", mockStoreMethods });

      const dataTable = wrapper.findComponent({ name: "DataTable" });
      expect(dataTable.props("items")).toEqual(mockPendingDevices);
    });

    it("Renders DeviceActionButton for accept", () => {
      wrapper.unmount();
      mountWrapper({ status: "pending", header: "secondary" });

      const acceptButtons = wrapper.findAllComponents('[data-test="DeviceActionButtonAccept-component"]');
      expect(acceptButtons.length).toBeGreaterThan(0);
    });

    it("Renders DeviceActionButton for reject", () => {
      wrapper.unmount();
      mountWrapper({ status: "pending", header: "secondary" });

      const rejectButtons = wrapper.findAllComponents('[data-test="deviceActionButtonReject-component"]');
      expect(rejectButtons.length).toBeGreaterThan(0);
    });
  });

  describe("Sorting", () => {
    it("Fetches devices with sort parameters when sorting", async () => {
      await flushPromises();

      const dataTable = wrapper.findComponent({ name: "DataTable" });
      await dataTable.vm.$emit("update:sort", "name");
      await flushPromises();

      expect(storeMethods.fetchDevices).toHaveBeenCalledWith(
        expect.objectContaining({
          sortField: "name",
          sortOrder: "asc",
        }),
      );
    });

    it("Toggles sort order on subsequent sorts", async () => {
      await flushPromises();

      const dataTable = wrapper.findComponent({ name: "DataTable" });

      await dataTable.vm.$emit("update:sort", "name");
      await flushPromises();

      await dataTable.vm.$emit("update:sort", "name");
      await flushPromises();

      expect(storeMethods.fetchDevices).toHaveBeenLastCalledWith(
        expect.objectContaining({
          sortField: "name",
          sortOrder: "desc",
        }),
      );
    });
  });

  describe("Pagination", () => {
    it("Fetches devices when page changes", async () => {
      await flushPromises();

      const dataTable = wrapper.findComponent({ name: "DataTable" });
      await dataTable.vm.$emit("update:page", 2);
      await flushPromises();

      expect(storeMethods.fetchDevices).toHaveBeenCalledWith(
        expect.objectContaining({
          page: 2,
        }),
      );
    });

    it("Fetches devices when items per page changes", async () => {
      await flushPromises();

      const dataTable = wrapper.findComponent({ name: "DataTable" });
      await dataTable.vm.$emit("update:itemsPerPage", 20);
      await flushPromises();

      expect(storeMethods.fetchDevices).toHaveBeenCalledWith(
        expect.objectContaining({
          perPage: 20,
        }),
      );
    });
  });

  describe("Device actions", () => {
    it("Refetches devices after successful update", async () => {
      await flushPromises();

      const renameComponent = wrapper.findComponent({ name: "DeviceRename" });
      await renameComponent.vm.$emit("update");
      await flushPromises();

      expect(storeMethods.fetchDevices).toHaveBeenCalled();
    });

    it("Refetches devices after tag update", async () => {
      await flushPromises();
      vi.clearAllMocks();

      const tagForm = wrapper.findComponent({ name: "TagFormUpdate" });
      await tagForm.vm.$emit("update");
      await flushPromises();

      expect(storeMethods.fetchDevices).toHaveBeenCalled();
    });

    it("Refetches devices after device deletion", async () => {
      await flushPromises();
      vi.clearAllMocks();

      const deleteComponent = wrapper.findComponent({ name: "DeviceDelete" });
      await deleteComponent.vm.$emit("update");
      await flushPromises();

      expect(storeMethods.fetchDevices).toHaveBeenCalled();
    });

    it("Refetches devices after device action", async () => {
      wrapper.unmount();
      mountWrapper({ status: "pending", header: "secondary" });

      await flushPromises();
      vi.clearAllMocks();

      const actionButton = wrapper.findComponent({ name: "DeviceActionButton" });
      await actionButton.vm.$emit("update");
      await flushPromises();

      expect(storeMethods.fetchDevices).toHaveBeenCalled();
    });
  });
});
