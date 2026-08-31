import { differenceInCalendarDays } from "date-fns";
import { type InstallKey } from "@/client";
import { formatDateShort } from "@/utils/date";

/**
 * The auto-managed system keys: every namespace has two, discriminated by `type` — `legacy` (devices
 * enrolled with only a tenant ID) and `pairing` (devices accepted through the pairing-code flow). A
 * user-created key is `user`. System keys are not presentable by an agent.
 */
export function isSystemKey(key: InstallKey): boolean {
  return key.type === "legacy" || key.type === "pairing";
}

/** The pairing system key specifically. */
export function isPairingKey(key: InstallKey): boolean {
  return key.type === "pairing";
}

/** The display name for a key: a friendly label for either system key, else the key's own name. */
export function installKeyDisplayName(key: InstallKey): string {
  if (isPairingKey(key)) return "Pairing code";
  if (isSystemKey(key)) return "Tenant-only registration";
  return key.name;
}

/**
 * How a device came to be enrolled. Legacy and pairing predate install keys, so neither has a
 * key to name; only the third does.
 */
export type EnrollmentSource =
  { kind: "legacy" } | { kind: "pairing" } | { kind: "key"; name: string };

/**
 * Resolve a device's enrollment source by matching its `install_key_id` digest
 * against the namespace's install keys: the pairing system key → pairing, the
 * legacy system key → legacy, a real key → its name, no digest or no match →
 * null (render as "—").
 */
export function resolveEnrollmentSource(
  installKeyId: string | undefined,
  installKeys: InstallKey[],
): EnrollmentSource | null {
  if (!installKeyId) return null;
  const match = installKeys.find((k) => k.id === installKeyId);
  if (!match) return null;
  if (isSystemKey(match))
    return isPairingKey(match) ? { kind: "pairing" } : { kind: "legacy" };
  return { kind: "key", name: match.name };
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
 * Shortest webhook timeout, in minutes.
 */
export const TIMEOUT_MIN = 1;
/**
 * Longest webhook timeout, in minutes. Beyond this an enrolment waiting on a webhook looks
 * hung rather than pending.
 */
export const TIMEOUT_MAX = 15;
/**
 * Shortest approval window, in hours.
 */
export const WINDOW_MIN_H = 1;
/**
 * Longest approval window, in hours.
 */
export const WINDOW_MAX_H = 24;

/**
 * Whether a string is a usable webhook URL. Only http and https: the server will call it, so a
 * scheme it cannot dial is rejected in the form rather than at enrolment time.
 */
export function isWebhookUrl(value: string): boolean {
  try {
    const parsed = new URL(value.trim());
    return parsed.protocol === "http:" || parsed.protocol === "https:";
  } catch {
    return false;
  }
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
  options: {
    secretOptional?: boolean;
    webhookTimeout?: number;
    webhookCallbackTtl?: number;
  } = {},
): string {
  if (mode === "webhook") {
    if (!isWebhookUrl(webhookUrl))
      return "Webhook URL must be an http or https URL.";
    if (!webhookSecret && !options.secretOptional)
      return "A signing secret is required for webhook mode.";
    const { webhookTimeout, webhookCallbackTtl } = options;
    if (
      webhookTimeout !== undefined &&
      (webhookTimeout < TIMEOUT_MIN || webhookTimeout > TIMEOUT_MAX)
    )
      return `Reply timeout must be ${TIMEOUT_MIN}–${TIMEOUT_MAX} seconds.`;
    if (webhookCallbackTtl !== undefined) {
      const hours = Math.round(webhookCallbackTtl / 3600);
      if (hours < WINDOW_MIN_H || hours > WINDOW_MAX_H)
        return `Callback window must be ${WINDOW_MIN_H}–${WINDOW_MAX_H} hours.`;
    }
  }
  if (mode === "allowlist" && macs.length === 0) {
    return "Add at least one MAC address for allowlist mode.";
  }
  return "";
}

function isInstallKeyExpired(key: InstallKey): boolean {
  return (
    key.expires_at != null && new Date(key.expires_at).getTime() <= Date.now()
  );
}

/**
 * The reasons an install key will not enrol anything. They are independent and can hold at once,
 * so the UI can say all of them rather than only the first.
 */
export interface KeyBlockers {
  revoked: boolean;
  disabled: boolean;
  expired: boolean;
  overused: boolean;
  inert: boolean;
}

/**
 * The independent reasons a key can't register right now — surfaced side by side, so a key that is
 * both expired and over its limit shows both.
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

/**
 * What kind of allowance a key has: one use, a set number, or no limit.
 */
export type UsageKind = "single" | "limited" | "unlimited";

/**
 * A key's allowance and how much of it is spent.
 */
export interface UsageInfo {
  kind: UsageKind;
  used: number;
  limit: number;
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
 * The expiry field of a create request. An empty or non-positive value means no expiry, which is
 * sent by omitting the field rather than by sending a zero.
 */
export function keyExpiryPayload(expiresIn: string): { expires_in?: number } {
  const days = Number(expiresIn);
  return days > 0 ? { expires_in: days } : {};
}

/**
 * The expiry field of an update request. Unlike create, clearing an expiry has to be said out
 * loud: null means "remove it", where omitting the field would mean "leave it alone".
 */
export function keyExpiryUpdatePayload(expiresIn: string): {
  expires_in: number | null;
} {
  const days = Number(expiresIn);
  return { expires_in: days > 0 ? days : null };
}

/**
 * How long is left before an expiry, and whether it has already passed. Both are returned
 * together because the caller needs the second to know how to render the first.
 */
export function getRemainingDays(expiresAt: string | null | undefined): {
  days: string;
  expired: boolean;
} {
  if (expiresAt == null) return { days: "-1", expired: false };
  const expired = new Date(expiresAt).getTime() <= Date.now();
  const remaining = differenceInCalendarDays(new Date(expiresAt), new Date());
  if (remaining < 1) return { days: "1", expired };
  return { days: String(remaining), expired: false };
}

/**
 * How loudly an expiry should read, from a fact to a warning.
 */
export type ExpiryTone = "muted" | "normal" | "warning" | "danger";

/**
 * An expiry as it is shown: the phrase, and how loudly to say it.
 */
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
  const label = formatDateShort(expiresAt);
  if (expiry.getTime() <= Date.now()) return { label, tone: "danger" };
  const tone: ExpiryTone =
    differenceInCalendarDays(expiry, new Date()) <= 7 ? "warning" : "normal";

  return { label, tone };
}
