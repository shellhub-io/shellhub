import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import ApiKeyEdit from "@/components/Team/ApiKeys/ApiKeyEdit.vue";
import { apiKeysApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useApiKeysStore from "@/store/modules/api_keys";

type ApiKeyEditWrapper = VueWrapper<InstanceType<typeof ApiKeyEdit>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Api Key Edit", () => {
  let wrapper: ApiKeyEditWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockApiKeysApi = new MockAdapter(apiKeysApi.getAxios());
  const apiKeysStore = useApiKeysStore();

  beforeEach(async () => {
    wrapper = mount(ApiKeyEdit, {
      global: {
        plugins: [vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        keyName: "fake-id",
        keyRole: "observer",
        hasAuthorization: true,
        disabled: false,
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
    expect(wrapper.find('[data-test="edit-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="edit-main-btn-title"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="edit-main-btn-title"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="key-name-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="edit-btn"]').exists()).toBe(true);
  });

  it("Successfully Edit Api Key", async () => {
    mockApiKeysApi.onPatch("http://localhost:3000/api/namespaces/api-key/fake-id").reply(200);

    const storeSpy = vi.spyOn(apiKeysStore, "editApiKey");

    await wrapper.findComponent('[data-test="edit-main-btn-title"]').trigger("click");
    await wrapper.findComponent('[data-test="key-name-text"]').setValue("fake-key-changed-name");

    await wrapper.findComponent('[data-test="edit-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith({
      key: "fake-id",
      name: "fake-key-changed-name",
      role: "observer",
    });
  });

  it("Fails to Edit Api Key", async () => {
    mockApiKeysApi.onPatch("http://localhost:3000/api/namespaces/api-key/fake-id").reply(400);

    await wrapper.findComponent('[data-test="edit-main-btn-title"]').trigger("click");

    await wrapper.findComponent('[data-test="edit-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to edit Api Key.");
  });

  it("Fails to Edit Api Key (409)", async () => {
    mockApiKeysApi.onPatch("http://localhost:3000/api/namespaces/api-key/fake-id").reply(409);

    await wrapper.findComponent('[data-test="edit-main-btn-title"]').trigger("click");

    await wrapper.findComponent('[data-test="key-name-text"]').setValue("fake-key");

    await wrapper.findComponent('[data-test="edit-btn"]').trigger("click");

    await flushPromises();

    expect(wrapper.vm.keyNameError).toBe("An API key with the same name already exists.");
  });
});
