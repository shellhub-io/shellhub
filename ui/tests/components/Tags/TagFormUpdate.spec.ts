import { createPinia, setActivePinia } from "pinia";
import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import TagFormUpdate from "@/components/Tags/TagFormUpdate.vue";
import { tagsApi } from "@/api/http";
import useTagsStore from "@/store/modules/tags";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

const devices = [
  {
    uid: "a582b47a42d",
    name: "39-5e-2a",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T18:58:53.276Z",
    online: false,
    namespace: "user",
    status: "accepted",
    tags: [{ name: "test1" }],
  },
  {
    uid: "a582b47a42e",
    name: "39-5e-2b",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T19:58:53.276Z",
    online: true,
    namespace: "user",
    status: "accepted",
    tags: [{ name: "test2" }],
  },
];

const tags = [
  { name: "tag1" },
  { name: "tag2" },
  { name: "tag3" },
];

describe("Tag Form Update", () => {
  let wrapper: VueWrapper<InstanceType<typeof TagFormUpdate>>;
  setActivePinia(createPinia());
  const tagsStore = useTagsStore();
  const vuetify = createVuetify();
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  mockTagsApi
    .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
    .reply(200, tags);
  localStorage.setItem("tenant", "fake-tenant-data");

  beforeEach(() => {
    wrapper = mount(TagFormUpdate, {
      attachTo: document.body,
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        deviceUid: devices[0].uid,
        tagsList: devices[0].tags,
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

  it("Renders the component data table", async () => {
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    await wrapper.findComponent('[data-test="open-tags-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="deviceTag-autocomplete"]').trigger("click");

    expect(wrapper.find('[data-test="has-tags-verification"]').exists()).toBe(true);

    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    expect(formDialog.exists()).toBe(true);
    expect(formDialog.props("title")).toBe("Edit tags");
    expect(formDialog.props("icon")).toBe("mdi-tag");
    expect(formDialog.props("confirmText")).toBe("");
    expect(formDialog.props("cancelText")).toBe("Close");
    expect(formDialog.props("confirmDisabled")).toBe(true);

    // Content inside the dialog
    expect(dialog.find('[data-test="deviceTag-autocomplete"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("Successfully add tags", async () => {
    mockTagsApi
      .onPost("http://localhost:3000/api/namespaces/fake-tenant-data/devices/a582b47a42d/tags/tag-test-1")
      .reply(200);

    const tagsSpy = vi.spyOn(tagsStore, "pushTagToDevice");

    await wrapper.findComponent('[data-test="open-tags-btn"]').trigger("click");
    await wrapper.vm.updateTags("tag-test-1");

    await flushPromises();

    expect(tagsSpy).toHaveBeenCalledWith({
      tenant: "fake-tenant-data",
      uid: "a582b47a42d",
      name: "tag-test-1",
    });
  });

  it("Successfully removes tags", async () => {
    mockTagsApi
      .onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/devices/a582b47a42d/tags/test1")
      .reply(200);

    const tagsSpy = vi.spyOn(tagsStore, "removeTagFromDevice");

    await wrapper.findComponent('[data-test="open-tags-btn"]').trigger("click");
    await wrapper.vm.updateTags("test1");

    await flushPromises();

    expect(tagsSpy).toHaveBeenCalledWith({
      tenant: "fake-tenant-data",
      uid: "a582b47a42d",
      name: "test1",
    });
  });

  it("Successfully loads more tags", async () => {
    mockTagsApi
      .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
      .reply(200, tags);

    const tagsSpy = vi.spyOn(tagsStore, "autocomplete");

    await wrapper.findComponent('[data-test="open-tags-btn"]').trigger("click");

    await flushPromises();

    await wrapper.vm.loadTags();

    await flushPromises();

    expect(tagsSpy).toHaveBeenCalledWith({
      filter: "",
      page: 1,
      perPage: 10,
      tenant: "fake-tenant-data",
    });
    expect(wrapper.vm.tags).toEqual(tags);
  });
});
