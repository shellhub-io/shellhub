import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import TagSelector from "@/components/Tags/TagSelector.vue";
import { router } from "@/router";
import { devicesApi, tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

const stats = {
  registered_devices: 3,
  online_devices: 1,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

const tags = [
  { name: "tag1" },
  { name: "tag2" },
  { name: "tag3" },
  { name: "tag4" },
  { name: "tag5" },
  { name: "tag6" },
  { name: "tag7" },
  { name: "tag8" },
  { name: "tag9" },
  { name: "tag10" },
  { name: "tag11" },
];

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
    tags: ["test"],
  },
];

describe("Tag Selector", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TagSelector>>;

  setActivePinia(createPinia());

  const vuetify = createVuetify();

  let mockDevices: MockAdapter;
  let mockTags: MockAdapter;

  beforeEach(async () => {
    vi.useRealTimers();
    localStorage.setItem("tenant", "fake-tenant-data");

    mockDevices = new MockAdapter(devicesApi.getAxios());
    mockTags = new MockAdapter(tagsApi.getAxios());

    mockTags
      .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
      .reply(200, tags);
    mockDevices.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, devices);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);

    wrapper = mount(TagSelector, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
      props: {
        variant: "device",
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

  it("Renders components", async () => {
    expect(wrapper.findComponent('[data-test="tags-btn"]').exists()).toBe(true);
  });

  it("Succesfully load tags", async () => {
    mockTags
      .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
      .reply(200, tags);

    await wrapper.findComponent('[data-test="tags-btn"]').trigger("click");

    await flushPromises();

    expect(wrapper.vm.fetchedTags).toEqual(tags);
  });
});
