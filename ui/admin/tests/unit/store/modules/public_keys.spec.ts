import { describe, expect, it, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import usePublicKeysStore from "@admin/store/modules/public_keys";

vi.mock("@admin/store/api/public_keys", () => ({
  removePublicKey: vi.fn().mockResolvedValue({}),
}));

describe("PublicKeys Pinia Store", () => {
  let publicKeysStore: ReturnType<typeof usePublicKeysStore>;

  const numberPublicKeys = 2;
  const publicKeys = [
    {
      data: "BBGVvbmFyZG8=",
      fingerprint: "b8:26:d5",
      created_at: "2020-11-23T20:59:13.323Z",
      tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      name: "shellhub",
      username: "user1",
      filter: { hostname: "device1" },
    },
    {
      data: "AbGVvbmFyZG8=",
      fingerprint: "b7:25:f8",
      created_at: "2020-11-23T20:59:13.323Z",
      tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      name: "shellhub",
      username: "user2",
      filter: { hostname: "device2" },
    },
  ];
  const publicKey = publicKeys[1];

  beforeEach(() => {
    setActivePinia(createPinia());
    publicKeysStore = usePublicKeysStore();
  });

  it("returns public key default state", () => {
    expect(publicKeysStore.list).toEqual([]);
    expect(publicKeysStore.get).toEqual({});
    expect(publicKeysStore.getNumberPublicKeys).toEqual(0);
  });

  it("sets public keys and total count", () => {
    publicKeysStore.publicKeys = publicKeys;
    publicKeysStore.numberPublicKeys = numberPublicKeys;

    expect(publicKeysStore.list).toEqual(publicKeys);
    expect(publicKeysStore.getNumberPublicKeys).toEqual(numberPublicKeys);
  });

  it("sets a single public key", () => {
    publicKeysStore.publicKey = publicKey;
    expect(publicKeysStore.get).toEqual(publicKey);
  });

  it("removes a public key from list", async () => {
    publicKeysStore.publicKeys = [...publicKeys];
    await publicKeysStore.remove(publicKey.fingerprint);

    expect(publicKeysStore.list.length).toEqual(numberPublicKeys - 1);
    expect(publicKeysStore.list.find((k) => k.fingerprint === publicKey.fingerprint)).toBeUndefined();
  });

  it("clears public key object", () => {
    publicKeysStore.publicKey = publicKey;
    publicKeysStore.clearObjectPublicKey();
    expect(publicKeysStore.get).toEqual({});
  });

  it("clears public keys list", () => {
    publicKeysStore.publicKeys = publicKeys;
    publicKeysStore.numberPublicKeys = numberPublicKeys;

    publicKeysStore.clearListPublicKeys();

    expect(publicKeysStore.list).toEqual([]);
    expect(publicKeysStore.getNumberPublicKeys).toEqual(0);
  });
});
