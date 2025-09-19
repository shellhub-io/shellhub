import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { IAdminSession } from "@admin/interfaces/ISession";
import useSessionsStore from "@admin/store/modules/sessions";
import routes from "@admin/router";
import SessionDetails from "@admin/views/SessionDetails.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type SessionDetailsWrapper = VueWrapper<InstanceType<typeof SessionDetails>>;

const sessionDetail = {
  uid: "df33c82dcc7b401b0e4511fd9e0a86a48c5875da6091e89cf37cbbb38819e17e",
  device_uid: "cb1533e2e683aec21aee89b24ac4604b1a1955930362d33fb22e4e03fac52c75",
  device: {
    uid: "cb1533e2e683aec21aee89b24ac4604b1a1955930362d33fb22e4e03fac52c75",
    name: "08-97-98-68-7a-97",
    identity: { mac: "08:97:98:68:7a:97" },
    info: {
      id: "ubuntu",
      pretty_name: "Ubuntu 20.04.4 LTS",
      version: "latest",
      arch: "amd64",
      platform: "docker",
    },
    public_key: "---KEY---",
    tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    last_seen: "2022-06-06T19:14:54.051Z",
    online: true,
    namespace: "dev",
    status: "accepted",
    created_at: "2022-04-13T11:43:25.218Z",
    remote_addr: "172.22.0.1",
    position: { latitude: 0, longitude: 0 },
    tags: [{ name: "dev" }],
  },
  tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  username: "antony",
  ip_address: "172.22.0.1",
  started_at: "2022-05-20T14:03:25.837Z",
  last_seen: "2022-05-20T14:03:35.701Z",
  active: false,
  authenticated: true,
  recorded: true,
  type: "web",
  term: "xterm.js",
} as IAdminSession;

const mockRoute = {
  params: {
    id: sessionDetail.uid,
  },
};

describe("Session Details", () => {
  let wrapper: SessionDetailsWrapper;
  const pinia = createPinia();
  setActivePinia(pinia);
  const sessionsStore = useSessionsStore();
  sessionsStore.fetchSessionById = vi.fn().mockResolvedValue(sessionDetail);
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(SessionDetails, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
        mocks: {
          $route: mockRoute,
        },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Render the correct title", () => {
    expect(wrapper.find("h1").text()).toEqual("Session Details");
  });

  it("Should render the props of the Session in the Screen", async () => {
    wrapper.vm.session = sessionDetail;

    await flushPromises();

    expect(wrapper.find(`[data-test='${sessionDetail.uid}']`).text()).toContain(sessionDetail.uid);
    expect(wrapper.find(`[data-test='${sessionDetail.device_uid}']`).text()).toContain(sessionDetail.device_uid);
    expect(wrapper.find(`[data-test='${sessionDetail.tenant_id}']`).text()).toContain(sessionDetail.tenant_id);
    expect(wrapper.find(`[data-test='${sessionDetail.username}']`).text()).toContain(sessionDetail.username);
    expect(wrapper.find(`[data-test='${sessionDetail.ip_address}']`).text()).toContain(sessionDetail.ip_address);
    expect(wrapper.find(`[data-test='${sessionDetail.last_seen}']`).text()).toContain(sessionDetail.last_seen);
    expect(wrapper.find(`[data-test='${sessionDetail.active}']`).text()).toContain(String(sessionDetail.active));
    expect(wrapper.find(`[data-test='${sessionDetail.term}']`).text()).toContain(sessionDetail.term);
    expect(wrapper.find(`[data-test='${sessionDetail.type}']`).text()).toContain(sessionDetail.type);
  });
});
