import { differenceInDays } from "date-fns";
import { formatDateFull, formatDateShort } from "@/utils/date";
import { isSdkError } from "@/api/errors";
import type { SshIdentity } from "@/client";

/**
 * Whether an enroll failure means the key is already an identity. The endpoint
 * answers 409 for a fingerprint already taken in the namespace, which for a key
 * the caller holds privately can only be an enrollment of their own.
 */
export function isAlreadyEnrolled(err: unknown): boolean {
  return isSdkError(err) && err.status === 409;
}

/**
 * Turn the expiry form value into the create-request field. "-1" (never) omits
 * it entirely, so a key with no deadline carries no expires_at at all rather
 * than one far in the future.
 */
export function keyExpiryPayload(expiresIn: string): { expires_in?: number } {
  const days = Number(expiresIn);

  return days > 0 ? { expires_in: days } : {};
}

/**
 * The same expiry, plus the single-use flag that only a service-account key
 * carries.
 */
export function serviceAccountLifecyclePayload(
  expiresIn: string,
  singleUse: boolean,
): { single_use: boolean; expires_in?: number } {
  return { single_use: singleUse, ...keyExpiryPayload(expiresIn) };
}

/**
 * The wording used for one kind of identity source. revokeConsequence is separate from
 * description because revoking a browser key logs that browser out, and the confirm dialog has
 * to say so before it happens.
 */
export interface IdentitySourceCopy {
  label: string;
  notable: boolean;
  description: string;
  revokeConsequence: string;
}

const SOURCE_COPY: Record<string, IdentitySourceCopy> = {
  browser: {
    label: "Browser",
    notable: false,
    description:
      "Browser. The web terminal generated this key inside a browser. It cannot be exported, so it works only there, and it dies with that browser's site data.",
    revokeConsequence:
      "That browser registers a new key the next time it opens a web terminal.",
  },
  approval: {
    label: "At login",
    notable: false,
    description:
      "At login. A member approved this key when an SSH client offered one this namespace had not seen.",
    revokeConsequence: "The next login with it needs approval again.",
  },
  manual: {
    label: "Manual",
    notable: false,
    description:
      "Manual. Somebody pasted or uploaded this public key on the SSH Identities page.",
    revokeConsequence: "The next login with it needs approval again.",
  },
};

// The browser key belonging to the browser being used right now. Worth its own
// copy in both directions: it is the one row a person can act on without going
// anywhere, and the only one whose revocation costs them the terminal they are
// looking at.
const CURRENT_BROWSER_COPY: IdentitySourceCopy = {
  label: "This browser",
  notable: true,
  description:
    "This browser. The web terminal generated this key here. It cannot be exported, and clearing this site's data destroys it.",
  revokeConsequence:
    "This browser registers a new key the next time you open a web terminal.",
};

/**
 * Describe how an identity came to exist. Only a browser key is notable: it is
 * the one that behaves differently (bound to one browser, unexportable, gone
 * with the site data), while a key added by hand and one accepted at login are
 * both just a key on somebody's disk. An unknown source falls back to the
 * added-by-hand copy, which claims the least.
 */
export function sshIdentitySource(
  source: string | undefined,
  isCurrentBrowser = false,
): IdentitySourceCopy {
  if (source === "browser" && isCurrentBrowser) {
    return CURRENT_BROWSER_COPY;
  }

  return SOURCE_COPY[source ?? "manual"] ?? SOURCE_COPY.manual;
}

/**
 * Shorten a fingerprint to what a person actually reads to tell two keys apart,
 * keeping the algorithm prefix: that is how `ssh-keygen -lf` prints it, and a
 * bare base64 blob does not say what hashed it. Callers must keep the full
 * value reachable to copy, since a truncated fingerprint matches nothing.
 */
export function shortFingerprint(fingerprint: string): string {
  const [prefix, body] = fingerprint.includes(":")
    ? [
        `${fingerprint.slice(0, fingerprint.indexOf(":") + 1)}`,
        fingerprint.slice(fingerprint.indexOf(":") + 1),
      ]
    : ["", fingerprint];

  return body.length > 12 ? `${prefix}${body.slice(0, 12)}\u2026` : fingerprint;
}

/**
 * How loudly an identity's state should read: dead is spent or expired, armed is a warning worth
 * acting on, quiet is a fact about the key and nothing more.
 */
export type IdentityStatusTone = "dead" | "armed" | "quiet";

/**
 * How close an expiry has to be before it stops being metadata and starts being
 * a warning. Anything further out is just a fact about the key.
 */
const EXPIRY_WARNING_DAYS = 14;

/**
 * How an identity ends: already consumed, already expired, single-use, expiring on a date, or
 * never. value is the rendered phrase and title the tooltip carrying the exact timestamp.
 */
export interface IdentityEndOfLife {
  kind: "consumed" | "expired" | "single-use" | "expires" | "never";
  value: string;
  tone: IdentityStatusTone;
  title?: string;
}

type Lifecycle = Pick<SshIdentity, "expires_at" | "consumed_at" | "single_use">;

/**
 * When and how an enrolled key stops working. Everything that ends a key lands
 * here — a TTL, a burnt single-use, or nothing at all — because they occupy the
 * same slot on screen and a key can only end once.
 */
export function sshIdentityEndOfLife(
  identity: Lifecycle,
  now: number = Date.now(),
): IdentityEndOfLife {
  if (identity.consumed_at)
    return {
      kind: "consumed",
      value: "consumed",
      tone: "dead",
      title: formatDateFull(identity.consumed_at),
    };

  if (identity.expires_at && new Date(identity.expires_at).getTime() <= now)
    return {
      kind: "expired",
      value: "expired",
      tone: "dead",
      title: formatDateFull(identity.expires_at),
    };

  if (identity.single_use)
    return { kind: "single-use", value: "on first use", tone: "armed" };

  if (identity.expires_at) {
    const at = new Date(identity.expires_at);

    return {
      kind: "expires",
      value: formatDateShort(identity.expires_at),
      tone:
        differenceInDays(at, now) <= EXPIRY_WARNING_DAYS ? "armed" : "quiet",
      title: formatDateFull(identity.expires_at),
    };
  }

  return { kind: "never", value: "never", tone: "quiet" };
}

/**
 * The status chip for an identity — its label and how loudly to show it.
 */
export interface IdentityStatus {
  label: string;
  tone: IdentityStatusTone;
  title?: string;
}

const STATUS_LABEL: Record<IdentityEndOfLife["kind"], string | null> = {
  consumed: "Consumed",
  expired: "Expired",
  "single-use": "Single-use",
  expires: "Expires",
  never: null,
};

/**
 * The one status chip for a key's lifecycle, for screens that show it as a
 * badge rather than in a column of its own. A durable key yields null.
 */
export function sshIdentityStatus(
  identity: Lifecycle,
  now: number = Date.now(),
): IdentityStatus | null {
  const end = sshIdentityEndOfLife(identity, now);
  const label = STATUS_LABEL[end.kind];
  if (!label) return null;

  return {
    label: end.kind === "expires" ? `${label} ${end.value}` : label,
    tone: end.tone,
    title: end.title,
  };
}
