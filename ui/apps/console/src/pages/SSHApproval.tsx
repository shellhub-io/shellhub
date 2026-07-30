import { useState, useEffect, useRef } from "react";
import {
  useParams,
  useNavigate,
  useLocation,
  Navigate,
} from "react-router-dom";
import {
  KeyIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  NoSymbolIcon,
  ArrowDownIcon,
  ArrowsRightLeftIcon,
  ArrowRightIcon,
  ArrowLeftIcon,
} from "@heroicons/react/24/outline";
import { Button, Spinner } from "@shellhub/design-system/primitives";
import BaseDialog from "@/components/common/BaseDialog";
import InputField from "@/components/common/fields/InputField";
import apiClient from "@/api/client";
import { isSdkError } from "@/api/errors";
import { webTerminalReauth } from "@/client";
import { useOtpInput } from "@/hooks/useOtpInput";
import { useSSHApproval, ApprovalDetails } from "@/hooks/useSSHApproval";
import { useAuthStore } from "@/stores/authStore";
import { useNamespaces } from "@/hooks/useNamespaces";
import { useSwitchNamespace } from "@/hooks/useNamespaceMutations";
import { getInitials } from "@/utils/string";

/**
 * The two approvals a native login can wait on. Both act on the identity — one
 * creates it, the other refreshes its re-auth window — so both open over the
 * identity list, and the terminal banner links straight to the right one.
 */
const LIST = "/ssh-identities";

const FLOWS = {
  new: {
    label: "Add SSH key",
    path: (code: string) => `${LIST}/new/${code}`,
  },
  confirm: {
    label: "Confirm SSH login",
    path: (code: string) => `${LIST}/confirm/${code}`,
  },
} as const;

type Flow = keyof typeof FLOWS;

/**
 * Modal for a login the gateway is holding open. A native client is shown a URL
 * in its terminal and opens it here, already signed in, so the question is not
 * "was this you" but "which account does this key become". The screen leads with
 * the account the key attaches to, since the plausible mistake is deciding while
 * signed in as the wrong one. The same surface serves the re-auth, which creates
 * nothing.
 *
 * The web terminal mounts it inline over itself instead of routing to it, so the
 * code and the close handler can come in as props; on a route they come from the
 * URL.
 */
export default function SSHApproval({
  flow,
  code: codeProp,
  onClose,
  onDecided,
}: {
  flow: Flow;
  code?: string;
  onClose?: () => void;
  onDecided?: () => void;
}) {
  const params = useParams<{ code: string }>();
  const navigate = useNavigate();
  const routed = codeProp === undefined;
  const code = codeProp ?? params.code ?? "";
  const {
    phase,
    details,
    secondsLeft,
    totalSeconds,
    confirm,
    reject,
    markConfirmed,
    deciding,
    actionError,
  } = useSSHApproval(code);

  const close = () => {
    if (onClose) {
      onClose();

      return;
    }

    void navigate(LIST);
  };

  const reauth = flow === "confirm";

  // Mounted inline, the terminal behind this screen is the outcome: once the
  // decision lands it either continues or reports the refusal itself, so there
  // is nothing left here to read.
  const decided = (landed: boolean) => {
    if (landed && !routed) onDecided?.();
  };

  // Only the request itself says which of the two this is, and it arrives after
  // the fetch. A link opened on the wrong route (stale, or pasted from another
  // login) would otherwise render the wrong question. Mounted inline there is no
  // route to correct, and redirecting would drag the app off the terminal.
  const actual: Flow = details?.kind === "reauth" ? "confirm" : "new";
  if (routed && details && actual !== flow) {
    return <Navigate to={FLOWS[actual].path(code)} replace />;
  }

  return (
    <BaseDialog open onClose={close} size="lg" aria-label={FLOWS[flow].label}>
      <div className="p-8">
        {phase === "loading" && <StatusMessage label="Loading request..." />}

        {phase === "pending" && details && (
          <PendingRequest
            details={details}
            code={code}
            showCode={routed}
            secondsLeft={secondsLeft}
            totalSeconds={totalSeconds}
            deciding={deciding}
            actionError={actionError}
            onConfirm={() => void confirm().then(decided)}
            onReject={() => void reject().then(decided)}
            onReauthed={() => {
              markConfirmed();
              decided(true);
            }}
          />
        )}

        {phase === "confirmed" && (
          <ResultMessage
            tone="success"
            icon={<CheckCircleIcon className="w-7 h-7" strokeWidth={1.5} />}
            title={reauth ? "Re-authenticated" : "Key added"}
            description={
              reauth
                ? `Back to your terminal: the login continues on its own.${reauthWindowSentence(details?.reauthPeriod ?? 0)}`
                : "It's in your SSH Identities. Back to your terminal: the login continues on its own."
            }
            action={<DoneButton onClick={close} />}
          />
        )}

        {phase === "rejected" && (
          <ResultMessage
            tone="error"
            icon={<NoSymbolIcon className="w-7 h-7" strokeWidth={1.5} />}
            title="Rejected"
            description="The login won't go through. Nothing was changed."
            action={<DoneButton onClick={close} />}
          />
        )}

        {phase === "expired" && (
          <ResultMessage
            tone="error"
            icon={<XCircleIcon className="w-7 h-7" strokeWidth={1.5} />}
            title="Request expired"
            description="This request is no longer valid. Start the SSH login again to get a new one."
            action={<DoneButton onClick={close} />}
          />
        )}

        {phase === "error" && (
          <ResultMessage
            tone="error"
            icon={<XCircleIcon className="w-7 h-7" strokeWidth={1.5} />}
            title="Something went wrong"
            description="We couldn't load this request. Please try again."
            action={<DoneButton onClick={close} />}
          />
        )}
      </div>
    </BaseDialog>
  );
}

function DoneButton({ onClick }: { onClick: () => void }) {
  return (
    <Button variant="secondary" size="md" onClick={onClick}>
      Close
    </Button>
  );
}

function PendingRequest({
  details,
  code,
  showCode,
  secondsLeft,
  totalSeconds,
  deciding,
  actionError,
  onConfirm,
  onReject,
  onReauthed,
}: {
  details: ApprovalDetails;
  code: string;
  showCode: boolean;
  secondsLeft: number;
  totalSeconds: number;
  deciding: boolean;
  actionError: string;
  onConfirm: () => void;
  onReject: () => void;
  onReauthed: () => void;
}) {
  const reauth = details.kind === "reauth";
  // A re-auth asks two things, and they are read in order: is this login yours,
  // then prove it is you. Keeping the secret field beside details nobody has
  // read yet invites typing past the very check the details exist for.
  const [step, setStep] = useState<"review" | "verify">("review");
  const { elsewhere, currentName, switching, switchToTarget } =
    useNamespaceMismatch(details.namespace);
  const verifying = reauth && !elsewhere && step === "verify";

  if (verifying) {
    return (
      <div>
        <h2 className="text-lg font-semibold text-text-primary mb-2">
          Prove it&apos;s you
        </h2>
        {/* One line of what is being released, so proving a factor never gets
            detached from the login it belongs to. */}
        <p className="text-sm text-text-secondary mb-5">
          The login on{" "}
          <span className="text-text-primary font-medium">
            {details.deviceName || "this device"}
          </span>{" "}
          as{" "}
          <span className="text-text-primary font-medium">
            {details.username}
          </span>{" "}
          goes through once you do.
        </p>

        <Countdown secondsLeft={secondsLeft} totalSeconds={totalSeconds} />

        <ReauthFactor
          fingerprint={details.fingerprint}
          approvalCode={details.code || code}
          onDone={onReauthed}
          onBack={() => setStep("review")}
        />
      </div>
    );
  }

  return (
    <div>
      <h2 className="text-lg font-semibold text-text-primary mb-2">
        {reauth
          ? "Re-authenticate to continue?"
          : "Add this SSH key to your identities?"}
      </h2>
      <p className="text-sm text-text-secondary leading-relaxed mb-5">
        {reauth
          ? `An access policy on ${details.deviceName || "this device"} asks you to re-authenticate before the login goes through.`
          : "The SSH login waiting in your terminal used a key ShellHub doesn't know yet."}
      </p>

      {/* Only a new identity attaches something, so only it draws the
          key-to-account binding. Its absence is what tells a re-auth that
          nothing is being created. */}
      {!reauth && (
        <BindingCard
          details={details}
          elsewhere={elsewhere}
          currentName={currentName}
        />
      )}

      <p className="text-sm text-text-secondary leading-relaxed border-l-2 border-primary pl-3.5 mb-4">
        {reauth
          ? `This confirms the key, not just this login.${reauthWindowSentence(details.reauthPeriod)}`
          : "Next time you connect with this key, you won't be asked again. Revoke it anytime in SSH Identities."}
      </p>

      <details
        className="rounded-xl border border-border bg-black/20 mb-4"
        open={reauth}
      >
        <summary className="cursor-pointer list-none px-3.5 py-2 text-sm text-text-secondary hover:text-text-primary">
          Login details
        </summary>
        <dl className="px-3.5 pb-2.5">
          {/* The code proves the page belongs to the login printed in the
              terminal. Mounted inline there is no banner to compare it with, and
              a code the user never saw only puzzles. */}
          {showCode && (
            <SpecRow label="code" value={details.code || code} mono />
          )}
          <SpecRow label="source ip" value={details.ipAddress} />
          <SpecRow
            label="device / login"
            value={[details.deviceName, details.username]
              .filter(Boolean)
              .join(" / ")}
          />
          <SpecRow
            label="requested"
            value={formatRequestedAt(details.requestedAt)}
          />
        </dl>
      </details>

      <p className="text-xs text-text-muted leading-relaxed mb-5">
        If you didn't start this login, reject it.
      </p>

      {actionError && (
        <p className="text-sm text-accent-red mb-4 animate-shake" role="alert">
          {actionError}
        </p>
      )}

      <Countdown secondsLeft={secondsLeft} totalSeconds={totalSeconds} />

      <div className="flex justify-end gap-2">
        <Button
          variant="ghost"
          size="md"
          disabled={deciding}
          icon={<NoSymbolIcon className="w-4 h-4" strokeWidth={2} />}
          onClick={onReject}
        >
          Reject
        </Button>
        {elsewhere ? (
          <Button
            variant="primary"
            size="md"
            loading={switching}
            icon={<ArrowsRightLeftIcon className="w-4 h-4" strokeWidth={2} />}
            onClick={switchToTarget}
          >
            Switch to {details.namespace}
          </Button>
        ) : reauth ? (
          // A re-auth is authentication, so it asks for a factor rather than a
          // click. Without one it would only be a button that says yes to itself.
          <Button
            variant="primary"
            size="md"
            icon={<ArrowRightIcon className="w-4 h-4" strokeWidth={2} />}
            onClick={() => setStep("verify")}
          >
            Continue
          </Button>
        ) : (
          <Button
            variant="primary"
            size="md"
            loading={deciding}
            icon={<CheckCircleIcon className="w-4 h-4" strokeWidth={2} />}
            onClick={onConfirm}
          >
            Add key
          </Button>
        )}
      </div>
    </div>
  );
}

/** How long the gateway will keep holding the login open. */
function Countdown({
  secondsLeft,
  totalSeconds,
}: {
  secondsLeft: number;
  totalSeconds: number;
}) {
  return (
    <>
      <div className="flex items-center gap-1.5 text-2xs font-mono text-text-muted">
        <ClockIcon className="w-3.5 h-3.5" strokeWidth={1.5} />
        Expires in {secondsLeft}s
      </div>
      <div className="h-0.5 mt-2 mb-5 rounded-sm bg-border overflow-hidden">
        <div
          className="h-full bg-primary transition-[width] duration-1000 ease-linear"
          style={{
            width: `${totalSeconds > 0 ? (secondsLeft / totalSeconds) * 100 : 0}%`,
          }}
        />
      </div>
    </>
  );
}

/**
 * The factor a re-auth demands, and the submit that releases the held login.
 * Which factor depends on how the account signs in: an SSO bounce for identity
 * provider users, an authenticator code for MFA, the account password otherwise.
 *
 * The same call proves the factor and releases the login, so the two cannot
 * drift apart — and the key must be the caller's own, which is what authorizes
 * the release.
 */
function ReauthFactor({
  fingerprint,
  approvalCode,
  onDone,
  onBack,
}: {
  fingerprint: string;
  approvalCode: string;
  onDone: () => void;
  onBack: () => void;
}) {
  const mfaEnabled = useAuthStore((s) => s.mfaEnabled);
  const origin = useAuthStore((s) => s.origin);
  // SSO users have no local password or ShellHub MFA — their second factor lives
  // at the identity provider, so the step-up re-authenticates there (ForceAuthn).
  const isSso = origin === "saml";
  const otp = useOtpInput(6);
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const formRef = useRef<HTMLFormElement>(null);

  // Landing here with the field ready gives back most of the click the extra
  // step costs. Done on the DOM rather than with autoFocus, which the house
  // lint bans, and a no-op for SSO, which has no field to fill.
  useEffect(() => {
    formRef.current?.querySelector<HTMLInputElement>("input")?.focus();
  }, []);

  // The IdP re-auth runs in a popup; it lands on /sso-reauth, which posts the
  // outcome back here.
  useEffect(() => {
    if (!isSso) return undefined;

    const onMessage = (event: MessageEvent) => {
      if (event.origin !== window.location.origin) return;
      if (event.data === "sso-reauth-ok") {
        onDone();
      } else if (event.data === "sso-reauth-error") {
        setError(
          "Re-authentication with your provider failed. Please try again.",
        );
        setSubmitting(false);
      }
    };
    window.addEventListener("message", onMessage);

    return () => window.removeEventListener("message", onMessage);
  }, [isSso, onDone]);

  const canSubmit =
    isSso || (mfaEnabled ? otp.isComplete : password.length > 0);

  const startSso = () => {
    if (submitting) return;
    setSubmitting(true);
    setError(null);
    apiClient
      .get<{ url: string }>("/api/user/saml/reauth", {
        params: { fingerprint, approval_code: approvalCode },
      })
      .then(({ data }) => {
        if (!window.open(data.url, "sso-reauth", "width=520,height=680")) {
          setError(
            "Pop-up blocked. Allow pop-ups for this site and try again.",
          );
          setSubmitting(false);
        }
      })
      .catch(() => {
        setError("Could not start SSO re-authentication.");
        setSubmitting(false);
      });
  };

  const submit = async () => {
    if (!canSubmit || submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      await webTerminalReauth({
        body: {
          ...(mfaEnabled ? { code: otp.getValue() } : { password }),
          fingerprint,
          approval_code: approvalCode,
        },
        throwOnError: true,
      });
      onDone();
    } catch (err) {
      // A wrong factor is a 403 (the session is still valid — a 401 would trip
      // the global logout-on-401 interceptor and sign the user out).
      setError(
        isSdkError(err) && err.status === 403
          ? mfaEnabled
            ? "Invalid code. It changes every 30 seconds — enter the current one."
            : "Incorrect password."
          : "Re-authentication failed. Please try again.",
      );
      if (mfaEnabled) otp.reset();
      setSubmitting(false);
    }
  };

  return (
    <form
      ref={formRef}
      onSubmit={(e) => {
        e.preventDefault();
        if (isSso) startSso();
        else void submit();
      }}
      className="space-y-4"
    >
      {isSso ? (
        <p className="text-sm text-text-muted">
          A window opens for you to sign in again with your identity provider.
          The login continues once you finish.
        </p>
      ) : mfaEnabled ? (
        <div>
          <p className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3 text-center">
            Verification Code
          </p>
          <div
            className="flex gap-2 justify-center"
            role="group"
            aria-label="Verification Code"
            onPaste={otp.handlePaste}
          >
            {otp.code.map((digit, index) => (
              <input
                key={index}
                ref={(el) => {
                  otp.inputRefs.current[index] = el;
                }}
                type="text"
                inputMode="numeric"
                maxLength={1}
                value={digit}
                aria-label={`Digit ${index + 1}`}
                onChange={(e) => otp.handleChange(index, e.target.value)}
                onKeyDown={(e) => otp.handleKeyDown(index, e)}
                className="w-10 h-10 text-center text-base font-mono bg-background border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
              />
            ))}
          </div>
        </div>
      ) : (
        <InputField
          id="reauth-password"
          type="password"
          label="Account password"
          value={password}
          onChange={setPassword}
          placeholder="Enter your password"
        />
      )}

      {error && (
        <p className="text-sm text-accent-red animate-shake" role="alert">
          {error}
        </p>
      )}

      <div className="flex justify-end gap-2">
        <Button
          type="button"
          variant="ghost"
          size="md"
          disabled={submitting}
          icon={<ArrowLeftIcon className="w-4 h-4" strokeWidth={2} />}
          onClick={onBack}
        >
          Back
        </Button>
        <Button
          type="submit"
          variant="primary"
          size="md"
          loading={submitting}
          disabled={!canSubmit}
          icon={<CheckCircleIcon className="w-4 h-4" strokeWidth={2} />}
        >
          Re-authenticate
        </Button>
      </div>
    </form>
  );
}

/**
 * Reports the namespace this request belongs to when it is not the one the
 * console is showing, along with a switch into it. The decision would succeed
 * either way (the API authorizes against the request's namespace, not the
 * session's), but the console would then be listing a different namespace's
 * identities, so the key would land somewhere the user cannot see. Namespace
 * names address devices in an SSHID, so they are unique and enough to find the
 * tenant among the memberships the user already has.
 */
function useNamespaceMismatch(namespace: string) {
  const currentTenant = useAuthStore((s) => s.tenant);
  const { namespaces } = useNamespaces();
  const switchNs = useSwitchNamespace();
  const { pathname } = useLocation();

  const target = namespaces.find((ns) => ns.name === namespace);
  const elsewhere =
    !!namespace && !!target && target.tenant_id !== currentTenant;

  return {
    elsewhere,
    currentName:
      namespaces.find((ns) => ns.tenant_id === currentTenant)?.name ?? "",
    switching: switchNs.isPending,
    switchToTarget: () => {
      if (target)
        void switchNs.mutateAsync({
          tenantId: target.tenant_id,
          redirectTo: pathname,
        });
    },
  };
}

/**
 * The key on top, the account below it, and an arrow between: the screen draws
 * the binding it is asking about. The account half carries the visual weight
 * because signing in as the wrong account is the mistake this screen prevents.
 */
function BindingCard({
  details,
  elsewhere,
  currentName,
}: {
  details: ApprovalDetails;
  elsewhere: boolean;
  currentName: string;
}) {
  const { name, email, user, logout } = useAuthStore();
  const navigate = useNavigate();
  const { code = "" } = useParams<{ code: string }>();

  const displayName = name || user || email || "Your account";

  const switchAccount = () => {
    logout();
    void navigate(
      `/login?redirect=${encodeURIComponent(FLOWS.new.path(code))}`,
    );
  };

  return (
    <div className="rounded-xl border border-border-light overflow-hidden mb-4">
      <div className="flex items-center gap-3 px-4 py-3 bg-card">
        <span className="grid place-items-center w-9 h-9 rounded-lg bg-hover-medium border border-border-light shrink-0">
          <KeyIcon className="w-4 h-4 text-text-secondary" strokeWidth={1.5} />
        </span>
        <div className="min-w-0">
          <p className="font-mono text-2xs uppercase tracking-label text-text-muted">
            the key, from {details.ipAddress || "an unknown address"}
          </p>
          <p className="font-mono text-xs text-text-secondary truncate">
            {details.fingerprint || "—"}
          </p>
        </div>
      </div>

      <div className="relative h-px bg-border">
        <span className="absolute left-5 top-1/2 -translate-y-1/2 grid place-items-center w-6 h-6 rounded-full bg-card border border-border-light">
          <ArrowDownIcon className="w-3 h-3 text-primary" strokeWidth={2} />
        </span>
      </div>

      <div className="flex items-center gap-3.5 px-4 py-4 bg-gradient-to-br from-primary-200 to-primary-50 border-t border-primary/25">
        <span className="grid place-items-center w-12 h-12 rounded-xl bg-primary/20 border border-primary/40 text-white font-mono font-semibold shrink-0">
          {getInitials(displayName) || "?"}
        </span>
        <div className="min-w-0">
          <p className="font-mono text-2xs uppercase tracking-label text-primary mb-0.5">
            added to
          </p>
          <p className="text-base font-semibold text-text-primary truncate">
            {displayName}
          </p>
          {email && (
            <p className="font-mono text-xs text-text-secondary truncate">
              {email}
            </p>
          )}
        </div>
      </div>

      {/* An identity is scoped to one namespace, and this one comes from the SSHID
          the login used, so it is not necessarily the one being browsed. Wear the
          namespace selector's badge so it reads as a namespace, not a bare word. */}
      {details.namespace && (
        <div
          className={`flex items-center gap-3.5 px-4 py-3 border-t ${elsewhere ? "border-accent-yellow/30 bg-accent-yellow/[0.07]" : "border-border bg-black/20"}`}
        >
          <span
            className={`grid place-items-center w-9 h-9 rounded-lg border font-mono text-2xs font-bold shrink-0 ${elsewhere ? "bg-accent-yellow/15 border-accent-yellow/30 text-accent-yellow" : "bg-primary/15 border-primary/20 text-primary"}`}
          >
            {getInitials(details.namespace)}
          </span>
          <div className="min-w-0">
            <p
              className={`font-mono text-2xs uppercase tracking-label ${elsewhere ? "text-accent-yellow" : "text-text-muted"}`}
            >
              {elsewhere ? "not your active namespace" : "in the namespace"}
            </p>
            <p className="text-sm font-semibold text-text-primary truncate">
              {details.namespace}
            </p>
            {elsewhere && currentName && (
              <p className="text-2xs text-text-secondary truncate mt-0.5">
                Active namespace: {currentName}
              </p>
            )}
          </div>
        </div>
      )}

      <div className="flex justify-end px-4 py-2 border-t border-border bg-black/20">
        <button
          type="button"
          onClick={switchAccount}
          className="text-xs text-primary underline underline-offset-2 hover:text-primary-600"
        >
          Not you? Sign in as someone else
        </button>
      </div>
    </div>
  );
}

/**
 * How long re-authenticating lasts. The stamp lands on the identity, not on this
 * login, so within the window other logins with the same key go straight
 * through. A policy with no period asks every time, and then there is nothing to
 * promise.
 */
function reauthWindowSentence(periodSeconds: number) {
  if (periodSeconds <= 0) return "";

  const hours = Math.round(periodSeconds / 3600);
  const window =
    hours >= 1
      ? `${hours} ${hours === 1 ? "hour" : "hours"}`
      : `${Math.max(1, Math.round(periodSeconds / 60))} minutes`;

  return ` Connecting with it again won't ask for the next ${window}.`;
}

function formatRequestedAt(value: string) {
  const ms = Date.parse(value);
  if (Number.isNaN(ms)) return value || "—";
  return new Date(ms).toLocaleString();
}

function StatusMessage({ label }: { label: string }) {
  return (
    <div
      className="flex flex-col items-center gap-3 py-6"
      role="status"
      aria-live="polite"
    >
      <Spinner size="2xl" />
      <p className="text-sm text-text-muted">{label}</p>
    </div>
  );
}

function SpecRow({
  label,
  value,
  mono,
}: {
  label: string;
  value?: string;
  mono?: boolean;
}) {
  return (
    <div className="flex items-center justify-between gap-4 py-1.5 border-t border-border/70 first:border-t-0">
      <dt className="font-mono text-2xs uppercase tracking-wider text-text-muted">
        {label}
      </dt>
      <dd
        className={`font-mono text-xs truncate ${mono ? "text-primary tracking-widest" : "text-text-primary"}`}
      >
        {value || "—"}
      </dd>
    </div>
  );
}

const TONES = {
  error: {
    ring: "bg-accent-red/10 border-accent-red/20",
    icon: "text-accent-red",
  },
  success: {
    ring: "bg-accent-green/10 border-accent-green/20",
    icon: "text-accent-green",
  },
} as const;

function ResultMessage({
  tone,
  icon,
  title,
  description,
  action,
}: {
  tone: keyof typeof TONES;
  icon: React.ReactNode;
  title: string;
  description: React.ReactNode;
  action?: React.ReactNode;
}) {
  return (
    <div className="text-center animate-slide-up">
      <div
        className={`inline-flex items-center justify-center w-14 h-14 rounded-2xl border mb-5 ${TONES[tone].ring}`}
      >
        <span className={TONES[tone].icon}>{icon}</span>
      </div>
      <h2 className="text-lg font-semibold text-text-primary mb-2">{title}</h2>
      <p className="text-sm text-text-secondary leading-relaxed mb-6">
        {description}
      </p>
      {action}
    </div>
  );
}
