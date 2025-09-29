import { flushPromises, DOMWrapper, mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, vi, beforeEach } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import TagRemove from "@/components/Tags/TagRemove.vue";
import { router } from "@/router";
import { tagsApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useTagsStore from "@/store/modules/tags";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Tag Remove", () => {
  let wrapper: ReturnType<typeof mount>;
  let mockTagsApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    mockTagsApi = new MockAdapter(tagsApi.getAxios());
    localStorage.setItem("tenant", "fake-tenant-data");

    wrapper = mount(TagRemove, {
      global: {
        plugins: [createVuetify(), router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        tagName: "tag-test",
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

  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Renders dialog and controls", async () => {
    const body = new DOMWrapper(document.body);
    await flushPromises();

    await wrapper.find('[data-test="open-tag-remove"]').trigger("click");
    await flushPromises();

    expect(wrapper.find('[data-test="mdi-information-list-item"]').exists()).toBe(true);
    // New MessageDialog selectors
    expect(body.find('[data-test="delete-tag-dialog"]').exists()).toBe(true);
    expect(body.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(body.find('[data-test="confirm-btn"]').exists()).toBe(true);
  });

  it("Successfully removes tag", async () => {
    mockTagsApi
      .onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/tags/tag-test")
      .reply(200);

    const tagsStore = useTagsStore();
    const storeSpy = vi.spyOn(tagsStore, "removeTag");

    await wrapper.find('[data-test="open-tag-remove"]').trigger("click");
    await flushPromises();

    const messageDialogStub = wrapper.findComponent({ name: "MessageDialog" });
    await messageDialogStub.vm.$emit("confirm");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      tenant: "fake-tenant-data",
      currentName: "tag-test",
    });

    expect(mockSnackbar.showSuccess).toHaveBeenCalled();
  });

  it("Shows error snackbar on failure", async () => {
    mockTagsApi
      .onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/tags/tag-test")
      .reply(409);

    await wrapper.find('[data-test="open-tag-remove"]').trigger("click");
    await flushPromises();

    await new DOMWrapper(document.body).find('[data-test="confirm-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to remove tag.");
  });
});
