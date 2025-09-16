import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";
import useSessionsStore from "@admin/store/modules/sessions";
import SessionList from "@admin/components/Sessions/SessionList.vue";
import routes from "@admin/router";
import { IAdminSession } from "@admin/interfaces/ISession";
import { SnackbarPlugin } from "@/plugins/snackbar";

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
  setActivePinia(createPinia());
  const sessionStore = useSessionsStore();
  const vuetify = createVuetify();

  sessionStore.sessions = sessions as IAdminSession[];
  sessionStore.fetchSessionList = vi.fn();
  sessionStore.sessionCount = sessions.length;

  beforeEach(() => {
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
    const sessionList = wrapper.find("[data-test='session-list']");
    expect(sessionList.exists()).toBe(true);
    expect(sessionStore.sessions).toEqual(sessions);
    expect(sessionStore.sessionCount).toBe(sessions.length);
  });
});
