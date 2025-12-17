import { createVuetify } from "vuetify";
import { flushPromises, mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { IAdminSession } from "@admin/interfaces/ISession";
import useSessionsStore from "@admin/store/modules/sessions";
import routes from "@admin/router";
import SessionDetails from "@admin/views/SessionDetails.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

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
  position: { latitude: 0, longitude: 0 },
} as IAdminSession;

const mockRoute = { params: { id: sessionDetail.uid } };

describe("Session Details", async () => {
  const pinia = createPinia();
  setActivePinia(pinia);

  const sessionsStore = useSessionsStore();
  sessionsStore.fetchSessionById = vi.fn().mockResolvedValue(sessionDetail);

  const vuetify = createVuetify();

  const wrapper = mount(SessionDetails, {
    global: {
      plugins: [pinia, vuetify, routes, SnackbarPlugin],
      mocks: {
        $route: mockRoute,
      },
    },
  });

  await flushPromises();

  it("Displays the session UID in the card title", () => {
    expect(wrapper.find(".text-h6").text()).toBe(sessionDetail.uid);
  });

  it("Displays session UID", () => {
    const uidField = wrapper.find('[data-test="session-uid-field"]');
    expect(uidField.text()).toContain("UID:");
    expect(uidField.text()).toContain(sessionDetail.uid);
  });

  it("Displays device with router link", () => {
    const deviceField = wrapper.find('[data-test="session-device-field"]');
    expect(deviceField.text()).toContain("Device:");
    expect(deviceField.find("a").exists()).toBe(true);
  });

  it("Displays username", () => {
    const usernameField = wrapper.find('[data-test="session-username-field"]');
    expect(usernameField.text()).toContain("Username:");
    expect(usernameField.text()).toContain(sessionDetail.username);
  });

  it("Displays IP address", () => {
    const ipField = wrapper.find('[data-test="session-ip-field"]');
    expect(ipField.text()).toContain("IP Address:");
    expect(ipField.text()).toContain(sessionDetail.ip_address);
  });

  it("Displays session type", () => {
    const typeField = wrapper.find('[data-test="session-type-field"]');
    expect(typeField.text()).toContain("Type:");
    expect(typeField.text()).toContain(sessionDetail.type);
  });

  it("Displays namespace with router link", () => {
    const namespaceField = wrapper.find('[data-test="session-namespace-field"]');
    expect(namespaceField.text()).toContain("Namespace:");
    expect(namespaceField.find("a").exists()).toBe(true);
  });

  it("Displays authenticated status as Yes/No", () => {
    const authField = wrapper.find('[data-test="session-authenticated-field"]');
    expect(authField.text()).toContain("Authenticated:");
    expect(authField.text()).toContain("Yes");
  });

  it("Displays recorded status as Yes/No", () => {
    const recordedField = wrapper.find('[data-test="session-recorded-field"]');
    expect(recordedField.text()).toContain("Recorded:");
    expect(recordedField.text()).toContain("Yes");
  });

  it("Displays terminal", () => {
    const termField = wrapper.find('[data-test="session-terminal-field"]');
    expect(termField.text()).toContain("Terminal:");
    expect(termField.text()).toContain(sessionDetail.term);
  });

  it("Displays started at date", () => {
    const startedField = wrapper.find('[data-test="session-started-field"]');
    expect(startedField.text()).toContain("Started At:");
  });

  it("Displays last seen date", () => {
    const lastSeenField = wrapper.find('[data-test="session-last-seen-field"]');
    expect(lastSeenField.text()).toContain("Last Seen:");
  });

  it("Shows error message when session data is empty", async () => {
    sessionsStore.fetchSessionById = vi.fn().mockResolvedValue({});
    const errorWrapper = mount(SessionDetails, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
        mocks: { $route: mockRoute },
      },
    });
    await flushPromises();
    expect(errorWrapper.text()).toContain("Something is wrong, try again!");
  });
});
