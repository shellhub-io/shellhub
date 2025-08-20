import { describe, expect, it, beforeEach } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import usePrivateKeysStore from "@/store/modules/private_keys";

describe("PrivateKey Store", () => {
  setActivePinia(createPinia());
  const privateKeysStore = usePrivateKeysStore();

  beforeEach(() => {
    localStorage.clear();
    privateKeysStore.privateKeys = [];
  });

  it("Returns private keys with default variables", () => {
    expect(privateKeysStore.privateKeys).toEqual([]);
  });

  it("Gets private keys from localStorage", () => {
    const privateKeys = [{ id: 1, data: "data1", name: "name1", hasPassphrase: false, fingerprint: "fp1" }];
    localStorage.setItem("privateKeys", JSON.stringify(privateKeys));
    privateKeysStore.getPrivateKeyList();
    expect(privateKeysStore.privateKeys).toEqual(privateKeys);
  });

  it("Adds a private key", () => {
    const privateKey = { data: "data1", name: "name1", hasPassphrase: false, fingerprint: "fp1" };

    privateKeysStore.addPrivateKey(privateKey);
    privateKeysStore.getPrivateKeyList();

    expect(privateKeysStore.privateKeys).toContainEqual({ ...privateKey, id: 1 });
  });

  it("Edits a private key", () => {
    const privateKey = { id: 1, data: "data1-updated", name: "name1-updated", hasPassphrase: true, fingerprint: "fp1-updated" };

    privateKeysStore.editPrivateKey(privateKey);

    const privateKeys = JSON.parse(localStorage.getItem("privateKeys") as string);
    expect(privateKeys).toContainEqual(privateKey);
    expect(privateKeysStore.privateKeys).toContainEqual(privateKey);
  });

  it("Removes a private key from localStorage", () => {
    const id = 1;
    privateKeysStore.deletePrivateKey(id);
    const privateKeys = JSON.parse(localStorage.getItem("privateKeys") as string);

    expect(privateKeys).not.toContainEqual({ id });
    expect(privateKeysStore.privateKeys).not.toContainEqual({ id });
  });
});
