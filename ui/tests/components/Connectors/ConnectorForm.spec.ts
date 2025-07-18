import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import ConnectorForm from "@/components/Connector/ConnectorForm.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { envVariables } from "@/envVariables";

type ConnectorFormWrapper = VueWrapper<InstanceType<typeof ConnectorForm>>;

describe("Connector Form", () => {
  let wrapper: ConnectorFormWrapper;

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

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(ConnectorForm, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      props: {
        isEditing: false,
        storeMethod: vi.fn(),
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("renders the component", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="connector-form-card"]').exists()).toBe(true);
    expect(dialog.find('[data-test="address-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="port-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="save-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("validates the address field", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    const addressField = wrapper.findComponent('[data-test="address-text"]');

    await addressField.setValue("invalid ip");
    await flushPromises();
    expect(wrapper.findComponent('[data-test="address-text"]').text()).toContain("Invalid IP address format");
  });

  it("validates the port field", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    const addressField = wrapper.findComponent('[data-test="port-text"]');

    await addressField.setValue("invalid port");
    await flushPromises();
    expect(wrapper.findComponent('[data-test="port-text"]').text()).toContain("this must be a `number` type");
  });
});
