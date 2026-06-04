export const LABEL_BASE =
  "block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";
export const LABEL = `${LABEL_BASE} mb-1.5`;

export const INPUT_BASE =
  "w-full px-3.5 py-2.5 bg-card rounded-lg text-text-primary placeholder:text-text-secondary text-ellipsis focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed";

const BORDER_OK = "border border-border";
const BORDER_ERROR = "border border-accent-red/50";

export const INPUT = `${INPUT_BASE} ${BORDER_OK} text-sm`;
export const INPUT_ERROR = `${INPUT_BASE} ${BORDER_ERROR} text-sm`;
export const INPUT_MONO = `${INPUT_BASE} ${BORDER_OK} font-mono text-xs`;
export const INPUT_MONO_ERROR = `${INPUT_BASE} ${BORDER_ERROR} font-mono text-xs`;

export const INPUT_READONLY =
  "opacity-60 cursor-default focus:ring-0 focus:border-border";
