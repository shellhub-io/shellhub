import { useEffect, useRef, useState } from "react";

const SAVED_FOR_MS = 2000;

/**
 * Runs the save behind a setting and reports how it went: failure, the message to show when the
 * save throws, or saved, which holds for two seconds after a save succeeds so the control can say
 * so and then settle.
 */
export function useSettingSave(failure: string) {
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined);

  useEffect(() => () => clearTimeout(timer.current), []);

  const run = async (save: () => Promise<unknown>) => {
    clearTimeout(timer.current);
    setError("");
    setSaved(false);
    try {
      await save();
      setSaved(true);
      timer.current = setTimeout(() => setSaved(false), SAVED_FOR_MS);
    } catch {
      setError(failure);
    }
  };

  return { error, saved, run };
}
