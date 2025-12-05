import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { VLayout } from "vuetify/components";
import InvitationsMenu from "@/components/Invitations/InvitationsMenu.vue";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useInvitationsStore from "@/store/modules/invitations";
import { IInvitation } from "@/interfaces/IInvitation";

const Component = {
  template: "<v-layout><InvitationsMenu v-model=\"show\" /></v-layout>",
  props: ["modelValue"],
  data: () => ({
    show: true,
  }),
};

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

const mockInvitations: IInvitation[] = [
  {
    status: "pending",
    role: "operator",
    invited_by: "admin",
    expires_at: "2025-12-31T23:59:59Z",
    created_at: "2025-12-01T00:00:00Z",
    updated_at: "2025-12-01T00:00:00Z",
    status_updated_at: "2025-12-01T00:00:00Z",
    namespace: {
      tenant_id: "tenant1",
      name: "Namespace 1",
    },
    user: {
      id: "user1",
      email: "user@example.com",
    },
  },
];

const vuetify = createVuetify();

const mountWrapper = () => mount(Component, {
  global: {
    plugins: [vuetify],
    provide: { [SnackbarInjectionKey]: mockSnackbar },
    components: { "v-layout": VLayout, InvitationsMenu },
    stubs: { teleport: true },
  },
  props: { modelValue: true },
  attachTo: document.body,
});

describe("InvitationsMenu", () => {
  let wrapper: VueWrapper<unknown>;
  let menu: VueWrapper<InstanceType<typeof InvitationsMenu>>;

  setActivePinia(createPinia());
  const invitationsStore = useInvitationsStore();
  invitationsStore.fetchUserPendingInvitationList = vi.fn().mockResolvedValue(Promise.resolve(mockInvitations));

  it("Opens drawer when icon is clicked", async () => {
    wrapper = mountWrapper();
    await wrapper.find('[data-test="invitations-menu-icon"]').trigger("click");
    menu = wrapper.findComponent(InvitationsMenu);
    menu.vm.isDrawerOpen = false;
    const icon = wrapper.find('[data-test="invitations-menu-icon"]');
    await icon.trigger("click");
    await flushPromises();

    expect(menu.vm.isDrawerOpen).toBe(true);
    const drawerComponent = wrapper.find('[data-test="invitations-drawer"]');
    expect(drawerComponent.exists()).toBe(true);
  });

  it("Fetches invitations on mount", async () => {
    const storeSpy = vi.spyOn(invitationsStore, "fetchUserPendingInvitationList");
    wrapper.unmount();
    wrapper = mountWrapper();
    await flushPromises();
    expect(storeSpy).toHaveBeenCalled();
  });
});
