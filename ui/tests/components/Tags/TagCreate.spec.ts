import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import TagCreate from "@/components/Tags/TagCreate.vue";
import { router } from "@/router";
import { tagsApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useTagsStore from "@/store/modules/tags";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Tag Form Create", () => {
  let wrapper: VueWrapper<InstanceType<typeof TagCreate>>;
  setActivePinia(createPinia());
  const tagsStore = useTagsStore();
  const vuetify = createVuetify();
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  localStorage.setItem("tenant", "fake-tenant-data");

  beforeEach(() => {
    wrapper = mount(TagCreate, {
      global: {
        plugins: [vuetify, router],
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

  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Renders the component table", async () => {
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.showDialog = true;
    await flushPromises();
    expect(wrapper.findComponent('[data-test="tag-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="create-btn"]').exists()).toBe(true);
  });

  it("Successfully create tag", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();

    mockTagsApi.onPost("http://localhost:3000/api/namespaces/fake-tenant-data/tags").reply(200);

    const tagsSpy = vi.spyOn(tagsStore, "createTag");

    await wrapper.findComponent('[data-test="tag-field"]').setValue("tag-test2");

    await wrapper.findComponent('[data-test="create-btn"]').trigger("click");

    await flushPromises();

    expect(tagsSpy).toHaveBeenCalledWith({
      tenant: "fake-tenant-data",
      name: "tag-test2",
    });
  });

  it("Fails to create tag", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();

    mockTagsApi
      .onPost("http://localhost:3000/api/namespaces/fake-tenant-data/tags")
      .reply(409);

    await wrapper.findComponent('[data-test="tag-field"]').setValue("duplicate-tag");

    await wrapper.findComponent('[data-test="create-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to create tag.");
  });
});
