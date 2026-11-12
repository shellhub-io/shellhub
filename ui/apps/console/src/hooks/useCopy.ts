import { createContext, useContext } from "react";

/**
 * How a copy failure is reported. The warning is raised at the app level rather than by each
 * button, so the explanation appears once however many are on screen.
 */
export interface ClipboardContextValue {
  triggerWarning: () => void;
}

/**
 * Carries the copy-failure reporter down the tree.
 */
export const ClipboardContext = createContext<ClipboardContextValue | null>(
  null,
);

/**
 * Returns the `triggerWarning` function from the nearest `<ClipboardProvider>`.
 *
 * Throws when called outside a provider.
 */
export function useClipboardWarning(): () => void {
  const ctx = useContext(ClipboardContext);
  if (!ctx) throw new Error("useClipboardWarning must be used within <ClipboardProvider>");
  return ctx.triggerWarning;
}

/**
 * Returns a function that copies text to the clipboard, reporting an insecure context or a
 * refused write through the nearest `<ClipboardProvider>`. Throws when called outside a provider.
 */
export function useCopy(): (text: string) => void {
  const ctx = useContext(ClipboardContext);
  if (!ctx) throw new Error("useCopy must be used within <ClipboardProvider>");

  const { triggerWarning } = ctx;
  return (text: string) => {
    if (!globalThis.isSecureContext) {
      triggerWarning();
      return;
    }
    navigator.clipboard.writeText(text).catch(triggerWarning);
  };
}
