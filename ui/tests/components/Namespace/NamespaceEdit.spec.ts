import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import NamespaceEdit from "@/components/Namespace/NamespaceEdit.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

type NamespaceEditWrapper = VueWrapper<InstanceType<typeof NamespaceEdit>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Namespace Edit", () => {
  let wrapper: NamespaceEditWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    settings: {
      session_record: true,
      connection_announcement: "",
    },
    max_devices: 3,
    devices_count: 3,
    created_at: "",
  };

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;
    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockNamespace.onGet("http://localhost:3000/api/namespaces?filter=&page=1&per_page=10").reply(200, [namespaceData]);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(NamespaceEdit, {
      global: {
        plugins: [[store, key], vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", async () => {
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.showDialog = true;
    await flushPromises();
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="change-connection-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="connection-announcement-text"]').exists()).toBe(true);
  });

  it("Successfully changes connection_announcement data", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    const changeNamespaceData = {
      tenant_id: "fake-tenant-data",
      settings: {
        connection_announcement: "test",
      },
    };

    mockNamespace.onPut("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, changeNamespaceData);

    await wrapper.findComponent('[data-test="connection-announcement-text"]').setValue("test");

    const changeDataSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="change-connection-btn"]').trigger("click");

    await nextTick();
    await flushPromises();
    expect(changeDataSpy).toHaveBeenCalledWith("namespaces/put", changeNamespaceData);
  });

  it("Fails to change namespace data", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    mockNamespace.onPut("http://localhost:3000/api/namespaces/fake-tenant-data").reply(403);

    await wrapper.findComponent('[data-test="change-connection-btn"]').trigger("click");

    await nextTick();
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while updating the connection announcement.");
  });
});
