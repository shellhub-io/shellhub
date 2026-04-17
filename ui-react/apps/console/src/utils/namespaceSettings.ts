import type { NamespaceSettings } from "../client";

export function normalizeNamespaceSettings(
  settings?: Partial<NamespaceSettings> | null,
): NamespaceSettings {
  return {
    session_record: settings?.session_record ?? false,
    connection_announcement: settings?.connection_announcement ?? "",
    allow_password: settings?.allow_password ?? true,
    allow_public_key: settings?.allow_public_key ?? true,
    allow_root: settings?.allow_root ?? true,
    allow_empty_passwords: settings?.allow_empty_passwords ?? true,
    allow_tty: settings?.allow_tty ?? true,
    allow_tcp_forwarding: settings?.allow_tcp_forwarding ?? true,
    allow_web_endpoints: settings?.allow_web_endpoints ?? true,
    allow_sftp: settings?.allow_sftp ?? true,
    allow_agent_forwarding: settings?.allow_agent_forwarding ?? true,
  };
}
