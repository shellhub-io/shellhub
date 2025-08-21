import { describe, expect, it, beforeEach, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { sshApi } from "@/api/http";
import usePublicKeysStore from "@/store/modules/public_keys";

const mockPublicKey = {
  data: "test-key",
  fingerprint: "fake-fingerprint",
  created_at: "2020-05-01T00:00:00.000Z",
  tenant_id: "fake-tenant",
  name: "example",
  filter: {
    hostname: ".*",
  },
  username: ".*",
};

describe("Public Keys Store", () => {
  localStorage.setItem("tenant", "fake-tenant");
  setActivePinia(createPinia());
  const publicKeysStore = usePublicKeysStore();
  const mockSshApi = new MockAdapter(sshApi.getAxios());

  mockSshApi.onGet("http://localhost:3000/api/sshkeys/public-keys?page=1&per_page=10").reply(200, [mockPublicKey], { "x-total-count": 1 });

  beforeEach(() => {
    publicKeysStore.publicKeys = [];
  });

  it("should have initial state values", () => {
    expect(publicKeysStore.publicKeys).toEqual([]);
    expect(publicKeysStore.publicKeyCount).toEqual(0);
  });

  it("should fetch public keys successfully", async () => {
    await publicKeysStore.fetchPublicKeyList();
    expect(publicKeysStore.publicKeys).toEqual([mockPublicKey]);
    expect(publicKeysStore.publicKeyCount).toEqual([mockPublicKey].length);
  });

  it("should create public key successfully", async () => {
    mockSshApi.onPost("http://localhost:3000/api/sshkeys/public-keys").reply(200);

    const storeSpy = vi.spyOn(publicKeysStore, "createPublicKey");
    await publicKeysStore.createPublicKey(mockPublicKey);

    expect(storeSpy).toHaveBeenCalledWith(mockPublicKey);
  });

  it("should update public key successfully", async () => {
    const updatedKey = { ...mockPublicKey, name: "updated-name" };

    mockSshApi.onPut(`http://localhost:3000/api/sshkeys/public-keys/${updatedKey.fingerprint}`).reply(200);

    const storeSpy = vi.spyOn(publicKeysStore, "updatePublicKey");
    await publicKeysStore.updatePublicKey(updatedKey);
    expect(storeSpy).toHaveBeenCalledWith(updatedKey);
  });

  it("should delete public key successfully", async () => {
    publicKeysStore.publicKeys = [mockPublicKey];
    const { fingerprint } = mockPublicKey;

    mockSshApi.onDelete(`http://localhost:3000/api/sshkeys/public-keys/${encodeURIComponent(fingerprint)}`).reply(200);

    await publicKeysStore.deletePublicKey(fingerprint);

    expect(publicKeysStore.publicKeys).toEqual([]);
    expect(publicKeysStore.publicKeyCount).toEqual(0);
  });
});
