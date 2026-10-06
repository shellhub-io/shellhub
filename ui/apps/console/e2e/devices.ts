import { generateKeyPairSync, randomBytes } from "node:crypto";
import { buildShortId } from "./seed";

export function buildDeviceAuthRequest() {
  const { publicKey } = generateKeyPairSync("rsa", {
    modulusLength: 2048,
    publicKeyEncoding: { type: "spki", format: "pem" },
  });
  return {
    hostname: `e2e-device-${buildShortId()}`,
    identity: {
      mac: [0x02, ...randomBytes(5)]
        .map((byte) => byte.toString(16).padStart(2, "0"))
        .join(":"),
    },
    info: {
      id: "debian",
      pretty_name: "Debian GNU/Linux 12",
      version: "12",
      arch: "x86_64",
      platform: "native" as const,
    },
    public_key: publicKey,
  };
}
