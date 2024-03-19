import { afterEach, describe, expect, it } from "vitest";
import { store } from "@/store";

describe("PrivateKey Vuex Module", () => {
  afterEach(() => {
    localStorage.clear();
  });

  it("Returns private keys with default variables", () => {
    expect(store.getters["privateKey/list"]).toEqual([]);
    expect(store.getters["privateKey/getNumberPrivateKeys"]).toEqual(0);
  });

  it("Commits fetchPrivateKey mutation", async () => {
    const privateKeys = [{ id: 1, data: "data1", name: "name1" }];
    store.commit("privateKey/fetchPrivateKey", privateKeys);
    expect(store.getters["privateKey/list"]).toEqual(privateKeys);
    expect(store.getters["privateKey/getNumberPrivateKeys"]).toEqual(privateKeys.length);
  });

  it("Commits setPrivateKey mutation", async () => {
    const privateKey = { id: 2, data: "data2", name: "name2" };
    store.commit("privateKey/setPrivateKey", privateKey);
    expect(store.getters["privateKey/list"]).toContain(privateKey);
    expect(store.getters["privateKey/getNumberPrivateKeys"]).toEqual(store.getters["privateKey/list"].length);
  });

  it("Commits editPrivateKey mutation", async () => {
    const privateKey = { id: 1, data: "data1-updated", name: "name1-updated" };
    store.commit("privateKey/editPrivateKey", privateKey);
    expect(store.getters["privateKey/list"]).toContainEqual(privateKey);
  });

  it("Commits removePrivateKey mutation", async () => {
    const id = 1;
    store.commit("privateKey/removePrivateKey", id);
    expect(store.getters["privateKey/list"]).not.toContainEqual({ id });
    expect(store.getters["privateKey/getNumberPrivateKeys"]).toEqual(store.getters["privateKey/list"].length);
  });

  it("Fetches private keys from localStorage", async () => {
    const privateKeys = [{ id: 1, data: "data1", name: "name1" }];
    localStorage.setItem("privateKeys", JSON.stringify(privateKeys));
    await store.dispatch("privateKey/fetch");
    expect(store.getters["privateKey/list"]).toEqual(privateKeys);
    expect(store.getters["privateKey/getNumberPrivateKeys"]).toEqual(privateKeys.length);
  });

  it("Sets a private key in localStorage", async () => {
    const privateKey = { data: "data1", name: "name1" };
    await store.dispatch("privateKey/set", privateKey);
    await store.dispatch("privateKey/fetch");
    const privateKeys = localStorage.getItem("privateKeys");
    let parsedPrivateKeys = [];
    if (privateKeys !== null) {
      parsedPrivateKeys = JSON.parse(privateKeys);
    }
    expect(parsedPrivateKeys).toContainEqual(privateKey);
    expect(store.getters["privateKey/list"]).toContainEqual(privateKey);
    expect(store.getters["privateKey/getNumberPrivateKeys"]).toEqual(parsedPrivateKeys.length);
  });

  it("Edits a private key in localStorage", async () => {
    const privateKey = { id: 1, data: "data1-updated", name: "name1-updated" };
    await store.dispatch("privateKey/edit", privateKey);
    const privateKeys = localStorage.getItem("privateKeys");
    let parsedPrivateKeys = [];
    if (privateKeys !== null) {
      parsedPrivateKeys = JSON.parse(privateKeys);
    }
    expect(parsedPrivateKeys).toContainEqual(privateKey);
    expect(store.getters["privateKey/list"]).toContainEqual(privateKey);
  });

  it("Removes a private key from localStorage", async () => {
    const id = 1;
    await store.dispatch("privateKey/remove", id);
    const privateKeys = localStorage.getItem("privateKeys");
    let parsedPrivateKeys = [];
    if (privateKeys !== null) {
      parsedPrivateKeys = JSON.parse(privateKeys);
    }
    expect(parsedPrivateKeys).not.toContainEqual({ id });
    expect(store.getters["privateKey/list"]).not.toContainEqual({ id });
    expect(store.getters["privateKey/getNumberPrivateKeys"]).toEqual(parsedPrivateKeys.length);
  });
});
