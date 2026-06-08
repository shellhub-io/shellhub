/**
 * VaultAutoLockBanner — ephemeral toast that fires when the vault is locked by
 * the idle auto-lock timer.
 *
 * Distinguished from VaultLockedBanner: that component is a persistent, inline
 * page banner rendered inside the vault page when the vault is locked (with an
 * "Unlock" CTA). This component is a floating toast that appears once — briefly —
 * whenever an auto-lock event happens (nonce bump), regardless of the current page.
 * It owns its own dismiss and auto-dismiss logic.
 */
import { useEffect, useRef, useState } from "react";
import { useVaultStore } from "@/stores/vaultStore";
import Alert from "@/components/common/Alert";

const AUTO_DISMISS_MS = 6000;

export default function VaultAutoLockBanner() {
  const autoLockNonce = useVaultStore((s) => s.autoLockNonce);

  // Initialize seen to the current nonce so the first render never fires the toast
  const seenNonce = useRef(autoLockNonce);
  const [visible, setVisible] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (autoLockNonce === seenNonce.current) return;

    // A new auto-lock event occurred — update seen and show the toast
    seenNonce.current = autoLockNonce;
    setVisible(true);

    // Clear any pending auto-dismiss from a previous show
    if (timerRef.current !== null) {
      clearTimeout(timerRef.current);
    }

    timerRef.current = setTimeout(() => {
      setVisible(false);
      timerRef.current = null;
    }, AUTO_DISMISS_MS);
  }, [autoLockNonce]);

  // Clear the timer on unmount to prevent state updates after unmount
  useEffect(() => {
    return () => {
      if (timerRef.current !== null) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  if (!visible) return null;

  return (
    <div className="fixed bottom-4 right-4 z-[75] w-80">
      <Alert
        variant="warning"
        onDismiss={() => {
          if (timerRef.current !== null) {
            clearTimeout(timerRef.current);
            timerRef.current = null;
          }
          setVisible(false);
        }}
      >
        Vault locked due to inactivity.
      </Alert>
    </div>
  );
}
