import { describe, expect, it } from "vitest";
import { store } from "../../../src/store";

describe("publicKeys", () => {
  const numberPublicKeys = 2;

  const publicKeys = [
    {
      id: "02143",
      data: "BBGVvbmFyZG8=",
      fingerprint: "b8:26:d5",
      created_at: "2020-11-23T20:59:13.323Z",
      tenant_id: "00000000-0000-4000-0000-000000000000",
      name: "shellhub",
    },
    {
      id: "07845",
      data: "AbGVvbmFyZG8=",
      fingerprint: "b7:25:f8",
      created_at: "2020-11-23T20:59:13.323Z",
      tenant_id: "00000000-0000-4000-0000-000000000000",
      name: "shellhub",
    },
  ];

  const publicKey = {
    id: "078954",
    data: "AbGVvbmFyZG8=",
    fingerprint: "b7:25:f8",
    created_at: "2020-11-23T20:59:13.323Z",
    tenant_id: "00000000-0000-4000-0000-000000000000",
    name: "shellhub",
  };

  const pagePerpageInitialValue = {
    page: 1,
    perPage: 10,
  };

  const pagePerpageValue = {
    page: 2,
    perPage: 50,
  };

  it("Return public key default variables", () => {
    expect(store.getters["publicKeys/list"]).toEqual([]);
    expect(store.getters["publicKeys/get"]).toEqual({});
    expect(store.getters["publicKeys/getNumberPublicKeys"]).toEqual(0);
    expect(store.getters["publicKeys/getPage"]).toEqual(
      pagePerpageInitialValue.page,
    );
    expect(store.getters["publicKeys/getPerPage"]).toEqual(
      pagePerpageInitialValue.perPage,
    );
  });
  it("Verify initial state change for setPublicKeys mutation", () => {
    store.commit("publicKeys/setPublicKeys", {
      data: publicKeys,
      headers: { "x-total-count": numberPublicKeys },
    });
    expect(store.getters["publicKeys/list"]).toEqual(publicKeys);
    expect(store.getters["publicKeys/getNumberPublicKeys"]).toEqual(
      numberPublicKeys,
    );
  });
  it("Verify inital state change for setPublicKey mutation", () => {
    store.commit("publicKeys/setPublicKey", { data: publicKey });
    expect(store.getters["publicKeys/get"]).toEqual(publicKey);
  });
  it("Verify inital state change for setPagePerpage mutation", () => {
    store.commit("publicKeys/setPagePerpage", pagePerpageValue);
    expect(store.getters["publicKeys/getPage"]).toEqual(pagePerpageValue.page);
    expect(store.getters["publicKeys/getPerPage"]).toEqual(
      pagePerpageValue.perPage,
    );
  });
  it("Verify inital state change for resetPagePerpage mutation", () => {
    store.commit("publicKeys/resetPagePerpage");
    expect(store.getters["publicKeys/getPage"]).toEqual(
      pagePerpageInitialValue.page,
    );
    expect(store.getters["publicKeys/getPerPage"]).toEqual(
      pagePerpageInitialValue.perPage,
    );
  });
  it("Verify remove public key item from list for removePublicKey mutation", () => {
    store.commit("publicKeys/removePublicKey", publicKey.id);
    expect(store.getters["publicKeys/list"].length).toEqual(
      numberPublicKeys - 1,
    );
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
