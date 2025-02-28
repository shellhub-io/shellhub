import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import TagRemove from "@/components/Tags/TagRemove.vue";
import { router } from "@/router";
import { namespacesApi, devicesApi, tagsApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

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

const members = [
  {
    id: "xxxxxxxx",
    username: "test",
    role: "owner",
  },
];

const namespaceData = {
  name: "test",
  owner: "xxxxxxxx",
  tenant_id: "fake-tenant-data",
  members,
  max_devices: 3,
  devices_count: 3,
  devices: 2,
  created_at: "",
};

const stats = {
  registered_devices: 3,
  online_devices: 1,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Tag Remove", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TagRemove>>;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;
  let mockDevices: MockAdapter;
  let mockTags: MockAdapter;

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    setActivePinia(createPinia());

    localStorage.setItem("tenant", "fake-tenant-data");

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());
    mockTags = new MockAdapter(tagsApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockDevices.onGet("http://localhost:3000/api/devices?filter=&page=1&per_page=10&status=accepted").reply(200, devices);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);

    wrapper = mount(TagRemove, {
      global: {
        plugins: [vuetify, router],
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

    mockTags.onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/tags/tag-test").reply(200);

    const tagsSpy = vi.spyOn(tagsApi, "deleteTag");

    await wrapper.findComponent('[data-test="open-tag-remove"]').trigger("click");

    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");

    await flushPromises();

    expect(tagsSpy).toHaveBeenCalledWith("fake-tenant-data", "tag-test");
  });

  it("Failed to remove tags", async () => {
    mockTags.onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/tags/tag-test").reply(409);

    await wrapper.findComponent('[data-test="open-tag-remove"]').trigger("click");

    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to remove tag.");
  });
});
