import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";
import useSessionsStore from "@admin/store/modules/sessions";
import { SnackbarPlugin } from "@/plugins/snackbar";
import SessionList from "../../../../../src/components/Sessions/SessionList.vue";
import routes from "../../../../../src/router";

type SessionListWrapper = VueWrapper<InstanceType<typeof SessionList>>;

const sessions = [
  {
    uid: "8c354a00f50",
    device_uid: "a582b47a42d",
    device: {
      uid: "a582b47a42d",
      name: "39-5e-2a",
      identity: { mac: "00:00:00:00:00:00" },
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
      identity: { mac: "00:00:00:00:00:00" },
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

describe("Sessions List", () => {
  let wrapper: SessionListWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());

    const vuetify = createVuetify();

    const sessionStore = useSessionsStore();

    sessionStore.sessions = sessions;
    sessionStore.numberSessions = sessions.length;

    vi.spyOn(sessionStore, "fetch").mockResolvedValue(false);

    wrapper = mount(SessionList, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with session data", () => {
    const dt = wrapper.find("[data-test='session-list']");
    expect(dt.exists()).toBe(true);

    const store = useSessionsStore();
    expect(store.getSessions).toEqual(sessions);
    expect(store.getNumberSessions).toBe(sessions.length);
  });
});
