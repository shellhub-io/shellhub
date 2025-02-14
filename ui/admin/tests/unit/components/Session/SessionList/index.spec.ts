import { createVuetify } from "vuetify";
import { createStore } from "vuex";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SessionList from "../../../../../src/components/Sessions/SessionList.vue";
import { key } from "../../../../../src/store";
import routes from "../../../../../src/router";

const headers = [
  { text: "Active", value: "active" },
  { text: "Id", value: "uid" },
  { text: "Device", value: "device" },
  { text: "Username", value: "username" },
  { text: "Authenticated", value: "authenticated" },
  { text: "IP Address", value: "ip_address" },
  { text: "Started", value: "started_at" },
  { text: "Last Seen", value: "last_seen" },
  { text: "Actions", value: "actions" },
];

const sessions = [
  {
    uid: "8c354a00f50",
    device_uid: "a582b47a42d",
    device: {
      uid: "a582b47a42d",
      name: "39-5e-2a",
      identity: {
        mac: "00:00:00:00:00:00",
      },
      info: {
        id: "debian",
        pretty_name: "Debian GNU/Linux 10 (buster)",
        version: "v0.2.5",
      },
      public_key: "----- PUBLIC KEY -----",
      tenant_id: "00000000",
      last_seen: "2020-05-18T13:27:02.498Z",
      online: false,
      namespace: "user",
    },
    tenant_id: "00000000",
    username: "user",
    ip_address: "000.000.000.000",
    started_at: "2020-05-18T12:30:28.824Z",
    last_seen: "2020-05-18T12:30:30.205Z",
    active: false,
    authenticated: false,
  },
  {
    uid: "8c354a00f51",
    device_uid: "a582b47a42d",
    device: {
      uid: "a582b47a42d",
      name: "b4-2e-99",
      identity: {
        mac: "00:00:00:00:00:00",
      },
      info: {
        id: "debian",
        pretty_name: "Debian GNU/Linux 10 (buster)",
        version: "v0.2.5",
      },
      public_key: "----- PUBLIC KEY -----",
      tenant_id: "00000000",
      last_seen: "2020-05-18T13:27:02.498Z",
      online: false,
      namespace: "user",
    },
    tenant_id: "00000000",
    username: "user",
    ip_address: "000.000.000.000",
    started_at: "2020-05-18T12:30:28.824Z",
    last_seen: "2020-05-18T12:30:30.205Z",
    active: false,
    authenticated: false,
  },
];

const store = createStore({
  state: {
    sessions,
  },
  getters: {
    "sessions/sessions": (state) => state.sessions,
    "sessions/numberSessions": (state) => state.sessions.length,
  },
  actions: {
    "sessions/fetch": () => vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("Sessions List", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(SessionList, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    const dt = wrapper.find("[data-test]");
    expect(dt.attributes()["data-test"]).toBe("session-list");
    expect(wrapper.vm.headers).toEqual(headers);
    expect(wrapper.vm.sessions).toEqual(sessions);
    expect(wrapper.vm.loading).toEqual(false);
    expect(wrapper.vm.itemsPerPage).toEqual(10);
  });
});
