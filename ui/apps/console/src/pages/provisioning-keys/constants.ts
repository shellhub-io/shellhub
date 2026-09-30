/**
 * What each enrolment mode is called and what it does, in one place, so the selector, the key
 * list and a key's page describe a mode identically.
 */
export const MODE_INFO: Record<
  string,
  { label: string; description: string; outcome: string }
> = {
  automatic: {
    label: "Automatic",
    description: "Accept every device that registers with this key.",
    outcome: "accepted the moment it connects",
  },
  manual: {
    label: "Manual",
    description:
      "Leave registered devices pending for you to review and accept.",
    outcome: "left pending for you to accept",
  },
  webhook: {
    label: "Webhook",
    description:
      "Ask your endpoint at registration whether to accept, reject, or leave the device pending.",
    outcome: "decided by your integrator",
  },
  allowlist: {
    label: "Identity allowlist",
    description:
      "Accept a device only when the identity it reports is on the list below; reject the rest.",
    outcome: "accepted if the identity it reports is allowed",
  },
};

/**
 * The description for a mode, falling back to automatic for one this build does not know — a
 * key made by a newer server still renders rather than showing a blank cell.
 */
export function modeInfo(mode: string) {
  return MODE_INFO[mode] ?? MODE_INFO.automatic;
}

/**
 * The pairing code described the way a mode is. It is no mode a key can be given, so it stays out
 * of MODE_INFO and the mode picker built from it.
 */
export const PAIRING_INFO = {
  label: "Pairing code",
  description: "Accept each device by the code its agent prints.",
  outcome: "accepted by the code its agent prints",
};
