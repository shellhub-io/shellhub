import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import TagFormUpdate from "@/components/Tags/TagFormUpdate.vue";
import { router } from "@/router";
import { namespacesApi, devicesApi, tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

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

const members = [
  {
    id: "xxxxxxxx",
    username: "test",
    role: "owner",
  },
];

const billingData = {
  active: false,
  status: "canceled",
  customer_id: "cus_test",
  subscription_id: "sub_test",
  current_period_end: 2068385820,
  created_at: "",
  updated_at: "",
  invoices: [],
};

const namespaceData = {
  name: "test",
  owner: "xxxxxxxx",
  tenant_id: "fake-tenant-data",
  members,
  max_devices: 3,
  devices_count: 3,
  devices: 2,
  created_at: "",
  billing: billingData,
};

const authData = {
  status: "",
  token: "",
  user: "test",
  name: "test",
  tenant: "fake-tenant-data",
  email: "test@test.com",
  id: "xxxxxxxx",
  role: "owner",
};

const customerData = {
  id: "cus_test",
  name: "test",
  email: "test@test.com",
  payment_methods: [
    {
      id: "test_id",
      number: "xxxxxxxxxxxx4242",
      brand: "visa",
      exp_month: 3,
      exp_year: 2029,
      cvc: "",
      default: true,
    },
  ],
};

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

describe("Tag Form Update", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TagFormUpdate>>;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;
  let mockDevices: MockAdapter;
  let mockTags: MockAdapter;

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);

    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());
    mockTags = new MockAdapter(tagsApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockTags
      .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
      .reply(200, tags);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("customer/setCustomer", customerData);
    store.commit("devices/setDeviceChooserStatus", true);

    wrapper = mount(TagFormUpdate, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      attachTo: el,
      props: {
        deviceUid: "a582b47a42d",
        tagsList: devices[0].tags,
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

  it("Renders the component data table", async () => {
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    await wrapper.findComponent('[data-test="open-tags-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="deviceTag-autocomplete"').trigger("click");
    expect(wrapper.find('[data-test="hastags-verification"]').exists()).toBe(true);
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="deviceTag-autocomplete"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("Successfully add tags", async () => {
    mockTags
      .onPost("http://localhost:3000/api/namespaces/fake-tenant-data/devices/a582b47a42d/tags/tag-test-1")
      .reply(200);

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="open-tags-btn"]').trigger("click");
    await wrapper.vm.updateTags("tag-test-1");

    await flushPromises();

    expect(StoreSpy).toHaveBeenCalledWith("tags/pushTagToDevice", {
      tenant: "fake-tenant-data",
      uid: "a582b47a42d",
      name: "tag-test-1",
    });
  });

  it("Succesfully remove tags", async () => {
    mockTags
      .onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/devices/a582b47a42d/tags/test1")
      .reply(200);

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="open-tags-btn"]').trigger("click");
    await wrapper.vm.updateTags("test1");

    await flushPromises();

    expect(StoreSpy).toHaveBeenCalledWith("tags/removeTagFromDevice", {
      tenant: "fake-tenant-data",
      uid: "a582b47a42d",
      name: "test1",
    });
  });

  it("Succesfully load more tags", async () => {
    mockTags
      .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
      .reply(200, tags.pop());

    mockTags
      .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=2&per_page=10")
      .reply(200, tags);

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="open-tags-btn"]').trigger("click");

    await flushPromises();

    await wrapper.vm.loadTags();

    await flushPromises();

    expect(StoreSpy).toHaveBeenCalledWith("tags/autocomplete", {
      filter: "",
      page: 1,
      perPage: 10,
      tenant: "fake-tenant-data",
    });
    expect(StoreSpy).toHaveBeenCalledWith("tags/autocomplete", {
      filter: "",
      page: 2,
      perPage: 10,
      tenant: "fake-tenant-data",
    });
    expect(wrapper.vm.fetchedTags).toEqual(tags);
  });

  it("Succesfully remove tags", async () => {
    mockTags
      .onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/devices/a582b47a42d/tags/test1")
      .reply(200);

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="open-tags-btn"]').trigger("click");
    expect(wrapper.vm.selectedTags).toEqual(["test1"]);
    await wrapper.vm.removeTag("test1");

    await flushPromises();

    expect(StoreSpy).toHaveBeenCalledWith("tags/removeTagFromDevice", {
      tenant: "fake-tenant-data",
      uid: "a582b47a42d",
      name: "test1",
    });
    expect(wrapper.vm.selectedTags).toEqual([]);
  });
});
