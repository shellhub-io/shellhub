import { getConfig } from "../../env";

interface ErrorEntry {
  message: string;
  reconnect: boolean;
  hints: string[];
  links?: Array<{ label: string; to: string }>;
}

export interface TerminalError {
  title: string;
  message: string;
  reconnect: boolean;
  hints: string[];
  links: Array<{ label: string; to: string }>;
}

// Keys must match error strings from ssh/web/errors.go exactly.
// If backend error messages change, update these keys accordingly.
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
};

// Values match ssh/web/messages.go messageKind iota.
// SIGNATURE (3) is omitted — handled internally during
// public-key auth and never reaches the terminal UI.
export const WS_KIND = { INPUT: 1, RESIZE: 2, ERROR: 4 } as const;

export const HTTP_CONNECT_ERROR: TerminalError = {
  title: "Connection failed",
  message: "Could not start the session.",
  reconnect: true,
  hints: [
    "The ShellHub server may be temporarily unavailable. Try again in a moment.",
  ],
  links: [],
};

export const WS_CLOSE_ERROR: TerminalError = {
  title: "Disconnected",
  message: "The session has ended.",
  reconnect: false,
  hints: [],
  links: [],
};

export const WS_NETWORK_ERROR: TerminalError = {
  title: "Connection error",
  message: "Could not reach the device.",
  reconnect: false,
  hints: [
    "Check your network connection and make sure the ShellHub server is running.",
  ],
  links: [],
};

export function parseMessage(
  data: string,
): { kind: number; data: string } | null {
  try {
    const msg = JSON.parse(data);
    if (
      typeof msg === "object" &&
      msg !== null &&
      typeof msg.kind === "number" &&
      typeof msg.data === "string"
    ) {
      return { kind: msg.kind, data: msg.data };
    }
  } catch {
    // Not JSON — regular text frame
  }
  return null;
}

export function resolveError(raw: string, deviceUid: string): TerminalError {
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

  const hasFirewall = getConfig().cloud || getConfig().enterprise;

  const hints = [...entry.hints];
  if (hasFirewall) {
    hints.push("A firewall rule may be blocking this connection.");
  }

  const links = (entry.links ?? []).map((l) => ({
    ...l,
    to: l.to.split("$uid").join(deviceUid),
  }));
  if (hasFirewall) {
    links.push({ label: "Firewall rules", to: "/firewall/rules" });
  }

  return {
    title: "Connection failed",
    message: entry.message,
    reconnect: entry.reconnect,
    hints,
    links,
  };
}
