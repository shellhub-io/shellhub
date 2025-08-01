import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import ApiKeyGenerate from "@/components/Team/ApiKeys/ApiKeyGenerate.vue";
import { namespacesApi, usersApi, apiKeysApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

type ApiKeyGenerateWrapper = VueWrapper<InstanceType<typeof ApiKeyGenerate>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Api Key Generate", () => {
  let wrapper: ApiKeyGenerateWrapper;

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

  const getKeyResponse = [
    {
      name: "aaaa2",
      tenant_id: "00000-0000-0000-0000-00000000000",
      role: "administrator",
      created_by: "66562f80daba745a106393b5",
      created_at: "2024-06-07T12:10:56.531Z",
      updated_at: "2024-06-07T12:31:03.505Z",
      expires_in: 1720354256,
    },
    {
      name: "aaaa2",
      tenant_id: "00000-0000-0000-0000-00000000000",
      role: "administrator",
      created_by: "66562f80daba745a106393b5",
      created_at: "2024-06-07T12:10:56.531Z",
      updated_at: "2024-06-07T12:31:03.505Z",
      expires_in: 1720354256,
    },
  ];

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockApiKeys = new MockAdapter(apiKeysApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockApiKeys.onGet("http://localhost:3000/api/namespaces/api-key?page=1&per_page=10").reply(200, getKeyResponse, { "x-total-count": 2 });

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("apiKeys/setKeyList", { data: getKeyResponse, headers: { "x-total-count": 2 } });

    wrapper = mount(ApiKeyGenerate, {
      global: {
        plugins: [[store, key], vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
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
    expect(wrapper.find('[data-test="api-key-generate-main-btn"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="api-key-generate-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="api-key-generate-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="key-name-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="api-key-generate-date"]').exists()).toBe(true);
    expect(dialog.find('[data-test="successKey-alert"]').exists()).toBe(false);
    expect(dialog.find('[data-test="keyResponse-text"]').exists()).toBe(false);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="add-btn"]').exists()).toBe(true);
  });

  it("Successfully Generate Api Key", async () => {
    mockApiKeys.onPost("http://localhost:3000/api/namespaces/api-key").reply(200, { id: "fake-id" });

    const storeSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="key-name-text"]').setValue("my api key");

    await wrapper.findComponent('[data-test="add-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith("apiKeys/generateApiKey", {
      name: "my api key",
      role: "administrator",
      expires_at: 30,
      tenant: "fake-tenant",
    });
  });

  it("Fails to Generate Api Key", async () => {
    mockApiKeys.onPost("http://localhost:3000/api/namespaces/api-key").reply(500);

    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="key-name-text"]').setValue("my api key");

    await wrapper.findComponent('[data-test="add-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to generate API Key.");

    expect(wrapper.vm.errorMessage).toBe("An error occurred while generating your API key. Please try again later.");

    expect(dialog.find('[data-test="fail-message-alert"]').exists()).toBe(true);
  });

  it("Fails to Generate Api Key (400)", async () => {
    mockApiKeys.onPost("http://localhost:3000/api/namespaces/api-key").reply(400);

    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="key-name-text"]').setValue("");

    await wrapper.findComponent('[data-test="add-btn"]').trigger("click");

    await flushPromises();

    expect(wrapper.vm.errorMessage).toBe("Please provide a name for the API key.");

    expect(dialog.find('[data-test="fail-message-alert"]').exists()).toBe(true);
  });

  it("Fails to Generate Api Key (401)", async () => {
    mockApiKeys.onPost("http://localhost:3000/api/namespaces/api-key").reply(401);

    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");

    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="key-name-text"]').setValue("my api key");

    await wrapper.findComponent('[data-test="add-btn"]').trigger("click");

    await flushPromises();

    expect(dialog.find('[data-test="fail-message-alert"]').exists()).toBe(true);
  });

  it("Fails to Generate Api Key (409)", async () => {
    mockApiKeys.onPost("http://localhost:3000/api/namespaces/api-key").reply(409);

    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="key-name-text"]').setValue("my api key");

    await wrapper.findComponent('[data-test="add-btn"]').trigger("click");
    await flushPromises();

    expect(wrapper.vm.errorMessage).toBe("An API key with the same name already exists.");

    expect(dialog.find('[data-test="fail-message-alert"]').exists()).toBe(true);
  });
});
