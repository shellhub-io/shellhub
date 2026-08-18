import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  createWebSshSession,
  getBillingPortal,
  getSamlReauthUrl,
} from "../unspeccedRoutes";

const mockGet = vi.fn();
const mockPost = vi.fn();

vi.mock("@/client/client.gen", () => ({
  client: {
    get: (...args: unknown[]): unknown => mockGet(...args),
    post: (...args: unknown[]): unknown => mockPost(...args),
  },
}));

beforeEach(() => {
  vi.clearAllMocks();
});

describe("getBillingPortal", () => {
  it("posts to the billing portal route and returns its payload", async () => {
    mockPost.mockResolvedValue({ data: { url: "https://portal.example/s/1" } });

    const result = await getBillingPortal();

    expect(mockPost).toHaveBeenCalledWith({
      url: "/api/billing/portal",
      throwOnError: true,
    });
    expect(result).toEqual({ url: "https://portal.example/s/1" });
  });

  it("propagates the error the client throws", async () => {
    mockPost.mockRejectedValue(new Error("network failure"));

    await expect(getBillingPortal()).rejects.toThrow("network failure");
  });
});

describe("getSamlReauthUrl", () => {
  it("gets the SAML reauth route with the fingerprint and approval code", async () => {
    mockGet.mockResolvedValue({ data: { url: "https://idp.example/sso" } });

    const result = await getSamlReauthUrl({
      fingerprint: "SHA256:abc",
      approval_code: "code-1",
    });

    expect(mockGet).toHaveBeenCalledWith({
      url: "/api/user/saml/reauth",
      query: { fingerprint: "SHA256:abc", approval_code: "code-1" },
      throwOnError: true,
    });
    expect(result).toEqual({ url: "https://idp.example/sso" });
  });

  it("propagates the error the client throws", async () => {
    mockGet.mockRejectedValue(new Error("network failure"));

    await expect(
      getSamlReauthUrl({ fingerprint: "SHA256:abc", approval_code: "code-1" }),
    ).rejects.toThrow("network failure");
  });
});

describe("createWebSshSession", () => {
  it("posts the password credentials and returns the session token", async () => {
    mockPost.mockResolvedValue({ data: { token: "session-token" } });

    const result = await createWebSshSession({
      device: "device-uid",
      username: "root",
      password: "secret",
    });

    expect(mockPost).toHaveBeenCalledWith({
      url: "/ws/ssh/session",
      body: { device: "device-uid", username: "root", password: "secret" },
      parseAs: "json",
      throwOnError: true,
    });
    expect(result).toEqual({ token: "session-token" });
  });

  it("posts the public key alongside the fingerprint when one is given", async () => {
    mockPost.mockResolvedValue({ data: { token: "session-token" } });

    await createWebSshSession({
      device: "device-uid",
      username: "root",
      fingerprint: "SHA256:abc",
      public_key: "ssh-rsa AAAA",
    });

    expect(mockPost).toHaveBeenCalledWith({
      url: "/ws/ssh/session",
      body: {
        device: "device-uid",
        username: "root",
        fingerprint: "SHA256:abc",
        public_key: "ssh-rsa AAAA",
      },
      parseAs: "json",
      throwOnError: true,
    });
  });

  it("propagates the error the client throws", async () => {
    mockPost.mockRejectedValue(new Error("network failure"));

    await expect(
      createWebSshSession({ device: "device-uid", username: "root" }),
    ).rejects.toThrow("network failure");
  });
});
