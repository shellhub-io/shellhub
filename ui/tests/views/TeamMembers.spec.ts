import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import TeamMembers from "@/views/TeamMembers.vue";
import { store, key } from "@/store";
import { SnackbarPlugin } from "@/plugins/snackbar";

type TeamMembersWrapper = VueWrapper<InstanceType<typeof TeamMembers>>;

describe("Team Members", () => {
  let wrapper: TeamMembersWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  const members = [
    {
      id: "507f1f77bcf86cd799439011",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
    data: {
      name: "test",
      owner: "test",
      tenant_id: "fake-tenant",
      members,
      settings: {
        session_record: true,
        connection_announcement: "",
      },
      max_devices: 3,
      devices_count: 3,
      created_at: "",
    },
  };

  beforeEach(async () => {
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(TeamMembers, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
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

  it("Renders the template with data", async () => {
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="member-invite"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="member-list"]').exists()).toBe(true);
  });
});
