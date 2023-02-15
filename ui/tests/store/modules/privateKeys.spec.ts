import { describe, expect, it, vi } from "vitest";
import { store } from "../../../src/store";

describe("PrivateKeys", () => {
  const numberPrivateKeys = 3;

  const privateKeys = [
    {
      name: "key1",
      data: "BBGVvbmFyZG8=",
    },
    {
      name: "key2",
      data: "AbGVvbmFyZG8=",
    },
    {
      name: "key3",
      data: "CbGVvbmFyZG8=",
    },
  ];
  const privateKey = {
    name: "key4",
    data: "AbGVvbmFyZG7=",
  };
  const privateKey2 = {
    name: "key2",
    data: "AbGVvbmFyZG8=",
  };

  it("Return private key default variables", () => {
    expect(store.getters["privateKey/list"]).toEqual([]);
    expect(store.getters["privateKey/getNumberPrivateKeys"]).toEqual(0);
  });

  it("Verify inital state change for setPrivateKey mutation", () => {
    store.commit("privateKey/setPrivateKey", privateKey);
    expect(store.getters["privateKey/list"]).toEqual([privateKey]);
    expect(store.getters["privateKey/getNumberPrivateKeys"]).toEqual(1);
  });
  it("Verify remove private key item from list for removePrivateKey mutation", () => {
    const currentPrivateKeys = store.getters["privateKey/list"];
    const currentNumberPrivateKeys =
      store.getters["privateKey/getNumberPrivateKeys"];

    store.commit("privateKey/removePrivateKey", privateKey2.data);
    expect(store.getters["privateKey/list"]).toEqual(
      currentPrivateKeys.filter((pk) => pk.data !== privateKey2.data)
    );
    expect(store.getters["privateKey/getNumberPrivateKeys"]).toEqual(
      currentNumberPrivateKeys - 1
    );
  });

  it("Verify fetch private key item from list for fetchPrivateKey mutation", () => {
    store.commit("privateKey/fetchPrivateKey", [privateKey2]);
    expect(store.getters["privateKey/list"]).toEqual([privateKey2]);
    expect(store.getters["privateKey/getNumberPrivateKeys"]).toEqual(1);
  });

  // TODO: edit private Key test after the types is defined
});
