import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";

// ─── Context ──────────────────────────────────────────────────────────────────

export interface ClipboardContextValue {
  triggerWarning: () => void;
}

export const ClipboardContext = createContext<ClipboardContextValue | null>(null);

// ─── Hook ─────────────────────────────────────────────────────────────────────

interface UseCopyResult {
  /** Call with the text to copy. Shows the warning dialog when clipboard access
   *  is unavailable (insecure context or API error). */
  copy: (text: string) => void;
  /** True for 1500 ms after a successful copy. Use for inline visual feedback. */
  copied: boolean;
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
