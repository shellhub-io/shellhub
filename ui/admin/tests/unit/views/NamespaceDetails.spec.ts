import { describe, expect, it, beforeEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import { formatFullDateTime } from "@/utils/date";
import useNamespacesStore from "@admin/store/modules/namespaces";
import NamespaceDetails from "@admin/views/NamespaceDetails.vue";
import { mockNamespace } from "../mocks";
import { afterEach } from "vitest";

vi.mock("@admin/store/api/namespaces");

describe("NamespaceDetails", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceDetails>>;

  const devicesCount = mockNamespace.devices_accepted_count
    + mockNamespace.devices_pending_count
    + mockNamespace.devices_rejected_count;

  const mountWrapper = async (mockError?: Error) => {
    const router = createCleanAdminRouter();
    await router.push({ name: "namespaceDetails", params: { id: mockNamespace.tenant_id } });
    await router.isReady();

    wrapper = mountComponent(NamespaceDetails, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: { adminNamespaces: mockError ? {} : { namespace: mockNamespace } },
        stubActions: !mockError,
      },
    });

    const namespacesStore = useNamespacesStore();
    if (mockError) vi.mocked(namespacesStore.fetchNamespaceById).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when namespace loads successfully", () => {
    beforeEach(() => mountWrapper());

    it("displays the title", () => {
      expect(wrapper.find("h1").text()).toBe("Namespace Details");
    });

    it("displays namespace name in card title and field", () => {
      expect(wrapper.find(".text-h6").text()).toContain(mockNamespace.name);
      const nameField = wrapper.find('[data-test="namespace-name-field"]');
      expect(nameField.text()).toContain("Name:");
      expect(nameField.text()).toContain(mockNamespace.name);
    });

    it("shows type chip with correct value", () => {
      const typeChip = wrapper.find('[data-test="namespace-type-chip"]');
      expect(typeChip.exists()).toBe(true);
      expect(typeChip.text()).toBe(mockNamespace.type);
    });

    it("displays tenant id", () => {
      const tenantField = wrapper.find('[data-test="namespace-tenant-id-field"]');
      expect(tenantField.text()).toContain("Tenant ID:");
      expect(tenantField.text()).toContain(mockNamespace.tenant_id);
    });

    it("displays owner with link", () => {
      const ownerField = wrapper.find('[data-test="namespace-owner-field"]');
      expect(ownerField.text()).toContain("Owner:");
      const link = ownerField.find("a");
      expect(link.exists()).toBe(true);
      expect(link.text()).toBe(mockNamespace.owner);
    });

    it("displays total devices count", () => {
      const devicesField = wrapper.find('[data-test="namespace-devices-field"]');
      expect(devicesField.text()).toContain("Total Devices:");
      expect(devicesField.text()).toContain(String(devicesCount));
    });

    it("displays devices breakdown with correct counts", () => {
      const breakdown = wrapper.find('[data-test="namespace-devices-breakdown"]');
      const accepted = breakdown.find('[data-test="namespace-devices-accepted"]');
      expect(accepted.text()).toContain("Accepted:");
      expect(accepted.text()).toContain(String(mockNamespace.devices_accepted_count));

      const pending = breakdown.find('[data-test="namespace-devices-pending"]');
      expect(pending.text()).toContain("Pending:");
      expect(pending.text()).toContain(String(mockNamespace.devices_pending_count));

      const rejected = breakdown.find('[data-test="namespace-devices-rejected"]');
      expect(rejected.text()).toContain("Rejected:");
      expect(rejected.text()).toContain(String(mockNamespace.devices_rejected_count));
    });

    it("displays created at date", () => {
      const createdField = wrapper.find('[data-test="namespace-created-field"]');
      expect(createdField.text()).toContain("Created:");
      expect(createdField.text()).toContain(formatFullDateTime(mockNamespace.created_at));
    });

    it("displays max devices", () => {
      const maxDevicesField = wrapper.find('[data-test="namespace-max-devices-field"]');
      expect(maxDevicesField.text()).toContain("Max Devices:");
      expect(maxDevicesField.text()).toContain(String(mockNamespace.max_devices));
    });

    it("displays session record setting", () => {
      const sessionRecordField = wrapper.find('[data-test="namespace-session-record-field"]');
      expect(sessionRecordField.text()).toContain("Session Record:");
      expect(sessionRecordField.text()).toContain("Enabled");
    });

    it("displays connection announcement", () => {
      const announcementField = wrapper.find('[data-test="namespace-connection-announcement-field"]');
      expect(announcementField.text()).toContain("Connection Announcement:");
      expect(announcementField.text()).toContain(mockNamespace.settings.connection_announcement);
    });

    it("displays members section", () => {
      const membersSection = wrapper.find('[data-test="namespace-members-section"]');
      expect(membersSection.exists()).toBe(true);
      expect(membersSection.text()).toContain(`Members (${mockNamespace.members.length})`);
    });

    it("displays members list with all member items", () => {
      const membersList = wrapper.find('[data-test="namespace-members-list"]');
      expect(membersList.exists()).toBe(true);
      const memberItems = wrapper.findAll('[data-test="namespace-member-item"]');
      expect(memberItems.length).toBe(mockNamespace.members.length);
    });

    it("displays member roles with correct values", () => {
      const roles = wrapper.findAll('[data-test="namespace-member-role"]');
      expect(roles.length).toBe(mockNamespace.members.length);
      expect(roles[0].text()).toContain("owner");
    });

    it("displays member statuses with correct values", () => {
      const statuses = wrapper.findAll('[data-test="namespace-member-status"]');
      expect(statuses.length).toBe(mockNamespace.members.length);
      expect(statuses[0].text()).toBe(mockNamespace.members[0].status);
    });

    it("displays member ids", () => {
      const memberIds = wrapper.findAll('[data-test="namespace-member-id"]');
      expect(memberIds.length).toBe(mockNamespace.members.length);
      expect(memberIds[0].text()).toContain(mockNamespace.members[0].id);
    });

    it("displays member added dates", () => {
      const addedDates = wrapper.findAll('[data-test="namespace-member-added"]');
      expect(addedDates.length).toBeGreaterThan(0);
      expect(addedDates[0].text()).toContain("Added:");
      expect(addedDates[0].text()).toContain(formatFullDateTime(mockNamespace.members[0].added_at));
    });
  });

  describe("when namespace fails to load", () => {
    it("shows error snackbar", async () => {
      await mountWrapper(createAxiosError(404, "Not Found"));

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to fetch namespace details.");
    });
  });
});
