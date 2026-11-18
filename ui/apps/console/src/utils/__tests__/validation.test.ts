import { describe, it, expect } from "vitest";
import {
  isGoRegex,
  validatePattern,
  suggestNamespaceName,
  validateEmail,
  validateIdentifier,
  validateKeyName,
  validateName,
  validateNamespaceName,
  validatePassword,
  validateUsername,
} from "../validation";

describe("isGoRegex", () => {
  it.each([
    ["a plain pattern", "^root$"],
    ["a flag group", "(?i)root"],
    ["a scoped flag group", "(?i:a)b"],
    ["a Go named group", "(?P<n>x)"],
    ["an octal escape", "\\12"],
    ["a lookahead-looking class", "[(?=]"],
    ["an escaped parenthesis before ?=", "a\\(?=b"],
    ["a quoted lookahead", "\\Q(?=\\E"],
    ["a quoted bracket left open", "\\Qa[b"],
    ["an empty flag group", "(?)a"],
    ["a repeated start anchor", "^*"],
    ["a repeated end anchor", "a$+"],
    ["an optional word boundary", "\\b?x"],
    ["a class opening with ] and holding a parenthesis", "[]a(x]"],
    ["a Unicode class before a dash in a class", "[\\p{L}-0-9]+"],
    ["a one-letter Unicode class before a dash", "[\\pL-0]"],
    ["a repeated hex code point", "\\x{41}+"],
    ["a POSIX class followed by a parenthesis", "[[:alpha:](]"],
    ["a POSIX class followed by repeat characters", "[[:alnum:]+*]"],
    ["a repeat after a flag group", "a*(?i)*"],
    ["a repeat after an empty quote", "a+\\Q\\E+"],
  ])("accepts %s", (_, pattern) => {
    expect(isGoRegex(pattern)).toBe(true);
  });

  it.each([
    ["a lookahead", "^(?!root).*"],
    ["a lookbehind", "(?<=a)b"],
    ["a backreference", "(\\w)\\1"],
    ["an unclosed class", "[a-"],
    ["an unclosed group", "("],
  ])("rejects %s", (_, pattern) => {
    expect(isGoRegex(pattern)).toBe(false);
  });
});

describe("validatePattern", () => {
  it("asks for a pattern when there is none", () => {
    expect(validatePattern("  ", "Username")).toBe("Username pattern is required");
  });

  it("refuses a pattern Go cannot compile", () => {
    expect(validatePattern("[a-", "Hostname")).toBe(
      "Hostname must be a valid regular expression",
    );
  });

  it("checks the trimmed pattern, which is what the form sends", () => {
    expect(validatePattern("a\\ ", "Username")).toBe(
      "Username must be a valid regular expression",
    );
  });

  it("accepts a valid pattern", () => {
    expect(validatePattern(" root|admin ", "Username")).toBeNull();
  });
});

describe("validateNamespaceName", () => {
  describe("valid names", () => {
    it.each([
      "abc",
      "a-b",
      "my-namespace",
      "a1b2c3",
      "1abc",
      "abc1",
      "a--b",
      "a".repeat(30),
    ])("accepts %p", (name) => {
      expect(validateNamespaceName(name)).toBeNull();
    });
  });

  describe("length rules", () => {
    it.each(["", "a", "ab"])("rejects %p as too short", (name) => {
      expect(validateNamespaceName(name)).toBe(
        "Name must be at least 3 characters",
      );
    });

    it("rejects names longer than 30 characters", () => {
      expect(validateNamespaceName("a".repeat(31))).toBe(
        "Name must be at most 30 characters",
      );
    });
  });

  describe("character rules", () => {
    it.each([
      "NameWithCaps",
      "has space",
      "has.dot",
      "has@at",
      "has#hash",
      "has$dollar",
      "has_underscore",
      "has/slash",
      "açento",
    ])("rejects %p", (name) => {
      expect(validateNamespaceName(name)).toBe(
        "Only lowercase letters, numbers, and hyphens (cannot start or end with hyphen)",
      );
    });
  });

  describe("structural rules", () => {
    it.each(["-leading", "trailing-", "-both-"])("rejects %p", (name) => {
      expect(validateNamespaceName(name)).toBe(
        "Only lowercase letters, numbers, and hyphens (cannot start or end with hyphen)",
      );
    });
  });
});

describe("suggestNamespaceName", () => {
  it.each([
    { username: "joaocel", expected: "joaocel" },
    { username: "Joao.Cel_01", expected: "joao-cel-01" },
    { username: "__ana--placa!!", expected: "ana-placa" },
    { username: `${"a".repeat(29)}_b`, expected: "a".repeat(29) },
    { username: "a.b", expected: "a-b" },
    { username: "ab", expected: "" },
    { username: "!!", expected: "" },
  ])("$username → $expected", ({ username, expected }) => {
    expect(suggestNamespaceName(username)).toBe(expected);
  });

  it.each(["joaocel", "Joao.Cel_01", `${"a".repeat(29)}_b`, "x".repeat(40)])(
    "suggests a name the namespace rules accept for %s",
    (username) => {
      expect(validateNamespaceName(suggestNamespaceName(username))).toBeNull();
    },
  );
});

describe("validateName", () => {
  it.each(["a", "Alice Smith", "X", "a".repeat(64), "  Alice  "])(
    "accepts %p",
    (value) => {
      expect(validateName(value)).toBeNull();
    },
  );

  it.each(["", "   "])("rejects empty/whitespace %p as required", (value) => {
    expect(validateName(value)).toBe("Name is required");
  });

  it("rejects names longer than 64 characters", () => {
    expect(validateName("a".repeat(65))).toBe(
      "Name must be at most 64 characters",
    );
  });
});

describe("validateKeyName", () => {
  it.each(["ci-runners", "edge_01", "abc", "a".repeat(20)])(
    "accepts %p",
    (value) => {
      expect(validateKeyName(value)).toBeNull();
    },
  );

  it.each([
    ["ab", "Name must be at least 3 characters."],
    ["a".repeat(21), "Name must be at most 20 characters."],
    ["has space", "Name can only contain letters, numbers, - and _."],
    ["bang!", "Name can only contain letters, numbers, - and _."],
  ])("rejects %p", (value, message) => {
    expect(validateKeyName(value)).toBe(message);
  });
});

describe("validateUsername", () => {
  it.each([
    "alice",
    "alice123",
    "alice.smith",
    "alice_smith",
    "alice-smith",
    "user@example",
    "abc",
    "a".repeat(32),
  ])("accepts %p", (value) => {
    expect(validateUsername(value)).toBeNull();
  });

  it.each(["", "   "])("rejects empty/whitespace %p as required", (value) => {
    expect(validateUsername(value)).toBe("Username is required");
  });

  it.each([
    "ab", // too short
    "a".repeat(33), // too long
    "Alice", // uppercase
    "ALICE",
    "has space",
    "has#hash",
    "has/slash",
    "has!bang",
  ])("rejects %p with the hint message", (value) => {
    expect(validateUsername(value)).toMatch(/^3-32 characters/);
  });
});

describe("validateEmail", () => {
  it.each([
    "alice@example.com",
    "user.name@example.co.uk",
    "user+tag@example.org",
  ])("accepts %p", (value) => {
    expect(validateEmail(value)).toBeNull();
  });

  it.each(["", "   "])("rejects empty/whitespace %p as required", (value) => {
    expect(validateEmail(value)).toBe("Email is required");
  });

  it.each(["plainstring", "no@domain", "missing.at.sign", "two@@signs.com"])(
    "rejects %p",
    (value) => {
      expect(validateEmail(value)).toBe("Enter a valid email address");
    },
  );
});

describe("validatePassword", () => {
  it.each(["12345", "a".repeat(32), "P@ssw0rd!"])("accepts %p", (value) => {
    expect(validatePassword(value)).toBeNull();
  });

  it.each(["", "1234", "a".repeat(33)])("rejects %p", (value) => {
    expect(validatePassword(value)).toBe(
      "Password must be 5–32 characters long",
    );
  });
});

describe("validateIdentifier", () => {
  it.each([
    "  alice  ",
    "alice.smith",
    "  alice@example.com  ",
    "user+tag@example.org",
  ])("accepts whitespace-padded valid username/email %p", (value) => {
    expect(validateIdentifier(value)).toBeNull();
  });

  it.each(["", "   "])("rejects empty/whitespace-only %p as required", (value) => {
    expect(validateIdentifier(value)).toBe("Username or email is required");
  });

  it.each(["!", "a", "  !  ", "has CAPS", "has space@no.tld"])(
    "rejects invalid identifier %p",
    (value) => {
      expect(validateIdentifier(value)).toBe(
        "Enter a valid username or email",
      );
    },
  );
});
