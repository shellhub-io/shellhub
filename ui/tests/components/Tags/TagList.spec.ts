import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import TagList from "@/components/Tags/TagList.vue";
import { router } from "@/router";
import { namespacesApi, devicesApi, tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

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

const tags = [{ name: "123x" }, { name: "newtag" }];

describe("Tag List", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TagList>>;

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
    mockTags
      .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
      .reply(200, tags);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);

    wrapper = mount(TagList, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
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
    await flushPromises();
    expect(wrapper.find('[data-test="tag-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="tag-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="tag-list-actions"]').exists()).toBe(true);
  });
});
