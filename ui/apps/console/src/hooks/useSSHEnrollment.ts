import { useCallback, useEffect, useRef, useState } from "react";
import {
  getSshEnrollment,
  confirmSshEnrollment,
  rejectSshEnrollment,
} from "@/client";
import { isSdkError } from "@/api/errors";

/**
 * Drives the SSH key enrollment modal: fetch the pending request the gateway is
 * holding open, then let the user confirm or reject it. This is self-service —
 * the person who started the SSH login enrolls their own key.
 *
 *  loading   → fetching the request details
 *  pending   → request is open; show details + Confirm/Reject + countdown
 *  confirmed → the key was enrolled (or a prior decision confirmed it)
 *  rejected  → the request was rejected
 *  expired   → the countdown elapsed, or the code is unknown/expired (404)
 *  error     → the request could not be fetched (non-404)
 *
 * Enrollments live for a short window. The details response carries no expiry,
 * so the countdown runs a fixed TTL from `requested_at` (falling back to now).
 */
export type EnrollmentPhase =
  "loading" | "pending" | "confirmed" | "rejected" | "expired" | "error";

export interface EnrollmentDetails {
  sshid: string;
  deviceName: string;
  username: string;
  ipAddress: string;
  requestedAt: string;
  fingerprint: string;
  code: string;
  /** true for a first-time key enrollment, false for a step-up confirmation. */
  enroll: boolean;
}

const ENROLLMENT_TTL_SECONDS = 90;

export function useSSHEnrollment(code: string) {
  const [phase, setPhase] = useState<EnrollmentPhase>("loading");
  const [details, setDetails] = useState<EnrollmentDetails | null>(null);
  const [secondsLeft, setSecondsLeft] = useState(ENROLLMENT_TTL_SECONDS);
  const [deciding, setDeciding] = useState(false);
  const [actionError, setActionError] = useState("");
  const expiresAtRef = useRef(0);

  // Fetch the request once on mount (or when the code changes).
  useEffect(() => {
    let cancelled = false;

    void (async () => {
      if (!code) {
        setPhase("expired");
        return;
      }

      setPhase("loading");
      try {
        const { data } = await getSshEnrollment({
          path: { code },
          throwOnError: true,
        });
        if (cancelled) return;

        const requestedAt = data.requested_at ?? "";
        setDetails({
          sshid: data.sshid ?? "",
          deviceName: data.device_name ?? "",
          username: data.username ?? "",
          ipAddress: data.ip_address ?? "",
          requestedAt,
          fingerprint: data.fingerprint ?? "",
          code: data.code ?? code,
          enroll: data.enroll ?? true,
        });

        if (data.state === "confirmed") {
          setPhase("confirmed");
          return;
        }
        if (data.state === "rejected") {
          setPhase("rejected");
          return;
        }

        const requestedMs = Date.parse(requestedAt);
        const base = Number.isNaN(requestedMs) ? Date.now() : requestedMs;
        expiresAtRef.current = base + ENROLLMENT_TTL_SECONDS * 1000;
        setSecondsLeft(
          Math.max(0, Math.ceil((expiresAtRef.current - Date.now()) / 1000)),
        );
        setPhase("pending");
      } catch (err) {
        if (cancelled) return;
        // 404 means unknown/expired/already-decided — treat all as expired.
        setPhase(isSdkError(err) && err.status === 404 ? "expired" : "error");
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [code]);

  // Count down while pending; at zero the window has closed.
  useEffect(() => {
    if (phase !== "pending") return undefined;

    const tick = () => {
      const left = Math.max(
        0,
        Math.ceil((expiresAtRef.current - Date.now()) / 1000),
      );
      setSecondsLeft(left);
      if (left <= 0) setPhase("expired");
    };

    tick();
    const id = window.setInterval(tick, 1000);
    return () => window.clearInterval(id);
  }, [phase]);

  const decide = useCallback(
    async (kind: "confirm" | "reject") => {
      if (!code || deciding) return;
      setDeciding(true);
      setActionError("");
      try {
        const call =
          kind === "confirm" ? confirmSshEnrollment : rejectSshEnrollment;
        await call({ path: { code }, throwOnError: true });
        setPhase(kind === "confirm" ? "confirmed" : "rejected");
      } catch (err) {
        // The window closed under us while deciding.
        if (isSdkError(err) && err.status === 404) {
          setPhase("expired");
          return;
        }
        setActionError("Something went wrong. Please try again.");
      } finally {
        setDeciding(false);
      }
    },
    [code, deciding],
  );

  const confirm = useCallback(() => decide("confirm"), [decide]);
  const reject = useCallback(() => decide("reject"), [decide]);

  return {
    phase,
    details,
    secondsLeft,
    confirm,
    reject,
    deciding,
    actionError,
  };
}
