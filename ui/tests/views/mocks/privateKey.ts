import { IPrivateKey } from "@/interfaces/IPrivateKey";

/**
 * Mock private key data for testing.
 * Provides a complete private key object with all required fields.
 */
export const mockPrivateKey: IPrivateKey = {
  id: 1,
  name: "test-key",
  data: "fake-data",
  hasPassphrase: false,
  fingerprint: "aa:bb:cc:dd",
};

/**
 * Mock private keys array for testing lists.
 * Provides multiple private keys for list/table testing scenarios.
 */
export const mockPrivateKeys: IPrivateKey[] = [mockPrivateKey];
