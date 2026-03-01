import { describe, it, expect, beforeEach } from "vitest";
import {
  deriveKey,
  encrypt,
  decrypt,
  createVaultMeta,
  verifyPassword,
  setSessionKey,
  getSessionKey,
  clearSessionKey,
  VAULT_META_KEY,
  VAULT_DATA_KEY,
  LEGACY_KEYS_KEY,
} from "../vault-crypto";
import type { VaultMeta, VaultData } from "../../types/vault";


/** Generate a random 16-byte salt synchronously using the test environment's
 *  Web Crypto (provided by jsdom / @vitest/browser). */
function randomSalt(): Uint8Array {
  return crypto.getRandomValues(new Uint8Array(16));
}

/** Build a minimal but valid VaultMeta by running the real creation path with
 *  a reduced iteration count so tests stay fast. */
async function buildMeta(
  password: string,
  iterations = 1,
): Promise<VaultMeta> {
  const salt = randomSalt();
  const key = await deriveKey(password, salt, iterations);

  const encoder = new TextEncoder();
  const verifierPlaintext = "shellhub-vault-ok";
  const iv = crypto.getRandomValues(new Uint8Array(12));

  const cipherBuf = await crypto.subtle.encrypt(
    { name: "AES-GCM", iv },
    key,
    encoder.encode(verifierPlaintext),
  );

  function toBase64(buf: ArrayBuffer | Uint8Array): string {
    const bytes = buf instanceof Uint8Array ? buf : new Uint8Array(buf);
    let s = "";
    for (let i = 0; i < bytes.byteLength; i++) s += String.fromCharCode(bytes[i]);
    return btoa(s);
  }

  return {
    version: 1,
    salt: toBase64(salt),
    iterations,
    verifier: toBase64(cipherBuf),
    verifierIv: toBase64(iv),
  };
}

describe("exported constants", () => {
  it("exports VAULT_META_KEY", () => {
    expect(VAULT_META_KEY).toBe("shellhub-vault-meta");
  });

  it("exports VAULT_DATA_KEY", () => {
    expect(VAULT_DATA_KEY).toBe("shellhub-vault-data");
  });

  it("exports LEGACY_KEYS_KEY", () => {
    expect(LEGACY_KEYS_KEY).toBe("privateKeys");
  });
});

describe("session key management", () => {
  beforeEach(() => {
    clearSessionKey();
  });

  it("getSessionKey returns null before anything is set", () => {
    expect(getSessionKey()).toBeNull();
  });

  it("setSessionKey stores the key and getSessionKey returns it", async () => {
    const key = await crypto.subtle.generateKey(
      { name: "AES-GCM", length: 256 },
      false,
      ["encrypt", "decrypt"],
    );

    setSessionKey(key);

    expect(getSessionKey()).toBe(key);
  });

  it("clearSessionKey sets the session key back to null", async () => {
    const key = await crypto.subtle.generateKey(
      { name: "AES-GCM", length: 256 },
      false,
      ["encrypt", "decrypt"],
    );

    setSessionKey(key);
    clearSessionKey();

    expect(getSessionKey()).toBeNull();
  });

  it("setSessionKey overwrites a previously set key", async () => {
    const key1 = await crypto.subtle.generateKey(
      { name: "AES-GCM", length: 256 },
      false,
      ["encrypt", "decrypt"],
    );
    const key2 = await crypto.subtle.generateKey(
      { name: "AES-GCM", length: 256 },
      false,
      ["encrypt", "decrypt"],
    );

    setSessionKey(key1);
    setSessionKey(key2);

    expect(getSessionKey()).toBe(key2);
  });
});

// deriveKey

describe("deriveKey", () => {
  it("returns a CryptoKey", async () => {
    const key = await deriveKey("password", randomSalt(), 1);

    expect(key).toBeDefined();
    expect(key.type).toBe("secret");
  });

  it("returns a key usable for AES-GCM encrypt and decrypt", async () => {
    const key = await deriveKey("password", randomSalt(), 1);

    expect(key.usages).toContain("encrypt");
    expect(key.usages).toContain("decrypt");
  });

  it("produces the same key from the same password and salt", async () => {
    const salt = randomSalt();
    const key1 = await deriveKey("stable-password", salt, 1);
    const key2 = await deriveKey("stable-password", salt, 1);

    // Export both and compare bytes
    const raw1 = await crypto.subtle.exportKey("raw", key1).catch(() => null);
    const raw2 = await crypto.subtle.exportKey("raw", key2).catch(() => null);

    // Keys are non-extractable by design; verify they behave identically
    // by encrypting with one and decrypting with the other.
    const plaintext = new TextEncoder().encode("test");
    const iv = crypto.getRandomValues(new Uint8Array(12));

    const ciphertext = await crypto.subtle.encrypt({ name: "AES-GCM", iv }, key1, plaintext);
    const decrypted = await crypto.subtle.decrypt({ name: "AES-GCM", iv }, key2, ciphertext);

    expect(new TextDecoder().decode(decrypted)).toBe("test");

    // Raw export is always null for non-extractable keys — just confirm it fails gracefully
    expect(raw1).toBeNull();
    expect(raw2).toBeNull();
  });

  it("produces a different key for a different password", async () => {
    const salt = randomSalt();
    const key1 = await deriveKey("password-A", salt, 1);
    const key2 = await deriveKey("password-B", salt, 1);

    const plaintext = new TextEncoder().encode("hello");
    const iv = crypto.getRandomValues(new Uint8Array(12));

    const ciphertext = await crypto.subtle.encrypt({ name: "AES-GCM", iv }, key1, plaintext);

    await expect(
      crypto.subtle.decrypt({ name: "AES-GCM", iv }, key2, ciphertext),
    ).rejects.toThrow();
  });

  it("produces a different key for a different salt", async () => {
    const key1 = await deriveKey("password", randomSalt(), 1);
    const key2 = await deriveKey("password", randomSalt(), 1);

    const plaintext = new TextEncoder().encode("hello");
    const iv = crypto.getRandomValues(new Uint8Array(12));

    const ciphertext = await crypto.subtle.encrypt({ name: "AES-GCM", iv }, key1, plaintext);

    await expect(
      crypto.subtle.decrypt({ name: "AES-GCM", iv }, key2, ciphertext),
    ).rejects.toThrow();
  });

  it("accepts a custom iteration count", async () => {
    // Just verify it does not throw and returns a usable key
    const key = await deriveKey("password", randomSalt(), 500);
    expect(key.type).toBe("secret");
  });
});

// encrypt / decrypt round-trip

describe("encrypt and decrypt", () => {
  async function makeKey(): Promise<CryptoKey> {
    return deriveKey("test-password", randomSalt(), 1);
  }

  it("encrypt returns an object with iv and ciphertext fields", async () => {
    const key = await makeKey();
    const result = await encrypt(key, "hello world");

    expect(result).toHaveProperty("iv");
    expect(result).toHaveProperty("ciphertext");
    expect(typeof result.iv).toBe("string");
    expect(typeof result.ciphertext).toBe("string");
  });

  it("iv and ciphertext are valid base64 strings", async () => {
    const key = await makeKey();
    const result = await encrypt(key, "hello");

    expect(() => atob(result.iv)).not.toThrow();
    expect(() => atob(result.ciphertext)).not.toThrow();
  });

  it("decrypt recovers the original plaintext", async () => {
    const key = await makeKey();
    const plaintext = "my secret message";

    const data = await encrypt(key, plaintext);
    const recovered = await decrypt(key, data);

    expect(recovered).toBe(plaintext);
  });

  it("round-trips an empty string", async () => {
    const key = await makeKey();
    const data = await encrypt(key, "");
    expect(await decrypt(key, data)).toBe("");
  });

  it("round-trips a multi-line string with special characters", async () => {
    const key = await makeKey();
    const value = "line1\nline2\ttabbed\u00e9\u00e0";
    const data = await encrypt(key, value);
    expect(await decrypt(key, data)).toBe(value);
  });

  it("uses a fresh IV on each encrypt call", async () => {
    const key = await makeKey();
    const { iv: iv1 } = await encrypt(key, "same text");
    const { iv: iv2 } = await encrypt(key, "same text");

    // With 96-bit random IVs collisions are practically impossible
    expect(iv1).not.toBe(iv2);
  });

  it("decrypt throws when ciphertext is tampered", async () => {
    const key = await makeKey();
    const data = await encrypt(key, "secret");

    // Flip the last character of the base64-encoded ciphertext
    const tampered: VaultData = {
      ...data,
      ciphertext: data.ciphertext.slice(0, -1) + (data.ciphertext.endsWith("A") ? "B" : "A"),
    };

    await expect(decrypt(key, tampered)).rejects.toThrow();
  });

  it("decrypt throws when iv is tampered", async () => {
    const key = await makeKey();
    const data = await encrypt(key, "secret");

    const tampered: VaultData = {
      ...data,
      iv: data.iv.slice(0, -1) + (data.iv.endsWith("A") ? "B" : "A"),
    };

    await expect(decrypt(key, tampered)).rejects.toThrow();
  });

  it("decrypt throws when iv is invalid base64", async () => {
    const key = await makeKey();
    const data = await encrypt(key, "secret");

    const tampered: VaultData = { ...data, iv: "!!!not-base64!!!" };

    await expect(decrypt(key, tampered)).rejects.toThrow();
  });

  it("decrypt throws when ciphertext is invalid base64", async () => {
    const key = await makeKey();
    const data = await encrypt(key, "secret");

    const tampered: VaultData = { ...data, ciphertext: "!!!not-base64!!!" };

    await expect(decrypt(key, tampered)).rejects.toThrow();
  });

  it("decrypt throws when the wrong key is used", async () => {
    const key1 = await makeKey();
    const key2 = await makeKey();

    const data = await encrypt(key1, "secret");

    await expect(decrypt(key2, data)).rejects.toThrow();
  });
});

// createVaultMeta

describe("createVaultMeta", () => {
  it("returns a meta object and a derived key", async () => {
    const { meta, derivedKey } = await createVaultMeta("my-password");

    expect(meta).toBeDefined();
    expect(derivedKey).toBeDefined();
    expect(derivedKey.type).toBe("secret");
  });

  it("meta has version 1", async () => {
    const { meta } = await createVaultMeta("pass");
    expect(meta.version).toBe(1);
  });

  it("meta contains salt, verifier, verifierIv as valid base64", async () => {
    const { meta } = await createVaultMeta("pass");

    expect(() => atob(meta.salt)).not.toThrow();
    expect(() => atob(meta.verifier)).not.toThrow();
    expect(() => atob(meta.verifierIv)).not.toThrow();
  });

  it("meta.iterations equals 600000", async () => {
    const { meta } = await createVaultMeta("pass");
    expect(meta.iterations).toBe(600_000);
  });

  it("different calls produce different salts", async () => {
    const { meta: m1 } = await createVaultMeta("pass");
    const { meta: m2 } = await createVaultMeta("pass");

    expect(m1.salt).not.toBe(m2.salt);
  });

  it("the returned derivedKey can encrypt and decrypt data", async () => {
    const { derivedKey } = await createVaultMeta("pass");
    const data = await encrypt(derivedKey, "hello");
    expect(await decrypt(derivedKey, data)).toBe("hello");
  });
});

// verifyPassword

describe("verifyPassword", () => {
  it("returns the derived key when password matches", async () => {
    const meta = await buildMeta("correct-password");
    const key = await verifyPassword("correct-password", meta);

    expect(key).toBeDefined();
    expect(key.type).toBe("secret");
  });

  it("the returned key can decrypt data encrypted during vault creation", async () => {
    const password = "my-vault-password";
    const { meta, derivedKey: originalKey } = await createVaultMeta(password);
    const encrypted = await encrypt(originalKey, "stored value");

    // Simulate unlocking after a page reload — re-derive the key from meta
    const unlockedKey = await verifyPassword(password, meta);
    const decrypted = await decrypt(unlockedKey, encrypted);

    expect(decrypted).toBe("stored value");
  });

  it("throws when the wrong password is supplied", async () => {
    const meta = await buildMeta("correct-password");

    await expect(verifyPassword("wrong-password", meta)).rejects.toThrow();
  });

  it("throws when meta is tampered (verifier flipped)", async () => {
    const meta = await buildMeta("correct-password");

    const tampered: VaultMeta = {
      ...meta,
      verifier: meta.verifier.slice(0, -1) + (meta.verifier.endsWith("A") ? "B" : "A"),
    };

    await expect(verifyPassword("correct-password", tampered)).rejects.toThrow();
  });

  it("throws when verifierIv is tampered", async () => {
    const meta = await buildMeta("correct-password");

    const tampered: VaultMeta = {
      ...meta,
      verifierIv: meta.verifierIv.slice(0, -1) + (meta.verifierIv.endsWith("A") ? "B" : "A"),
    };

    await expect(verifyPassword("correct-password", tampered)).rejects.toThrow();
  });

  it("throws when salt is replaced (wrong derived key)", async () => {
    const meta = await buildMeta("correct-password");

    const newSalt = crypto.getRandomValues(new Uint8Array(16));
    function toBase64(buf: Uint8Array): string {
      let s = "";
      for (let i = 0; i < buf.byteLength; i++) s += String.fromCharCode(buf[i]);
      return btoa(s);
    }

    const tampered: VaultMeta = { ...meta, salt: toBase64(newSalt) };

    await expect(verifyPassword("correct-password", tampered)).rejects.toThrow();
  });

  it("respects the iterations field stored in meta", async () => {
    // Build meta with a non-default iteration count and verify it still works
    const meta = await buildMeta("pass", 2);
    expect(meta.iterations).toBe(2);

    const key = await verifyPassword("pass", meta);
    expect(key.type).toBe("secret");
  });
});
