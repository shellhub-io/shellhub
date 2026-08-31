import { isEnterpriseOrCloud } from "@/env";

interface ErrorEntry {
  title?: string;
  message: string;
  reconnect: boolean;
  hints: string[];
  links?: Array<{ label: string; to: string }>;
}

/**
 * A terminal failure as the banner shows it. reconnect says whether retrying could help, which
 * decides if the button is offered at all; hints and links are the ways out.
 */
export interface TerminalError {
  title: string;
  message: string;
  reconnect: boolean;
  hints: string[];
  links: Array<{ label: string; to: string }>;
}

const errorMap: Record<string, ErrorEntry> = {
  "failed to authenticate to device": {
    message: "The username or password is incorrect.",
    reconnect: true,
    hints: [
      "The username must match the OS user on the device, not your ShellHub account.",
    ],
  },
  "failed to find the device": {
    message: "Cannot reach this device.",
    reconnect: false,
    hints: ["The device may be offline or removed from ShellHub."],
    links: [{ label: "Device details", to: "/devices/$uid" }],
  },
  "failed to connect to device": {
    message: "Cannot reach this device.",
    reconnect: false,
    hints: [
      "Make sure the device is powered on and the ShellHub agent is running.",
    ],
    links: [{ label: "Device details", to: "/devices/$uid" }],
  },
  "failed to create a session between the server to the agent": {
    message: "Connected to the device but could not start a session.",
    reconnect: false,
    hints: [
      "This is usually temporary. Wait a few seconds and try again.",
      "If it persists, the ShellHub agent on the device may need to be restarted.",
    ],
  },
  "failed to get the shell to agent": {
    message: "Connected to the device but it refused to open a shell.",
    reconnect: false,
    hints: [
      "Check that the user has a valid login shell (not /usr/sbin/nologin or /bin/false).",
      "On containerized agents, make sure a shell is available inside the container.",
    ],
  },
  "failed to request the pty to agent": {
    message: "Connected to the device but terminal allocation was denied.",
    reconnect: false,
    hints: [
      "The device may have too many open terminals. Close unused sessions and try again.",
    ],
  },
  "failed to find the credentials": {
    message: "Your session credentials have expired.",
    reconnect: false,
    hints: [
      "This happens when the connection takes too long to establish. Close and open a new terminal.",
    ],
  },
  "failed to get auth data from key": {
    message: "The selected public key could not be used for authentication.",
    reconnect: false,
    hints: [
      "The key may have been deleted or modified since you selected it.",
      "Verify the key is still registered in your SSH key settings.",
    ],
  },
  "failed to use the public key for this action": {
    message: "This public key is not authorized for this device.",
    reconnect: false,
    hints: [
      "Public keys must be associated with the target device or its tags.",
    ],
  },
  "connections using public keys are not permitted when the agent version is 0.5.x or earlier":
    {
      message: "This device does not support public key authentication.",
      reconnect: true,
      hints: [
        "The ShellHub agent is v0.5.x or earlier. Update to v0.6.0+ for public key support, or reconnect using a password.",
      ],
      links: [{ label: "Device details", to: "/devices/$uid" }],
    },
  "access to the device has been denied": {
    title: "Access denied",
    message: "You do not have permission to access this device.",
    reconnect: false,
    hints: ["Access may be restricted by a billing limit or namespace policy."],
  },
  "invalid sshid format": {
    title: "Invalid connection identifier",
    message: "The SSH connection identifier is malformed.",
    reconnect: false,
    hints: [
      "Use the format username@namespace.device@host - for example: root@dev.agent@localhost.",
    ],
  },
};

/**
 * The WebSocket frame kinds. The values match the messageKind iota in ssh/web/messages.go, so
 * reordering them there without changing them here silently misroutes every frame.
 */
export const WS_KIND = {
  INPUT: 1,
  RESIZE: 2,
  SIGNATURE: 3,
  ERROR: 4,
  SESSION: 5,
  REAUTH: 6,
} as const;

/**
 * The session could not be started at all. Worth retrying, since the cause is usually transient.
 */
export const HTTP_CONNECT_ERROR: TerminalError = {
  title: "Connection failed",
  message: "Could not start the session.",
  reconnect: true,
  hints: [
    "The ShellHub server may be temporarily unavailable. Try again in a moment.",
  ],
  links: [],
};

/**
 * The session ended cleanly. Not an error to retry — the shell exited.
 */
export const WS_CLOSE_ERROR: TerminalError = {
  title: "Disconnected",
  message: "The session has ended.",
  reconnect: false,
  hints: [],
  links: [],
};

/**
 * The connection dropped. Reconnecting is not offered, because the device is the likely cause
 * and an immediate retry would fail the same way.
 */
export const WS_NETWORK_ERROR: TerminalError = {
  title: "Connection error",
  message: "Could not reach the device.",
  reconnect: false,
  hints: [
    "Check your network connection and make sure the ShellHub server is running.",
  ],
  links: [],
};

/**
 * The device asked for a fresh re-authentication and the user dismissed it. Retrying is
 * offered, since it will prompt again.
 */
export const WS_REAUTH_CANCELLED: TerminalError = {
  title: "Re-authentication needed",
  message: "This device requires a fresh re-authentication to connect.",
  reconnect: true,
  hints: [],
  links: [],
};

/**
 * Reads a WebSocket frame as a protocol message, or null when it is not one. A frame that is
 * not JSON is ordinary terminal output, not a failure, which is why this returns null rather
 * than throwing.
 */
export function parseMessage(
  data: string,
): { kind: number; data: string } | null {
  try {
    const msg: unknown = JSON.parse(data);
    if (
      typeof msg === "object" &&
      msg !== null &&
      "kind" in msg &&
      "data" in msg &&
      typeof (msg as { kind: unknown }).kind === "number" &&
      typeof (msg as { data: unknown }).data === "string"
    ) {
      return {
        kind: (msg as { kind: number }).kind,
        data: (msg as { data: string }).data,
      };
    }
  } catch {
    return null;
  }
  return null;
}

/**
 * Turns a server error string into the banner's contents. Identity mode changes the advice —
 * the same failure has a different remedy when access is granted by identity rather than by
 * key — so it is a parameter rather than read here.
 */
export function resolveError(
  raw: string,
  deviceUid: string,
  isIdentityMode: boolean,
): TerminalError {
  const entry = errorMap[raw];
  if (!entry) {
    return {
      title: "Connection failed",
      message: "An unexpected error occurred.",
      reconnect: false,
      hints: [],
      links: [],
    };
  }

  const hints = entry.hints.map((h) =>
    isIdentityMode ? h.replace("namespace policy", "access policy") : h,
  );

  const links = (entry.links ?? []).map((l) => ({
    ...l,
    to: l.to.replace("$uid", encodeURIComponent(deviceUid)),
  }));

  if (isEnterpriseOrCloud()) {
    hints.push(
      isIdentityMode
        ? "An access policy may be blocking this connection."
        : "A firewall rule may be blocking this connection.",
    );
    links.push({
      label: isIdentityMode ? "Access policies" : "Firewall rules",
      to: "/firewall-rules",
    });
  }

  return {
    title: entry.title ?? "Connection failed",
    message: entry.message,
    reconnect: entry.reconnect,
    hints,
    links,
  };
}
