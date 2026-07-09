import { describe, it, expect } from "vitest";
import {
  keySchema,
  buildKeyBody,
  buildKeyDefaults,
  type KeyFormValues,
  type KeyMode,
} from "../keySchema";
import type { PublicKeyResponse } from "@/client";

function makeValues(overrides: Partial<KeyFormValues> = {}): KeyFormValues {
  return {
    name: "My Key",
    data: "",
    usernameOption: "all",
    username: "",
    filterOption: "all",
    hostname: "",
    tags: [],
    ...overrides,
  };
}

/** First validation message per field for the given mode. */
function errorsFor(
  mode: KeyMode,
  values: KeyFormValues,
): Partial<Record<keyof KeyFormValues, string>> {
  const result = keySchema(mode).safeParse(values);
  if (result.success) return {};

  const errors: Partial<Record<keyof KeyFormValues, string>> = {};
  for (const issue of result.error.issues) {
    const key = issue.path[0] as keyof KeyFormValues;
    if (key && errors[key] === undefined) errors[key] = issue.message;
  }
  return errors;
}

function makeResponse(
  overrides: Partial<PublicKeyResponse> = {},
): PublicKeyResponse {
  return {
    name: "My Key",
    data: btoa("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItest comment"),
    fingerprint: "aa:bb:cc:dd",
    created_at: "2024-01-01T00:00:00Z",
    tenant_id: "tenant-abc",
    filter: { hostname: ".*", tags: [] },
    username: ".*",
    ...overrides,
  };
}

describe("keySchema — name", () => {
  it("rejects an empty name", () => {
    expect(errorsFor("create", makeValues({ name: "" })).name).toBeDefined();
  });

  it("rejects a whitespace-only name", () => {
    expect(errorsFor("create", makeValues({ name: "   " })).name).toBeDefined();
  });

  it("rejects a name longer than 64 characters", () => {
    expect(
      errorsFor("create", makeValues({ name: "a".repeat(65) })).name,
    ).toBeDefined();
  });

  it("accepts a name exactly 64 characters long", () => {
    expect(
      errorsFor("create", makeValues({ name: "a".repeat(64) })).name,
    ).toBeUndefined();
  });

  it("accepts a non-empty name within the limit", () => {
    expect(
      errorsFor("create", makeValues({ name: "My SSH Key" })).name,
    ).toBeUndefined();
  });
});

describe("keySchema — data (create mode)", () => {
  it("rejects an empty data field in create mode", () => {
    expect(errorsFor("create", makeValues({ data: "" })).data).toBeDefined();
  });

  it("rejects an invalid public key format in create mode", () => {
    expect(
      errorsFor("create", makeValues({ data: "not-a-valid-key" })).data,
    ).toBeDefined();
  });

  it("accepts a valid ssh-ed25519 public key in create mode", () => {
    expect(
      errorsFor(
        "create",
        makeValues({
          data: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItest comment",
        }),
      ).data,
    ).toBeUndefined();
  });

  it("accepts a valid ssh-rsa public key in create mode", () => {
    expect(
      errorsFor(
        "create",
        makeValues({
          data: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7testkey comment",
        }),
      ).data,
    ).toBeUndefined();
  });
});

describe("keySchema — data (edit mode)", () => {
  it("does not validate the data field in edit mode (empty is fine)", () => {
    expect(errorsFor("edit", makeValues({ data: "" })).data).toBeUndefined();
  });

  it("does not validate the data field in edit mode (invalid value is fine)", () => {
    expect(
      errorsFor("edit", makeValues({ data: "garbage" })).data,
    ).toBeUndefined();
  });
});

describe("keySchema — username", () => {
  it("requires a non-empty username when usernameOption is 'username'", () => {
    expect(
      errorsFor(
        "create",
        makeValues({ usernameOption: "username", username: "" }),
      ).username,
    ).toBeDefined();
  });

  it("rejects a whitespace-only username when usernameOption is 'username'", () => {
    expect(
      errorsFor(
        "create",
        makeValues({ usernameOption: "username", username: "   " }),
      ).username,
    ).toBeDefined();
  });

  it("accepts a non-empty username when usernameOption is 'username'", () => {
    expect(
      errorsFor(
        "create",
        makeValues({ usernameOption: "username", username: "root" }),
      ).username,
    ).toBeUndefined();
  });

  it("ignores username validation when usernameOption is 'all'", () => {
    expect(
      errorsFor("create", makeValues({ usernameOption: "all", username: "" }))
        .username,
    ).toBeUndefined();
  });
});

describe("keySchema — hostname", () => {
  it("requires a non-empty hostname when filterOption is 'hostname'", () => {
    expect(
      errorsFor(
        "create",
        makeValues({ filterOption: "hostname", hostname: "" }),
      ).hostname,
    ).toBeDefined();
  });

  it("rejects a whitespace-only hostname when filterOption is 'hostname'", () => {
    expect(
      errorsFor(
        "create",
        makeValues({ filterOption: "hostname", hostname: "   " }),
      ).hostname,
    ).toBeDefined();
  });

  it("accepts a non-empty hostname pattern when filterOption is 'hostname'", () => {
    expect(
      errorsFor(
        "create",
        makeValues({ filterOption: "hostname", hostname: "web-.*" }),
      ).hostname,
    ).toBeUndefined();
  });

  it("ignores hostname validation when filterOption is not 'hostname'", () => {
    expect(
      errorsFor("create", makeValues({ filterOption: "all", hostname: "" }))
        .hostname,
    ).toBeUndefined();
  });
});

describe("keySchema — tags", () => {
  it("rejects 0 tags when filterOption is 'tags'", () => {
    expect(
      errorsFor("create", makeValues({ filterOption: "tags", tags: [] })).tags,
    ).toBeDefined();
  });

  it("rejects more than 3 tags when filterOption is 'tags'", () => {
    expect(
      errorsFor(
        "create",
        makeValues({ filterOption: "tags", tags: ["t1", "t2", "t3", "t4"] }),
      ).tags,
    ).toBeDefined();
  });

  it.each([
    { count: 1, tags: ["t1"] },
    { count: 2, tags: ["t1", "t2"] },
    { count: 3, tags: ["t1", "t2", "t3"] },
  ])("accepts $count tag(s) when filterOption is 'tags'", ({ tags }) => {
    expect(
      errorsFor("create", makeValues({ filterOption: "tags", tags })).tags,
    ).toBeUndefined();
  });

  it("ignores tag count validation when filterOption is not 'tags'", () => {
    expect(
      errorsFor("create", makeValues({ filterOption: "all", tags: [] })).tags,
    ).toBeUndefined();
  });
});

describe("buildKeyBody", () => {
  it("base64-encodes the data field", () => {
    const raw = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItest";
    const body = buildKeyBody(makeValues({ name: "k", data: raw }));
    expect(body.data).toBe(btoa(raw));
  });

  it("trims the name before encoding", () => {
    const body = buildKeyBody(
      makeValues({ name: "  My Key  ", data: "ssh-ed25519 AAAA..." }),
    );
    expect(body.name).toBe("My Key");
  });

  it("sets username to '.*' when usernameOption is 'all'", () => {
    const body = buildKeyBody(makeValues({ usernameOption: "all" }));
    expect(body.username).toBe(".*");
  });

  it("uses the specific username when usernameOption is 'username'", () => {
    const body = buildKeyBody(
      makeValues({ usernameOption: "username", username: "admin" }),
    );
    expect(body.username).toBe("admin");
  });

  it("builds filter with hostname when filterOption is 'hostname'", () => {
    const body = buildKeyBody(
      makeValues({ filterOption: "hostname", hostname: "web-.*" }),
    );
    expect(body.filter).toEqual({ hostname: "web-.*" });
  });

  it("builds filter with tags when filterOption is 'tags'", () => {
    const body = buildKeyBody(
      makeValues({ filterOption: "tags", tags: ["prod", "db"] }),
    );
    expect(body.filter).toEqual({ tags: ["prod", "db"] });
  });

  it("builds filter with hostname='.*' when filterOption is 'all'", () => {
    const body = buildKeyBody(makeValues({ filterOption: "all" }));
    expect(body.filter).toEqual({ hostname: ".*" });
  });
});

describe("buildKeyDefaults", () => {
  it("round-trips a key with username='.*' and hostname-filter '.*' as 'all' mode", () => {
    const defaults = buildKeyDefaults(makeResponse());
    expect(defaults.name).toBe("My Key");
    expect(defaults.usernameOption).toBe("all");
    expect(defaults.username).toBe("");
    expect(defaults.filterOption).toBe("all");
    expect(defaults.hostname).toBe("");
    expect(defaults.tags).toEqual([]);
  });

  it("decodes base64 data into the data field", () => {
    const raw = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItest comment";
    const defaults = buildKeyDefaults(makeResponse({ data: btoa(raw) }));
    expect(defaults.data).toBe(raw);
  });

  it("falls back to raw data when base64 decoding fails", () => {
    const raw = "ssh-ed25519 AAAA...";
    const defaults = buildKeyDefaults(makeResponse({ data: raw }));
    expect(defaults.data).toBe(raw);
  });

  it("round-trips a key with a specific username", () => {
    const defaults = buildKeyDefaults(makeResponse({ username: "root" }));
    expect(defaults.usernameOption).toBe("username");
    expect(defaults.username).toBe("root");
  });

  it("round-trips a key with a hostname filter (non '.*')", () => {
    const defaults = buildKeyDefaults(
      makeResponse({ filter: { hostname: "web-.*", tags: [] } }),
    );
    expect(defaults.filterOption).toBe("hostname");
    expect(defaults.hostname).toBe("web-.*");
    expect(defaults.tags).toEqual([]);
  });

  it("round-trips a key with a tags filter", () => {
    const defaults = buildKeyDefaults(
      makeResponse({
        filter: {
          hostname: undefined,
          tags: [
            {
              name: "prod",
              tenant_id: "t",
              created_at: "2024-01-01",
              updated_at: "2024-01-01",
            },
            {
              name: "db",
              tenant_id: "t",
              created_at: "2024-01-01",
              updated_at: "2024-01-01",
            },
          ],
        },
      }),
    );
    expect(defaults.filterOption).toBe("tags");
    expect(defaults.tags).toEqual(["prod", "db"]);
    expect(defaults.hostname).toBe("");
  });

  it("treats filter.tags present (non-empty) as tags mode even without hostname", () => {
    const defaults = buildKeyDefaults(
      makeResponse({
        filter: {
          hostname: undefined,
          tags: [
            {
              name: "edge",
              tenant_id: "t",
              created_at: "2024-01-01",
              updated_at: "2024-01-01",
            },
          ],
        },
      }),
    );
    expect(defaults.filterOption).toBe("tags");
  });

  it("treats hostname='.*' with empty tags as 'all' mode", () => {
    const defaults = buildKeyDefaults(
      makeResponse({ filter: { hostname: ".*", tags: [] } }),
    );
    expect(defaults.filterOption).toBe("all");
    expect(defaults.hostname).toBe("");
  });
});
