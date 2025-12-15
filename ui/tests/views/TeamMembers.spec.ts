import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import TeamMembers from "@/views/TeamMembers.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";
import { INamespaceMember } from "@/interfaces/INamespace";

type TeamMembersWrapper = VueWrapper<InstanceType<typeof TeamMembers>>;

describe("Team Members", () => {
  let wrapper: TeamMembersWrapper;
  setActivePinia(createPinia());
  const namespacesStore = useNamespacesStore();
  const vuetify = createVuetify();

  const members = [
    {
      id: "507f1f77bcf86cd799439011",
      role: "owner" as const,
    },
  ] as INamespaceMember[];

  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    max_devices: 3,
    devices_count: 3,
    created_at: "",
    billing: null,
    settings: {
      session_record: true,
    },
    devices_accepted_count: 3,
    devices_rejected_count: 0,
    devices_pending_count: 0,
    type: "team" as const,
  };

  beforeEach(() => {
    namespacesStore.currentNamespace = namespaceData;

    wrapper = mount(TeamMembers, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="member-list"]').exists()).toBe(true);
  });
});
