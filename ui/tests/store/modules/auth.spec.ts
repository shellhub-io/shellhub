import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { flushPromises } from "@vue/test-utils";
import { store } from "@/store";
import { mfaApi, usersApi, apiKeysApi } from "@/api/http";

describe("Auth", () => {
  let mock: MockAdapter;
  let mockUser: MockAdapter;
  let mockApiKeys: MockAdapter;

  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant");
    mock = new MockAdapter(mfaApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockApiKeys = new MockAdapter(apiKeysApi.getAxios());
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    mock.reset();
  });

  it("Return authentication default variables", () => {
    expect(store.getters["auth/link_mfa"]).toEqual("");
    expect(store.getters["auth/mfaStatus"]).toEqual({ enable: false, validate: false });
    expect(store.getters["auth/recoveryCodes"]).toEqual([]);
    expect(store.getters["auth/secret"]).toEqual("");
    expect(store.getters["auth/showRecoveryModal"]).toEqual(false);
    expect(store.getters["auth/getSortStatusField"]).toEqual(undefined);
    expect(store.getters["auth/getSortStatusString"]).toEqual("asc");
    expect(store.getters["auth/apiKey"]).toEqual("");
    expect(store.getters["auth/apiKeyList"]).toEqual([]);
    expect(store.getters["auth/getNumberApiKeys"]).toEqual(0);
  });

  it("Test disableMfa action", async () => {
    const reqSpy = vi.spyOn(store, "dispatch");

    // Mock the API call for disabling MFA
    mock.onPost("http://localhost:3000/api/mfa/disable").reply(200);

    // Trigger the disableMfa action
    await store.dispatch("auth/disableMfa");

    // Check if the state has been updated correctly
    expect(reqSpy).toHaveBeenCalled();
    // Check other related state values as needed
    expect(store.getters["auth/isMfa"]).toEqual(false);
  });

  it("Test enableMfa action", async () => {
    // Mock the API call for enabling MFA
    const enableMfaResponse = {
      token: "token",
    };

    const enableMfaData = {
      token_mfa: "000000",
      secret: "OYDXN4MO2S2JTASNBG5AD54FVT7A5GVH",
      codes: [
        "HW2wlxV40B",
        "2xsmMUHHHb",
        "DTQgVsaVac",
        "KXPBoXvuWD",
        "QQYTPfotBi",
        "XWiKBEPyb4",
      ],
    };

    const reqSpy = vi.spyOn(store, "dispatch");

    mock.onPost("http://localhost:3000/api/mfa/enable").reply(200, enableMfaResponse);

    // Trigger the enableMfa action
    await store.dispatch("auth/enableMfa", enableMfaData);
    await flushPromises();
    // Check if the state has been updated correctly
    expect(reqSpy).toHaveBeenCalledWith("auth/enableMfa", {
      token_mfa: "000000",
      secret: "OYDXN4MO2S2JTASNBG5AD54FVT7A5GVH",
      codes: [
        "HW2wlxV40B",
        "2xsmMUHHHb",
        "DTQgVsaVac",
        "KXPBoXvuWD",
        "QQYTPfotBi",
        "XWiKBEPyb4",
      ],
    });
  });

  it("Test validateMfa action", async () => {
    // Mock the API call for validating MFA
    const validateMfaResponse = {
      token: "token",
    };
    const validateMfaData = { code: "000000" };

    const reqSpy = vi.spyOn(store, "dispatch");

    mock.onPost("http://localhost:3000/api/mfa/auth").reply(200, validateMfaResponse);

    // Trigger the validateMfa action
    await store.dispatch("auth/validateMfa", validateMfaData);
    await flushPromises();

    expect(reqSpy).toHaveBeenCalledWith("auth/validateMfa", { code: "000000" });
    // Check if the state has been updated correctly
    expect(store.getters["auth/stateToken"]).toEqual(validateMfaResponse.token);
  });

  it("Test recoveryMfa action", async () => {
    // Mock the API call for recovering MFA
    const recoveryMfaResponse = {
      token: "token",
    };
    const recoveryMfaData = { code: "000000" };

    const reqSpy = vi.spyOn(store, "dispatch");

    mock.onPost("http://localhost:3000/api/mfa/recovery").reply(200, recoveryMfaResponse);

    // Trigger the recoveryMfa action
    await store.dispatch("auth/recoverLoginMfa", recoveryMfaData);
    await flushPromises();

    expect(reqSpy).toHaveBeenCalledWith("auth/recoverLoginMfa", { code: "000000" });

    // Check if the state has been updated correctly
    expect(store.getters["auth/stateToken"]).toEqual(recoveryMfaResponse.token);
    expect(store.getters["auth/showRecoveryModal"]).toEqual(true);
  });

  it("Test generateMfa action", async () => {
    // Mock the API call for generating MFA
    const generateMfaResponse = {
      secret: "secret-mfa",
      link: "link-mfa",
      codes: [
        "HW2wlxV40B",
        "2xsmMUHHHb",
        "DTQgVsaVac",
        "KXPBoXvuWD",
        "QQYTPfotBi",
        "XWiKBEPyb4",
      ],
    };

    const reqSpy = vi.spyOn(store, "dispatch");

    mock.onGet("http://localhost:3000/api/mfa/generate").reply(200, generateMfaResponse);

    // Trigger the generateMfa action
    await store.dispatch("auth/generateMfa");

    expect(reqSpy).toHaveBeenCalledWith("auth/generateMfa");

    // Check if the state has been updated correctly
    expect(store.getters["auth/link_mfa"]).toEqual(generateMfaResponse.link);
    expect(store.getters["auth/secret"]).toEqual(generateMfaResponse.secret);
    expect(store.getters["auth/recoveryCodes"]).toEqual(generateMfaResponse.codes);
  });

  it("Test getUserStatus action", async () => {
    // Mock the API call for getting MFA status
    const getUserStatusResponse = {
      mfa: true,
      token: "token",
      id: "userId",
      user: "username",
      name: "testname",
      email: "test@test.com",
      tenant: "fake-tenant",
      role: "administrator",
    };

    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, getUserStatusResponse);

    const reqSpy = vi.spyOn(store, "dispatch");

    // Trigger the getMfaStatus action
    await store.dispatch("auth/getUserInfo");

    expect(reqSpy).toHaveBeenCalledWith("auth/getUserInfo");

    // Check if the state has been updated correctly
    expect(store.getters["auth/stateToken"]).toEqual(getUserStatusResponse.token);
    expect(store.getters["auth/currentUser"]).toEqual(getUserStatusResponse.user);
    expect(store.getters["auth/currentName"]).toEqual(getUserStatusResponse.name);
    expect(store.getters["auth/tenant"]).toEqual(getUserStatusResponse.tenant);
    expect(store.getters["auth/email"]).toEqual(getUserStatusResponse.email);
    expect(store.getters["auth/id"]).toEqual(getUserStatusResponse.id);
    expect(store.getters["auth/role"]).toEqual(getUserStatusResponse.role);
    expect(store.getters["auth/mfaStatus"]).toEqual(getUserStatusResponse.mfa);
  });

  it("Test GenerateApiKey action", async () => {
    // Mock the API call for GenerateApiKey
    const generateApiResponse = {
      key: "fake-api-key",
    };
    const generateApiData = {
      tenant: "fake-tenant",
      name: "my api key",
      expires_at: 30,
    };

    const reqSpy = vi.spyOn(store, "dispatch");

    mockApiKeys.onPost("http://localhost:3000/api/namespaces/fake-tenant/api-key").reply(200, generateApiResponse);

    // Trigger the GenerateApiKey action
    await store.dispatch("auth/generateApiKey", generateApiData);
    await flushPromises();

    expect(reqSpy).toHaveBeenCalledWith("auth/generateApiKey", generateApiData);
    // Check if the state has been updated correctly
    expect(store.getters["auth/apiKey"]).toEqual(generateApiResponse);
  });

  it("Test Get Api Keys action", async () => {
    // Mock the API call for getApiKey
    const getApiResponse = [
      {
        id: "3e5a5194-9dec-4a32-98db-7434c6d49df1",
        tenant_id: "fake-tenant",
        user_id: "507f1f77bcf86cd799439011",
        name: "my api key",
        expires_in: 1707958989,
      },
    ];

    const getApiData = {
      tenant: "fake-tenant",
    };

    const reqSpy = vi.spyOn(store, "dispatch");

    mockApiKeys.onGet("http://localhost:3000/api/namespaces/fake-tenant/api-key").reply(200, getApiResponse, { "x-total-count": 1 });

    // Trigger the getApiKey action
    await store.dispatch("auth/getApiKey", getApiData);
    await flushPromises();

    expect(reqSpy).toHaveBeenCalledWith("auth/getApiKey", getApiData);
    // Check if the state has been updated correctly
    expect(store.getters["auth/apiKeyList"]).toEqual(getApiResponse);
    expect(store.getters["auth/getNumberApiKeys"]).toEqual(1);
  });

  it("Test editApiKey action", async () => {
    // Mock the API call for editApiKey
    const editApiResponse = {
      id: "fake-api-key",
      tenant_id: "fake-tenant",
      user_id: "507f1f77bcf86cd799439011",
      name: "my api key 2",
      expires_in: 1707958989,
    };

    const editApiData = {
      tenant: "fake-tenant",
      id: "fake-api-key",
      name: "my api key 2",
    };

    const reqSpy = vi.spyOn(store, "dispatch");

    mockApiKeys.onPatch("http://localhost:3000/api/namespaces/fake-tenant/api-key/fake-api-key").reply(200, editApiResponse);

    // Trigger the editApiKey action
    await store.dispatch("auth/editApiKey", editApiData);
    await flushPromises();

    expect(reqSpy).toHaveBeenCalledWith("auth/editApiKey", editApiData);
  });

  it("Test removeApiKey action", async () => {
    // Mock the API call for removeApiKey
    const removeApiData = {
      tenant: "fake-tenant",
      id: "fake-api-key",
    };

    const reqSpy = vi.spyOn(store, "dispatch");

    mockApiKeys.onDelete("http://localhost:3000/api/namespaces/fake-tenant/api-key/fake-api-key").reply(200);

    // Trigger the editApiKey action
    await store.dispatch("auth/removeApiKey", removeApiData);
    await flushPromises();

    expect(reqSpy).toHaveBeenCalledWith("auth/removeApiKey", removeApiData);
  });
});
