import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { setActivePinia, createPinia } from "pinia";
import ApiKeyDelete from "@/components/Team/ApiKeys/ApiKeyDelete.vue";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useApiKeysStore from "@/store/modules/api_keys";
import { apiKeysApi } from "@/api/http";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

type ApiKeyDeleteWrapper = VueWrapper<InstanceType<typeof ApiKeyDelete>>;

describe("Api Key Delete", () => {
  let wrapper: ApiKeyDeleteWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockApiKeysApi = new MockAdapter(apiKeysApi.getAxios());
  const apiKeysStore = useApiKeysStore();

  beforeEach(async () => {
    wrapper = mount(ApiKeyDelete, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        keyId: "fake-id",
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
    expect(wrapper.find('[data-test="delete-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="delete-main-btn-title"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="delete-main-btn-title"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="delete-btn"]').exists()).toBe(true);
  });

  it("Successfully Delete Api Key", async () => {
    mockApiKeysApi.onDelete("http://localhost:3000/api/namespaces/api-key/fake-id").reply(200);

    const storeSpy = vi.spyOn(apiKeysStore, "removeApiKey");

    await wrapper.findComponent('[data-test="delete-main-btn-title"]').trigger("click");

    await wrapper.findComponent('[data-test="delete-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith({
      key: "fake-id",
    });
  });

  it("Fails to delete Api Key", async () => {
    mockApiKeysApi.onDelete("http://localhost:3000/api/namespaces/api-key/fake-id").reply(404);

    await wrapper.findComponent('[data-test="delete-main-btn-title"]').trigger("click");

    await wrapper.findComponent('[data-test="delete-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete Api Key.");
  });
});
