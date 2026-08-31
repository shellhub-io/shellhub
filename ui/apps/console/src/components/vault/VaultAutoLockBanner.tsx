import { useEffect, useRef, useState } from "react";
import { useVaultStore } from "@/stores/vaultStore";
import { Callout } from "@shellhub/design-system/primitives";

const AUTO_DISMISS_MS = 6000;

export default function VaultAutoLockBanner() {
  const autoLockNonce = useVaultStore((s) => s.autoLockNonce);

  const seenNonce = useRef(autoLockNonce);
  const [visible, setVisible] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (autoLockNonce === seenNonce.current) return;

    seenNonce.current = autoLockNonce;
    setVisible(true);

    if (timerRef.current !== null) {
      clearTimeout(timerRef.current);
    }

    timerRef.current = setTimeout(() => {
      setVisible(false);
      timerRef.current = null;
    }, AUTO_DISMISS_MS);
  }, [autoLockNonce]);

  useEffect(() => {
    return () => {
      if (timerRef.current !== null) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  if (!visible) return null;

  return (
    <div className="fixed bottom-4 right-4 z-toast w-80">
      <Callout
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
      </Callout>
    </div>
  );
}
