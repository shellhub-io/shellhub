import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import InvitationsMenuItem from "@/components/Invitations/InvitationsMenuItem.vue";
import { IInvitation } from "@/interfaces/IInvitation";
import { formatFullDateTime } from "@/utils/date";
import { SnackbarPlugin } from "@/plugins/snackbar";

type InvitationsMenuItemWrapper = VueWrapper<InstanceType<typeof InvitationsMenuItem>>;

const mockInvitation: IInvitation = {
  status: "pending",
  role: "operator",
  invited_by: "638af3e2c3a5f90008c8b456",
  expires_at: "2025-12-31T23:59:59Z",
  created_at: "2025-12-01T00:00:00Z",
  updated_at: "2025-12-01T00:00:00Z",
  status_updated_at: "2025-12-01T00:00:00Z",
  namespace: {
    tenant_id: "tenant1",
    name: "Test Namespace",
  },
  user: {
    id: "user1",
    email: "user@example.com",
  },
};

describe("InvitationsMenuItem", () => {
  let wrapper: InvitationsMenuItemWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(InvitationsMenuItem, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
      props: {
        invitation: mockInvitation,
      },
    });
  });

  it("Displays namespace name", () => {
    expect(wrapper.text()).toContain("Test Namespace");
  });

  it("Displays role", () => {
    expect(wrapper.text()).toContain(mockInvitation.role);
  });

  it("Displays invitation description", () => {
    const formattedCreatedAt = formatFullDateTime(mockInvitation.created_at);
    expect(wrapper.text()).toContain(`Invited by ${mockInvitation.invited_by} at ${formattedCreatedAt}`);
  });

  it("Emits update event when invitation is accepted/declined successfully", () => {
    wrapper.vm.handleSuccess();
    expect(wrapper.emitted("update")).toBeTruthy();
  });
});
