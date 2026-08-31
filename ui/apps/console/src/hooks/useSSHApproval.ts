import { useCallback, useEffect, useRef, useState } from "react";
import { getSshApproval } from "@/client";
import { isSdkError } from "@/api/errors";
import {
  useConfirmSSHApproval,
  useRejectSSHApproval,
} from "./useSSHIdentityMutations";

/**
 * Drives the SSH login approval modal: fetch the request the gateway is holding
 * open, then let the user confirm or reject it. This is self-service — the
 * person who started the SSH login decides on their own login.
 *
 *  loading   → fetching the request details
 *  pending   → request is open; show details + Confirm/Reject + countdown
 *  confirmed → the approval went through (or a prior decision confirmed it)
 *  rejected  → the request was rejected
 *  expired   → the countdown elapsed, or the code is unknown/expired (404)
 *  error     → the request could not be fetched (non-404)
 */
export type ApprovalPhase =
  "loading" | "pending" | "confirmed" | "rejected" | "expired" | "error";

/** What confirming does. The whole screen branches on this. */
export type ApprovalKind = "identity" | "reauth";

export interface ApprovalDetails {
  sshid: string;
  deviceName: string;
  username: string;
  ipAddress: string;
  requestedAt: string;
  fingerprint: string;
  code: string;
  kind: ApprovalKind;
  /**
   * The policy's re-auth window in seconds, on a reauth approval. Zero means the
   * policy asks every time.
   */
  reauthPeriod: number;
  /** Namespace the key lands in, which may differ from the one being browsed. */
  namespace: string;
}

/** Fallback window when the server does not report one. */
const APPROVAL_TTL_SECONDS = 90;

export function useSSHApproval(code: string) {
  const [phase, setPhase] = useState<ApprovalPhase>("loading");
  const [details, setDetails] = useState<ApprovalDetails | null>(null);
  const [secondsLeft, setSecondsLeft] = useState(APPROVAL_TTL_SECONDS);
  const [totalSeconds, setTotalSeconds] = useState(APPROVAL_TTL_SECONDS);
  const [deciding, setDeciding] = useState(false);
  const [actionError, setActionError] = useState("");
  const expiresAtRef = useRef(0);
  const confirmMutation = useConfirmSSHApproval();
  const rejectMutation = useRejectSSHApproval();

  useEffect(() => {
    let cancelled = false;

    void (async () => {
      if (!code) {
        setPhase("expired");
        return;
      }

      setPhase("loading");
      try {
        const { data } = await getSshApproval({
          path: { code },
          throwOnError: true,
        });
        if (cancelled) return;

        setDetails({
          sshid: data.sshid ?? "",
          deviceName: data.device_name ?? "",
          username: data.username ?? "",
          ipAddress: data.ip_address ?? "",
          requestedAt: data.requested_at ?? "",
          fingerprint: data.fingerprint ?? "",
          code: data.code ?? code,
          kind: data.kind === "reauth" ? "reauth" : "identity",
          reauthPeriod: data.reauth_period ?? 0,
          namespace: data.namespace ?? "",
        });

        if (data.state === "confirmed") {
          setPhase("confirmed");
          return;
        }
        if (data.state === "rejected") {
          setPhase("rejected");
          return;
        }

        const left = data.expires_in_seconds ?? APPROVAL_TTL_SECONDS;
        expiresAtRef.current = Date.now() + left * 1000;
        setSecondsLeft(left);
        setTotalSeconds(left);
        setPhase("pending");
      } catch (err) {
        if (cancelled) return;
        setPhase(isSdkError(err) && err.status === 404 ? "expired" : "error");
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [code]);

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
    async (decision: "confirm" | "reject") => {
      if (!code || deciding) return false;
      setDeciding(true);
      setActionError("");
      try {
        const mutation =
          decision === "confirm" ? confirmMutation : rejectMutation;
        await mutation.mutateAsync({ path: { code } });
        setPhase(decision === "confirm" ? "confirmed" : "rejected");

        return true;
      } catch (err) {
        if (isSdkError(err) && err.status === 404) {
          setPhase("expired");
          return false;
        }
        setActionError("Something went wrong. Please try again.");

        return false;
      } finally {
        setDeciding(false);
      }
    },
    [code, deciding, confirmMutation, rejectMutation],
  );

  const confirm = useCallback(() => decide("confirm"), [decide]);
  const reject = useCallback(() => decide("reject"), [decide]);

  return {
    phase,
    details,
    secondsLeft,
    totalSeconds,
    confirm,
    reject,
    markConfirmed: useCallback(() => setPhase("confirmed"), []),
    deciding,
    actionError,
  };
}
