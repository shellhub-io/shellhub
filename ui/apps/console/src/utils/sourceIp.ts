/**
 * The result of reading a Source IP field as it is typed. incomplete and invalid are kept apart
 * on purpose: half an address is not an error to show yet, a malformed one is.
 */
export type SourceIpParse =
  | { status: "empty" }
  | { status: "incomplete"; note: string }
  | { status: "invalid"; note: string }
  | {
      status: "valid" | "host" | "any";
      value: string;
      label: string;
      note?: string;
    };

/**
 * What an already-stored CIDR represents, for the chip shown beside it. The empty string is for
 * a value that does not parse at all.
 */
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

/**
 * Reads a Source IP entry as the user types it.
 *
 * A bare address resolves to a host route — /32 or /128 — because that is what the backend
 * stores for one; a full CIDR is classified; 0.0.0.0/0 is called out as "any", since matching
 * every address is rarely what was meant. Input that cannot become an address is invalid,
 * while input that is merely unfinished stays incomplete, so the field does not scold someone
 * mid-keystroke.
 */
export function parseSourceIp(raw: string): SourceIpParse {
  const s = raw.trim();
  if (!s) return { status: "empty" };

  const cidr = s.match(/^(.+)\/(\d{1,3})$/);
  const addrPart = cidr ? cidr[1] : s;
  const hasLeadingZero =
    (cidr && cidr[2].length > 1 && cidr[2][0] === "0") ||
    (/^\d{1,3}(\.\d{1,3}){3}$/.test(addrPart) &&
      addrPart.split(".").some((p) => p.length > 1 && p[0] === "0"));
  if (hasLeadingZero)
    return { status: "invalid", note: "leading zeros are not allowed" };

  if (cidr) {
    const bits = Number(cidr[2]);
    const v4 = ipv4Octets(addrPart);
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
    if (looksIpv6(addrPart)) {
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

  if (/^\d{1,3}(\.\d{1,3}){3}$/.test(s))
    return { status: "invalid", note: "octet out of range (max 255)" };
  if (/^[\d./:a-fA-F]+$/.test(s))
    return { status: "incomplete", note: "incomplete IP address" };
  return { status: "invalid", note: "not a valid IP or CIDR" };
}

/**
 * Labels a stored CIDR for its chip. Unlike parseSourceIp this assumes a settled value, so
 * anything unparseable is the empty string rather than an error to explain.
 */
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
