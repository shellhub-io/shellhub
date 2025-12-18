import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { sessionsApi } from "@/api/http";
import useSessionsStore from "@/store/modules/sessions";
import { ISession } from "@/interfaces/ISession";
import { buildUrl } from "../../utils/url";

const mockSessionBase: ISession = {
  uid: "session-uid-123",
  device_uid: "device-uid-456",
  device: {
    uid: "device-uid-456",
    name: "my-device",
    identity: {
      mac: "00:1A:2B:3C:4D:5E",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 20.3",
      version: "20.3",
      arch: "x86_64",
      platform: "docker",
    },
    public_key: "ssh-rsa AAAAB3NzaC1...",
    tenant_id: "tenant-id-789",
    last_seen: "2026-01-01T12:00:00.000Z",
    status_updated_at: "2026-01-01T12:00:00.000Z",
    online: true,
    namespace: "production",
    status: "accepted",
    created_at: "2026-01-01T00:00:00.000Z",
    remote_addr: "192.168.1.100",
    position: { latitude: 0, longitude: 0 },
    tags: [{
      name: "server",
      tenant_id: "tenant-id-789",
      created_at: "2026-01-01T00:00:00.000Z",
      updated_at: "2026-01-01T00:00:00.000Z",
    }, {
      name: "production",
      tenant_id: "tenant-id-789",
      created_at: "2026-01-01T00:00:00.000Z",
      updated_at: "2026-01-01T00:00:00.000Z",
    }],
  },
  tenant_id: "tenant-id-789",
  username: "admin",
  ip_address: "192.168.1.50",
  started_at: "2026-01-01T10:00:00.000Z",
  last_seen: "2026-01-01T12:00:00.000Z",
  active: true,
  authenticated: true,
  recorded: true,
  type: "shell",
  term: "xterm-256color",
  position: { latitude: 0, longitude: 0 },
};

describe("Sessions Store", () => {
  let sessionsStore: ReturnType<typeof useSessionsStore>;
  let mockSessionsApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    sessionsStore = useSessionsStore();
    mockSessionsApi = new MockAdapter(sessionsApi.getAxios());
  });

  afterEach(() => { mockSessionsApi.reset(); });

  describe("Initial State", () => {
    it("should have empty sessions array", () => {
      expect(sessionsStore.sessions).toEqual([]);
    });

    it("should have empty session object", () => {
      expect(sessionsStore.session).toEqual({});
    });

    it("should have zero session count", () => {
      expect(sessionsStore.sessionCount).toBe(0);
    });
  });

  describe("fetchSessionList", () => {
    const baseUrl = "http://localhost:3000/api/sessions";

    it("should fetch sessions list successfully with default pagination", async () => {
      const sessionList = [mockSessionBase];

      mockSessionsApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(200, sessionList, { "x-total-count": "1" });

      await expect(sessionsStore.fetchSessionList()).resolves.not.toThrow();

      expect(sessionsStore.sessions).toEqual(sessionList);
      expect(sessionsStore.sessionCount).toBe(1);
    });

    it("should fetch sessions list successfully with custom pagination", async () => {
      const sessionList = [mockSessionBase];

      mockSessionsApi
        .onGet(buildUrl(baseUrl, { page: "2", per_page: "20" }))
        .reply(200, sessionList, { "x-total-count": "15" });

      await expect(sessionsStore.fetchSessionList({ page: 2, perPage: 20 })).resolves.not.toThrow();

      expect(sessionsStore.sessions).toEqual(sessionList);
      expect(sessionsStore.sessionCount).toBe(15);
    });

    it("should fetch empty sessions list successfully", async () => {
      mockSessionsApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(200, [], { "x-total-count": "0" });

      await expect(sessionsStore.fetchSessionList()).resolves.not.toThrow();

      expect(sessionsStore.sessions).toEqual([]);
      expect(sessionsStore.sessionCount).toBe(0);
    });

    it("should reset state and throw on not found error when fetching sessions list", async () => {
      mockSessionsApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(404, { message: "Sessions not found" });

      await expect(sessionsStore.fetchSessionList()).rejects.toBeAxiosErrorWithStatus(404);

      expect(sessionsStore.sessions).toEqual([]);
      expect(sessionsStore.sessionCount).toBe(0);
    });

    it("should reset state and throw on server error when fetching sessions list", async () => {
      mockSessionsApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(500);

      await expect(sessionsStore.fetchSessionList()).rejects.toBeAxiosErrorWithStatus(500);

      expect(sessionsStore.sessions).toEqual([]);
      expect(sessionsStore.sessionCount).toBe(0);
    });

    it("should reset state and throw on network error when fetching sessions list", async () => {
      mockSessionsApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .networkError();

      await expect(sessionsStore.fetchSessionList()).rejects.toThrow("Network Error");

      expect(sessionsStore.sessions).toEqual([]);
      expect(sessionsStore.sessionCount).toBe(0);
    });
  });

  describe("getSession", () => {
    const baseGetSessionUrl = (sessionUid: string) => `http://localhost:3000/api/sessions/${sessionUid}`;

    it("should get session successfully and update state", async () => {
      const sessionUid = "session-uid-123";

      mockSessionsApi
        .onGet(baseGetSessionUrl(sessionUid))
        .reply(200, mockSessionBase);

      await expect(sessionsStore.getSession(sessionUid)).resolves.not.toThrow();

      expect(sessionsStore.session).toEqual(mockSessionBase);
    });

    it("should reset state and throw on not found error when getting session", async () => {
      const sessionUid = "non-existent-session";

      mockSessionsApi
        .onGet(baseGetSessionUrl(sessionUid))
        .reply(404, { message: "Session not found" });

      await expect(sessionsStore.getSession(sessionUid)).rejects.toBeAxiosErrorWithStatus(404);

      expect(sessionsStore.session).toEqual({});
    });

    it("should reset state and throw on server error when getting session", async () => {
      const sessionUid = "session-uid-123";

      mockSessionsApi
        .onGet(baseGetSessionUrl(sessionUid))
        .reply(500);

      await expect(sessionsStore.getSession(sessionUid)).rejects.toBeAxiosErrorWithStatus(500);

      expect(sessionsStore.session).toEqual({});
    });

    it("should reset state and throw on network error when getting session", async () => {
      const sessionUid = "session-uid-123";

      mockSessionsApi
        .onGet(baseGetSessionUrl(sessionUid))
        .networkError();

      await expect(sessionsStore.getSession(sessionUid)).rejects.toThrow("Network Error");

      expect(sessionsStore.session).toEqual({});
    });
  });

  describe("getSessionLogs", () => {
    const getLogsUrl = (sessionUid: string) => `http://localhost:3000/api/sessions/${sessionUid}/records/0`;

    it("should get session logs successfully", async () => {
      const sessionUid = "session-uid-123";
      const mockLogs = "session log data output";

      mockSessionsApi
        .onGet(getLogsUrl(sessionUid))
        .reply(200, mockLogs);

      const logs = await sessionsStore.getSessionLogs(sessionUid);

      expect(logs).toBe(mockLogs);
    });

    it("should get empty session logs successfully", async () => {
      const sessionUid = "session-uid-123";

      mockSessionsApi
        .onGet(getLogsUrl(sessionUid))
        .reply(200, "");

      const logs = await sessionsStore.getSessionLogs(sessionUid);

      expect(logs).toBe("");
    });

    it("should handle not found error when getting session logs", async () => {
      const sessionUid = "non-existent-session";

      mockSessionsApi
        .onGet(getLogsUrl(sessionUid))
        .reply(404, { message: "Session logs not found" });

      await expect(sessionsStore.getSessionLogs(sessionUid)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when getting session logs", async () => {
      const sessionUid = "session-uid-123";

      mockSessionsApi
        .onGet(getLogsUrl(sessionUid))
        .reply(500);

      await expect(sessionsStore.getSessionLogs(sessionUid)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when getting session logs", async () => {
      const sessionUid = "session-uid-123";

      mockSessionsApi
        .onGet(getLogsUrl(sessionUid))
        .networkError();

      await expect(sessionsStore.getSessionLogs(sessionUid)).rejects.toThrow("Network Error");
    });
  });

  describe("closeSession", () => {
    const closeUrl = (sessionUid: string) => `http://localhost:3000/api/sessions/${sessionUid}/close`;

    it("should close session successfully", async () => {
      const sessionData = {
        uid: "session-uid-123",
        device_uid: "device-uid-456",
      };

      mockSessionsApi
        .onPost(closeUrl(sessionData.uid))
        .reply(200);

      await expect(sessionsStore.closeSession(sessionData)).resolves.not.toThrow();
    });

    it("should handle not found error when closing session", async () => {
      const sessionData = {
        uid: "non-existent-session",
        device_uid: "device-uid-456",
      };

      mockSessionsApi
        .onPost(closeUrl(sessionData.uid))
        .reply(404, { message: "Session not found" });

      await expect(sessionsStore.closeSession(sessionData)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when closing session", async () => {
      const sessionData = {
        uid: "session-uid-123",
        device_uid: "device-uid-456",
      };

      mockSessionsApi
        .onPost(closeUrl(sessionData.uid))
        .reply(500);

      await expect(sessionsStore.closeSession(sessionData)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when closing session", async () => {
      const sessionData = {
        uid: "session-uid-123",
        device_uid: "device-uid-456",
      };

      mockSessionsApi
        .onPost(closeUrl(sessionData.uid))
        .networkError();

      await expect(sessionsStore.closeSession(sessionData)).rejects.toThrow("Network Error");
    });
  });

  describe("deleteSessionLogs", () => {
    beforeEach(() => {
      sessionsStore.session = { ...mockSessionBase, recorded: true };
    });

    it("should delete session logs successfully and update recorded state", async () => {
      const sessionUid = "session-uid-123";

      mockSessionsApi
        .onDelete(`http://localhost:3000/api/sessions/${sessionUid}/records/0`)
        .reply(200);

      await expect(sessionsStore.deleteSessionLogs(sessionUid)).resolves.not.toThrow();

      expect(sessionsStore.session.recorded).toBe(false);
    });

    it("should handle not found error when deleting session logs", async () => {
      const sessionUid = "non-existent-session";

      mockSessionsApi
        .onDelete(`http://localhost:3000/api/sessions/${sessionUid}/records/0`)
        .reply(404, { message: "Session logs not found" });

      await expect(sessionsStore.deleteSessionLogs(sessionUid)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when deleting session logs", async () => {
      const sessionUid = "session-uid-123";

      mockSessionsApi
        .onDelete(`http://localhost:3000/api/sessions/${sessionUid}/records/0`)
        .reply(500);

      await expect(sessionsStore.deleteSessionLogs(sessionUid)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when deleting session logs", async () => {
      const sessionUid = "session-uid-123";

      mockSessionsApi
        .onDelete(`http://localhost:3000/api/sessions/${sessionUid}/records/0`)
        .networkError();

      await expect(sessionsStore.deleteSessionLogs(sessionUid)).rejects.toThrow("Network Error");
    });
  });
});
