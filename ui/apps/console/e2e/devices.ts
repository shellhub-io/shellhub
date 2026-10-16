import type { NamespaceCredential } from "./seed";
import {
  generateKeyPairSync,
  randomBytes,
  sign,
  type KeyObject,
} from "node:crypto";
import { authDevice, createDeviceAuthChallenge } from "@/client";
import { buildRequestContext } from "./api";
import { buildShortId } from "./seed";

const devicePrivateKeys = new Map<string, KeyObject>();

function sshString(value: Buffer | string) {
  const bytes = typeof value === "string" ? Buffer.from(value) : value;
  const length = Buffer.alloc(4);
  length.writeUInt32BE(bytes.length);
  return Buffer.concat([length, bytes]);
}

interface SignableDeviceAuth {
  public_key: string;
  tenant_id?: string;
  provisioning_key?: string;
}

export async function signDeviceAuth<T extends SignableDeviceAuth>(body: T) {
  const privateKey = devicePrivateKeys.get(body.public_key);
  if (!privateKey) {
    throw new Error(
      "the request's public key was not made by buildDeviceAuthRequest, so nothing can sign it",
    );
  }

  const { data } = await createDeviceAuthChallenge({
    ...buildRequestContext(),
  });
  if (!data.challenge) {
    throw new Error("expected the server to issue a device auth challenge");
  }

  const statement = Buffer.concat([
    sshString("shellhub-device-auth-v1"),
    sshString(data.challenge),
    sshString(body.tenant_id ?? ""),
    sshString(body.provisioning_key ?? ""),
    sshString(body.public_key),
  ]);
  const signature = Buffer.concat([
    sshString("rsa-sha2-256"),
    sshString(sign("sha256", statement, privateKey)),
  ]);

  return {
    ...body,
    challenge: data.challenge,
    signature: signature.toString("base64"),
  };
}

export function buildDeviceAuthRequest() {
  const { publicKey, privateKey } = generateKeyPairSync("rsa", {
    modulusLength: 2048,
    publicKeyEncoding: { type: "spki", format: "pem" },
  });
  devicePrivateKeys.set(publicKey, privateKey);
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

export async function enrollDevice({
  tenant,
  provisioningKey,
}: NamespaceCredential) {
  const body = await signDeviceAuth({
    ...buildDeviceAuthRequest(),
    ...(provisioningKey
      ? { provisioning_key: provisioningKey }
      : { tenant_id: tenant }),
  });
  const { data } = await authDevice({ ...buildRequestContext(), body });
  if (!data.uid || !data.token) {
    throw new Error(`expected ${body.hostname} to enroll into ${tenant}`);
  }
  return { uid: data.uid, token: data.token, name: body.hostname };
}
