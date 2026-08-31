/**
 * Up to two initials for an avatar. Splits on the separators that appear in names, emails and
 * usernames alike, so "ada.lovelace@example.com" gives AL rather than one letter.
 */
export function getInitials(name: string): string {
  return name
    .split(/[\s\-_@.]+/)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? "")
    .join("");
}
