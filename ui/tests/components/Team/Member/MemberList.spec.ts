import { setActivePinia, createPinia } from "pinia";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import { store, key } from "@/store";
import MemberList from "@/components/Team/Member/MemberList.vue";
import { namespacesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type MemberListWrapper = VueWrapper<InstanceType<typeof MemberList>>;

describe("Member List", () => {
  let wrapper: MemberListWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  const namespaceData = {
    name: "user",
    owner: "xxxxxxxx",
    tenant_id: "fake-tenant-data",
    members: [
      {
        id: "xxxxxxxx",
        username: "test",
        email: "test@test.com",
        role: "owner",
        added_at: "2024-01-01T12:00:00Z",
      },
    ],
    max_devices: 3,
    devices_count: 3,
    devices: 2,
    created_at: "",
    billing: {},
  };

  beforeEach(async () => {
    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = shallowMount(MemberList, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component HTML", async () => {
    expect(wrapper.findComponent('[data-test="member-table"]').exists()).toBe(true);
  });
});
