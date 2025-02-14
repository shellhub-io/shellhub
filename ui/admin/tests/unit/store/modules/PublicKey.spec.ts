import { describe, expect, it } from "vitest";
import { store } from "../../../../src/store";

describe("PublicKeys", () => {
  const numberPublicKeys = 2;
  const publicKeys = [
    {
      data: "BBGVvbmFyZG8=",
      fingerprint: "b8:26:d5",
      created_at: "2020-11-23T20:59:13.323Z",
      tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      name: "shellhub",
    },
    {
      data: "AbGVvbmFyZG8=",
      fingerprint: "b7:25:f8",
      created_at: "2020-11-23T20:59:13.323Z",
      tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      name: "shellhub",
    },
  ];
  const publicKey = {
    data: "AbGVvbmFyZG8=",
    fingerprint: "b7:25:f8",
    created_at: "2020-11-23T20:59:13.323Z",
    tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    name: "shellhub",
  };

  it("Return public key default variables", () => {
    expect(store.getters["publicKeys/list"]).toEqual([]);
    expect(store.getters["publicKeys/get"]).toEqual({});
    expect(store.getters["publicKeys/getNumberPublicKeys"]).toEqual(0);
  });

  it("Verify initial state change for setPublicKeys mutation", () => {
    store.commit("publicKeys/setPublicKeys", {
      data: publicKeys,
      headers: { "x-total-count": numberPublicKeys },
    });
    expect(store.getters["publicKeys/list"]).toEqual(publicKeys);
    expect(store.getters["publicKeys/getNumberPublicKeys"]).toEqual(numberPublicKeys);
  });
  it("Verify inital state change for setPublicKey mutation", () => {
    store.commit("publicKeys/setPublicKey", { data: publicKey });
    expect(store.getters["publicKeys/get"]).toEqual(publicKey);
  });
  it("Verify remove public key item from list for removePublicKey mutation", () => {
    store.commit("publicKeys/removePublicKey", publicKey.tenant_id);
    expect(store.getters["publicKeys/list"].length).toEqual(numberPublicKeys - 1);
  });
  it("Verify changed public key object state for clearObjectPublicKey mutation", () => {
    store.commit("publicKeys/clearObjectPublicKey");
    expect(store.getters["publicKeys/get"]).toEqual({});
  });
  it("Verify changed firewall list state for clearListPublicKey mutation", () => {
    store.commit("publicKeys/clearListPublicKeys");
    expect(store.getters["publicKeys/list"]).toEqual([]);
    expect(store.getters["publicKeys/getNumberPublicKeys"]).toEqual(0);
  });
});
