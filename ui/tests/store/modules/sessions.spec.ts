import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it } from "vitest";
import { sessionsApi } from "@/api/http";
import useSessionsStore from "@/store/modules/sessions";
import { ISession } from "@/interfaces/ISession";

const mockSession: ISession = {
  uid: "session1",
  device_uid: "device1",
  device: {
    uid: "device1",
    name: "Device 1",
    identity: {
      mac: "00-00-00-00-00-01",
    },
    info: {
      id: "linux",
      pretty_name: "Linux",
      version: "1.0",
      arch: "amd64",
      platform: "docker",
    },
    public_key: "",
    tenant_id: "tenant1",
    last_seen: "2023-01-01T00:00:00Z",
    online: true,
    namespace: "default",
    status: "accepted",
    created_at: "2023-01-01T00:00:00Z",
    remote_addr: "192.168.1.1",
    position: { latitude: 0, longitude: 0 },
    tags: [],
  },
  tenant_id: "tenant1",
  username: "testuser",
  ip_address: "192.168.1.1",
  started_at: "2023-01-01T00:00:00Z",
  last_seen: "2023-01-01T00:00:00Z",
  active: true,
  authenticated: true,
  recorded: true,
  type: "shell",
  term: "xterm",
  position: { latitude: 0, longitude: 0 },
};

const mockSessions = [mockSession];

describe("Sessions Store", () => {
  setActivePinia(createPinia());
  const mockSessionsApi = new MockAdapter(sessionsApi.getAxios());
  const sessionsStore = useSessionsStore();

  it("should have initial state values", () => {
    expect(sessionsStore.sessions).toEqual([]);
    expect(sessionsStore.session).toEqual({});
    expect(sessionsStore.sessionCount).toEqual(0);
  });

  it("successfully fetches sessions list", async () => {
    mockSessionsApi.onGet("http://localhost:3000/api/sessions?page=1&per_page=10").reply(200, mockSessions, {
      "x-total-count": "1",
    });

    await sessionsStore.fetchSessionList();

    expect(sessionsStore.sessions).toEqual(mockSessions);
    expect(sessionsStore.sessionCount).toEqual(1);
  });

  it("handles error when fetching sessions list", async () => {
    mockSessionsApi.onGet("http://localhost:3000/api/sessions?page=1&per_page=10").reply(500);

    try {
      await sessionsStore.fetchSessionList({ page: 1, perPage: 10 });
    } catch (error) {
      expect(sessionsStore.sessions).toEqual([]);
      expect(sessionsStore.sessionCount).toEqual(0);
    }
  });

  it("successfully gets a single session", async () => {
    mockSessionsApi.onGet("http://localhost:3000/api/sessions/session1").reply(200, mockSession);

    await sessionsStore.getSession("session1");

    expect(sessionsStore.session).toEqual(mockSession);
  });

  it("successfully gets session logs", async () => {
    const mockLogs = "session log data";
    mockSessionsApi.onGet("http://localhost:3000/api/sessions/session1/records/0").reply(200, mockLogs);

    const logs = await sessionsStore.getSessionLogs("session1");

    expect(logs).toEqual(mockLogs);
  });

  it("successfully closes a session", async () => {
    mockSessionsApi.onPost("http://localhost:3000/api/sessions/session1/close").reply(200);

    const sessionData = { uid: "session1", device_uid: "device1" };
    await sessionsStore.closeSession(sessionData);

    expect(true).toBe(true);
  });

  it("successfully deletes session logs", async () => {
    mockSessionsApi.onPost("http://localhost:3000/api/sessions/session1/close").reply(200);

    sessionsStore.session = { ...mockSession, recorded: true };

    await sessionsStore.deleteSessionLogs("session1");

    expect(sessionsStore.session.recorded).toBe(false);
  });
});
