import { describe, expect, it } from "vitest";
import { normalizeNamespaceSettings } from "../namespaceSettings";

describe("normalizeNamespaceSettings", () => {
  it("fills all namespace settings defaults", () => {
    expect(normalizeNamespaceSettings()).toEqual({
      session_record: false,
      connection_announcement: "",
      device_auto_accept: false,
      allow_password: true,
      allow_public_key: true,
      allow_root: true,
      allow_empty_passwords: true,
      allow_tty: true,
      allow_tcp_forwarding: true,
      allow_web_endpoints: true,
      allow_sftp: true,
      allow_agent_forwarding: true,
    });
  });

  it("preserves provided values", () => {
    expect(
      normalizeNamespaceSettings({
        session_record: true,
        connection_announcement: "hello",
        device_auto_accept: true,
        allow_password: false,
        allow_public_key: false,
        allow_root: false,
        allow_empty_passwords: false,
        allow_tty: false,
        allow_tcp_forwarding: false,
        allow_web_endpoints: false,
        allow_sftp: false,
        allow_agent_forwarding: false,
      }),
    ).toEqual({
      session_record: true,
      connection_announcement: "hello",
      device_auto_accept: true,
      allow_password: false,
      allow_public_key: false,
      allow_root: false,
      allow_empty_passwords: false,
      allow_tty: false,
      allow_tcp_forwarding: false,
      allow_web_endpoints: false,
      allow_sftp: false,
      allow_agent_forwarding: false,
    });
  });
});
