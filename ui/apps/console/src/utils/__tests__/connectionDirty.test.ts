import { describe, it, expect } from "vitest";
import { connectionDirty } from "../connectionDirty";

const record = {
  username: "root",
  auth_method: "key",
  key_fingerprint: "SHA256:abc",
};

describe("connectionDirty", () => {
  it("is clean when the form matches the saved connection", () => {
    expect(
      connectionDirty(
        { username: "root", authMethod: "key", keyFingerprint: "SHA256:abc" },
        record,
      ),
    ).toBe(false);
  });

  it("ignores surrounding whitespace on the username", () => {
    expect(
      connectionDirty(
        {
          username: "  root  ",
          authMethod: "key",
          keyFingerprint: "SHA256:abc",
        },
        record,
      ),
    ).toBe(false);
  });

  it("is dirty when the username changes", () => {
    expect(
      connectionDirty(
        { username: "deploy", authMethod: "key", keyFingerprint: "SHA256:abc" },
        record,
      ),
    ).toBe(true);
  });

  it("is dirty when the auth method changes", () => {
    expect(
      connectionDirty(
        { username: "root", authMethod: "password", keyFingerprint: "" },
        record,
      ),
    ).toBe(true);
  });

  it("is dirty when a different key is selected", () => {
    expect(
      connectionDirty(
        { username: "root", authMethod: "key", keyFingerprint: "SHA256:xyz" },
        record,
      ),
    ).toBe(true);
  });

  it("is dirty when a manual key is pasted", () => {
    expect(
      connectionDirty(
        { username: "root", authMethod: "key", keyFingerprint: "manual" },
        record,
      ),
    ).toBe(true);
  });

  it("is not dirty when no usable key is selected yet (blank fingerprint)", () => {
    // Opening a key connection before the vault resolves the key must not read
    // as a change.
    expect(
      connectionDirty(
        { username: "root", authMethod: "key", keyFingerprint: "" },
        record,
      ),
    ).toBe(false);
  });

  it("treats an empty saved auth method as password", () => {
    expect(
      connectionDirty(
        { username: "root", authMethod: "password", keyFingerprint: "" },
        { username: "root", auth_method: "", key_fingerprint: "" },
      ),
    ).toBe(false);
  });
});
