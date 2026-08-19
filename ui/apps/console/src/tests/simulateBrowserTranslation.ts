/**
 * Rewrites a rendered tree the way Chrome's auto-translation does: every non-blank text node
 * is moved inside a fresh <font> wrapper. The text node itself survives, so React still
 * updates it, but it is no longer a child of the element React recorded as its parent — the
 * next insertBefore or removeChild against that parent then throws NotFoundError.
 *
 * Use it to prove a subtree keeps re-rendering when a translated DOM is underneath it.
 */
export function simulateBrowserTranslation(root: HTMLElement): void {
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
  const texts: Text[] = [];

  while (walker.nextNode()) texts.push(walker.currentNode as Text);

  for (const text of texts) {
    if (!text.textContent?.trim()) continue;
    const font = document.createElement("font");
    text.parentNode?.insertBefore(font, text);
    font.appendChild(text);
  }
}
