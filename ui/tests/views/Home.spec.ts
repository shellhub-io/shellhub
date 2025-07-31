import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import Home from "@/views/Home.vue";
import { devicesApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type HomeWrapper = VueWrapper<InstanceType<typeof Home>>;

describe("Home", () => {
  let wrapper: HomeWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    settings: {
      session_record: true,
    },
    max_devices: 3,
    devices_count: 3,
    created_at: "",
  };

  const statsMock = {
    registered_devices: 0,
    online_devices: 0,
    active_sessions: 0,
    pending_devices: 0,
    rejected_devices: 0,
  };

  const res = {
    data: [namespaceData],
    headers: {
      "x-total-count": 1,
    },
  };

  beforeEach(async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/stats").reply(200, statsMock);
    store.commit("namespaces/setNamespaces", res);

    wrapper = mount(Home, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", async () => {
    expect(wrapper.find('[data-test="home-card"]').exists()).toBe(true);
    wrapper.vm.hasStatus = true; // Set the conditional validation to true so it can show the error card.
    await nextTick();
    expect(wrapper.find('[data-test="home-failed"]').exists()).toBe(true);
  });

  it("Displays error message if API call fails with 403 status code", async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/stats").reply(403);

    await flushPromises();

    expect(wrapper.find('[data-test="home-failed"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="home-failed"]').text()).toContain(
      "Something is wrong, try again !",
    );
  });
});
