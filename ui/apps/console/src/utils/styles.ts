/**
 * The field-label type style, without spacing. Compose with a margin when the label sits above
 * its control.
 */
export const LABEL_BASE =
  "block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";
/**
 * A field label with the standard gap beneath it. The usual choice; LABEL_BASE is for a layout
 * that sets its own spacing.
 */
export const LABEL = `${LABEL_BASE} mb-1.5`;

/**
 * Everything an input needs except its border and type scale, so the four variants below differ
 * only in those and cannot drift apart in anything else.
 */
export const INPUT_BASE =
  "w-full px-3.5 py-2.5 bg-card rounded-lg text-text-primary placeholder:text-text-secondary text-ellipsis focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed";

const BORDER_OK = "border border-border";
const BORDER_ERROR = "border border-accent-red/50";

/**
 * A text input in its normal state.
 */
export const INPUT = `${INPUT_BASE} ${BORDER_OK} text-sm`;
/**
 * A text input showing a validation error.
 */
export const INPUT_ERROR = `${INPUT_BASE} ${BORDER_ERROR} text-sm`;
/**
 * A monospaced input, for a key, a hash or an address, where character shape has to be legible.
 */
export const INPUT_MONO = `${INPUT_BASE} ${BORDER_OK} font-mono text-xs`;
/**
 * A monospaced input showing a validation error.
 */
export const INPUT_MONO_ERROR = `${INPUT_BASE} ${BORDER_ERROR} font-mono text-xs`;

/**
 * Added to an input that is displayed but not editable. It dims and drops the focus ring, so the
 * field does not invite typing that goes nowhere.
 */
export const INPUT_READONLY =
  "opacity-60 cursor-default focus:ring-0 focus:border-border";
