import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import NamespaceGenerateApiKey from "@/components/Namespace/NamespaceGenerateApiKey.vue";
import { namespacesApi, usersApi, apiKeysApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { INotificationsError } from "@/interfaces/INotifications";

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

type NamespaceGenerateApiKeyWrapper = VueWrapper<InstanceType<typeof NamespaceGenerateApiKey>>;

describe("Namespace Api Key Generate", () => {
  let wrapper: NamespaceGenerateApiKeyWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  let mockApiKeys: MockAdapter;

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

  const session = true;

  const getKeyResponse = [
    {
      id: "3e5a5194-9dec-4a32-98db-7434c6d49df1",
      tenant_id: "fake-tenant",
      user_id: "507f1f77bcf86cd799439011",
      name: "my api key",
      expires_in: 1707958989,
    },
    {
      id: "3e5a5194-9dec-4a32-98db-7434c6d49df2",
      tenant_id: "fake-tenant",
      user_id: "507f1f77bcf86cd799439011",
      name: "my api key",
      expires_in: 1707958989,
    },
  ];

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockApiKeys = new MockAdapter(apiKeysApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockApiKeys.onGet("http://localhost:3000/api/namespaces/fake-tenant/api-key").reply(200, getKeyResponse, { "x-total-count": 2 });

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("security/setSecurity", session);
    store.commit("auth/setKeyList", { data: getKeyResponse, headers: { "x-total-count": 2 } });

    wrapper = mount(NamespaceGenerateApiKey, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
        config: {
          errorHandler: () => { /* ignore global error handler */ },
        },
      },
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
    expect(wrapper.find('[data-test="namespace-generate-main-btn"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="namespace-generate-main-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="namespace-generate-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="namespace-generate-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="key-name-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="namespace-generate-date"]').exists()).toBe(true);
    expect(dialog.find('[data-test="successKey-alert"]').exists()).toBe(false);
    expect(dialog.find('[data-test="keyResponse-text"]').exists()).toBe(false);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="add-btn"]').exists()).toBe(true);
  });

  it("Successfully Generate Api Key", async () => {
    mockApiKeys.onPost("http://localhost:3000/api/namespaces/fake-tenant/api-key").reply(200);

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="namespace-generate-main-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="key-name-text"]').setValue("my api key");

    await wrapper.findComponent('[data-test="add-btn"]').trigger("click");
    await flushPromises();
    expect(StoreSpy).toHaveBeenCalledWith("auth/generateApiKey", {
      name: "my api key",
      expires_at: 30,
      tenant: "fake-tenant",
    });
    expect(dialog.find('[data-test="successKey-alert"]').exists()).toBe(true);
    expect(dialog.find('[data-test="keyResponse-text"]').exists()).toBe(true);
  });

  it("Fails to Generate Api Key", async () => {
    mockApiKeys.onPost("http://localhost:3000/api/namespaces/fake-tenant/api-key").reply(404);

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="namespace-generate-main-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="key-name-text"]').setValue("my api key");

    await wrapper.findComponent('[data-test="add-btn"]').trigger("click");
    await flushPromises();
    expect(StoreSpy).toHaveBeenCalledWith(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.generateKey,
    );
    expect(dialog.find('[data-test="failMessage-alert"]').exists()).toBe(true);
  });
});