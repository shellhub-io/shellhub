import { useEffect, useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import {
  CpuChipIcon,
  CheckCircleIcon,
  XCircleIcon,
  ArrowRightIcon,
  CommandLineIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { resolveDeviceLoginCode, acceptDevicePairing } from "@/client";
import { isSdkError } from "@/api/errors";
import { useAuthStore } from "@/stores/authStore";
import { useAcceptDevice } from "@/hooks/useDeviceMutations";
import { useSwitchNamespace } from "@/hooks/useNamespaceMutations";
import { useNamespaces } from "@/hooks/useNamespaces";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import RadioCard from "@/components/common/fields/RadioCard";
import PairingCodeForm from "@/components/common/PairingCodeForm";
import { Button, Spinner } from "@shellhub/design-system/primitives";
import { getInitials } from "@/utils/string";
import { getAcceptDeviceErrorMessage } from "@/utils/deviceErrors";

type DevicePreview = NonNullable<
  Awaited<ReturnType<typeof resolveDeviceLoginCode>>["data"]
>;

type Branch =
  | { kind: "loading" }
  | { kind: "missing-code" }
  | { kind: "error" }
  | { kind: "switching" }
  | { kind: "ready"; device: DevicePreview }
  | { kind: "pick-namespace"; device: DevicePreview }
  | { kind: "already-accepted"; device: DevicePreview }
  | { kind: "success"; device: DevicePreview }
  | {
      kind: "pairing-success";
      device: DevicePreview;
      uid: string;
      tenantId: string;
      namespace: string;
    };

/**
 * The end-to-end accept flow for a device login/pairing code: resolve the code,
 * preview the device, pick a namespace (for pairing codes), and accept.
 *
 * `code` is owned by the caller — the /accept-device page reads it from the URL;
 * the pairing modal keeps it in state. Passing `onCodeChange` puts the flow in
 * embedded (modal) mode: the no-code state collects the code in place instead of
 * navigating, and "back"/"try again" reset the code instead of leaving the page.
 */
export default function AcceptDeviceFlow({
  code,
  onCodeChange,
}: {
  code: string;
  onCodeChange?: (code: string) => void;
}) {
  const navigate = useNavigate();
  const authToken = useAuthStore((s) => s.token);
  const authTenant = useAuthStore((s) => s.tenant);

  const embedded = onCodeChange !== undefined;

  const acceptDevice = useAcceptDevice();
  const switchNamespace = useSwitchNamespace();

  const [branch, setBranch] = useState<Branch>({ kind: "loading" });
  const [actionError, setActionError] = useState("");
  const [selectedTenant, setSelectedTenant] = useState("");
  const [accepting, setAccepting] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function resolve() {
      if (!code) {
        if (!cancelled) setBranch({ kind: "missing-code" });
        return;
      }

      if (!authToken) {
        // The modal is only shown to signed-in users; the page sends anonymous
        // visitors through login and back.
        if (!embedded) {
          void navigate(
            `/login?redirect=${encodeURIComponent(`/accept-device?code=${code}`)}`,
          );
        }
        return;
      }

      try {
        const { data } = await resolveDeviceLoginCode({
          path: { code },
          throwOnError: true,
        });
        if (cancelled) return;

        // A pairing code belongs to a tenant-less agent: the device does not
        // exist yet and the user picks the namespace here.
        if (data.kind === "pairing") {
          setBranch({ kind: "pick-namespace", device: data });
          return;
        }

        if (data.status === "accepted") {
          setBranch({ kind: "already-accepted", device: data });
          return;
        }

        // Accepting requires a session scoped to the device's namespace.
        // useSwitchNamespace mints a namespace token and hard-navigates to the
        // accept page, so the resolve re-runs with the right tenant.
        if (data.tenant_id && data.tenant_id !== authTenant) {
          setBranch({ kind: "switching" });
          await switchNamespace.mutateAsync({
            tenantId: data.tenant_id,
            redirectTo: `/accept-device?code=${code}`,
          });
          return;
        }

        setBranch({ kind: "ready", device: data });
      } catch (err) {
        if (cancelled) return;
        // A stale session token 401s here; send the user through login and back
        // instead of mislabeling it as an expired code.
        if (isSdkError(err) && err.status === 401 && !embedded) {
          void navigate(
            `/login?redirect=${encodeURIComponent(`/accept-device?code=${code}`)}`,
          );
          return;
        }
        setBranch({ kind: "error" });
      }
    }

    void resolve();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [code, authToken, authTenant]);

  const handleAccept = async (device: DevicePreview) => {
    if (!device.uid) return;
    setActionError("");
    try {
      await acceptDevice.mutateAsync({ path: { uid: device.uid } });
      setBranch({ kind: "success", device });
    } catch (err) {
      setActionError(getAcceptDeviceErrorMessage(err));
    }
  };

  const handleAcceptPairing = async (device: DevicePreview) => {
    if (!selectedTenant) return;
    setActionError("");
    setAccepting(true);
    try {
      const { data } = await acceptDevicePairing({
        path: { code },
        body: { tenant_id: selectedTenant },
        throwOnError: true,
      });
      setBranch({
        kind: "pairing-success",
        device,
        uid: data.uid ?? "",
        tenantId: data.tenant_id ?? "",
        namespace: data.namespace ?? "",
      });
    } catch (err) {
      setActionError(getAcceptDeviceErrorMessage(err));
    } finally {
      setAccepting(false);
    }
  };

  return (
    <>
      {branch.kind === "loading" && <StatusMessage label="Checking code..." />}

      {branch.kind === "switching" && (
        <StatusMessage label="Switching namespace..." />
      )}

      {branch.kind === "missing-code" && (
        <div className="animate-fade-in">
          <div className="text-center mb-6">
            <div className="w-14 h-14 mx-auto mb-4 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center">
              <CommandLineIcon
                className="w-7 h-7 text-primary"
                strokeWidth={1.5}
              />
            </div>
            <h2 className="text-lg font-semibold text-text-primary">
              Claim a device
            </h2>
            <p className="text-sm text-text-muted mt-1">
              Enter the code your device is showing. Not showing one? Run{" "}
              <CommandChip /> to generate it.
            </p>
          </div>

          <PairingCodeForm onSubmit={onCodeChange} submitLabel="Claim device" />
        </div>
      )}

      {branch.kind === "error" && (
        <ResultMessage
          tone="error"
          icon={<XCircleIcon className="w-7 h-7" strokeWidth={1.5} />}
          title="Invalid or Expired Code"
          description={
            <>
              This code is invalid or has expired. Run <CommandChip /> on the
              device to get a new one.
            </>
          }
          action={
            onCodeChange ? (
              <Button
                variant="secondary"
                size="md"
                onClick={() => onCodeChange("")}
              >
                Enter another code
              </Button>
            ) : (
              <Button
                variant="secondary"
                size="md"
                as={Link}
                to="/accept-device"
              >
                Enter another code
              </Button>
            )
          }
        />
      )}

      {branch.kind === "already-accepted" && (
        <ResultMessage
          tone="success"
          icon={<CheckCircleIcon className="w-7 h-7" strokeWidth={1.5} />}
          title="Device Already Accepted"
          description={
            <>
              <span className="font-mono text-text-primary">
                {branch.device.name}
              </span>{" "}
              is already accepted into this namespace.
            </>
          }
          action={<ViewDeviceLink uid={branch.device.uid} />}
        />
      )}

      {branch.kind === "success" && (
        <ResultMessage
          tone="success"
          icon={<CheckCircleIcon className="w-7 h-7" strokeWidth={1.5} />}
          title="Device Accepted"
          description={
            <>
              <span className="font-mono text-text-primary">
                {branch.device.name}
              </span>{" "}
              joined{" "}
              <span className="text-text-primary font-medium">
                {branch.device.namespace}
              </span>
              . You can return to your terminal.
            </>
          }
          action={<ViewDeviceLink uid={branch.device.uid} />}
        />
      )}

      {branch.kind === "ready" && (
        <div className="text-center">
          <Reveal delay={0}>
            <div className="relative inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-primary/10 border border-primary/20 mb-5">
              <div className="absolute inset-0 rounded-2xl bg-primary/10 blur-xl animate-pulse-subtle" />
              <CpuChipIcon
                className="relative w-8 h-8 text-primary"
                strokeWidth={1.25}
              />
            </div>
          </Reveal>

          <Reveal delay={60}>
            <h2 className="text-lg font-semibold text-text-primary mb-2">
              Accept this device?
            </h2>
            <p className="text-sm text-text-secondary leading-relaxed mb-6">
              A device is asking to join{" "}
              <span className="text-text-primary font-medium">
                {branch.device.namespace}
              </span>
              . Review its identity before accepting.
            </p>
          </Reveal>

          <Reveal delay={120}>
            <dl className="text-left text-sm bg-surface/60 border border-border rounded-xl divide-y divide-border/70 overflow-hidden mb-6">
              <SpecRow label="hostname" value={branch.device.name} />
              <SpecRow label="os" value={branch.device.info?.pretty_name} />
              <SpecRow label="mac" value={branch.device.identity?.mac} />
              <SpecRow label="namespace" value={branch.device.namespace} />
              <div className="flex items-center justify-between gap-4 px-4 py-2.5">
                <dt className="font-mono text-2xs uppercase tracking-wider text-text-muted">
                  status
                </dt>
                <dd>
                  <span className="inline-flex items-center gap-1.5 rounded-full bg-accent-yellow/10 border border-accent-yellow/20 px-2.5 py-0.5 font-mono text-2xs text-accent-yellow">
                    <span className="w-1.5 h-1.5 rounded-full bg-accent-yellow animate-pulse-subtle" />
                    pending
                  </span>
                </dd>
              </div>
            </dl>
          </Reveal>

          {actionError && (
            <p
              className="text-sm text-accent-red mb-4 animate-shake"
              role="alert"
            >
              {actionError}
            </p>
          )}

          <Reveal delay={180}>
            <Button
              variant="primary"
              size="md"
              fullWidth
              loading={acceptDevice.isPending}
              icon={<CheckCircleIcon className="w-4 h-4" strokeWidth={2} />}
              onClick={() => void handleAccept(branch.device)}
            >
              Accept device
            </Button>

            <CancelRow onReset={onCodeChange} />
          </Reveal>
        </div>
      )}

      {branch.kind === "pick-namespace" && (
        <div className="text-center">
          <Reveal delay={0}>
            <div className="relative inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-primary/10 border border-primary/20 mb-5">
              <div className="absolute inset-0 rounded-2xl bg-primary/10 blur-xl animate-pulse-subtle" />
              <CpuChipIcon
                className="relative w-8 h-8 text-primary"
                strokeWidth={1.25}
              />
            </div>
          </Reveal>

          <Reveal delay={60}>
            <h2 className="text-lg font-semibold text-text-primary mb-2">
              Accept this device?
            </h2>
            <p className="text-sm text-text-secondary leading-relaxed mb-6">
              A device is asking to join one of your namespaces. Review its
              identity and choose where it belongs.
            </p>
          </Reveal>

          <Reveal delay={120}>
            <dl className="text-left text-sm bg-surface/60 border border-border rounded-xl divide-y divide-border/70 overflow-hidden mb-6">
              <SpecRow label="hostname" value={branch.device.name} />
              <SpecRow label="os" value={branch.device.info?.pretty_name} />
              <SpecRow label="mac" value={branch.device.identity?.mac} />
            </dl>
          </Reveal>

          <Reveal delay={180}>
            <div className="text-left mb-6">
              <NamespacePicker
                value={selectedTenant}
                onChange={setSelectedTenant}
                preferredTenant={authTenant ?? ""}
              />
            </div>
          </Reveal>

          {actionError && (
            <p
              className="text-sm text-accent-red mb-4 animate-shake"
              role="alert"
            >
              {actionError}
            </p>
          )}

          <Reveal delay={240}>
            <Button
              variant="primary"
              size="md"
              fullWidth
              loading={accepting}
              disabled={!selectedTenant}
              icon={<CheckCircleIcon className="w-4 h-4" strokeWidth={2} />}
              onClick={() => void handleAcceptPairing(branch.device)}
            >
              Accept device
            </Button>

            <CancelRow onReset={onCodeChange} />
          </Reveal>
        </div>
      )}

      {branch.kind === "pairing-success" && (
        <ResultMessage
          tone="success"
          icon={<CheckCircleIcon className="w-7 h-7" strokeWidth={1.5} />}
          title="Device Accepted"
          description={
            <>
              <span className="font-mono text-text-primary">
                {branch.device.name}
              </span>{" "}
              joined{" "}
              <span className="text-text-primary font-medium">
                {branch.namespace}
              </span>
              . The agent will connect automatically. You can return to your
              terminal.
            </>
          }
          action={
            branch.uid ? (
              <Button
                variant="successSoft"
                size="md"
                iconRight={
                  <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
                }
                onClick={() =>
                  void switchNamespace.mutateAsync({
                    tenantId: branch.tenantId,
                    redirectTo: `/devices/${branch.uid}`,
                  })
                }
              >
                View device
              </Button>
            ) : undefined
          }
        />
      )}
    </>
  );
}

/** Cancel affordance for the ready/pick states: reset to code entry when
 * embedded (modal), otherwise leave to the dashboard. */
function CancelRow({ onReset }: { onReset?: (code: string) => void }) {
  return (
    <div className="text-center mt-4">
      {onReset ? (
        <button
          type="button"
          onClick={() => onReset("")}
          className="text-2xs text-text-muted hover:text-text-secondary transition-colors"
        >
          Use a different code
        </button>
      ) : (
        <Link
          to="/dashboard"
          className="text-2xs text-text-muted hover:text-text-secondary transition-colors"
        >
          Cancel
        </Link>
      )}
    </div>
  );
}

/** Namespace selector for pairing codes: the device has no namespace yet, so
 * the user picks one of theirs. Preselects the session's namespace. */
function NamespacePicker({
  value,
  onChange,
  preferredTenant,
}: {
  value: string;
  onChange: (tenant: string) => void;
  preferredTenant: string;
}) {
  const { namespaces, isLoading } = useNamespaces();

  useEffect(() => {
    if (value || namespaces.length === 0) return;
    const preferred = namespaces.find((n) => n.tenant_id === preferredTenant);
    onChange(preferred?.tenant_id ?? namespaces[0].tenant_id ?? "");
  }, [value, namespaces, preferredTenant, onChange]);

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-sm text-text-muted">
        <Spinner size="sm" />
        Loading namespaces...
      </div>
    );
  }

  if (namespaces.length === 0) {
    return (
      <p className="text-sm text-text-secondary">
        You don&apos;t belong to any namespace yet. Create one in the console
        and try again.
      </p>
    );
  }

  return (
    <RadioGroupField label="Namespace" value={value} onChange={onChange}>
      {namespaces.map((namespace) => (
        <RadioCard
          key={namespace.tenant_id}
          value={namespace.tenant_id ?? ""}
          icon={
            <span className="w-6 h-6 rounded bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-2xs font-bold font-mono">
              {getInitials(namespace.name ?? "")}
            </span>
          }
          label={namespace.name ?? ""}
          description={namespace.tenant_id ?? ""}
        />
      ))}
    </RadioGroupField>
  );
}

/** Staggered slide-up reveal for the ready-state composition. */
function Reveal({
  delay,
  children,
}: {
  delay: number;
  children: React.ReactNode;
}) {
  return (
    <div
      className="animate-slide-up [animation-fill-mode:backwards]"
      style={{ animationDelay: `${delay}ms` }}
    >
      {children}
    </div>
  );
}

function CommandChip() {
  return (
    <code className="inline-flex items-center gap-1 rounded bg-surface border border-border px-1.5 py-0.5 font-mono text-xs text-text-primary whitespace-nowrap">
      <CommandLineIcon
        className="w-3.5 h-3.5 text-text-muted"
        strokeWidth={1.5}
      />
      shellhub-agent login
    </code>
  );
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

function ViewDeviceLink({ uid }: { uid?: string }) {
  if (!uid) return null;
  return (
    <Button
      as={Link}
      to={`/devices/${uid}`}
      variant="successSoft"
      size="md"
      iconRight={<ArrowRightIcon className="w-4 h-4" strokeWidth={2} />}
    >
      View device
    </Button>
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
        className={cn(
          "inline-flex items-center justify-center w-14 h-14 rounded-2xl border mb-5",
          TONES[tone].ring,
        )}
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
