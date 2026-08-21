import { DEFAULT_FONT_SIZE } from "@/stores/terminalThemeStore";

type FontSizeIntent = "increase" | "decrease" | "reset";

// "=" is accepted alongside "+" because the two share a key on a US layout and
// browsers report either depending on whether Shift was held. "_" is
// deliberately absent: "-" needs no Shift, so accepting it would only steal
// ctrl+_ (0x1F, readline/emacs undo) from the remote shell. Numpad add and
// subtract already arrive as "+" and "-".
const INTENT_BY_KEY: Record<string, FontSizeIntent> = {
  "+": "increase",
  "=": "increase",
  "-": "decrease",
  0: "reset",
};

const STEP = 1;

/** The size the given keystroke asks for, or null if it asks for nothing. */
export const nextFontSize = (current: number, event: KeyboardEvent): number | null => {
  if (event.altKey) return null;
  if (!event.ctrlKey && !event.metaKey) return null;
  if (!Object.hasOwn(INTENT_BY_KEY, event.key)) return null;

  switch (INTENT_BY_KEY[event.key]) {
    case "increase":
      return current + STEP;
    case "decrease":
      return current - STEP;
    case "reset":
      return DEFAULT_FONT_SIZE;
  }
};
