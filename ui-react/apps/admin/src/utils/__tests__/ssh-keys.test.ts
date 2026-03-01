import { describe, it, expect } from "vitest";
import { Buffer } from "buffer";
import {
  validatePrivateKey,
  getFingerprint,
  generateSignature,
} from "../ssh-keys";

// Real keys generated with ssh-keygen for test purposes.
// These keys are not used anywhere — they are test fixtures only.

const RSA_PRIVATE_KEY = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEAyhBNwbjqaHDTBXC+Oib7K3CM1Wr2Y14m5hCaBPvnNnuPuVLEPBG3
kc9pe9/pYdCA3t5MCI+xJcz1d2rk/REsP3f3kkwCpUZR1AAfhuB9J2sPb1oj+MnEyhu6vr
oGus8vY8VmyEyAd+CFeheudfRRnFe083NKYzPJIsIskae+whvTrr8FTC/3n543I8eigL4G
45xzi2KadH+jpdWh3j6vrxo2cimJ0d53xfhZcSOFoeJ5FYNheIff/Tohx56+YkUWVHa849
7SOVSiRV8vID0aCMM139Epip6P/lwevnkT5t64hpd1oLtt2pKUqmHXGuUTWX6Uj51hagdd
csi8/deVJQAAA8gocJc1KHCXNQAAAAdzc2gtcnNhAAABAQDKEE3BuOpocNMFcL46JvsrcI
zVavZjXibmEJoE++c2e4+5UsQ8EbeRz2l73+lh0IDe3kwIj7ElzPV3auT9ESw/d/eSTAKl
RlHUAB+G4H0naw9vWiP4ycTKG7q+uga6zy9jxWbITIB34IV6F6519FGcV7Tzc0pjM8kiwi
yRp77CG9OuvwVML/efnjcjx6KAvgbjnHOLYpp0f6Ol1aHePq+vGjZyKYnR3nfF+FlxI4Wh
4nkVg2F4h9/9OiHHnr5iRRZUdrzj3tI5VKJFXy8gPRoIwzXf0SmKno/+XB6+eRPm3riGl3
Wgu23akpSqYdca5RNZfpSPnWFqB11yyLz915UlAAAAAwEAAQAAAQBA9Cq6qkGX3yTGa+6K
bPtyhFSRSjf3y00/yXhpP7ycc+3IigWQtbVHxt/GtTyld8vlWJoiamZsm00Q3SjM/Nc3c8
3+ljaHgGpPdtarmcBBipXEmrpTgDClM3K9WrlNFcrLDKq0mnr1jhP6ImCvhCEz27HMTHRO
HVh7tuVy0PRuGsdSYcA6NR40jL4UawlFQ2yQGZgCRKWPLwXsUbSCpM2KcYR7F93oFyRV8n
aIeIk16uWOWspSy1rxQTOLkB7sgns0ywo6QdhhQfWRpuOx7m8xtDNJxEwBHHy3ScGlSejR
q6ZkCIc4eRqPQlxGjGy8pDwwg2NwueQn+J7SdOIH+wHFAAAAgHkWfOPO7WEWy+vvJAN+OZ
hstL6VcswMi0WUy0CxQmJcb9Ooj+6L5IRLU8rccsaDhM6V0Z3/w+sdfX58OoSx8NISCjyb
1uJXkasUxEM80cRUNUkk0ZGgBraBlHgpE/aP8O9OOdDPhMCrZmoI6SyNd4E/MhnCWDQBWn
PRabecsScQAAAAgQDyqqLBwueLgaroErwFLZaEDTfasG9pElTwiQ0Ai0SvVaUgJZsn4lUd
+EUtZ89GFhvpgGVZlx656PhrOi2kzMzqJluV0T2yBXj7+AjzaL7yaos13Ye7rtHomnMKzH
7Ow+6TSMr+02XM6eRGfPS0nj688jXJAEgWZsC6pJ6KYTMQOwAAAIEA1SqLnVmbiPT/831H
ekaRolDw5j0wRPx5Z6ukyOLGq5m3AtMYfJjYfVu7QlwkYvNX47fFD1WlTYpbfUKURfEsZe
Kd3kfKz4B4a79sdFupKHqodtQm31m+8udZh9XjZjTU8v95O+P+VJYp7ir8yemNvphh45UZ
6LglBt+quGp9+h8AAAAQbHVpemhmNDJAYXNwaXJlNQECAw==
-----END OPENSSH PRIVATE KEY-----`;

const ED25519_PRIVATE_KEY = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACDEA2hVbe89aWtEu5lfi1/6jwsLcQR+xSw05ml6lQ1NJwAAAJig+OfcoPjn
3AAAAAtzc2gtZWQyNTUxOQAAACDEA2hVbe89aWtEu5lfi1/6jwsLcQR+xSw05ml6lQ1NJw
AAAECFbTseFm8LyJakOYHfUn0yPPfBJulXelhjfPv1iW0JRsQDaFVt7z1pa0S7mV+LX/qP
CwtxBH7FLDTmaXqVDU0nAAAAEGx1aXpoZjQyQGFzcGlyZTUBAgMEBQ==
-----END OPENSSH PRIVATE KEY-----`;

// Passphrase: "secret123"
const RSA_ENCRYPTED_KEY = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAACmFlczI1Ni1jdHIAAAAGYmNyeXB0AAAAGAAAABD5T/wMMB
ZOOYSdNWIsoS6dAAAAGAAAAAEAAAEXAAAAB3NzaC1yc2EAAAADAQABAAABAQCy8KnBju/I
Or2wUwGnjf4g8oHdqlpPkvteevzdq79bv7ntKSn9dSpoZi+0ueQ+e7OtPz0JpaoJhxZEa7
0X/LrZqhbcI3cB2zKeTxqyk1xhF1MzGr4WWz77hvrDA9URQ6v0VDPhmRWbKTQ1cQPZ8p4S
zdL9kkJ5v6vQIe+D8EpRP1ZpHKV5S9k5yU6Oj5slYweLmdFe4Zl8Yg7FQCJn1LvdOfbRlO
eCvEg5Fyy7IfExThAs6W1xhtLe3cUp205snN/lc28S7q4RF+Fa2oKta4FUgSZaA7g/M9km
oJ0UsvGP2aFKaLsoRm+rNRthnLpAQRAeZB7J85xk4YJf8FYwtczrAAAD0COJKccbDXjYHd
z4/oEZD4v1+X2fPdzQAB+xLofv9U2eVgcP/BuL9zk4ut15FL1c7U7D2V96dyX+O4n9+3fq
iadfMtB92R/bpD7IMDsrJrZn4SBhPdVU30vT9iMFxwrPiRvj91L0Pe6MWibtX1V76SZAlF
6Nny2f4cmAfhU8yT5wrpI27RmgcY5CaQBlp2FeQM+nQUCyqoyJvvWYg2DTp0JLVIuWUyi7
v2qbBheVyn1xh7TJpg7wcyrHj0uexSFvYV7yQMZPM8jfYhQT2+mGzDG7WpgdVGjDHWYK/6
uBMWaNH+oQEuKJgUfJc9oa2UaacWZARJyimYjbn+RGI13TaNTZXSDFSsUZ/4URCcSWcmDu
w0rXJ4Xl0s53tR1ttV40upyBNlbFJUIioAS6aYhyTBPae2YeBpK4w7KEyndiLRxXjg0Wkn
ZQXQvyxhXBptgeiULil0q4yJaSZjV3FsijpWXciuylu5n8FR3vDvNzbbB4AGnIKpLGc/xk
SlydsKfcKACNihjG8UXxQX3S8UrQyJmIbdW1hMz7F9ZGmqtz1RGhfLt1uL3tilN+ZO99HY
FvClil7AyIs8geD/8wmHHLb6tmv6nsfpeTNxXS0xjZlFMr9K2Al+R19+ScmKZj22pnLVGY
zZhMs8ZxkU6ikN9Zg3qrphyYvl7HNmiVBGbiXQTxVZFL8v4UZP3BzSh1XB//BndJ4a1N4S
wR9nBfCq5LLJfbx7iwraKdTbg1XV7NWtJOaA7EvF29ARcfoUImtxJzCBlVD6pYAAA1fi01
9WUl2pa900v7+alWHpozYITrmii8njBf6yO2QmqGvnY08QOXIscVwoySRFd5FmUEA0vK00
hqg/AXbZU2tmYFXl2RzJQRdREF2hUx53u4wz2ZmWjkBLyx+ODYrcEpvGFYUyXRW8x/gdfQ
ptv9jESiq6uyP51cKAWuBBplaJugy14pWxdjLGIEr1fOq+DO8cs0GzlNkBfsRxzvsHP2OW
YFrpyz/RaaD9jsJ9T51NuCZiCuSjbu1i1J9VrYuGHj43TsNUlge9VQo4ta50UKhq1LRAQO
RaSdR2TnBtnsucvplSlrYtYa0Trwtm+YkTiJ67M2/cLrotiTOOquTl+FPJMzU7pQGvYR+1
oLLpNP/mt42RvzW/yQUE3D8JQWIFLpbGNjRrKhAWOy+L5Q3q6xPX8XwAkDJ6FAn9vN4Zur
TgHt+qll0hxVvcoRaYR4GFdAoQpPytZMECRSeSFrZR8pD/IyjDwDzC72xta21ZUYOL3FtU
Gnnatxi+SZxAeDVWnSUiKKGfT5PRc=
-----END OPENSSH PRIVATE KEY-----`;

// Passphrase: "secret123"
const ED25519_ENCRYPTED_KEY = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAACmFlczI1Ni1jdHIAAAAGYmNyeXB0AAAAGAAAABCbK93oo1
vNPXO72fUi8VWlAAAAGAAAAAEAAAAzAAAAC3NzaC1lZDI1NTE5AAAAILHBFe9KWaIvJIaB
hfmdz/we7LCGSzk7O4IxoFf7/4RyAAAAoPLCNp8d4HDzALhi3Z2PlxJ67mnZ3XUTgfrXxF
6kB2VxXTohWOPPp0h5hyLhiD+VsXt/iDAiFp/wavhFrxhWSJLLmCAIpajCnGw12Rsi9QH+
GkE70ctnQuo+5POhex1UuvPQURGQr0chUy88oLjb0PHS4aiLQs6z7NTIdmUug5w3dz1IgK
5zL4wYcw95UO/AcHqh3dt4o5XFwB7G1iLu5tA=
-----END OPENSSH PRIVATE KEY-----`;

describe("validatePrivateKey", () => {
  describe("unencrypted keys", () => {
    it("accepts a valid RSA private key", () => {
      const result = validatePrivateKey(RSA_PRIVATE_KEY);
      expect(result).toEqual({ valid: true, encrypted: false });
    });

    it("accepts a valid ED25519 private key", () => {
      const result = validatePrivateKey(ED25519_PRIVATE_KEY);
      expect(result).toEqual({ valid: true, encrypted: false });
    });
  });

  describe("encrypted keys", () => {
    it("recognises an encrypted RSA key without a passphrase as valid-encrypted", () => {
      const result = validatePrivateKey(RSA_ENCRYPTED_KEY);
      expect(result).toEqual({ valid: true, encrypted: true });
    });

    it("recognises an encrypted ED25519 key without a passphrase as valid-encrypted", () => {
      const result = validatePrivateKey(ED25519_ENCRYPTED_KEY);
      expect(result).toEqual({ valid: true, encrypted: true });
    });
  });

  describe("invalid input", () => {
    it("rejects an empty string", () => {
      const result = validatePrivateKey("");
      expect(result).toEqual({ valid: false, error: "Invalid private key format." });
    });

    it("rejects arbitrary text", () => {
      const result = validatePrivateKey("not a key at all");
      expect(result).toEqual({ valid: false, error: "Invalid private key format." });
    });

    it("rejects a truncated PEM block", () => {
      const truncated = RSA_PRIVATE_KEY.slice(0, 100);
      const result = validatePrivateKey(truncated);
      expect(result).toEqual({ valid: false, error: "Invalid private key format." });
    });

    it("rejects a public key presented as a private key", () => {
      // A minimal OpenSSH public key line — not a private key PEM block.
      const publicKey =
        "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI fake+public+key test@host";
      const result = validatePrivateKey(publicKey);
      expect(result).toEqual({ valid: false, error: "Invalid private key format." });
    });

    it("rejects a PEM header without a body", () => {
      const bare =
        "-----BEGIN OPENSSH PRIVATE KEY-----\n-----END OPENSSH PRIVATE KEY-----";
      const result = validatePrivateKey(bare);
      expect(result).toEqual({ valid: false, error: "Invalid private key format." });
    });
  });
});

describe("getFingerprint", () => {
  it("returns an MD5 fingerprint string for an RSA key", () => {
    const fingerprint = getFingerprint(RSA_PRIVATE_KEY);
    // sshpk formats MD5 fingerprints as 16 colon-separated lowercase hex pairs
    // without a leading "MD5:" prefix, e.g. "8d:c7:f8:50:..."
    expect(fingerprint).toMatch(/^[0-9a-f]{2}(:[0-9a-f]{2}){15}$/);
  });

  it("returns an MD5 fingerprint string for an ED25519 key", () => {
    const fingerprint = getFingerprint(ED25519_PRIVATE_KEY);
    expect(fingerprint).toMatch(/^[0-9a-f]{2}(:[0-9a-f]{2}){15}$/);
  });

  it("returns the same fingerprint on repeated calls for the same key", () => {
    const first = getFingerprint(RSA_PRIVATE_KEY);
    const second = getFingerprint(RSA_PRIVATE_KEY);
    expect(first).toBe(second);
  });

  it("returns different fingerprints for different keys", () => {
    const rsa = getFingerprint(RSA_PRIVATE_KEY);
    const ed = getFingerprint(ED25519_PRIVATE_KEY);
    expect(rsa).not.toBe(ed);
  });

  it("decrypts an encrypted RSA key with the correct passphrase and returns a fingerprint", () => {
    const fingerprint = getFingerprint(RSA_ENCRYPTED_KEY, "secret123");
    expect(fingerprint).toMatch(/^[0-9a-f]{2}(:[0-9a-f]{2}){15}$/);
  });

  it("decrypts an encrypted ED25519 key with the correct passphrase and returns a fingerprint", () => {
    const fingerprint = getFingerprint(ED25519_ENCRYPTED_KEY, "secret123");
    expect(fingerprint).toMatch(/^[0-9a-f]{2}(:[0-9a-f]{2}){15}$/);
  });

  it("throws when an encrypted key is provided without a passphrase", () => {
    expect(() => getFingerprint(RSA_ENCRYPTED_KEY)).toThrow();
  });

  it("throws when an encrypted key is provided with the wrong passphrase", () => {
    expect(() => getFingerprint(RSA_ENCRYPTED_KEY, "wrongpassphrase")).toThrow();
  });
});

describe("generateSignature", () => {
  const CHALLENGE = Buffer.from("test-challenge-data");

  describe("RSA keys", () => {
    it("returns a non-empty base64 string for an RSA key", () => {
      const sig = generateSignature(RSA_PRIVATE_KEY, CHALLENGE);
      expect(typeof sig).toBe("string");
      expect(sig.length).toBeGreaterThan(0);
      // base64 character set
      expect(sig).toMatch(/^[A-Za-z0-9+/]+=*$/);
    });

    it("produces the same signature for the same key and challenge", () => {
      const sig1 = generateSignature(RSA_PRIVATE_KEY, CHALLENGE);
      const sig2 = generateSignature(RSA_PRIVATE_KEY, CHALLENGE);
      // RSA PKCS#1 with SHA-1 is deterministic
      expect(sig1).toBe(sig2);
    });

    it("produces different signatures for different challenges", () => {
      const sig1 = generateSignature(RSA_PRIVATE_KEY, Buffer.from("challenge-one"));
      const sig2 = generateSignature(RSA_PRIVATE_KEY, Buffer.from("challenge-two"));
      expect(sig1).not.toBe(sig2);
    });

    it("signs with an encrypted RSA key given the correct passphrase", () => {
      const sig = generateSignature(RSA_ENCRYPTED_KEY, CHALLENGE, "secret123");
      expect(typeof sig).toBe("string");
      expect(sig.length).toBeGreaterThan(0);
    });
  });

  describe("ED25519 keys", () => {
    it("returns a non-empty string for an ED25519 key", () => {
      const sig = generateSignature(ED25519_PRIVATE_KEY, CHALLENGE);
      expect(typeof sig).toBe("string");
      expect(sig.length).toBeGreaterThan(0);
    });

    it("signs with an encrypted ED25519 key given the correct passphrase", () => {
      const sig = generateSignature(ED25519_ENCRYPTED_KEY, CHALLENGE, "secret123");
      expect(typeof sig).toBe("string");
      expect(sig.length).toBeGreaterThan(0);
    });
  });

  describe("error cases", () => {
    it("throws when the key is encrypted and no passphrase is provided", () => {
      expect(() => generateSignature(RSA_ENCRYPTED_KEY, CHALLENGE)).toThrow();
    });

    it("throws when the key is encrypted and the wrong passphrase is provided", () => {
      expect(() =>
        generateSignature(RSA_ENCRYPTED_KEY, CHALLENGE, "wrongpassphrase"),
      ).toThrow();
    });

    it("throws when passed an invalid key string", () => {
      expect(() => generateSignature("not-a-key", CHALLENGE)).toThrow();
    });
  });
});
