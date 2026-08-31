import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";

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

interface UseCopyResult {
  copy: (text: string) => void;
  copied: boolean;
}

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
 * Safe clipboard copy with automatic insecure-context handling.
 *
 * Must be used within `<ClipboardProvider>`.
 *
 * ```tsx
 * const { copy, copied } = useCopy();
 * <button onClick={() => copy(deviceId)}>{copied ? "Copied!" : "Copy"}</button>
 * ```
 */
export function useCopy(): UseCopyResult {
  const ctx = useContext(ClipboardContext);
  if (!ctx) throw new Error("useCopy must be used within <ClipboardProvider>");

  const { triggerWarning } = ctx;
  const [copied, setCopied] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const copy = useCallback(
    (text: string) => {
      if (!globalThis.isSecureContext) {
        triggerWarning();
        return;
      }

      navigator.clipboard.writeText(text).then(
        () => {
          if (timerRef.current) clearTimeout(timerRef.current);
          setCopied(true);
          timerRef.current = setTimeout(() => setCopied(false), 1500);
        },
        () => triggerWarning(),
      );
    },
    [triggerWarning],
  );

  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  return { copy, copied };
}
