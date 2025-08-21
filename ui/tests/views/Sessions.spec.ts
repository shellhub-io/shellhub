import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import Sessions from "@/views/Sessions.vue";
import { sessionsApi } from "@/api/http";
import { store, key } from "@/store";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";

type SessionsWrapper = VueWrapper<InstanceType<typeof Sessions>>;

describe("Sessions View", () => {
  let wrapper: SessionsWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  const mockSessionsApi = new MockAdapter(sessionsApi.getAxios());

  const mockSession = {
    uid: "1",
    device_uid: "1",
    device: {
      uid: "1",
      name: "00-00-00-00-00-01",
      identity: {
        mac: "00-00-00-00-00-01",
      },
      info: {
        id: "manjaro",
        pretty_name: "Manjaro Linux",
        version: "latest",
        arch: "amd64",
        platform: "docker",
      },
      public_key: "",
      tenant_id: "fake-tenant-data",
      last_seen: "0",
      online: true,
      namespace: "dev",
      status: "accepted",
      status_updated_at: "0",
      created_at: "0",
      remote_addr: "192.168.0.1",
      position: { latitude: 0, longitude: 0 },
      tags: [],
      public_url: false,
      public_url_address: "",
      acceptable: false,
    },
    tenant_id: "fake-tenant-data",
    username: "test",
    ip_address: "192.168.0.1",
    started_at: "",
    last_seen: "",
    active: false,
    authenticated: true,
    recorded: true,
    type: "none",
    term: "none",
    position: { longitude: 0, latitude: 0 },

  };

  beforeEach(async () => {
    mockSessionsApi.onGet("http://localhost:3000/api/sessions?page=1&per_page=10").reply(200, [mockSession], { "x-total-count": "1" });

    wrapper = mount(Sessions, {
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

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="sessions-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sessions-list"]').exists()).toBe(true);
  });

  it("Renders the SessionList component", () => {
    expect(wrapper.findComponent({ name: "SessionList" }).exists()).toBe(true);
  });

  it("Shows the no items message when there are no sessions", async () => {
    mockSessionsApi.onGet("http://localhost:3000/api/sessions?page=1&per_page=10").reply(200, [], { "x-total-count": "0" });
    wrapper.unmount();
    wrapper = mount(Sessions, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
    await flushPromises();
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-items-message-component"]').text()).toContain("Looks like you don't have any Sessions");
  });
});
