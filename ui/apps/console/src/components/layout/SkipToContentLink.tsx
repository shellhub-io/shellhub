/**
 * Skip-to-main-content link for keyboard users (WCAG 2.4.1 Bypass Blocks).
 * Rendered as the first child of each authenticated layout shell so it is the
 * first Tab stop; visually offscreen (translated up) until focused, when it
 * slides into view at top-left. Activating it moves focus to `#main-content`.
 *
 * Trade-off: when the mobile sidebar drawer is open it sets `aria-modal`, so
 * screen readers treat this link (outside the drawer) as inert until the drawer
 * closes. That is acceptable — the skip link is not meaningful while the nav
 * drawer itself is open.
 */
export default function SkipToContentLink() {
  return (
    <a
      href="#main-content"
      className="fixed left-4 top-4 z-[200] -translate-y-20 rounded-md border border-border bg-surface px-4 py-2 text-sm font-medium text-text-primary transition-transform focus:translate-y-0"
    >
      Skip to main content
    </a>
  );
}
