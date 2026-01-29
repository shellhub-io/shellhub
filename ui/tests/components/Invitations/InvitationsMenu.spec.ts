import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { VLayout } from "vuetify/components";
import { mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import { mockInvitations } from "@tests/mocks";
import InvitationsMenu from "@/components/Invitations/InvitationsMenu.vue";
import useInvitationsStore from "@/store/modules/invitations";
import { IInvitation } from "@/interfaces/IInvitation";

const Component = {
  template: "<v-layout><InvitationsMenu v-model=\"show\" /></v-layout>",
  props: ["modelValue"],
  data: () => ({
    show: true,
  }),
};

describe("InvitationsMenu", () => {
  let wrapper: VueWrapper<unknown>;
  let invitationsStore: ReturnType<typeof useInvitationsStore>;

  const mountWrapper = (invitations: IInvitation[] = mockInvitations) => {
    wrapper = mountComponent(Component, {
      global: {
        components: { "v-layout": VLayout, InvitationsMenu },
        stubs: { teleport: true },
      },
      props: { modelValue: true },
      attachTo: document.body,
      piniaOptions: { initialState: { invitations: { pendingInvitations: invitations } } },
    });

    invitationsStore = useInvitationsStore();
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Badge and icon", () => {
    it("Renders invitations menu icon", () => {
      mountWrapper();

      expect(wrapper.find('[data-test="invitations-menu-icon"]').exists()).toBe(true);
    });

    it("Shows badge when there are pending invitations", () => {
      mountWrapper();

      const badge = wrapper.find('[data-test="invitations-menu-badge"]');
      expect(badge.exists()).toBe(true);
      expect(badge.text()).toBe("2");
    });

    it("Does not show badge when there are no pending invitations", () => {
      mountWrapper([]);

      const badge = wrapper.find('[data-test="invitations-menu-badge"] .v-badge__badge');
      expect(badge.attributes("style")).toContain("display: none;");
    });
  });

  describe("Drawer toggle", () => {
    beforeEach(() => mountWrapper());

    it("Opens drawer when icon is clicked", async () => {
      await wrapper.find('[data-test="invitations-menu-icon"]').trigger("click");
      await flushPromises();

      expect(wrapper.find('[data-test="invitations-drawer"]').exists()).toBe(true);
    });

    it("Fetches invitations when drawer is opened", async () => {
      await wrapper.find('[data-test="invitations-menu-icon"]').trigger("click");
      await flushPromises();

      expect(invitationsStore.fetchUserPendingInvitationList).toHaveBeenCalled();
    });
  });

  describe("Invitations list", () => {
    beforeEach(() => mountWrapper());

    it("Displays invitations list when there are pending invitations", async () => {
      await wrapper.find('[data-test="invitations-menu-icon"]').trigger("click");
      await flushPromises();

      expect(wrapper.find('[data-test="invitations-list"]').exists()).toBe(true);
    });

    it("Renders invitation items", async () => {
      await wrapper.find('[data-test="invitations-menu-icon"]').trigger("click");
      await flushPromises();

      const items = wrapper.findAllComponents({ name: "InvitationsMenuItem" });
      expect(items).toHaveLength(2);
    });
  });

  describe("Empty state", () => {
    it("Shows empty state when there are no pending invitations", async () => {
      mountWrapper([]);

      await wrapper.find('[data-test="invitations-menu-icon"]').trigger("click");
      await flushPromises();

      expect(wrapper.find('[data-test="empty-state"]').exists()).toBe(true);
      expect(wrapper.text()).toContain("No pending invitations");
    });
  });

  describe("Error handling", () => {
    beforeEach(() => mountWrapper());
    it("Handles fetch error", async () => {
      vi.mocked(invitationsStore.fetchUserPendingInvitationList).mockRejectedValueOnce(
        createAxiosError(500, "Internal server error"),
      );

      await wrapper.find('[data-test="invitations-menu-icon"]').trigger("click");
      await flushPromises();

      expect(invitationsStore.fetchUserPendingInvitationList).toHaveBeenCalled();
    });

    it("Handles 403 permission error", async () => {
      vi.mocked(invitationsStore.fetchUserPendingInvitationList).mockRejectedValueOnce(
        createAxiosError(403, "Permission denied"),
      );

      await wrapper.find('[data-test="invitations-menu-icon"]').trigger("click");
      await flushPromises();

      expect(invitationsStore.fetchUserPendingInvitationList).toHaveBeenCalled();
    });
  });
});
