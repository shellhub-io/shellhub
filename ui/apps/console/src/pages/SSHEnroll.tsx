import { useParams, useNavigate } from "react-router-dom";
import {
  FingerPrintIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  NoSymbolIcon,
} from "@heroicons/react/24/outline";
import { Button, Spinner } from "@shellhub/design-system/primitives";
import BaseDialog from "@/components/common/BaseDialog";
import { useSSHEnrollment, EnrollmentDetails } from "@/hooks/useSSHEnrollment";

/**
 * Modal for enrolling an SSH key for a pure-OpenSSH login the gateway is holding
 * open. It opens over the SSH Identities page (route /ssh-identities/enroll/
 * <code>): the user is shown that URL in their terminal, opens it here (already
 * logged in), confirms the presented key's fingerprint and source IP are theirs
 * by matching the correlation code with their terminal, and enrolls the key
 * (trust-on-first-use). The same surface serves the per-session step-up.
 */
export default function SSHEnroll() {
  const { code = "" } = useParams<{ code: string }>();
  const navigate = useNavigate();
  const {
    phase,
    details,
    secondsLeft,
    confirm,
    reject,
    deciding,
    actionError,
  } = useSSHEnrollment(code);

  const close = () => {
    void navigate("/ssh-identities");
  };

  return (
    <BaseDialog open onClose={close} size="md" aria-label="Enroll SSH key">
      <div className="p-8">
        {phase === "loading" && <StatusMessage label="Loading request..." />}

        {phase === "pending" && details && (
          <PendingRequest
            details={details}
            secondsLeft={secondsLeft}
            deciding={deciding}
            actionError={actionError}
            onConfirm={() => void confirm()}
            onReject={() => void reject()}
          />
        )}

        {phase === "confirmed" && (
          <ResultMessage
            tone="success"
            icon={<CheckCircleIcon className="w-7 h-7" strokeWidth={1.5} />}
            title="Enrolled"
            description="The key is now linked to your account. Return to your terminal — the connection will continue."
            action={<DoneButton onClick={close} />}
          />
        )}

        {phase === "rejected" && (
          <ResultMessage
            tone="error"
            icon={<NoSymbolIcon className="w-7 h-7" strokeWidth={1.5} />}
            title="Rejected"
            description="The request was rejected. The connection will not be established."
            action={<DoneButton onClick={close} />}
          />
        )}

        {phase === "expired" && (
          <ResultMessage
            tone="error"
            icon={<XCircleIcon className="w-7 h-7" strokeWidth={1.5} />}
            title="Request Expired"
            description="This enrollment request has expired or is no longer valid. Start the SSH login again to get a new one."
            action={<DoneButton onClick={close} />}
          />
        )}

        {phase === "error" && (
          <ResultMessage
            tone="error"
            icon={<XCircleIcon className="w-7 h-7" strokeWidth={1.5} />}
            title="Something Went Wrong"
            description="We couldn't load this enrollment request. Please try again."
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
  secondsLeft,
  deciding,
  actionError,
  onConfirm,
  onReject,
}: {
  details: EnrollmentDetails;
  secondsLeft: number;
  deciding: boolean;
  actionError: string;
  onConfirm: () => void;
  onReject: () => void;
}) {
  const stepUp = !details.enroll;

  return (
    <div className="text-center">
      <div className="relative inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-primary/10 border border-primary/20 mb-5">
        <div className="absolute inset-0 rounded-2xl bg-primary/10 blur-xl animate-pulse-subtle" />
        <FingerPrintIcon
          className="relative w-8 h-8 text-primary"
          strokeWidth={1.25}
        />
      </div>

      <h2 className="text-lg font-semibold text-text-primary mb-2">
        {stepUp ? "Confirm this session?" : "Enroll this key?"}
      </h2>
      <p className="text-sm text-text-secondary leading-relaxed mb-5">
        {stepUp
          ? "This session needs your confirmation. Only continue if the fingerprint and source below match your terminal."
          : "Linking this key to your account signs you in with it from now on. Only enroll if the fingerprint and source below match your terminal."}
      </p>

      {/* Fingerprint + source IP — the substantive anti-phishing check. */}
      <dl className="text-left text-sm bg-surface/60 border border-border rounded-xl divide-y divide-border/70 overflow-hidden mb-4">
        <SpecRow label="fingerprint" value={details.fingerprint} />
        <SpecRow label="source ip" value={details.ipAddress} />
        <SpecRow label="username" value={details.username} />
        <SpecRow label="device" value={details.deviceName} />
        <SpecRow
          label="requested"
          value={formatRequestedAt(details.requestedAt)}
        />
      </dl>

      {/* Correlation code — visually match against the terminal banner. */}
      {details.code && (
        <div className="mb-4">
          <p className="text-2xs font-mono uppercase tracking-wider text-text-muted mb-1.5">
            Match this code with your terminal
          </p>
          <div className="inline-flex items-center justify-center px-4 py-2 rounded-xl bg-primary/10 border border-primary/20">
            <span className="font-mono text-xl font-semibold tracking-[0.2em] text-primary">
              {details.code}
            </span>
          </div>
        </div>
      )}

      <div className="flex items-center justify-center gap-1.5 text-2xs font-mono text-text-muted mb-5">
        <ClockIcon className="w-3.5 h-3.5" strokeWidth={1.5} />
        Expires in {secondsLeft}s
      </div>

      {actionError && (
        <p className="text-sm text-accent-red mb-4 animate-shake" role="alert">
          {actionError}
        </p>
      )}

      <div className="flex items-center gap-3">
        <Button
          variant="secondary"
          size="md"
          fullWidth
          disabled={deciding}
          icon={<NoSymbolIcon className="w-4 h-4" strokeWidth={2} />}
          onClick={onReject}
        >
          Reject
        </Button>
        <Button
          variant="primary"
          size="md"
          fullWidth
          loading={deciding}
          icon={<CheckCircleIcon className="w-4 h-4" strokeWidth={2} />}
          onClick={onConfirm}
        >
          {stepUp ? "Confirm" : "Enroll key"}
        </Button>
      </div>
    </div>
  );
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

function SpecRow({ label, value }: { label: string; value?: string }) {
  return (
    <div className="flex items-center justify-between gap-4 px-4 py-2.5">
      <dt className="font-mono text-2xs uppercase tracking-wider text-text-muted">
        {label}
      </dt>
      <dd className="font-mono text-text-primary truncate">{value || "—"}</dd>
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
      <h2 className="text-lg font-semibold text-text-primary mb-3">{title}</h2>
      <p className="text-sm text-text-secondary leading-relaxed mb-6">
        {description}
      </p>
      {action}
    </div>
  );
}
