import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { adminApi } from "@/api/http";
import useSessionsStore from "@admin/store/modules/sessions";
import { IAdminSession } from "@admin/interfaces/ISession";
import { buildUrl } from "@tests/utils/url";

const mockSessionBase: IAdminSession = {
  uid: "session-uid-123",
  device_uid: "device-uid-456",
  device: {
    uid: "device-uid-456",
    name: "admin-device",
    identity: {
      mac: "00:1A:2B:3C:4D:5E",
    },
    info: {
      id: "debian",
      pretty_name: "Debian GNU/Linux 11",
      version: "11",
      arch: "x86_64",
      platform: "docker",
    },
    public_key: "ssh-rsa AAAAB3NzaC1...",
    tenant_id: "tenant-id-789",
    last_seen: "2026-01-01T12:00:00.000Z",
    status_updated_at: "2026-01-01T12:00:00.000Z",
    online: true,
    namespace: "admin-namespace",
    status: "accepted",
    created_at: "2026-01-01T00:00:00.000Z",
    remote_addr: "192.168.1.100",
    position: { latitude: 0, longitude: 0 },
    tags: [{
      name: "admin",
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

describe("Admin Sessions Store", () => {
  let sessionsStore: ReturnType<typeof useSessionsStore>;
  let mockAdminApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    sessionsStore = useSessionsStore();
    mockAdminApi = new MockAdapter(adminApi.getAxios());
  });

  afterEach(() => { mockAdminApi.reset(); });

  describe("Initial State", () => {
    it("should have empty sessions array", () => {
      expect(sessionsStore.sessions).toEqual([]);
    });

    it("should have zero session count", () => {
      expect(sessionsStore.sessionCount).toBe(0);
    });
  });

  describe("fetchSessionList", () => {
    const baseUrl = "http://localhost:3000/admin/api/sessions";

    it("should fetch sessions list successfully with pagination", async () => {
      const sessionList = [mockSessionBase];

      mockAdminApi.onGet(buildUrl(baseUrl, { page: "1", per_page: "10" })).reply(200, sessionList, { "x-total-count": "1" });

      await expect(sessionsStore.fetchSessionList({ perPage: 10, page: 1 })).resolves.not.toThrow();

      expect(sessionsStore.sessions).toEqual(sessionList);
      expect(sessionsStore.sessionCount).toBe(1);
    });

    it("should fetch sessions list with multiple sessions", async () => {
      const sessionList = [
        mockSessionBase,
        { ...mockSessionBase, uid: "session-uid-456" },
      ];

      mockAdminApi.onGet(buildUrl(baseUrl, { page: "2", per_page: "20" })).reply(200, sessionList, { "x-total-count": "2" });

      await expect(sessionsStore.fetchSessionList({ perPage: 20, page: 2 })).resolves.not.toThrow();

      expect(sessionsStore.sessions).toEqual(sessionList);
      expect(sessionsStore.sessionCount).toBe(2);
    });

    it("should fetch empty sessions list successfully", async () => {
      mockAdminApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(200, [], { "x-total-count": "0" });

      await expect(sessionsStore.fetchSessionList({ perPage: 10, page: 1 })).resolves.not.toThrow();

      expect(sessionsStore.sessions).toEqual([]);
      expect(sessionsStore.sessionCount).toBe(0);
    });

    it("should throw on server error when fetching sessions list", async () => {
      mockAdminApi.onGet(buildUrl(baseUrl, { page: "1", per_page: "10" })).reply(500);

      await expect(sessionsStore.fetchSessionList({ perPage: 10, page: 1 })).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when fetching sessions list", async () => {
      mockAdminApi.onGet(buildUrl(baseUrl, { page: "1", per_page: "10" })).networkError();

      await expect(sessionsStore.fetchSessionList({ perPage: 10, page: 1 })).rejects.toThrow("Network Error");
    });
  });

  describe("fetchSessionById", () => {
    const generateGetSessionUrl = (sessionUid: string) => `http://localhost:3000/admin/api/sessions/${sessionUid}`;

    it("should fetch session by id successfully and return data", async () => {
      const sessionUid = "session-uid-123";

      mockAdminApi.onGet(generateGetSessionUrl(sessionUid)).reply(200, mockSessionBase);

      await sessionsStore.fetchSessionById(sessionUid);

      expect(sessionsStore.session).toEqual(mockSessionBase);
    });

    it("should throw on not found error when fetching session by id", async () => {
      const sessionUid = "non-existent-session";

      mockAdminApi.onGet(generateGetSessionUrl(sessionUid)).reply(404, { message: "Session not found" });

      await expect(sessionsStore.fetchSessionById(sessionUid)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when fetching session by id", async () => {
      const sessionUid = "session-uid-123";

      mockAdminApi.onGet(generateGetSessionUrl(sessionUid)).networkError();

      await expect(sessionsStore.fetchSessionById(sessionUid)).rejects.toThrow("Network Error");
    });
  });
});
