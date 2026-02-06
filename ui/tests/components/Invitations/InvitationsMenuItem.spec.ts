import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { mockInvitation } from "@tests/mocks";
import InvitationsMenuItem from "@/components/Invitations/InvitationsMenuItem.vue";
import { IInvitation } from "@/interfaces/IInvitation";
import { formatFullDateTime } from "@/utils/date";

describe("InvitationsMenuItem", () => {
  let wrapper: VueWrapper<InstanceType<typeof InvitationsMenuItem>>;

  const mountWrapper = (invitation: IInvitation = mockInvitation) => {
    wrapper = mountComponent(InvitationsMenuItem, { props: { invitation } });
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Invitation display", () => {
    it("Displays namespace name", () => {
      expect(wrapper.text()).toContain("Test Namespace");
    });

    it("Displays role with icon", () => {
      expect(wrapper.text()).toContain("operator");
    });

    it("Displays administrator role correctly", () => {
      wrapper.unmount();
      mountWrapper({
        ...mockInvitation,
        role: "administrator",
      });

      expect(wrapper.text()).toContain("admin");
    });

    it("Displays observer role correctly", () => {
      wrapper.unmount();
      mountWrapper({
        ...mockInvitation,
        role: "observer",
      });

      expect(wrapper.text()).toContain("observer");
    });
  });

  describe("Expand/collapse details", () => {
    it("Shows invitation details when expanded", async () => {
      const expandBtn = wrapper.find("button .mdi-chevron-down"); // Expand button ⌄
      await expandBtn.trigger("click");

      const formattedCreatedAt = formatFullDateTime(mockInvitation.created_at);
      expect(wrapper.text()).toContain(`Invited by ${mockInvitation.invited_by} at ${formattedCreatedAt}`);
    });
  });

  describe("Action buttons", () => {
    it("Renders accept button", () => {
      const acceptComponent = wrapper.findComponent({ name: "InvitationAccept" });
      expect(acceptComponent.exists()).toBe(true);
    });

    it("Renders decline button", () => {
      const declineComponent = wrapper.findComponent({ name: "InvitationDecline" });
      expect(declineComponent.exists()).toBe(true);
    });

    it("Passes correct props to InvitationAccept", () => {
      const acceptComponent = wrapper.findComponent({ name: "InvitationAccept" });
      expect(acceptComponent.props("tenant")).toBe("tenant1");
      expect(acceptComponent.props("namespaceName")).toBe("Test Namespace");
      expect(acceptComponent.props("role")).toBe("operator");
    });

    it("Passes correct props to InvitationDecline", () => {
      const declineComponent = wrapper.findComponent({ name: "InvitationDecline" });
      expect(declineComponent.props("tenant")).toBe("tenant1");
      expect(declineComponent.props("namespaceName")).toBe("Test Namespace");
    });
  });
});
