import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import ConnectorDelete from "@/components/Connector/ConnectorDelete.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

type ConnectorDeleteWrapper = VueWrapper<InstanceType<typeof ConnectorDelete>>;

describe("Connector Delete", () => {
  let wrapper: ConnectorDeleteWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  const members = [
    {
      id: "507f1f77bcf86cd799439011",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
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
  };

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant",
    email: "test@test.com",
    id: "507f1f77bcf86cd799439011",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    wrapper = mount(ConnectorDelete, {
      global: {
        plugins: [[store, key], vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        uid: "fake-fingerprint",
        hasAuthorization: true,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", async () => {
    expect(wrapper.find('[data-test="connector-remove-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="remove-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="remove-title"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="connector-remove-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="text-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="remove-btn"]').exists()).toBe(true);
  });

  it("Successfully removes a connector", async () => {
    await wrapper.setProps({ uid: "fake-fingerprint" });
    await wrapper.findComponent('[data-test="connector-remove-btn"]').trigger("click");
    mockNamespace.onDelete("http://localhost:3000/api/connector/fake-fingerprint").reply(200);
    const removeSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");
    expect(removeSpy).toHaveBeenCalledWith("connectors/remove", "fake-fingerprint");
  });

  it("Shows error snackbar if removing a connector fails", async () => {
    await wrapper.setProps({ uid: "fake-fingerprint" });
    await wrapper.findComponent('[data-test="connector-remove-btn"]').trigger("click");
    mockNamespace.onDelete("http://localhost:3000/api/connector/fake-fingerprint").reply(404); // non-existent key
    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to remove connector.");
  });
});
