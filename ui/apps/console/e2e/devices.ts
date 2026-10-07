import { generateKeyPairSync, randomBytes } from "node:crypto";
import { authDevice } from "@/client";
import { buildRequestContext } from "./api";
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

export async function enrollDevice(tenant: string) {
  const body = { ...buildDeviceAuthRequest(), tenant_id: tenant };
  const { data } = await authDevice({ ...buildRequestContext(), body });
  if (!data.uid || !data.token) {
    throw new Error(`expected ${body.hostname} to enroll into ${tenant}`);
  }
  return { uid: data.uid, token: data.token, name: body.hostname };
}
