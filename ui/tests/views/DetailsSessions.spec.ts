import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { createRouter, createWebHistory } from "vue-router";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import DetailsSessions from "@/views/DetailsSessions.vue";
import { sessionsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { routes } from "@/router";

type DetailsSessionsWrapper = VueWrapper<InstanceType<typeof DetailsSessions>>;

describe("Details Sessions", () => {
  let wrapper: DetailsSessionsWrapper;
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
      last_seen: "2025-01-02T00:00:00.000Z",
      online: true,
      namespace: "dev",
      status: "accepted",
      status_updated_at: "0",
      created_at: "2025-01-01T00:00:00.000Z",
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
    started_at: "2025-01-02T00:00:00.000Z",
    last_seen: "2025-01-02T00:00:00.000Z",
    active: false,
    authenticated: true,
    recorded: true,
    type: "none",
    term: "none",
    position: { longitude: 0, latitude: 0 },
  };

  const router = createRouter({
    history: createWebHistory(),
    routes,
  });

  beforeEach(async () => {
    router.push("/sessions/1");
    await router.isReady();

    mockSessionsApi.onGet("http://localhost:3000/api/sessions/1").reply(200, mockSession);

    wrapper = mount(DetailsSessions, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
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
    expect(wrapper.find('[data-test="session-uid-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="session-user-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="session-authenticated-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="session-ip-address-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="session-started-at-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="session-last-seen-field"]').exists()).toBe(true);
  });
});
