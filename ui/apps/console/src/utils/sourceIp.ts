// Parsing + classification for the access-policy Source IP field. The backend normalizes a
// bare IP to a /32 (or /128) host route; these helpers mirror that so the UI can show the
// user what will be stored, flag an all-IPs entry, and reject nonsense before submit.

export type SourceIpParse =
  | { status: "empty" }
  | { status: "incomplete" }
  | { status: "invalid"; note: string }
  // value is the canonical CIDR that will be stored; label describes it.
  | {
      status: "valid" | "host" | "any";
      value: string;
      label: string;
      note?: string;
    };

export type SourceIpKind = "host" | "private" | "public" | "any" | "ipv6" | "";

function ipv4Octets(s: string): number[] | null {
  const m = s.match(/^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/);
  if (!m) return null;
  const octets = m.slice(1).map(Number);
  return octets.some((n) => n > 255) ? null : octets;
}

function looksIpv6(s: string): boolean {
  return (
    s.includes(":") && /^[0-9a-fA-F:]+$/.test(s) && s.split(":").length <= 8
  );
}

function classifyV4(
  o: number[],
): "private network" | "loopback" | "unspecified" | "public" {
  if (
    o[0] === 10 ||
    (o[0] === 172 && o[1] >= 16 && o[1] <= 31) ||
    (o[0] === 192 && o[1] === 168)
  )
    return "private network";
  if (o[0] === 127) return "loopback";
  if (o[0] === 0) return "unspecified";
  return "public";
}

function addressCount(bits: number): string {
  const hosts = 2 ** (32 - bits);
  if (hosts === 1) return "1 address";
  if (hosts >= 1e6)
    return `${(hosts / 1e6).toFixed(hosts >= 1e7 ? 0 : 1)}M addresses`;
  if (hosts >= 1000) return `${Math.round(hosts / 1000)}K addresses`;
  return `${hosts} addresses`;
}

// parseSourceIp interprets a raw entry as the user types. A bare IP resolves to a host route
// (/32 or /128); a full CIDR is classified; "0.0.0.0/0" is flagged as "any"; anything that
// can't ever be an address is invalid, while still-being-typed input stays "incomplete".
export function parseSourceIp(raw: string): SourceIpParse {
  const s = raw.trim();
  if (!s) return { status: "empty" };

  const cidr = s.match(/^(.+)\/(\d{1,3})$/);
  if (cidr) {
    const [, addr, bitsStr] = cidr;
    const bits = Number(bitsStr);
    const v4 = ipv4Octets(addr);
    if (v4) {
      if (bits > 32)
        return {
          status: "invalid",
          note: `prefix /${bits} is out of range (max /32)`,
        };
      if (bits === 0)
        return {
          status: "any",
          value: "0.0.0.0/0",
          label: "= any IP",
          note: "same as leaving this empty",
        };
      return {
        status: "valid",
        value: s,
        label: `${classifyV4(v4)} · ${addressCount(bits)}`,
      };
    }
    if (looksIpv6(addr)) {
      if (bits > 128)
        return {
          status: "invalid",
          note: "prefix /" + bits + " is out of range (max /128)",
        };
      return { status: "valid", value: s, label: `IPv6 · /${bits}` };
    }
    return { status: "invalid", note: "not a valid network address" };
  }

  const v4 = ipv4Octets(s);
  if (v4)
    return {
      status: "host",
      value: `${s}/32`,
      label: `single host · ${classifyV4(v4)}`,
      note: `stored as ${s}/32`,
    };
  if (looksIpv6(s))
    return {
      status: "host",
      value: `${s}/128`,
      label: "single IPv6 host",
      note: `stored as ${s}/128`,
    };

  // A complete four-octet IPv4 that failed validation is wrong, not still-being-typed.
  if (/^\d{1,3}(\.\d{1,3}){3}$/.test(s))
    return { status: "invalid", note: "octet out of range (max 255)" };
  // Still typing something address-shaped (digits, dots, slash, hex/colon) — don't cry wolf yet.
  if (/^[\d./:a-fA-F]+$/.test(s)) return { status: "incomplete" };
  return { status: "invalid", note: "not a valid IP or CIDR" };
}

// sourceIpKind labels an already-stored CIDR for its chip badge.
export function sourceIpKind(cidr: string): SourceIpKind {
  const m = cidr.match(/^(.+)\/(\d+)$/);
  if (!m) return "";
  const [, addr, bitsStr] = m;
  if (cidr === "0.0.0.0/0") return "any";
  const v4 = ipv4Octets(addr);
  if (v4) {
    if (Number(bitsStr) === 32) return "host";
    return classifyV4(v4) === "public" ? "public" : "private";
  }
  return "ipv6";
}
