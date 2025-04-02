import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import SessionDetails from "../../../../src/views/SessionDetails.vue";

type SessionDetailsWrapper = VueWrapper<InstanceType<typeof SessionDetails>>;

const sessionDetail = {
  uid: "df33c82dcc7b401b0e4511fd9e0a86a48c5875da6091e89cf37cbbb38819e17e",
  device_uid: "cb1533e2e683aec21aee89b24ac4604b1a1955930362d33fb22e4e03fac52c75",
  device: {
    uid: "cb1533e2e683aec21aee89b24ac4604b1a1955930362d33fb22e4e03fac52c75",
    name: "08-97-98-68-7a-97",
    identity: {
      mac: "08:97:98:68:7a:97",
    },
    info: {
      id: "ubuntu",
      pretty_name: "Ubuntu 20.04.4 LTS",
      version: "latest",
      arch: "amd64",
      platform: "docker",
    },
    public_key: `-----BEGIN RSA PUBLIC KEY-----
      MIIBCgKCAQEAx2C95p3s9OpwHdSwV8xP5dS39jGBCM+VMChiqJViaaVoJJ2tTK/i
      zCEH6+jAuKSfvXjM3jql59RD0o7lFqd9bixiGN8/KvXZ/6hlDrdKniatIGHmGw6z
      N9EfKbTqJh0vHX/yRzRWlfAlLHoWjg0lV+Y6RpAiV1u6Gd4ZnDyz62u82fpQYqLu
      IFrhfOP52qbVZHMT6Vn/q8U26wysrDlVbF1k8RDR79Ib9i1Bu3mBPn0r5AEJOpQQ
      NqODe3Wjntgy8i0/iFaUV+9K17u50Pmm4uPfVfMEPmZSXpAwfpgWFPFInA9hLefq
      9XLjOj93MwVWN4iXLbLOoLI/9MQw5zZSYQIDAQAB
      -----END RSA PUBLIC KEY-----
    `,
    tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    last_seen: "2022-06-06T19:14:54.051Z",
    online: true,
    namespace: "dev",
    status: "accepted",
    created_at: "2022-04-13T11:43:25.218Z",
    remote_addr: "172.22.0.1",
    position: {
      latitude: 0,
      longitude: 0,
    },
    tags: ["dev"],
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
};

const mockRoute = {
  params: {
    id: sessionDetail.uid,
  },
};

describe("Session Details", () => {
  const store = createStore({
    state: {
      device: sessionDetail,
    },
    getters: {
      "sessions/session": () => sessionDetail,
    },
    actions: {
      "sessions/get": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });
  let wrapper: SessionDetailsWrapper;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(SessionDetails, {
      global: {
        plugins: [[store, key], vuetify, routes],
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

  it("Has the correct data", async () => {
    expect(wrapper.vm.session).toEqual(sessionDetail);
  });

  it("Render the correct title", () => {
    expect(wrapper.find("h1").text()).toEqual("Session Details");
  });

  it("Should render the props of the Session in the Screen", () => {
    expect(wrapper.find(`[data-test='${sessionDetail.uid}']`).text()).toContain(sessionDetail.uid);
    expect(wrapper.find(`[data-test='${sessionDetail.device_uid}']`).text()).toContain(sessionDetail.device_uid);
    expect(wrapper.find(`[data-test='${sessionDetail.tenant_id}']`).text()).toContain(sessionDetail.tenant_id);
    expect(wrapper.find(`[data-test='${sessionDetail.username}']`).text()).toContain(sessionDetail.username);
    expect(wrapper.find(`[data-test='${sessionDetail.ip_address}']`).text()).toContain(sessionDetail.ip_address);
    expect(wrapper.find(`[data-test='${sessionDetail.last_seen}']`).text()).toContain(sessionDetail.last_seen);
    expect(wrapper.find(`[data-test='${sessionDetail.active}']`).text()).toContain(sessionDetail.active);
    expect(wrapper.find(`[data-test='${sessionDetail.term}']`).text()).toContain(sessionDetail.term);
    expect(wrapper.find(`[data-test='${sessionDetail.type}']`).text()).toContain(sessionDetail.type);
  });
});
