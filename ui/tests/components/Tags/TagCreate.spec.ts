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

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

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
    tags: ["test"],
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
    tags: ["test2"],
  },
];

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Tag Form Create", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TagCreate>>;
  setActivePinia(createPinia());

  const tagsStore = useTagsStore();
  const vuetify = createVuetify();

  let mockTags: MockAdapter;

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);

    localStorage.setItem("tenant", "fake-tenant-data");

    mockTags = new MockAdapter(tagsApi.getAxios());

    mockTags.onGet("http://localhost:3000/api/devices?filter=&page=1&per_page=10&status=accepted").reply(200, devices);

    wrapper = mount(TagCreate, {
      global: {
        plugins: [vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
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

    mockTags.onPost("http://localhost:3000/api/namespaces/fake-tenant-data/tags").reply(200);

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

    mockTags.onPost("http://localhost:3000/api/namespaces/fake-tenant-data/tags").reply(409);

    await wrapper.findComponent('[data-test="tag-field"]').setValue("");

    await wrapper.findComponent('[data-test="create-btn"]').trigger("click");

    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to create tag.");
  });
});
