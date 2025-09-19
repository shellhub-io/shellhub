import { flushPromises, DOMWrapper, mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import TagRemove from "@/components/Tags/TagRemove.vue";
import { router } from "@/router";
import { tagsApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Tag Remove", async () => {
  setActivePinia(createPinia());
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  const vuetify = createVuetify();
  localStorage.setItem("tenant", "fake-tenant-data");

  const wrapper = mount(TagRemove, {
    global: {
      plugins: [vuetify, router],
      provide: { [SnackbarInjectionKey]: mockSnackbar },
    },
    props: {
      tagName: "tag-test",
      hasAuthorization: true,
    },
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
    await flushPromises();
    await wrapper.findComponent('[data-test="open-tag-remove"]').trigger("click");

    expect(wrapper.find('[data-test="mdi-information-list-item"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-card"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-subtitle"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="remove-btn"]').exists()).toBe(true);
  });

  it("Successfully remove tag", async () => {
    await flushPromises();

    mockTagsApi.onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/tags/tag-test").reply(200);

    const tagsSpy = vi.spyOn(tagsApi, "deleteTag");

    await wrapper.findComponent('[data-test="open-tag-remove"]').trigger("click");

    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");

    await flushPromises();

    expect(tagsSpy).toHaveBeenCalledWith("fake-tenant-data", "tag-test");
  });

  it("Failed to remove tags", async () => {
    mockTagsApi.onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/tags/tag-test").reply(409);

    await wrapper.findComponent('[data-test="open-tag-remove"]').trigger("click");

    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to remove tag.");
  });
});
