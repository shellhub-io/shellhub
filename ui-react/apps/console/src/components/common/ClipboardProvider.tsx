import {
  createContext,
  ReactNode,
  useCallback,
  useContext,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from "react";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import BaseDialog from "./BaseDialog";

// ─── Context ──────────────────────────────────────────────────────────────────

interface ClipboardContextValue {
  triggerWarning: () => void;
}

const ClipboardContext = createContext<ClipboardContextValue | null>(null);

// ─── Provider ─────────────────────────────────────────────────────────────────

/**
 * Mounts a single clipboard-warning dialog for the whole app.
 * Wrap the app root with this once; every `useCopy()` call shares it.
 */
export function ClipboardProvider({ children }: { children: ReactNode }) {
  const [showDialog, setShowDialog] = useState(false);
  const titleId = useId();
  const descId = useId();

  const triggerWarning = useCallback(() => setShowDialog(true), []);
  const handleClose = useCallback(() => setShowDialog(false), []);
  const ctxValue = useMemo(() => ({ triggerWarning }), [triggerWarning]);

  return (
    <ClipboardContext.Provider value={ctxValue}>
      {children}

      <BaseDialog
        open={showDialog}
        onClose={handleClose}
        size="sm"
        aria-labelledby={titleId}
        aria-describedby={descId}
      >
        {/* Header */}
        <div className="p-6 pb-0 flex items-center gap-3">
          <ExclamationTriangleIcon
            className="w-5 h-5 flex-shrink-0 text-accent-yellow"
            aria-hidden="true"
          />
          <h2 id={titleId} className="text-base font-semibold text-text-primary">
            Copying is not allowed
          </h2>
        </div>

        {/* Body */}
        <div className="px-6 pt-3 pb-6">
          <p id={descId} className="text-sm text-text-muted">
            Clipboard access is only permitted on secure (HTTPS) or localhost
            origins. Please ensure your instance is secure to enable clipboard
            features.
          </p>
        </div>

        {/* Footer */}
        <div className="flex justify-end px-6 py-4 border-t border-border">
          <button
            type="button"
            data-testid="copy-warning-ok-btn"
            onClick={handleClose}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
          >
            OK
          </button>
        </div>
      </BaseDialog>
    </ClipboardContext.Provider>
  );
}

// ─── Hook ─────────────────────────────────────────────────────────────────────

export interface UseCopyResult {
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
