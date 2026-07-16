import { addDays, differenceInCalendarDays, format } from "date-fns";
import { type InstallKey } from "@/client";

/** Midnight-UTC RFC3339 string for a date, so it round-trips through a day input. */
export function startOfDayUtc(date: Date): string {
  return new Date(`${date.toISOString().slice(0, 10)}T00:00:00Z`).toISOString();
}

/** The prefilled expiry for a new/never-off key: 30 days out at the start of that day. */
export function defaultExpiry(): string {
  return startOfDayUtc(addDays(new Date(), 30));
}

/**
 * The auto-managed "legacy" key: every namespace has exactly one key with
 * `system === true`, the source attributed to devices enrolled with only a
 * tenant ID. It cannot be edited or deleted (the API returns 403).
 */
export function isSystemKey(key: InstallKey): boolean {
  return key.system === true;
}

export type EnrollmentSource =
  { kind: "legacy" } | { kind: "key"; name: string };

/**
 * Resolve a device's enrollment source by matching its `install_key_id` digest
 * against the namespace's install keys: the system key → legacy/direct, a real
 * key → its name, no digest or no match → null (render as "—").
 */
export function resolveEnrollmentSource(
  installKeyId: string | undefined,
  installKeys: InstallKey[],
): EnrollmentSource | null {
  if (!installKeyId) return null;
  const match = installKeys.find((k) => k.id === installKeyId);
  if (!match) return null;
  return match.system ? { kind: "legacy" } : { kind: "key", name: match.name };
}

/** Split a MAC-allowlist textarea into a normalized, deduped list (lowercased, blanks dropped). */
export function parseAllowedMacs(text: string): string[] {
  const seen = new Set<string>();
  for (const line of text.split("\n")) {
    const mac = line.trim().toLowerCase();
    if (mac) seen.add(mac);
  }
  return [...seen];
}

/** Client-side name validation for an install key: 3-20 chars, alphanumerics plus - and _. */
export function validateName(value: string): string {
  if (value.length < 3) return "Name must be at least 3 characters.";
  if (value.length > 20) return "Name must be at most 20 characters.";
  if (!/^[a-zA-Z0-9_-]+$/.test(value))
    return "Name can only contain letters, numbers, - and _.";
  return "";
}

/**
 * Client-side check of a mode's required config, mirroring the API: webhook needs an https URL and a
 * secret; allowlist needs at least one MAC. Returns an error message, or "" when valid. Pass
 * `secretOptional` when editing a key already in webhook mode: its stored secret is write-only, so a
 * blank field keeps it rather than clearing it.
 */
export function validateModeConfig(
  mode: string,
  webhookUrl: string,
  webhookSecret: string,
  macs: string[],
  options: { secretOptional?: boolean } = {},
): string {
  if (mode === "webhook") {
    if (!/^https?:\/\//.test(webhookUrl.trim()))
      return "Webhook URL must be an http or https URL.";
    if (!webhookSecret && !options.secretOptional)
      return "A signing secret is required for webhook mode.";
  }
  if (mode === "allowlist" && macs.length === 0) {
    return "Add at least one MAC address for allowlist mode.";
  }
  return "";
}

export type InstallKeyStatus =
  "valid" | "disabled" | "expired" | "revoked" | "overused";

/**
 * Whether a key's absolute expiry has elapsed. A null `expires_at` never
 * expires.
 */
function isInstallKeyExpired(key: InstallKey): boolean {
  return (
    key.expires_at != null && new Date(key.expires_at).getTime() <= Date.now()
  );
}

/**
 * Derive a install key's status client-side from its fields, mirroring how the
 * API decides usability: revoked wins, then a reversible pause, then an elapsed
 * expiry, then an exhausted usage limit; otherwise the key is still valid.
 */
export function getInstallKeyStatus(key: InstallKey): InstallKeyStatus {
  if (key.revoked) return "revoked";
  if (key.disabled) return "disabled";
  if (isInstallKeyExpired(key)) return "expired";
  if (key.usage_limit > 0 && key.used_times >= key.usage_limit) {
    return "overused";
  }
  return "valid";
}

export interface KeyBlockers {
  revoked: boolean;
  disabled: boolean;
  expired: boolean;
  overused: boolean;
  /** Any blocker present: the key can no longer register devices. */
  inert: boolean;
}

/**
 * The independent reasons a key can't register right now. Unlike getInstallKeyStatus (which collapses
 * to a single highest-priority status), these are surfaced side by side, so a key that is both expired
 * and over its limit shows both — each in the column that caused it.
 */
export function getKeyBlockers(key: InstallKey): KeyBlockers {
  const revoked = !!key.revoked;
  const disabled = !!key.disabled;
  const expired = isInstallKeyExpired(key);
  const overused = key.usage_limit > 0 && key.used_times >= key.usage_limit;

  return {
    revoked,
    disabled,
    expired,
    overused,
    inert: revoked || disabled || expired || overused,
  };
}

export type UsageKind = "single" | "limited" | "unlimited";

export interface UsageInfo {
  kind: UsageKind;
  used: number;
  /** The enrollment cap: 1 for single-use, N for limited, 0 for unlimited. */
  limit: number;
  /** Fill fraction 0..1 for the meter; always 0 for unlimited (no cap to fill). */
  ratio: number;
  exhausted: boolean;
}

/**
 * Decode a key's enrollment budget for the usage meter. `usage_limit` is the
 * source of truth the API derives reusability from: 0 unlimited, 1 single-use,
 * N (>=2) limited to N devices.
 */
export function getUsageInfo(key: InstallKey): UsageInfo {
  const used = key.used_times;
  if (key.usage_limit === 0) {
    return { kind: "unlimited", used, limit: 0, ratio: 0, exhausted: false };
  }
  const limit = key.usage_limit;
  return {
    kind: limit === 1 ? "single" : "limited",
    used,
    limit,
    ratio: Math.min(1, used / limit),
    exhausted: used >= limit,
  };
}

/**
 * The second-line descriptor for the Enrollment cell: a non-valid key surfaces its state (paused,
 * revoked, ...) with a tone; a valid key returns null so the cell shows the mode's behavior instead.
 */
export function getInstallKeyStateLabel(
  status: InstallKeyStatus,
): { label: string; tone: "muted" | "warning" | "danger" } | null {
  switch (status) {
    case "disabled":
      return { label: "Disabled", tone: "muted" };
    case "expired":
      return { label: "Expired", tone: "warning" };
    case "overused":
      return { label: "Limit reached", tone: "warning" };
    case "revoked":
      return { label: "Revoked", tone: "danger" };
    default:
      return null;
  }
}

export type ExpiryTone = "muted" | "normal" | "warning" | "danger";

export interface ExpiryInfo {
  label: string;
  tone: ExpiryTone;
}

/**
 * Absolute, urgency-aware expiry for a card. Non-expiring keys (null
 * `expires_at`) read as a muted "Never"; otherwise the concrete date is always
 * shown, with urgency carried by colour: danger once elapsed, a warning within
 * the last week, normal beyond that.
 */
export function getExpiryInfo(
  expiresAt: string | null | undefined,
): ExpiryInfo {
  if (expiresAt == null) return { label: "Never", tone: "muted" };
  const expiry = new Date(expiresAt);
  const label = format(expiry, "MMM d, yyyy");
  if (expiry.getTime() <= Date.now()) return { label, tone: "danger" };
  const tone: ExpiryTone =
    differenceInCalendarDays(expiry, new Date()) <= 7 ? "warning" : "normal";

  return { label, tone };
}
