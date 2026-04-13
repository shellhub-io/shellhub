import { forwardRef, ReactNode, useCallback, useEffect, useImperativeHandle, useRef } from "react";
import { useCopy } from "@/hooks/useCopy";

export interface CopyWarningRenderProps {
  /** Triggers a clipboard copy. Shows the warning dialog if clipboard access
   *  is not available (insecure context or API error). */
  copy: (text: string) => void;
  /** True for 1500 ms after a successful copy. Use for inline visual feedback. */
  copied: boolean;
}

interface CopyWarningProps {
  /** Render prop — receives `{ copy, copied }`. */
  children: (props: CopyWarningRenderProps) => ReactNode;

  /** When provided, a global Ctrl+C listener is registered that copies this
   *  text. The keystroke is debounced (500 ms) and `e.preventDefault()` is
   *  called so the browser's native copy is suppressed. */
  macro?: string;

  /** When true, all copy logic is skipped — `copy()` becomes a no-op. */
  bypass?: boolean;
}

export interface CopyWarningHandle {
  /** Imperative copy — equivalent to calling the `copy` render prop. */
  copyFn: (text: string) => void;
}

/**
 * Render-prop wrapper for safe clipboard copy with a browser security warning.
 *
 * The warning dialog is provided globally by `<ClipboardProvider>` — this
 * component only manages the render-prop interface and optional keyboard shortcut.
 *
 * Usage (render prop):
 * ```tsx
 * <CopyWarning copiedItem="Tenant ID">
 *   {({ copy, copied }) => (
 *     <button onClick={() => copy(tenantId)}>
 *       {copied ? "Copied!" : "Copy"}
 *     </button>
 *   )}
 * </CopyWarning>
 * ```
 *
 * Usage (imperative, via ref):
 * ```tsx
 * const ref = useRef<CopyWarningHandle>(null);
 * <CopyWarning ref={ref}>…</CopyWarning>
 * ref.current?.copyFn(text);
 * ```
 *
 * For simpler cases without keyboard shortcuts or refs, prefer `useCopy()` directly.
 */
const CopyWarning = forwardRef<CopyWarningHandle, CopyWarningProps>(
  function CopyWarning({ children, macro, bypass }, ref) {
    const { copy: rawCopy, copied } = useCopy();
    const executedRef = useRef(false);

    const handleCopy = useCallback(
      (text: string) => {
        if (bypass) return;
        rawCopy(text);
      },
      [bypass, rawCopy],
    );

    // Global Ctrl+C shortcut — only registered when `macro` is provided.
    useEffect(() => {
      if (!macro) return;

      const handleKeyDown = (e: KeyboardEvent) => {
        if (bypass) return;
        if (executedRef.current) return;
        if (e.type !== "keydown" || !e.ctrlKey || e.key !== "c") return;

        executedRef.current = true;
        e.preventDefault();
        handleCopy(macro);
        setTimeout(() => {
          executedRef.current = false;
        }, 500);
      };

      document.addEventListener("keydown", handleKeyDown);
      return () => document.removeEventListener("keydown", handleKeyDown);
    }, [macro, bypass, handleCopy]);

    useImperativeHandle(ref, () => ({ copyFn: handleCopy }), [handleCopy]);

    return <>{children({ copy: handleCopy, copied })}</>;
  },
);

export default CopyWarning;
