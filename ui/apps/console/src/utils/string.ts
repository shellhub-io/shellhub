/**
 * A phrase with its first letter raised, for where a lowercase phrase or identifier opens a line.
 */
export function capitalize(phrase: string): string {
  return phrase.charAt(0).toUpperCase() + phrase.slice(1);
}

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
