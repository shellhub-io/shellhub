import { IPublicKey } from "@/interfaces/IPublicKey";

/**
 * Mock public key data for testing.
 * Provides a complete public key object with all required fields.
 */
export const mockPublicKey: IPublicKey = {
  data: "",
  fingerprint: "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:01",
  created_at: "2025-01-01T00:00:00.000Z",
  tenant_id: "00000000-0000-4000-0000-000000000000",
  name: "public-key-test",
  username: ".*",
  filter: { hostname: ".*" },
};

/**
 * Mock public keys array for testing lists.
 * Provides multiple public keys for list/table testing scenarios.
 */
export const mockPublicKeys: IPublicKey[] = [mockPublicKey];
