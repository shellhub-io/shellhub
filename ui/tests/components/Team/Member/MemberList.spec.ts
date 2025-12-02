import { setActivePinia, createPinia } from "pinia";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import MemberList from "@/components/Team/Member/MemberList.vue";
import { namespacesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";
import { INamespace } from "@/interfaces/INamespace";

type MemberListWrapper = VueWrapper<InstanceType<typeof MemberList>>;

describe("Member List", () => {
  let wrapper: MemberListWrapper;
  setActivePinia(createPinia());
  const namespacesStore = useNamespacesStore();
  const vuetify = createVuetify();

  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  const namespaceData = {
    name: "user",
    owner: "xxxxxxxx",
    tenant_id: "fake-tenant-data",
    members: [
      {
        id: "xxxxxxxx",
        email: "test@test.com",
        role: "owner" as const,
        added_at: "2024-01-01T12:00:00Z",
      },
    ],
    settings: {
      session_record: true,
    },
    max_devices: 3,
    devices_accepted_count: 2,
    devices_rejected_count: 0,
    devices_pending_count: 0,
    created_at: "",
    billing: null,
  };

  beforeEach(() => {
    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    namespacesStore.currentNamespace = namespaceData as INamespace;

    wrapper = shallowMount(MemberList, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component HTML", () => {
    expect(wrapper.findComponent('[data-test="member-table"]').exists()).toBe(true);
  });
});
