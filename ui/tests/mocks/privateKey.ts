import { IPrivateKey } from "@/interfaces/IPrivateKey";
import { generateKeyPairSync } from "crypto";

/**
 * Mock private key data for testing.
 * Provides a complete private key object with all required fields.
 */
export const mockPrivateKey: IPrivateKey = {
  id: 1,
  name: "test-key",
  data: generateKeyPairSync("ed25519").privateKey.export({
    type: "pkcs8",
    format: "pem",
  }) as string,
  hasPassphrase: false,
  fingerprint: "aa:bb:cc:dd",
};

/**
 * Mock private keys array for testing lists.
 * Provides multiple private keys for list/table testing scenarios.
 */
export const mockPrivateKeys: IPrivateKey[] = [mockPrivateKey];
