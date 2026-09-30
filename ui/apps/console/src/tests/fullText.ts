/**
 * A Testing Library matcher for text split across elements, like a count in its own span: it
 * matches the innermost element whose whole text reads exactly text, whatever its tag.
 */
export function fullText(text: string) {
  return (_: string, element: Element | null) =>
    !!element &&
    element.textContent === text &&
    Array.from(element.children).every((child) => child.textContent !== text);
}
