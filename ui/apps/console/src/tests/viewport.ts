/**
 * Replaces window.matchMedia, which jsdom lacks, with one whose every query matches while wide is
 * true. Hooks that read the breakpoint at import take the stub, so install it from vi.hoisted,
 * before the component under test is imported, and flip wide per case.
 */
export function installViewport() {
  const viewport = { wide: true };
  window.matchMedia = (query: string) =>
    ({
      get matches() {
        return viewport.wide;
      },
      media: query,
      addEventListener: () => {},
      removeEventListener: () => {},
    }) as unknown as MediaQueryList;
  return viewport;
}
