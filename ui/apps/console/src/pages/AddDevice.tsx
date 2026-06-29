import { useState, type JSX } from "react";
import { Link } from "react-router-dom";
import Breadcrumb from "@/components/common/Breadcrumb";
import { useAuthStore } from "../stores/authStore";
import {
  ArchiveBoxIcon,
  ArrowTopRightOnSquareIcon,
  BookOpenIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  ChevronUpIcon,
  CommandLineIcon,
  ComputerDesktopIcon,
  CpuChipIcon,
  CubeIcon,
  InformationCircleIcon,
  PlusIcon,
  ServerStackIcon,
  SparklesIcon,
  WrenchIcon,
} from "@heroicons/react/24/outline";
import CopyButton from "../components/common/CopyButton";
import InputField from "@/components/common/fields/InputField";
import NumericInput from "@/components/common/fields/NumericInput";
import RadioCard from "@/components/common/fields/RadioCard";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import { LABEL_BASE } from "@/utils/styles";
import {
  Button,
  Card,
  DockerIcon,
  WindowChrome,
} from "@shellhub/design-system/primitives";

/* ─── Types ─── */
type Method =
  | "auto"
  | "docker"
  | "podman"
  | "snap"
  | "standalone"
  | "wsl"
  | "yocto"
  | "buildroot"
  | "freebsd";

interface MethodInfo {
  id: Method;
  label: string;
  tag?: string;
  description: string;
  icon: JSX.Element;
  manual?: boolean;
  docsUrl?: string;
}

const INITIAL_VISIBLE = 3;

/* ─── Constants ─── */
const METHODS: MethodInfo[] = [
  {
    id: "auto",
    label: "Auto Detect",
    tag: "Recommended",
    description:
      "Automatically detects Docker, Snap, or Standalone and uses the best available method.",
    icon: <SparklesIcon className="w-5 h-5" />,
  },
  {
    id: "docker",
    label: "Docker",
    description:
      "Run the agent as a Docker container. Requires Docker daemon running on the host.",
    icon: <DockerIcon className="w-5 h-5" />,
  },
  {
    id: "standalone",
    label: "Standalone",
    description:
      "Install directly using runc and systemd. No container runtime required.",
    icon: <ServerStackIcon className="w-5 h-5" />,
  },
  {
    id: "podman",
    label: "Podman",
    description:
      "Alternative to Docker with rootless container capabilities. Requires Podman daemon.",
    icon: <CubeIcon className="w-5 h-5" />,
  },
  {
    id: "snap",
    label: "Snap",
    description:
      "Easy installation via Snap store with automatic updates. Requires snapd service.",
    icon: <ArchiveBoxIcon className="w-5 h-5" />,
  },
  {
    id: "wsl",
    label: "WSL",
    description:
      "Optimized for Windows Subsystem for Linux 2 with systemd and mirrored networking.",
    icon: <ComputerDesktopIcon className="w-5 h-5" />,
  },
  {
    id: "yocto",
    label: "Yocto Project",
    tag: "Manual",
    description:
      "For embedded Linux systems built with the Yocto build system.",
    manual: true,
    docsUrl: "https://docs.shellhub.io/overview/supported-platforms/yocto",
    icon: <CpuChipIcon className="w-5 h-5" />,
  },
  {
    id: "buildroot",
    label: "Buildroot",
    tag: "Manual",
    description: "For embedded Linux systems built with Buildroot toolchain.",
    manual: true,
    docsUrl: "https://docs.shellhub.io/overview/supported-platforms/buildroot",
    icon: <WrenchIcon className="w-5 h-5" />,
  },
  {
    id: "freebsd",
    label: "FreeBSD",
    tag: "Manual",
    description:
      "For FreeBSD systems. Requires ports tree and manual compilation.",
    manual: true,
    docsUrl: "https://docs.shellhub.io/overview/supported-platforms/freebsd",
    icon: <CommandLineIcon className="w-5 h-5" />,
  },
];

/* ─── Page ─── */
export default function AddDevice() {
  const { tenant } = useAuthStore();
  const [method, setMethod] = useState<Method>("auto");
  const [showAllMethods, setShowAllMethods] = useState(false);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [hostname, setHostname] = useState("");
  const [identity, setIdentity] = useState("");
  const [keepaliveInterval, setKeepaliveInterval] = useState("");
  const keepaliveIntervalError =
    keepaliveInterval && parseInt(keepaliveInterval, 10) < 1
      ? "Interval must be a positive number"
      : "";

  const selectedMethod = METHODS.find((m) => m.id === method)!;
  const baseMethods = METHODS.slice(0, INITIAL_VISIBLE);
  const selectedInBase = baseMethods.some((m) => m.id === method);
  const visibleMethods = showAllMethods
    ? METHODS
    : selectedInBase
      ? baseMethods
      : [...baseMethods, selectedMethod];

  const origin = window.location.origin;

  const buildCommand = () => {
    const parts = ["curl -sSf", `${origin}/install.sh`, "|"];
    if (method !== "auto") parts.push(`INSTALL_METHOD=${method}`);
    parts.push(`TENANT_ID=${tenant}`);
    parts.push(`SERVER_ADDRESS=${origin}`);
    if (hostname.trim()) parts.push(`PREFERRED_HOSTNAME=${hostname.trim()}`);
    if (identity.trim()) parts.push(`PREFERRED_IDENTITY=${identity.trim()}`);
    if (keepaliveInterval.trim() && !keepaliveIntervalError)
      parts.push(`KEEPALIVE_INTERVAL=${parseInt(keepaliveInterval, 10)}`);
    parts.push("sh");
    return parts.join(" ");
  };

  const command = buildCommand();

  return (
    <div className="max-w-2xl animate-fade-in">
      <Breadcrumb
        items={[{ label: "Devices", to: "/devices" }, { label: "Add Device" }]}
      />

      {/* Header */}
      <div className="flex items-start gap-4 mb-8">
        <div className="w-12 h-12 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0">
          <PlusIcon className="w-6 h-6 text-primary" />
        </div>
        <div>
          <h1 className="text-xl font-bold text-text-primary">Add Device</h1>
          <p className="text-sm text-text-muted mt-1">
            Install the ShellHub agent on your device to connect it to your
            namespace.
          </p>
        </div>
      </div>

      {/* Step 1 — Method */}
      <div className="mb-6">
        <div className="flex items-center gap-2.5 mb-3">
          <span className="w-5 h-5 rounded-full bg-primary/15 border border-primary/25 flex items-center justify-center text-2xs font-bold text-primary">
            1
          </span>
          <span id="add-device-method-label" className={LABEL_BASE}>
            Installation method
          </span>
        </div>
        <div className="space-y-2">
          <RadioGroupField
            labelledBy="add-device-method-label"
            value={method}
            onChange={setMethod}
          >
            {visibleMethods.map((m) => (
              <RadioCard
                key={m.id}
                value={m.id}
                icon={m.icon}
                label={m.label}
                description={m.description}
                adornment={
                  m.tag && (
                    <span
                      className={`px-1.5 py-0.5 text-3xs font-bold uppercase tracking-wider rounded border ${
                        m.tag === "Manual"
                          ? "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20"
                          : "bg-accent-green/15 text-accent-green border-accent-green/20"
                      }`}
                    >
                      {m.tag}
                    </span>
                  )
                }
              />
            ))}
          </RadioGroupField>

          {!showAllMethods && (
            <button
              type="button"
              onClick={() => setShowAllMethods(true)}
              className="flex items-center justify-center gap-1.5 w-full py-2 text-2xs font-mono text-text-muted hover:text-primary transition-colors"
            >
              <ChevronDownIcon className="w-3 h-3" strokeWidth={2} />
              Show all methods
            </button>
          )}
          {showAllMethods && (
            <button
              type="button"
              onClick={() => setShowAllMethods(false)}
              className="flex items-center justify-center gap-1.5 w-full py-2 text-2xs font-mono text-text-muted hover:text-primary transition-colors"
            >
              <ChevronUpIcon className="w-3 h-3" strokeWidth={2} />
              Show fewer
            </button>
          )}
        </div>
      </div>

      {/* Step 2 — Command or Manual Instructions */}
      <div className="mb-6">
        <div className="flex items-center gap-2.5 mb-3">
          <span className="w-5 h-5 rounded-full bg-primary/15 border border-primary/25 flex items-center justify-center text-2xs font-bold text-primary">
            2
          </span>
          <span className={LABEL_BASE}>
            {selectedMethod.manual
              ? "Follow the documentation"
              : "Run on your device"}
          </span>
        </div>

        {selectedMethod.manual ? (
          <Card className="p-5">
            <div className="flex items-start gap-3">
              <div className="w-9 h-9 rounded-lg bg-accent-yellow/10 border border-accent-yellow/20 flex items-center justify-center shrink-0">
                <BookOpenIcon className="w-4.5 h-4.5 text-accent-yellow" />
              </div>
              <div>
                <p className="text-sm text-text-primary font-medium mb-1">
                  Manual installation required
                </p>
                <p className="text-2xs text-text-muted leading-relaxed mb-3">
                  {selectedMethod.label} requires manual setup. Follow the
                  platform-specific documentation for step-by-step instructions.
                </p>
                <Button
                  variant="warningSoft"
                  as="a"
                  size="sm"
                  href={selectedMethod.docsUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  iconRight={
                    <ArrowTopRightOnSquareIcon
                      className="w-3 h-3"
                      strokeWidth={2}
                    />
                  }
                >
                  View {selectedMethod.label} guide
                </Button>
              </div>
            </div>
          </Card>
        ) : (
          <WindowChrome
            variant="terminal"
            size="sm"
            titleBarSlot={<CopyButton text={command} showLabel />}
          >
            <pre className="text-accent-cyan whitespace-pre-wrap break-all">
              <span className="text-text-muted select-none">$ </span>
              {command}
            </pre>
          </WindowChrome>
        )}
      </div>

      {/* Advanced options (only for non-manual methods) */}
      {!selectedMethod.manual && (
        <div className="mb-6">
          <button
            type="button"
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="flex items-center gap-1.5 text-2xs font-mono text-text-muted hover:text-text-secondary transition-colors"
          >
            <ChevronRightIcon
              className={`w-3 h-3 transition-transform ${showAdvanced ? "rotate-90" : ""}`}
              strokeWidth={2}
            />
            Advanced options
          </button>

          {showAdvanced && (
            <Card className="mt-3 p-4 space-y-4 animate-fade-in">
              <InputField
                id="add-device-hostname"
                label="Preferred Hostname"
                labelAdornment={
                  <span className="text-text-muted/50 normal-case tracking-normal text-2xs">
                    (optional)
                  </span>
                }
                value={hostname}
                onChange={setHostname}
                placeholder="my-device"
                hint="Override the device hostname reported to ShellHub."
              />
              <InputField
                id="add-device-identity"
                label="Preferred Identity"
                labelAdornment={
                  <span className="text-text-muted/50 normal-case tracking-normal text-2xs">
                    (optional)
                  </span>
                }
                value={identity}
                onChange={setIdentity}
                placeholder="server-01"
                hint="Set a custom identity string for the device."
              />
              <NumericInput
                id="add-device-keepalive"
                label="Keep Alive Interval"
                labelAdornment={
                  <span className="text-text-muted/50 normal-case tracking-normal text-2xs">
                    (optional)
                  </span>
                }
                value={keepaliveInterval}
                onChange={setKeepaliveInterval}
                placeholder="30"
                hint="Interval in seconds between keep-alive messages sent by the agent. Defaults to 30."
                error={keepaliveIntervalError || undefined}
              />
            </Card>
          )}
        </div>
      )}

      {/* Info */}
      <div className="flex items-start gap-3 bg-primary/[0.04] border border-primary/15 rounded-xl px-4 py-3.5 mb-6">
        <InformationCircleIcon className="w-4 h-4 text-primary shrink-0 mt-0.5" />
        <div className="text-xs text-text-secondary leading-relaxed">
          After installing, your device will appear in the{" "}
          <Link
            to="/devices"
            className="text-primary font-medium hover:text-primary/80 transition-colors"
          >
            Pending tab
          </Link>{" "}
          and must be accepted before you can connect to it.
        </div>
      </div>

      {/* Docs link */}
      <div className="flex items-center gap-2 text-2xs text-text-muted">
        <BookOpenIcon className="w-3.5 h-3.5" />
        <a
          href="https://docs.shellhub.io/user-guides/devices/adding"
          target="_blank"
          rel="noopener noreferrer"
          className="hover:text-primary transition-colors"
        >
          Check the documentation for more installation methods and details
        </a>
        <ArrowTopRightOnSquareIcon className="w-3 h-3" strokeWidth={2} />
      </div>
    </div>
  );
}
