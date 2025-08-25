import { setActivePinia, createPinia } from "pinia";
import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import TagEdit from "@/components/Tags/TagEdit.vue";
import { router } from "@/router";
import { tagsApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useTagsStore from "@/store/modules/tags";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Tag Form Edit", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TagEdit>>;
  setActivePinia(createPinia());
  const tagsStore = useTagsStore();
  const vuetify = createVuetify();
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());

  beforeEach(async () => {
    wrapper = mount(TagEdit, {
      global: {
        plugins: [[store, key], vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        tag: "tag-test",
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

  it("Renders the component table", async () => {
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    await wrapper.findComponent('[data-test="open-tag-edit"]').trigger("click");

    expect(wrapper.find('[data-test="mdi-information-list-item"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="tag-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="edit-btn"]').exists()).toBe(true);
  });

  it("Successfully edit tag", async () => {
    mockTagsApi.onPut("http://localhost:3000/api/tags/tag-test").reply(200);

    const storeSpy = vi.spyOn(tagsStore, "updateTag");

    await wrapper.findComponent('[data-test="open-tag-edit"]').trigger("click");

    await wrapper.findComponent('[data-test="tag-field"]').setValue("tag-test2");

    await wrapper.findComponent('[data-test="edit-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      oldTag: "tag-test",
      newTag: "tag-test2",
    });
  });

  it("Failed to edit tags", async () => {
    mockTagsApi.onPut("http://localhost:3000/api/tags/tag-test").reply(409);

    await wrapper.findComponent('[data-test="open-tag-edit"]').trigger("click");

    await wrapper.findComponent('[data-test="edit-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update tag.");
  });
});
