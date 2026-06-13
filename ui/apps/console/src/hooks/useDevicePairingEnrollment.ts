import { useCallback, useEffect, useRef, useState } from "react";
import { prepareDevicePairing, getDevicePairingStatus } from "@/client";

/**
 * Drives the Add Device "install once, page confirms itself" flow: mint (or
 * reuse) a pre-authorized pairing code, then poll until the device that claims
 * it comes online and is accepted automatically.
 *
 *  minting   → asking the server for a code (or validating a stored one)
 *  waiting   → code ready, watching for the device to connect
 *  connected → a device claimed the code and was accepted
 *  expired   → the code's window elapsed with no device
 *  error     → the code could not be minted (e.g. missing permission)
 *
 * The active code is cached in localStorage per namespace so a refresh reuses it
 * instead of minting a new one every time. A stored code is reused only while
 * it's still within its window and hasn't already been claimed.
 */
export type EnrollmentPhase =
  "minting" | "waiting" | "connected" | "expired" | "error";

export interface EnrollmentDevice {
  uid: string;
  name: string;
}

const POLL_INTERVAL_MS = 3000;
const DEFAULT_TTL_SECONDS = 600;
const STORAGE_PREFIX = "shellhub:pairing-enrollment:";

interface StoredCode {
  code: string;
  expiresAt: number;
}

function readStored(key: string): StoredCode | null {
  try {
    const raw = window.localStorage.getItem(key);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as StoredCode;
    if (
      typeof parsed.code !== "string" ||
      typeof parsed.expiresAt !== "number"
    ) {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

function writeStored(key: string, value: StoredCode) {
  try {
    window.localStorage.setItem(key, JSON.stringify(value));
  } catch {
    // storage may be unavailable (private mode/quota); the flow still works,
    // it just won't survive a refresh.
  }
}

function clearStored(key: string) {
  try {
    window.localStorage.removeItem(key);
  } catch {
    // ignore
  }
}

export function useDevicePairingEnrollment(enabled: boolean, tenant: string) {
  const [code, setCode] = useState("");
  const [phase, setPhase] = useState<EnrollmentPhase>("minting");
  const [device, setDevice] = useState<EnrollmentDevice | null>(null);
  const [secondsLeft, setSecondsLeft] = useState(0);
  const [nonce, setNonce] = useState(0);
  const expiresAtRef = useRef(0);

  const storageKey = STORAGE_PREFIX + tenant;

  // Acquire a code: reuse a stored one that is still valid and unclaimed,
  // otherwise mint a fresh one. Re-runs on regenerate (nonce) and namespace
  // change. A regenerate clears the stored code first, so it always mints.
  useEffect(() => {
    if (!enabled) return undefined;

    let cancelled = false;

    const start = (value: StoredCode) => {
      if (cancelled) return;
      setCode(value.code);
      expiresAtRef.current = value.expiresAt;
      setSecondsLeft(
        Math.max(0, Math.ceil((value.expiresAt - Date.now()) / 1000)),
      );
      setPhase("waiting");
    };

    const mint = async () => {
      try {
        const { data } = await prepareDevicePairing({ throwOnError: true });
        if (cancelled) return;

        const value: StoredCode = {
          code: data.code ?? "",
          expiresAt:
            Date.now() +
            (data.expires_in_seconds ?? DEFAULT_TTL_SECONDS) * 1000,
        };
        writeStored(storageKey, value);
        start(value);
      } catch {
        if (!cancelled) setPhase("error");
      }
    };

    void (async () => {
      setPhase("minting");
      setDevice(null);
      setCode("");

      const stored = readStored(storageKey);
      if (stored && stored.expiresAt > Date.now()) {
        // Only reuse if it hasn't already been claimed by a device.
        try {
          const { data } = await getDevicePairingStatus({
            path: { code: stored.code },
            throwOnError: true,
          });
          if (cancelled) return;

          if (data.status === "accepted") {
            clearStored(storageKey);
            await mint();
          } else {
            start(stored);
          }

          return;
        } catch {
          // Unknown/expired server-side; fall through to minting.
          clearStored(storageKey);
        }
      }

      await mint();
    })();

    return () => {
      cancelled = true;
    };
  }, [enabled, tenant, nonce, storageKey]);

  // Count down every second while waiting; at zero, check once more before
  // giving up (the device may have claimed the code in the final poll gap) and
  // then flip to expired.
  useEffect(() => {
    if (!enabled || phase !== "waiting") return undefined;

    let id = 0;
    let cancelled = false;

    const tick = () => {
      const left = Math.max(
        0,
        Math.ceil((expiresAtRef.current - Date.now()) / 1000),
      );
      setSecondsLeft(left);
      if (left > 0) return;

      window.clearInterval(id);
      void (async () => {
        try {
          const { data } = await getDevicePairingStatus({
            path: { code },
            throwOnError: true,
          });
          if (cancelled) return;

          if (data.status === "accepted") {
            clearStored(storageKey);
            setDevice({ uid: data.uid ?? "", name: data.name ?? "" });
            setPhase("connected");

            return;
          }
        } catch {
          // fall through to expired
        }

        if (!cancelled) setPhase("expired");
      })();
    };

    tick();
    id = window.setInterval(tick, 1000);

    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, [enabled, phase, code, storageKey]);

  // Poll the code's status until a device claims it.
  useEffect(() => {
    if (!enabled || phase !== "waiting" || !code) return undefined;

    let cancelled = false;

    const tick = async () => {
      try {
        const { data } = await getDevicePairingStatus({
          path: { code },
          throwOnError: true,
        });
        if (cancelled) return;

        if (data.status === "accepted") {
          clearStored(storageKey);
          setDevice({ uid: data.uid ?? "", name: data.name ?? "" });
          setPhase("connected");
        }
      } catch {
        // A miss is a not-yet-visible or expired code; the countdown ends the
        // wait, so keep polling until then.
      }
    };

    const id = window.setInterval(() => void tick(), POLL_INTERVAL_MS);

    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, [enabled, phase, code, storageKey]);

  const regenerate = useCallback(() => {
    clearStored(storageKey);
    setNonce((n) => n + 1);
  }, [storageKey]);

  return { code, phase, device, secondsLeft, regenerate };
}
