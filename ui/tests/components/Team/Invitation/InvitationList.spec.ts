import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { mockInvitation } from "@tests/mocks/invitation";
import * as hasPermissionModule from "@/utils/permission";
import InvitationList from "@/components/Team/Invitation/InvitationList.vue";
import useInvitationsStore from "@/store/modules/invitations";
import moment from "moment";

vi.mock("@/utils/permission");

describe("InvitationList", () => {
  let wrapper: VueWrapper<InstanceType<typeof InvitationList>>;
  let invitationsStore: ReturnType<typeof useInvitationsStore>;

  const mockInvitations = [
    mockInvitation,
    {
      ...mockInvitation,
      user: { id: "user2", email: "user2@example.com" },
      status: "accepted" as const,
    },
  ];

  const mountWrapper = ({
    invitations = mockInvitations,
    canEditInvitation = true,
    canCancelInvitation = true,
    canSendInvitation = true,
  } = {}) => {
    vi.mocked(hasPermissionModule.default).mockImplementation((permission: string) => {
      if (permission === "namespace:editInvitation") return canEditInvitation;
      if (permission === "namespace:cancelInvitation") return canCancelInvitation;
      if (permission === "namespace:addMember") return canSendInvitation;
      return false;
    });

    wrapper = mountComponent(InvitationList, {
      piniaOptions: {
        initialState: {
          invitations: { namespaceInvitations: invitations, invitationCount: invitations.length },
          auth: { tenantId: "tenant1" },
        },
      },
    });

    invitationsStore = useInvitationsStore();
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe("Component rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders data table", () => {
      const table = wrapper.find('[data-test="invitations-list"]');
      expect(table.exists()).toBe(true);
    });

    it("renders status filter select", () => {
      const statusSelect = wrapper.find('[data-test="invitation-status-select"]');
      expect(statusSelect.exists()).toBe(true);
    });

    it("displays invitation emails", () => {
      const rows = wrapper.findAll("tbody tr");
      expect(rows.length).toBeGreaterThan(0);
    });

    it("displays action menu for pending invitations", () => {
      const actionBtn = wrapper.find('[data-test="invitation-actions"]');
      expect(actionBtn.exists()).toBe(true);
    });
  });

  describe("Data fetching", () => {
    it("calls fetchNamespaceInvitationList with correct params", async () => {
      mountWrapper();
      await flushPromises();

      expect(invitationsStore.fetchNamespaceInvitationList).toHaveBeenCalledWith(
        "tenant1",
        1,
        10,
        expect.any(String),
      );
    });

    it("calls fetchNamespaceInvitationList when page changes", async () => {
      mountWrapper();
      vi.clearAllMocks();

      await wrapper.findComponent({ name: "DataTable" }).vm.$emit("update:page", 2);
      await flushPromises();

      expect(invitationsStore.fetchNamespaceInvitationList).toHaveBeenCalledWith(
        "tenant1",
        2,
        10,
        expect.any(String),
      );
    });

    it("calls fetchNamespaceInvitationList when items per page changes", async () => {
      mountWrapper();
      vi.clearAllMocks();

      await wrapper.findComponent({ name: "DataTable" }).vm.$emit("update:itemsPerPage", 20);
      await flushPromises();

      expect(invitationsStore.fetchNamespaceInvitationList).toHaveBeenCalledWith(
        "tenant1",
        1,
        20,
        expect.any(String),
      );
    });
  });

  describe("Status filtering", () => {
    beforeEach(() => mountWrapper());

    it("filters by pending status by default", () => {
      const statusSelect = wrapper.find('[data-test="invitation-status-select"] input').element as HTMLInputElement;
      expect(statusSelect.value).toBe("Pending");
    });

    it("calls fetchNamespaceInvitationList when status filter changes", async () => {
      const statusSelect = wrapper.findComponent({ name: "VSelect" });
      await statusSelect.vm.$emit("update:modelValue", "accepted");
      await flushPromises();

      expect(invitationsStore.fetchNamespaceInvitationList).toHaveBeenCalled();
    });

    it("resets page to 1 when filter changes", async () => {
      await wrapper.findComponent({ name: "DataTable" }).vm.$emit("update:page", 3);
      await flushPromises();

      const statusSelect = wrapper.findComponent({ name: "VSelect" });
      await statusSelect.vm.$emit("update:modelValue", "cancelled");
      await flushPromises();

      expect(invitationsStore.fetchNamespaceInvitationList).toHaveBeenCalledWith(
        "tenant1",
        1,
        10,
        expect.any(String),
      );
    });
  });

  describe("Invitation status display", () => {
    it("displays pending status with warning color", () => {
      const pendingInvitation = { ...mockInvitation, status: "pending" as const };
      mountWrapper({ invitations: [pendingInvitation] });

      const statusChip = wrapper.find(".v-chip");
      expect(statusChip.classes()).toContain("text-warning");
    });

    it("displays accepted status with success color", () => {
      const acceptedInvitation = { ...mockInvitation, status: "accepted" as const };
      mountWrapper({ invitations: [acceptedInvitation] });

      const statusChip = wrapper.find(".v-chip");
      expect(statusChip.classes()).toContain("text-success");
    });

    it("displays rejected status with error color", () => {
      const rejectedInvitation = { ...mockInvitation, status: "rejected" as const };
      mountWrapper({ invitations: [rejectedInvitation] });

      const statusChip = wrapper.find(".v-chip");
      expect(statusChip.classes()).toContain("text-error");
    });

    it("displays cancelled status with grey color", () => {
      const cancelledInvitation = { ...mockInvitation, status: "cancelled" as const };
      mountWrapper({ invitations: [cancelledInvitation] });

      const statusChip = wrapper.find(".v-chip");
      expect(statusChip.classes()).toContain("text-grey");
    });
  });

  describe("Expiration display", () => {
    it("shows expiration date for pending invitation", () => {
      const futureDate = moment().add(7, "days").toISOString();
      const pendingInvitation = { ...mockInvitation, status: "pending" as const, expires_at: futureDate };
      mountWrapper({ invitations: [pendingInvitation] });

      const dateCell = wrapper.find('[data-test="invitation-date-cell"]');
      expect(dateCell.exists()).toBe(true);
    });

    it("shows expired status for expired pending invitation", () => {
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };
      mountWrapper({ invitations: [expiredInvitation] });

      const dateCell = wrapper.find('[data-test="invitation-date-cell"]');
      expect(dateCell.text()).toContain("Expired at");
    });

    it("shows error styling for expired invitation", () => {
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };
      mountWrapper({ invitations: [expiredInvitation] });

      const dateCell = wrapper.find('[data-test="invitation-date-cell"] span');
      expect(dateCell.classes()).toContain("text-error");
    });

    it("shows status updated date for accepted invitation", () => {
      const acceptedInvitation = {
        ...mockInvitation,
        status: "accepted" as const,
        status_updated_at: "2025-12-15T00:00:00Z",
      };
      mountWrapper({ invitations: [acceptedInvitation] });

      const dateCell = wrapper.find('[data-test="invitation-date-cell"]');
      expect(dateCell.exists()).toBe(true);
    });
  });

  describe("Actions menu", () => {
    it("shows resend, edit, and cancel actions for pending invitation", () => {
      const pendingInvitation = { ...mockInvitation, status: "pending" as const };
      mountWrapper({ invitations: [pendingInvitation] });

      expect(wrapper.findComponent({ name: "InvitationResend" }).exists()).toBe(true);
      expect(wrapper.findComponent({ name: "InvitationEdit" }).exists()).toBe(true);
      expect(wrapper.findComponent({ name: "InvitationCancel" }).exists()).toBe(true);
    });

    it("shows resend action for cancelled invitation", () => {
      const cancelledInvitation = { ...mockInvitation, status: "cancelled" as const };
      mountWrapper({ invitations: [cancelledInvitation] });

      expect(wrapper.findComponent({ name: "InvitationResend" }).exists()).toBe(true);
    });

    it("does not show actions menu for accepted invitation", () => {
      const acceptedInvitation = { ...mockInvitation, status: "accepted" as const };
      mountWrapper({ invitations: [acceptedInvitation] });

      const actionBtn = wrapper.find('[data-test="invitation-actions"]');
      expect(actionBtn.exists()).toBe(false);
    });

    it("does not show actions menu for rejected invitation", () => {
      const rejectedInvitation = { ...mockInvitation, status: "rejected" as const };
      mountWrapper({ invitations: [rejectedInvitation] });

      const actionBtn = wrapper.find('[data-test="invitation-actions"]');
      expect(actionBtn.exists()).toBe(false);
    });
  });

  describe("Error handling", () => {
    it("shows error snackbar on 403 status", async () => {
      const error = createAxiosError(403, "Forbidden");

      mountWrapper();
      vi.mocked(invitationsStore.fetchNamespaceInvitationList).mockRejectedValueOnce(error);
      await flushPromises();

      await wrapper.findComponent({ name: "DataTable" }).vm.$emit("update:page", 2);
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("You don't have permission to access this resource.");
    });

    it("shows generic error snackbar for other errors", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      mountWrapper();
      vi.mocked(invitationsStore.fetchNamespaceInvitationList).mockRejectedValueOnce(error);
      await flushPromises();

      await wrapper.findComponent({ name: "DataTable" }).vm.$emit("update:page", 2);
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load the invitation list.");
    });
  });

  describe("Refresh functionality", () => {
    beforeEach(() => mountWrapper());

    it("calls fetchNamespaceInvitationList when getInvitations is called", async () => {
      vi.clearAllMocks();

      await wrapper.vm.getInvitations();
      await flushPromises();

      expect(invitationsStore.fetchNamespaceInvitationList).toHaveBeenCalled();
    });

    it("sets status filter to pending when setStatusFilterToPending is called", async () => {
      // Change filter first
      const statusSelect = wrapper.findComponent({ name: "VSelect" });
      await statusSelect.vm.$emit("update:modelValue", "accepted");
      await flushPromises();

      // Reset to pending
      wrapper.vm.setStatusFilterToPending();
      await flushPromises();

      const statusSelectInput = wrapper.find('[data-test="invitation-status-select"] input').element as HTMLInputElement;
      expect(statusSelectInput.value).toBe("Pending");
    });
  });

  describe("Permission-based actions", () => {
    it("enables all actions when user has all permissions", () => {
      mountWrapper({
        canEditInvitation: true,
        canCancelInvitation: true,
        canSendInvitation: true,
      });

      expect(wrapper.findComponent({ name: "InvitationEdit" }).exists()).toBe(true);
      expect(wrapper.findComponent({ name: "InvitationCancel" }).exists()).toBe(true);
      expect(wrapper.findComponent({ name: "InvitationResend" }).exists()).toBe(true);
    });
  });
});
