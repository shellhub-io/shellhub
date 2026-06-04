import { useState } from "react";

/**
 * Calls `onOpen` during render when `open` transitions from false to true.
 * Uses React's "adjusting state during render" pattern — no useEffect needed.
 */
export function useResetOnOpen(open: boolean, onOpen: () => void) {
  const [wasOpen, setWasOpen] = useState(false);
  if (open && !wasOpen) {
    setWasOpen(true);
    onOpen();
  } else if (!open && wasOpen) {
    setWasOpen(false);
  }
}
