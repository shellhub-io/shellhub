import { beforeEach, describe, expect, it } from "vitest";
import { store } from "../../../src/store";

const dataTest = {
  token: "token",
  user: "shellhub",
  name: "shellhub",
  tenant: "test",
  email: "shellgub@ossystem.com",
  id: "01",
  role: "admin",
};

describe("Auth", () => {
  beforeEach(() => {
    store.commit("auth/logout");
  });
  it("returns status", async () => {
    store.commit("auth/authSuccess", dataTest);
    const actual = await store.getters["auth/authStatus"];
    expect(actual).toEqual("success");
  });
  it("returns token", async () => {
    store.commit("auth/authSuccess", dataTest);
    const actual = await store.getters["auth/isLoggedIn"];
    expect(actual).toEqual(true);
  });
  it("returns user", async () => {
    store.commit("auth/authSuccess", dataTest);
    const actual = await store.getters["auth/currentUser"];
    expect(actual).toEqual(dataTest.user);
  });
  it("returns name", async () => {
    store.commit("auth/authSuccess", dataTest);
    const actual = await store.getters["auth/currentName"];
    expect(actual).toEqual(dataTest.name);
  });
  it("returns tenant", async () => {
    store.commit("auth/authSuccess", dataTest);
    const actual = await store.getters["auth/tenant"];
    expect(actual).toEqual(dataTest.tenant);
  });
  it("complete test", () => {
    const statusLoading = "loading";
    const statusError = "error";
    const statusSuccess = "success";
    const token = "eyJhbGciOiJSUzI1NiIsInR5c";
    const user = "user";
    const tenant = "00000000";

    store.commit("auth/authRequest");
    expect(store.getters["auth/authStatus"]).toEqual(statusLoading);

    store.commit("auth/authError");
    expect(store.getters["auth/authStatus"]).toEqual(statusError);

    store.commit("auth/authSuccess", { token, user, tenant });
    expect(store.getters["auth/authStatus"]).toEqual(statusSuccess);
    expect(store.getters["auth/isLoggedIn"]).toEqual(true);
    expect(store.getters["auth/currentUser"]).toEqual(user);
    expect(store.getters["auth/tenant"]).toEqual(tenant);

    store.commit("auth/logout");
    expect(store.getters["auth/authStatus"]).toEqual("");
    expect(store.getters["auth/isLoggedIn"]).toEqual(false);
    expect(store.getters["auth/currentUser"]).toEqual("");
    expect(store.getters["auth/tenant"]).toEqual("");
  });
});
