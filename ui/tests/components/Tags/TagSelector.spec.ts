import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import TagSelector from "@/components/Tags/TagSelector.vue";
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

describe("Tag Selector", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TagSelector>>;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;
  let mockDevices: MockAdapter;
  let mockTags: MockAdapter;

  beforeEach(async () => {
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

    wrapper = mount(TagSelector, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
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
    expect(wrapper.find('[data-test="tags-btn"]').exists()).toBe(true);
    wrapper.vm.menuOpen = true;
    await flushPromises();
    expect(wrapper.findComponent('[data-test="tag-item"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="load-more-tags-btn"]').exists()).toBe(true);
  });

  it("Succesfully load more tags", async () => {
    mockTags
      .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?page=1&per_page=10")
      .reply(200, tags);

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="tags-btn"]').trigger("click");

    await flushPromises();

    await wrapper.vm.loadMoreTags();

    await flushPromises();

    expect(StoreSpy).toHaveBeenCalledWith("tags/autocomplete", {
      page: 2,
      perPage: 10,
      tenant: "fake-tenant-data",
    });

    expect(wrapper.vm.fetchedTags).toEqual(tags);
  });
});
