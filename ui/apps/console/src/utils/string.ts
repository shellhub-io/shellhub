export function getInitials(name: string): string {
  return name
    .split(/[\s\-_@.]+/)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? "")
    .join("");
}
