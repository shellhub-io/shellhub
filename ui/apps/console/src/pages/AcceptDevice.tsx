import { useEffect, useState } from "react";
import { useNavigate, useSearchParams, Link } from "react-router-dom";
import {
  CpuChipIcon,
  CheckCircleIcon,
  XCircleIcon,
  ArrowRightIcon,
  CommandLineIcon,
} from "@heroicons/react/24/outline";
import { resolveDeviceLoginCode, acceptDevicePairing } from "@/client";
import { isSdkError } from "@/api/errors";
import { useAuthStore } from "@/stores/authStore";
import { useAcceptDevice } from "@/hooks/useDeviceMutations";
import { useSwitchNamespace } from "@/hooks/useNamespaceMutations";
import { useNamespaces } from "@/hooks/useNamespaces";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import RadioCard from "@/components/common/fields/RadioCard";
import Spinner from "@/components/common/Spinner";
import { getInitials } from "@/utils/string";

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

export default function AcceptDevice() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const authToken = useAuthStore((s) => s.token);
  const authTenant = useAuthStore((s) => s.tenant);

  const code = searchParams.get("code") ?? "";

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
        const redirectTarget = `/accept-device?${searchParams.toString()}`;
        void navigate(`/login?redirect=${encodeURIComponent(redirectTarget)}`);
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
        // useSwitchNamespace mints a namespace token and hard-navigates back
        // to this page, so the resolve re-runs with the right tenant.
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
        // A stale session token 401s here; send the user through login and
        // back instead of mislabeling it as an expired code.
        if (isSdkError(err) && err.status === 401) {
          const redirectTarget = `/accept-device?${searchParams.toString()}`;
          void navigate(
            `/login?redirect=${encodeURIComponent(redirectTarget)}`,
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
    } catch {
      setActionError("Failed to accept the device. Please try again.");
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
      if (isSdkError(err) && err.status === 403) {
        setActionError(
          "You don't have permission to accept devices into this namespace.",
        );
      } else if (isSdkError(err) && err.status === 402) {
        setActionError("This namespace reached its device limit.");
      } else if (isSdkError(err) && err.status === 409) {
        setActionError(
          "A device with this hostname already exists in this namespace.",
        );
      } else {
        setActionError("Failed to accept the device. Please try again.");
      }
    } finally {
      setAccepting(false);
    }
  };

  return (
    <div className="w-full max-w-md mx-auto animate-fade-in">
      <div className="bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm">
        {branch.kind === "loading" && (
          <StatusMessage label="Checking code..." />
        )}

        {branch.kind === "switching" && (
          <StatusMessage label="Switching namespace..." />
        )}

        {branch.kind === "missing-code" && (
          <ResultMessage
            tone="error"
            icon={<XCircleIcon className="w-7 h-7" strokeWidth={1.5} />}
            title="Invalid Link"
            description={
              <>
                This link is missing the device code. Run <CommandChip /> on the
                device and use the URL it prints.
              </>
            }
          />
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

            {/* Identity ledger — reads like the terminal it came from */}
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
              <button
                type="button"
                onClick={() => void handleAccept(branch.device)}
                disabled={acceptDevice.isPending}
                className="w-full bg-primary hover:bg-primary-600 text-white py-3 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 flex items-center justify-center gap-2"
              >
                {acceptDevice.isPending ? (
                  <>
                    <Spinner size="sm" tone="onPrimary" />
                    Accepting...
                  </>
                ) : (
                  <>
                    <CheckCircleIcon className="w-4 h-4" strokeWidth={2} />
                    Accept Device
                  </>
                )}
              </button>

              <div className="text-center mt-4">
                <Link
                  to="/dashboard"
                  className="text-2xs text-text-muted hover:text-text-secondary transition-colors"
                >
                  Cancel
                </Link>
              </div>
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
              <button
                type="button"
                onClick={() => void handleAcceptPairing(branch.device)}
                disabled={accepting || !selectedTenant}
                className="w-full bg-primary hover:bg-primary-600 text-white py-3 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 flex items-center justify-center gap-2"
              >
                {accepting ? (
                  <>
                    <Spinner size="sm" tone="onPrimary" />
                    Accepting...
                  </>
                ) : (
                  <>
                    <CheckCircleIcon className="w-4 h-4" strokeWidth={2} />
                    Accept Device
                  </>
                )}
              </button>

              <div className="text-center mt-4">
                <Link
                  to="/dashboard"
                  className="text-2xs text-text-muted hover:text-text-secondary transition-colors"
                >
                  Cancel
                </Link>
              </div>
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
                . The agent will connect automatically — you can return to your
                terminal.
              </>
            }
            action={
              branch.uid ? (
                <button
                  type="button"
                  onClick={() =>
                    void switchNamespace.mutateAsync({
                      tenantId: branch.tenantId,
                      redirectTo: `/devices/${branch.uid}`,
                    })
                  }
                  className="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-5 py-2.5 rounded-lg text-sm font-semibold transition-all shadow-lg shadow-primary/20"
                >
                  View Device
                  <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
                </button>
              ) : undefined
            }
          />
        )}
      </div>
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
        and open this link again.
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
    <Link
      to={`/devices/${uid}`}
      className="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-5 py-2.5 rounded-lg text-sm font-semibold transition-all shadow-lg shadow-primary/20"
    >
      View Device
      <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
    </Link>
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
