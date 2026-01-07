import { VueWrapper, flushPromises } from "@vue/test-utils";
import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import TeamInvitations from "@/views/TeamInvitations.vue";
import InvitationList from "@/components/Team/Invitation/InvitationList.vue";

describe("TeamInvitations View", () => {
  let wrapper: VueWrapper<InstanceType<typeof TeamInvitations>>;

  const mountWrapper = () => { wrapper = mountComponent(TeamInvitations); };

  afterEach(() => { wrapper?.unmount(); });

  describe("when page loads", () => {
    beforeEach(() => mountWrapper());

    it("renders the page header", () => {
      expect(wrapper.find('[data-test="title"]').text()).toContain("Invitations");
      expect(wrapper.text()).toContain("Team Management");
    });

    it("displays the invitation list", () => {
      expect(wrapper.find('[data-test="invitation-list"]').exists()).toBe(true);
    });
  });

  describe("when new invitation is created", () => {
    it("refreshes invitation list when MemberInvite emits update", async () => {
      mountWrapper();

      const invitationListRef = wrapper.vm.$refs.invitationListRef as InstanceType<typeof InvitationList>;
      expect(invitationListRef).toBeDefined();

      const setFilterSpy = vi.spyOn(invitationListRef, "setStatusFilterToPending");
      const getInvitationsSpy = vi.spyOn(invitationListRef, "getInvitations");

      const inviteComponent = wrapper.findComponent({ name: "MemberInvite" });
      inviteComponent.vm.$emit("update");
      await flushPromises();

      expect(setFilterSpy).toHaveBeenCalled();
      expect(getInvitationsSpy).toHaveBeenCalled();
    });
  });
});
