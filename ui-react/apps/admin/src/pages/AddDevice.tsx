import { useState } from "react";
import { Link } from "react-router-dom";
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
import { DockerIcon } from "../components/icons";
import CopyButton from "../components/common/CopyButton";

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

import { INPUT } from "../utils/styles";

const LABEL =
  "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";

/* ─── Page ─── */
export default function AddDevice() {
  const { tenant } = useAuthStore();
  const [method, setMethod] = useState<Method>("auto");
  const [showAllMethods, setShowAllMethods] = useState(false);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [hostname, setHostname] = useState("");
  const [identity, setIdentity] = useState("");

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
    parts.push("sh");
    return parts.join(" ");
  };

  const command = buildCommand();

  return (
    <div className="max-w-2xl animate-fade-in">
      {/* Breadcrumb */}
      <div className="flex items-center gap-1.5 mb-5">
        <Link
          to="/devices"
          className="text-2xs font-mono text-text-muted hover:text-primary transition-colors"
        >
          Devices
        </Link>
        <ChevronRightIcon
          className="w-3 h-3 text-text-muted/40"
          strokeWidth={2}
        />
        <span className="text-2xs font-mono text-text-secondary">
          Add Device
        </span>
      </div>

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
          <label className={LABEL}>Installation method</label>
        </div>
        <div className="space-y-2">
          {visibleMethods.map((m) => (
            <button
              key={m.id}
              type="button"
              onClick={() => {
                setMethod(m.id);
                setShowAllMethods(false);
              }}
              className={`flex items-start gap-3 w-full px-3.5 py-3 rounded-lg border text-left transition-all ${
                method === m.id
                  ? "bg-primary/[0.06] border-primary/30 ring-1 ring-primary/10"
                  : "bg-card border-border hover:bg-hover-subtle"
              }`}
            >
              <div
                className={`mt-0.5 shrink-0 w-4 h-4 rounded-full border-2 flex items-center justify-center transition-all ${method === m.id ? "border-primary" : "border-text-muted/40"}`}
              >
                {method === m.id && (
                  <div className="w-2 h-2 rounded-full bg-primary" />
                )}
              </div>
              <div className="flex items-start gap-2.5 min-w-0">
                <span
                  className={`mt-0.5 shrink-0 transition-colors ${method === m.id ? "text-primary" : "text-text-muted"}`}
                >
                  {m.icon}
                </span>
                <div className="min-w-0">
                  <span className="flex items-center gap-2">
                    <span
                      className={`text-sm font-medium transition-colors ${method === m.id ? "text-text-primary" : "text-text-secondary"}`}
                    >
                      {m.label}
                    </span>
                    {m.tag && (
                      <span
                        className={`px-1.5 py-0.5 text-3xs font-bold uppercase tracking-wider rounded border ${
                          m.tag === "Manual"
                            ? "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20"
                            : "bg-accent-green/15 text-accent-green border-accent-green/20"
                        }`}
                      >
                        {m.tag}
                      </span>
                    )}
                  </span>
                  <span className="block text-2xs text-text-muted mt-0.5">
                    {m.description}
                  </span>
                </div>
              </div>
            </button>
          ))}

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
          <label className={LABEL}>
            {selectedMethod.manual
              ? "Follow the documentation"
              : "Run on your device"}
          </label>
        </div>

        {selectedMethod.manual ? (
          <div className="bg-card border border-border rounded-xl p-5">
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
                <a
                  href={selectedMethod.docsUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1.5 px-3.5 py-2 bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20 rounded-lg text-xs font-medium hover:bg-accent-yellow/15 transition-all"
                >
                  View {selectedMethod.label} guide
                  <ArrowTopRightOnSquareIcon
                    className="w-3 h-3"
                    strokeWidth={2}
                  />
                </a>
              </div>
            </div>
          </div>
        ) : (
          <div className="bg-card border border-border rounded-xl overflow-hidden">
            {/* Terminal chrome */}
            <div className="flex items-center justify-between px-4 py-2.5 border-b border-border bg-surface/50">
              <div className="flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-full bg-accent-red/60" />
                <span className="w-2.5 h-2.5 rounded-full bg-accent-yellow/60" />
                <span className="w-2.5 h-2.5 rounded-full bg-accent-green/60" />
              </div>
              <span className="text-2xs font-mono text-text-muted/50">
                terminal
              </span>
              <CopyButton text={command} showLabel />
            </div>
            {/* Command */}
            <div className="p-4 overflow-x-auto">
              <pre className="text-xs font-mono text-accent-cyan leading-relaxed whitespace-pre-wrap break-all">
                <span className="text-text-muted select-none">$ </span>
                {command}
              </pre>
            </div>
          </div>
        )}
      </div>

      {/* Advanced options (only for non-manual methods) */}
      {!selectedMethod.manual && (
        <div className="mb-6">
          <button
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
            <div className="mt-3 bg-card border border-border rounded-xl p-4 space-y-4 animate-fade-in">
              <div>
                <label className={`block ${LABEL} mb-1.5`}>
                  Preferred Hostname{" "}
                  <span className="text-text-muted/50 normal-case tracking-normal">
                    (optional)
                  </span>
                </label>
                <input
                  type="text"
                  value={hostname}
                  onChange={(e) => setHostname(e.target.value)}
                  placeholder="my-device"
                  className={INPUT}
                />
                <p className="text-2xs text-text-muted/60 mt-1">
                  Override the device hostname reported to ShellHub.
                </p>
              </div>
              <div>
                <label className={`block ${LABEL} mb-1.5`}>
                  Preferred Identity{" "}
                  <span className="text-text-muted/50 normal-case tracking-normal">
                    (optional)
                  </span>
                </label>
                <input
                  type="text"
                  value={identity}
                  onChange={(e) => setIdentity(e.target.value)}
                  placeholder="server-01"
                  className={INPUT}
                />
                <p className="text-2xs text-text-muted/60 mt-1">
                  Set a custom identity string for the device.
                </p>
              </div>
            </div>
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
