export const EXPIRY_OPTIONS = [
  { label: "30 days", value: 30 },
  { label: "60 days", value: 60 },
  { label: "90 days", value: 90 },
  { label: "1 year", value: 365 },
  { label: "Never", value: -1 },
];

export const ROLES = ["administrator", "operator", "observer"] as const;

export function isExpired(expiresIn: number): boolean {
  if (expiresIn <= 0) return false;
  return Date.now() > expiresIn * 1000;
}

export function initials(name: string | undefined) {
  if (!name) return "?";
  return name
    .split(/[\s\-_@.]+/)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? "")
    .join("");
}
