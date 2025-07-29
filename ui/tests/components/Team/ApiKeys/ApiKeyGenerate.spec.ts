import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import ApiKeyGenerate from "@/components/Team/ApiKeys/ApiKeyGenerate.vue";
import { apiKeysApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useApiKeysStore from "@/store/modules/api_keys";

type ApiKeyGenerateWrapper = VueWrapper<InstanceType<typeof ApiKeyGenerate>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
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

describe("Api Key Generate", () => {
  let wrapper: ApiKeyGenerateWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockApiKeysApi = new MockAdapter(apiKeysApi.getAxios());
  const apiKeysStore = useApiKeysStore();

  beforeEach(async () => {
    store.commit("auth/authSuccess", authData);
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
    mockApiKeysApi.onPost("http://localhost:3000/api/namespaces/api-key").reply(200, { id: "fake-id" });

    const storeSpy = vi.spyOn(apiKeysStore, "generateApiKey");

    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="key-name-text"]').setValue("my api key");
    await wrapper.findComponent('[data-test="add-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith({
      name: "my api key",
      role: "administrator",
      expires_in: 30,
    });
  });

  it("Fails to Generate Api Key", async () => {
    mockApiKeysApi.onPost("http://localhost:3000/api/namespaces/api-key").reply(500);

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
    mockApiKeysApi.onPost("http://localhost:3000/api/namespaces/api-key").reply(400);

    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="key-name-text"]').setValue("");

    await wrapper.findComponent('[data-test="add-btn"]').trigger("click");

    await flushPromises();

    expect(wrapper.vm.errorMessage).toBe("Please provide a name for the API key.");

    expect(dialog.find('[data-test="fail-message-alert"]').exists()).toBe(true);
  });

  it("Fails to Generate Api Key (401)", async () => {
    mockApiKeysApi.onPost("http://localhost:3000/api/namespaces/api-key").reply(401);

    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");

    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="key-name-text"]').setValue("my api key");

    await wrapper.findComponent('[data-test="add-btn"]').trigger("click");

    await flushPromises();

    expect(dialog.find('[data-test="fail-message-alert"]').exists()).toBe(true);
  });

  it("Fails to Generate Api Key (409)", async () => {
    mockApiKeysApi.onPost("http://localhost:3000/api/namespaces/api-key").reply(409);

    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="key-name-text"]').setValue("my api key");

    await wrapper.findComponent('[data-test="add-btn"]').trigger("click");
    await flushPromises();

    expect(wrapper.vm.errorMessage).toBe("An API key with the same name already exists.");

    expect(dialog.find('[data-test="fail-message-alert"]').exists()).toBe(true);
  });
});
