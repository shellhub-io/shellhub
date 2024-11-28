import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PrivateKeyAdd from "@/components/PrivateKeys/PrivateKeyAdd.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type PrivateKeyAddWrapper = VueWrapper<InstanceType<typeof PrivateKeyAdd>>;

describe("Setting Private Keys", () => {
  const node = document.createElement("div");
  node.setAttribute("id", "app");
  document.body.appendChild(node);

  let wrapper: PrivateKeyAddWrapper;

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

  const session = true;

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("security/setSecurity", session);
    wrapper = mount(PrivateKeyAdd, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      attachTo: el,
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Renders components", async () => {
    wrapper.vm.dialog = true;
    await flushPromises();
    const dialog = new DOMWrapper(document.body);

    expect(dialog.find('[data-test="card-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="name-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="private-key-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="private-key-cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="private-key-save-btn"]').exists()).toBe(true);
  });

  it("Sets private key data error message", async () => {
    wrapper.vm.dialog = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("not-working-name");

    await wrapper.findComponent('[data-test="name-field"]').setValue("");

    await flushPromises();

    expect(wrapper.vm.nameError).toEqual("this is a required field");
  });

  it("Sets private key data error message", async () => {
    wrapper.vm.dialog = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="private-key-field"]').setValue("not-working-key");

    await wrapper.findComponent('[data-test="private-key-field"]').setValue("");

    await flushPromises();

    expect(wrapper.vm.privateKeyDataError).toEqual("this is a required field");
  });
});
