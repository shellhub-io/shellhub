import { describe, expect, it } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useSessionsStore from "@admin/store/modules/sessions";
import { IAdminSession } from "@admin/interfaces/ISession";

const session = {
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
} as IAdminSession;

const sessions = [
  { ...session },
  {
    ...session,
    device: {
      ...session.device,
      name: "b4-2e-99",
    },
  },
];

const sessionCount = 2;

describe("Sessions Pinia Store", () => {
  setActivePinia(createPinia());
  const sessionsStore = useSessionsStore();

  it("returns default session state", () => {
    expect(sessionsStore.sessions).toEqual([]);
    expect(sessionsStore.sessionCount).toEqual(0);
  });

  it("sets sessions and total count", () => {
    sessionsStore.sessions = sessions;
    sessionsStore.sessionCount = sessionCount;

    expect(sessionsStore.sessions).toEqual(sessions);
    expect(sessionsStore.sessionCount).toEqual(sessionCount);
  });
});
