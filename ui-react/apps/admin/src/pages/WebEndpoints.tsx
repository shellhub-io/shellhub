import { useEffect, useState, useRef, FormEvent } from "react";
import { useWebEndpointsStore } from "../stores/webEndpointsStore";
import { WebEndpoint } from "../types/webEndpoint";
import { Device } from "../types/device";
import { getDevices } from "../api/devices";
import PageHeader from "../components/common/PageHeader";
import Drawer from "../components/common/Drawer";
import ConfirmDialog from "../components/common/ConfirmDialog";
import { formatDate } from "../utils/date";
import { LABEL, INPUT_MONO } from "../utils/styles";
import { useClickOutside } from "../hooks/useClickOutside";
import {
  XMarkIcon,
  ServerStackIcon,
  ArrowsRightLeftIcon,
  ExclamationCircleIcon,
  GlobeAltIcon,
  LockClosedIcon,
  TrashIcon,
  LinkIcon,
  ClockIcon,
  PlusIcon,
  MagnifyingGlassIcon,
} from "@heroicons/react/24/outline";

import axios from "axios";

/* ─── Constants ─── */

const GO_ZERO_TIME = "0001-01-01T00:00:00Z";

function neverExpires(expiresIn: string): boolean {
  return !expiresIn || expiresIn === GO_ZERO_TIME;
}

function isExpired(expiresIn: string): boolean {
  if (neverExpires(expiresIn)) return false;
  return new Date(expiresIn).getTime() < Date.now();
}

function formatExpiration(expiresIn: string): string {
  if (neverExpires(expiresIn)) return "Never expires";
  const d = new Date(expiresIn);
  const now = new Date();
  if (d.getTime() < now.getTime()) {
    return `Expired ${formatDate(expiresIn)}`;
  }
  const diffMs = d.getTime() - now.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  if (diffMins < 60) return `${diffMins}m remaining`;
  const diffHrs = Math.floor(diffMins / 60);
  if (diffHrs < 24) return `${diffHrs}h remaining`;
  const diffDays = Math.floor(diffHrs / 24);
  if (diffDays < 30) return `${diffDays}d remaining`;
  return d.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function isValidIPv4(ip: string): boolean {
  const parts = ip.split(".");
  if (parts.length !== 4) return false;
  return parts.every((p) => {
    const n = Number(p);
    return /^\d{1,3}$/.test(p) && n >= 0 && n <= 255;
  });
}

function isValidIPv6(ip: string): boolean {
  return /^([0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}$/.test(ip) || ip === "::1";
}

function isValidHost(host: string): boolean {
  return isValidIPv4(host) || isValidIPv6(host);
}

function isValidFQDN(domain: string): boolean {
  return /^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/.test(
    domain,
  );
}

const MAX_CUSTOM_TTL = 9223372036;

/* ─── Device Selector ─── */
function DeviceSelector({
  selected,
  onChange,
  error,
}: {
  selected: Device | null;
  onChange: (device: Device | null) => void;
  error?: string;
}) {
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const wrapperRef = useRef<HTMLDivElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>();

  useClickOutside(wrapperRef, () => setOpen(false));

  useEffect(() => {
    if (!open) return;
    clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      setLoading(true);
      getDevices(1, 20, "accepted")
        .then(({ data: d }) => setDevices(d))
        .catch(() => setDevices([]))
        .finally(() => setLoading(false));
    }, 300);
    return () => clearTimeout(debounceRef.current);
  }, [open]);

  const handleSearch = (value: string) => {
    setSearch(value);
    clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      setLoading(true);
      getDevices(1, 20, "accepted")
        .then(({ data: d }) => {
          const filtered = value
            ? d.filter(
                (dev) =>
                  dev.name.toLowerCase().includes(value.toLowerCase()) ||
                  dev.uid.toLowerCase().includes(value.toLowerCase()),
              )
            : d;
          setDevices(filtered);
        })
        .catch(() => setDevices([]))
        .finally(() => setLoading(false));
    }, 300);
  };

  return (
    <div ref={wrapperRef} className="relative">
      <div
        className={`flex items-center min-h-[42px] px-3.5 py-2 bg-card border rounded-lg cursor-text transition-all ${
          open ? "border-primary/50 ring-1 ring-primary/20" : "border-border"
        } ${error ? "border-accent-red/50" : ""}`}
        onClick={() => setOpen(true)}
      >
        {selected ? (
          <div className="flex items-center gap-2 flex-1 min-w-0">
            <span
              className={`w-2 h-2 rounded-full shrink-0 ${selected.online ? "bg-accent-green" : "bg-text-muted/40"}`}
            />
            <span className="text-sm text-text-primary truncate">
              {selected.name}
            </span>
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                onChange(null);
                setSearch("");
              }}
              className="ml-auto shrink-0 p-0.5 text-text-muted hover:text-text-primary transition-colors"
            >
              <XMarkIcon className="w-3.5 h-3.5" strokeWidth={2} />
            </button>
          </div>
        ) : (
          <input
            type="text"
            value={search}
            onChange={(e) => handleSearch(e.target.value)}
            onFocus={() => setOpen(true)}
            placeholder="Search devices..."
            className="flex-1 bg-transparent text-sm text-text-primary placeholder:text-text-secondary outline-none"
          />
        )}
      </div>
      {error && <p className="mt-1 text-2xs text-accent-red">{error}</p>}
      {open && !selected && (
        <div className="absolute z-10 mt-1 w-full max-h-48 overflow-y-auto bg-surface border border-border rounded-lg shadow-xl">
          {loading ? (
            <div className="px-3 py-2 text-xs text-text-muted">
              Loading devices...
            </div>
          ) : devices.length === 0 ? (
            <div className="px-3 py-2 text-xs text-text-muted">
              No devices found
            </div>
          ) : (
            devices.map((dev) => (
              <button
                key={dev.uid}
                type="button"
                onClick={() => {
                  onChange(dev);
                  setOpen(false);
                  setSearch("");
                }}
                className="w-full text-left px-3 py-2 text-sm text-text-primary hover:bg-hover-medium transition-colors flex items-center gap-2"
              >
                <span
                  className={`w-2 h-2 rounded-full shrink-0 ${dev.online ? "bg-accent-green" : "bg-text-muted/40"}`}
                />
                <span className="truncate">{dev.name}</span>
                <span className="text-2xs text-text-muted font-mono ml-auto shrink-0">
                  {dev.uid.slice(0, 8)}
                </span>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}

/* ─── Timeout Selector ─── */
const EXPIRATION_PRESETS = [
  { label: "1 min", value: 60 },
  { label: "5 min", value: 300 },
  { label: "15 min", value: 900 },
  { label: "1 hour", value: 3600 },
  { label: "1 day", value: 86400 },
  { label: "1 week", value: 604800 },
  { label: "1 month", value: 2624016 },
] as const;

function TimeoutSelector({
  value,
  onChange,
}: {
  value: number;
  onChange: (v: number) => void;
}) {
  const hasExpiration = value !== -1;
  const [customMode, setCustomMode] = useState(false);
  const [customValue, setCustomValue] = useState("");
  const [customError, setCustomError] = useState<string | null>(null);

  const isPreset = EXPIRATION_PRESETS.some((p) => p.value === value);

  // Sync custom mode with external value (adjust during render)
  const syncKey = `${hasExpiration}-${isPreset}-${value}`;
  const [prevSyncKey, setPrevSyncKey] = useState(syncKey);
  if (syncKey !== prevSyncKey) {
    setPrevSyncKey(syncKey);
    if (!hasExpiration) {
      setCustomMode(false);
      setCustomValue("");
      setCustomError(null);
    } else if (!isPreset && value > 0) {
      setCustomMode(true);
      setCustomValue(String(value));
    }
  }

  const handleToggle = () => {
    if (hasExpiration) {
      onChange(-1);
    } else {
      onChange(3600); // default to 1 hour when enabling
    }
  };

  const handleCustomSubmit = () => {
    const n = parseInt(customValue, 10);
    if (isNaN(n) || n < 1) {
      setCustomError("Must be at least 1 second");
      return;
    }
    if (n > MAX_CUSTOM_TTL) {
      setCustomError(`Maximum is ${MAX_CUSTOM_TTL}`);
      return;
    }
    setCustomError(null);
    onChange(n);
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <label className={LABEL + " !mb-0"}>Set expiration</label>
        <button
          type="button"
          onClick={handleToggle}
          className={`relative w-9 h-5 rounded-full transition-colors ${hasExpiration ? "bg-primary" : "bg-border"}`}
        >
          <span
            className={`absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white transition-transform ${hasExpiration ? "translate-x-4" : ""}`}
          />
        </button>
      </div>

      {hasExpiration ? (
        <div className="space-y-2.5">
          <div className="flex flex-wrap gap-1.5">
            {EXPIRATION_PRESETS.map((preset) => (
              <button
                key={preset.value}
                type="button"
                onClick={() => {
                  setCustomMode(false);
                  setCustomError(null);
                  onChange(preset.value);
                }}
                className={`px-2.5 py-1.5 text-xs rounded-md border transition-all ${
                  !customMode && preset.value === value
                    ? "bg-primary/10 border-primary/30 text-primary font-medium"
                    : "bg-card border-border text-text-secondary hover:border-border-light hover:text-text-primary"
                }`}
              >
                {preset.label}
              </button>
            ))}
            <button
              type="button"
              onClick={() => {
                setCustomMode(true);
                setCustomValue("");
                setCustomError(null);
              }}
              className={`px-2.5 py-1.5 text-xs rounded-md border transition-all ${
                customMode
                  ? "bg-primary/10 border-primary/30 text-primary font-medium"
                  : "bg-card border-border text-text-secondary hover:border-border-light hover:text-text-primary"
              }`}
            >
              Custom
            </button>
          </div>

          {customMode && (
            <div>
              <input
                type="number"
                value={customValue}
                onChange={(e) => {
                  setCustomValue(e.target.value);
                  setCustomError(null);
                }}
                onBlur={handleCustomSubmit}
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleCustomSubmit();
                }}
                placeholder="Value in seconds"
                min={1}
                max={MAX_CUSTOM_TTL}
                className={INPUT_MONO}
                autoFocus
              />
              {customError && (
                <p className="mt-1 text-2xs text-accent-red">{customError}</p>
              )}
            </div>
          )}
        </div>
      ) : (
        <p className="text-2xs text-text-muted">
          This endpoint will never expire.
        </p>
      )}
    </div>
  );
}

/* ─── Endpoint Drawer ─── */
function EndpointDrawer({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const { create } = useWebEndpointsStore();

  const [device, setDevice] = useState<Device | null>(null);
  const [hostMode, setHostMode] = useState<"localhost" | "custom">("localhost");
  const [host, setHost] = useState("127.0.0.1");
  const [port, setPort] = useState("");
  const [ttl, setTtl] = useState(-1);
  const [tlsEnabled, setTlsEnabled] = useState(false);
  const [tlsVerify, setTlsVerify] = useState(false);
  const [tlsDomain, setTlsDomain] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    setDevice(null);
    setHostMode("localhost");
    setHost("127.0.0.1");
    setPort("");
    setTtl(-1);
    setTlsEnabled(false);
    setTlsVerify(false);
    setTlsDomain("");
    setError(null);
  }, [open]);

  const hostError =
    host && !isValidHost(host)
      ? "Enter a valid IPv4 or IPv6 address"
      : undefined;
  const portNum = parseInt(port, 10);
  const portError =
    port && (isNaN(portNum) || portNum < 1 || portNum > 65535)
      ? "Port must be 1-65535"
      : undefined;
  const tlsDomainError =
    tlsEnabled && tlsDomain && !isValidFQDN(tlsDomain)
      ? "Enter a valid domain (e.g. example.com)"
      : undefined;

  const confirmDisabled =
    !device ||
    !host.trim() ||
    !!hostError ||
    !port.trim() ||
    !!portError ||
    (tlsEnabled && tlsDomain.trim() !== "" && !!tlsDomainError);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (confirmDisabled) return;
    setError(null);
    setSubmitting(true);
    try {
      await create({
        uid: device!.uid,
        host: host.trim(),
        port: portNum,
        ttl,
        ...(tlsEnabled
          ? {
              tls: {
                enabled: true,
                verify: tlsVerify,
                domain: tlsDomain.trim(),
              },
            }
          : {}),
      });
      onClose();
    } catch (err: unknown) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        setError("A web endpoint with this configuration already exists.");
      } else {
        setError(
          err instanceof Error ? err.message : "Failed to create web endpoint",
        );
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="New Web Endpoint"
      subtitle="Tunnel HTTP traffic to a service on your device."
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            onClick={handleSubmit}
            disabled={submitting || confirmDisabled}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
          >
            {submitting ? (
              <span className="flex items-center gap-2">
                <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                Creating...
              </span>
            ) : (
              "Create Endpoint"
            )}
          </button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        {/* Device */}
        <div>
          <label className={LABEL}>Device</label>
          <DeviceSelector
            selected={device}
            onChange={setDevice}
            error={!device && error ? "Select a device" : undefined}
          />
        </div>

        {/* Target */}
        <div>
          <label className={LABEL}>Target</label>
          <div className="space-y-2">
            {/* Localhost card */}
            <div
              onClick={() => {
                setHostMode("localhost");
                setHost("127.0.0.1");
              }}
              className={`relative p-4 rounded-lg border transition-all cursor-pointer ${
                hostMode === "localhost"
                  ? "bg-primary/[0.06] border-primary/30"
                  : "bg-card/50 border-border hover:border-border-light"
              }`}
            >
              <div className="flex gap-3.5">
                <div
                  className={`shrink-0 w-10 h-10 rounded-lg flex items-center justify-center transition-colors ${
                    hostMode === "localhost"
                      ? "bg-primary/15"
                      : "bg-hover-medium"
                  }`}
                >
                  <ServerStackIcon
                    className={`w-5 h-5 transition-colors ${hostMode === "localhost" ? "text-primary" : "text-text-muted"}`}
                  />
                </div>
                <div className="min-w-0 flex-1">
                  <span
                    className={`text-sm font-semibold transition-colors ${hostMode === "localhost" ? "text-primary" : "text-text-primary"}`}
                  >
                    Localhost
                  </span>
                  <p className="text-2xs text-text-muted leading-relaxed mt-0.5">
                    Service running on the device itself.
                  </p>
                </div>
              </div>
              {hostMode === "localhost" && (
                <div
                  className="grid grid-cols-[1fr,100px] gap-2 mt-3 pt-3 border-t border-primary/10"
                  onClick={(e) => e.stopPropagation()}
                >
                  <input
                    type="text"
                    value="127.0.0.1"
                    readOnly
                    className={`${INPUT_MONO} opacity-60 cursor-default`}
                    tabIndex={-1}
                  />
                  <input
                    type="number"
                    value={port}
                    onChange={(e) => setPort(e.target.value)}
                    placeholder="Port"
                    min={1}
                    max={65535}
                    className={INPUT_MONO}
                  />
                  {portError && (
                    <p className="col-span-2 text-2xs text-accent-red">
                      {portError}
                    </p>
                  )}
                </div>
              )}
            </div>

            {/* Local network card */}
            <div
              onClick={() => {
                if (hostMode !== "custom") {
                  setHostMode("custom");
                  setHost("");
                }
              }}
              className={`relative p-4 rounded-lg border transition-all cursor-pointer ${
                hostMode === "custom"
                  ? "bg-primary/[0.06] border-primary/30"
                  : "bg-card/50 border-border hover:border-border-light"
              }`}
            >
              <div className="flex gap-3.5">
                <div
                  className={`shrink-0 w-10 h-10 rounded-lg flex items-center justify-center transition-colors ${
                    hostMode === "custom" ? "bg-primary/15" : "bg-hover-medium"
                  }`}
                >
                  <ArrowsRightLeftIcon
                    className={`w-5 h-5 transition-colors ${hostMode === "custom" ? "text-primary" : "text-text-muted"}`}
                  />
                </div>
                <div className="min-w-0 flex-1">
                  <span
                    className={`text-sm font-semibold transition-colors ${hostMode === "custom" ? "text-primary" : "text-text-primary"}`}
                  >
                    Local network
                  </span>
                  <p className="text-2xs text-text-muted leading-relaxed mt-0.5">
                    Proxy to another host on the device's local network.
                  </p>
                </div>
              </div>
              {hostMode === "custom" && (
                <div
                  className="grid grid-cols-[1fr,100px] gap-2 mt-3 pt-3 border-t border-primary/10"
                  onClick={(e) => e.stopPropagation()}
                >
                  <input
                    type="text"
                    value={host}
                    onChange={(e) => setHost(e.target.value)}
                    placeholder="e.g. 192.168.1.100"
                    className={INPUT_MONO}
                    autoFocus
                  />
                  <input
                    type="number"
                    value={port}
                    onChange={(e) => setPort(e.target.value)}
                    placeholder="Port"
                    min={1}
                    max={65535}
                    className={INPUT_MONO}
                  />
                  {hostError && (
                    <p className="col-span-2 text-2xs text-accent-red">
                      {hostError}
                    </p>
                  )}
                  {portError && (
                    <p className="col-span-2 text-2xs text-accent-red">
                      {portError}
                    </p>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Expiration */}
        <div className="border border-border rounded-lg p-4">
          <TimeoutSelector value={ttl} onChange={setTtl} />
        </div>

        {/* TLS Section */}
        <div className="border border-border rounded-lg p-4 space-y-3">
          <div className="flex items-center justify-between">
            <label className={LABEL + " !mb-0"}>
              Device service uses HTTPS
            </label>
            <button
              type="button"
              onClick={() => setTlsEnabled(!tlsEnabled)}
              className={`relative w-9 h-5 rounded-full transition-colors ${tlsEnabled ? "bg-primary" : "bg-border"}`}
            >
              <span
                className={`absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white transition-transform ${tlsEnabled ? "translate-x-4" : ""}`}
              />
            </button>
          </div>
          <p className="text-2xs text-text-muted leading-relaxed">
            Enable this if the service running on the device listens over HTTPS.
            The agent will connect using TLS instead of plain HTTP.
          </p>

          {tlsEnabled && (
            <div className="space-y-3 pt-3 border-t border-border/50">
              {/* Verify */}
              <label className="flex items-start gap-2.5 cursor-pointer">
                <input
                  type="checkbox"
                  checked={tlsVerify}
                  onChange={(e) => setTlsVerify(e.target.checked)}
                  className="w-4 h-4 mt-0.5 rounded border-border bg-card text-primary focus:ring-primary/20"
                />
                <div>
                  <span className="text-sm text-text-secondary">
                    Verify device certificate
                  </span>
                  <p className="text-2xs text-text-muted mt-0.5">
                    Reject connections if the device's TLS certificate is
                    invalid or self-signed.
                  </p>
                </div>
              </label>

              {/* Domain */}
              <div>
                <label className={LABEL}>Server Name (SNI)</label>
                <input
                  type="text"
                  value={tlsDomain}
                  onChange={(e) => setTlsDomain(e.target.value)}
                  placeholder="e.g. myservice.local"
                  className={INPUT_MONO}
                />
                <p className="mt-1 text-2xs text-text-muted">
                  Domain sent during the TLS handshake with the device service.
                </p>
                {tlsDomainError && (
                  <p className="mt-1 text-2xs text-accent-red">
                    {tlsDomainError}
                  </p>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Error */}
        {error && (
          <div className="flex items-start gap-2 px-3 py-2.5 bg-accent-red/[0.08] border border-accent-red/20 rounded-lg">
            <ExclamationCircleIcon
              className="w-4 h-4 shrink-0 text-accent-red mt-0.5"
              strokeWidth={2}
            />
            <p className="text-xs text-accent-red">{error}</p>
          </div>
        )}
      </form>
    </Drawer>
  );
}

/* ─── Endpoint Card ─── */
function EndpointCard({
  endpoint,
  onDelete,
}: {
  endpoint: WebEndpoint;
  onDelete: () => void;
}) {
  const expired = isExpired(endpoint.expires_in);
  const never = neverExpires(endpoint.expires_in);
  const fullUrl = `${window.location.protocol}//${endpoint.full_address}`;

  return (
    <div
      className={`group bg-card border rounded-xl p-4 hover:bg-hover-subtle transition-all ${
        expired
          ? "border-accent-yellow/30"
          : "border-border hover:border-border-light"
      }`}
    >
      <div className="flex items-start justify-between gap-3">
        {/* Left: icon + info */}
        <div className="flex items-start gap-3.5 min-w-0 flex-1">
          <div
            className={`shrink-0 w-9 h-9 rounded-lg flex items-center justify-center mt-0.5 ${
              expired ? "bg-accent-yellow/10" : "bg-primary/10"
            }`}
          >
            <GlobeAltIcon
              className={`w-4.5 h-4.5 ${expired ? "text-accent-yellow" : "text-primary"}`}
            />
          </div>
          <div className="min-w-0 flex-1">
            <h3 className="text-sm font-semibold text-text-primary truncate">
              {endpoint.device?.name || endpoint.device_uid}
            </h3>

            {/* URL */}
            <div className="mt-1">
              <a
                href={fullUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-2xs font-mono text-primary hover:text-primary/80 truncate block max-w-[320px] transition-colors"
                title={fullUrl}
              >
                {fullUrl}
              </a>
            </div>

            {/* Badges row */}
            <div className="flex flex-wrap items-center gap-2 mt-2.5">
              {/* Host:Port */}
              <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-hover-medium text-text-muted text-2xs rounded font-mono">
                <ServerStackIcon className="w-2.5 h-2.5" strokeWidth={2} />
                {endpoint.host}:{endpoint.port}
              </span>

              {/* Device-side TLS indicator */}
              {endpoint.tls?.enabled && (
                <span
                  className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-accent-green/10 text-accent-green text-2xs rounded font-medium"
                  title="The device service uses HTTPS"
                >
                  <LockClosedIcon className="w-2.5 h-2.5" strokeWidth={2} />
                  TLS
                </span>
              )}

              {/* Expired badge */}
              {expired && (
                <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-accent-yellow/10 text-accent-yellow text-2xs rounded font-medium">
                  <ExclamationCircleIcon
                    className="w-2.5 h-2.5"
                    strokeWidth={2}
                  />
                  Expired
                </span>
              )}

              {/* Expiration / Created */}
              <span className="text-2xs text-text-muted/60">
                {never
                  ? "Never expires"
                  : formatExpiration(endpoint.expires_in)}
                {" · "}
                {formatDate(endpoint.created_at)}
              </span>
            </div>
          </div>
        </div>

        {/* Right: delete action */}
        <div className="flex items-center shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            onClick={onDelete}
            className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/10 transition-all"
            title="Delete"
          >
            <TrashIcon className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}

/* ─── Page ─── */
function WebEndpointsContent() {
  const { webEndpoints, totalCount, loading, page, perPage, fetch, remove } =
    useWebEndpointsStore();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<{
    address: string;
    deviceName: string;
  } | null>(null);
  const [search, setSearch] = useState("");

  useEffect(() => {
    fetch();
  }, [fetch]);

  const openNew = () => {
    setDrawerOpen(true);
  };

  const closeDrawer = () => {
    setDrawerOpen(false);
  };

  const totalPages = Math.ceil(totalCount / perPage);

  const filtered = search
    ? webEndpoints.filter(
        (ep) =>
          (ep.device?.name || "")
            .toLowerCase()
            .includes(search.toLowerCase()) ||
          ep.full_address.toLowerCase().includes(search.toLowerCase()) ||
          ep.address.toLowerCase().includes(search.toLowerCase()),
      )
    : webEndpoints;

  return (
    <div>
      {/* Content */}
      {loading && webEndpoints.length === 0 ? (
        <div className="flex items-center justify-center py-16">
          <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
        </div>
      ) : webEndpoints.length === 0 ? (
        /* Empty state */
        <div className="relative -mx-8 -mt-8 min-h-[calc(100vh-3.5rem)] flex flex-col">
          {/* Background */}
          <div className="absolute inset-0 overflow-hidden pointer-events-none">
            <div className="absolute -top-32 left-1/3 w-[500px] h-[500px] bg-primary/5 rounded-full blur-[120px] animate-pulse-subtle" />
            <div
              className="absolute bottom-0 right-1/4 w-[400px] h-[400px] bg-accent-cyan/5 rounded-full blur-[100px] animate-pulse-subtle"
              style={{ animationDelay: "1s" }}
            />
            <div className="absolute inset-0 grid-bg opacity-30" />
          </div>

          <div className="relative z-10 flex-1 flex items-center justify-center px-8 py-12">
            <div className="w-full max-w-2xl animate-fade-in">
              {/* Header */}
              <div className="text-center mb-10">
                <div className="w-16 h-16 rounded-2xl bg-primary/10 border border-primary/20 flex items-center justify-center mx-auto mb-6 shadow-lg shadow-primary/5">
                  <GlobeAltIcon className="w-8 h-8 text-primary" />
                </div>

                <span className="inline-block text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
                  HTTP Tunneling
                </span>
                <h1 className="text-3xl font-bold text-text-primary mb-3">
                  Web Endpoints
                </h1>
                <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
                  Create unique URLs that tunnel HTTP traffic to services
                  running on your devices.
                </p>
              </div>

              {/* Highlights */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-10">
                {[
                  {
                    icon: <LinkIcon className="w-5 h-5" />,
                    title: "Direct Access",
                    description:
                      "Each endpoint gets a unique URL that routes directly to a host and port on your device.",
                  },
                  {
                    icon: <LockClosedIcon className="w-5 h-5" />,
                    title: "Device-side TLS",
                    description:
                      "Connect to HTTPS services on your devices — the agent handles the TLS handshake locally.",
                  },
                  {
                    icon: <ClockIcon className="w-5 h-5" />,
                    title: "Auto-Expiring",
                    description:
                      "Endpoints expire automatically after a configurable TTL, or run indefinitely.",
                  },
                ].map((h, idx) => (
                  <div
                    key={h.title}
                    className="bg-card/60 border border-border rounded-xl p-5 text-center animate-slide-up"
                    style={{ animationDelay: `${150 + idx * 100}ms` }}
                  >
                    <div className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center mx-auto mb-3 text-primary">
                      {h.icon}
                    </div>
                    <h3 className="text-sm font-semibold text-text-primary mb-1">
                      {h.title}
                    </h3>
                    <p className="text-xs text-text-muted leading-relaxed">
                      {h.description}
                    </p>
                  </div>
                ))}
              </div>

              {/* CTA */}
              <div
                className="text-center animate-slide-up"
                style={{ animationDelay: "450ms" }}
              >
                <button
                  onClick={openNew}
                  className="inline-flex items-center gap-2 px-6 py-3 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200 shadow-lg shadow-primary/20"
                >
                  <PlusIcon className="w-4 h-4" strokeWidth={2} />
                  Create your first endpoint
                </button>
                <p className="mt-4 text-2xs text-text-muted">
                  No VPN, no SSH port forwarding — just a URL.
                </p>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <>
          <PageHeader
            variant="decorated"
            icon={<GlobeAltIcon className="w-6 h-6" />}
            overline="Networking"
            title="Web Endpoints"
            description="Unique URLs that tunnel HTTP traffic to services on your devices."
          >
            <button
              onClick={openNew}
              className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200"
            >
              <PlusIcon className="w-4 h-4" strokeWidth={2} />
              New Endpoint
            </button>
          </PageHeader>

          {/* Search bar */}
          <div className="mb-4 animate-fade-in">
            <div className="relative max-w-sm">
              <MagnifyingGlassIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted pointer-events-none" />
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search by device or address..."
                className="w-full pl-9 pr-3.5 py-2 bg-card border border-border rounded-lg text-sm text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
              />
            </div>
          </div>

          {filtered.length === 0 ? (
            <div className="py-12 text-center animate-fade-in">
              <p className="text-sm text-text-muted">
                No endpoints matching &ldquo;{search}&rdquo;
              </p>
            </div>
          ) : (
            <>
              {/* Endpoint cards */}
              <div className="space-y-2 animate-fade-in">
                {filtered.map((ep) => (
                  <EndpointCard
                    key={ep.address}
                    endpoint={ep}
                    onDelete={() =>
                      setDeleteTarget({
                        address: ep.address,
                        deviceName: ep.device?.name || ep.device_uid,
                      })
                    }
                  />
                ))}
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-between mt-4 px-1">
                  <span className="text-xs font-mono text-text-muted">
                    {totalCount} endpoint{totalCount !== 1 ? "s" : ""}
                  </span>
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => fetch(page - 1)}
                      disabled={page <= 1}
                      className="px-2.5 py-1 text-xs font-medium text-text-secondary hover:text-text-primary disabled:opacity-soft disabled:cursor-not-allowed transition-colors"
                    >
                      Prev
                    </button>
                    <span className="text-xs font-mono text-text-muted px-2">
                      {page} / {totalPages}
                    </span>
                    <button
                      onClick={() => fetch(page + 1)}
                      disabled={page >= totalPages}
                      className="px-2.5 py-1 text-xs font-medium text-text-secondary hover:text-text-primary disabled:opacity-soft disabled:cursor-not-allowed transition-colors"
                    >
                      Next
                    </button>
                  </div>
                </div>
              )}
            </>
          )}
        </>
      )}

      {/* Drawer */}
      <EndpointDrawer open={drawerOpen} onClose={closeDrawer} />

      {/* Delete Dialog */}
      <ConfirmDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={async () => {
          await remove(deleteTarget!.address);
          setDeleteTarget(null);
        }}
        title="Delete Web Endpoint"
        description={
          <>
            Are you sure you want to delete the endpoint for{" "}
            <span className="font-medium text-text-primary">
              {deleteTarget?.deviceName}
            </span>
            ? This action cannot be undone.
          </>
        }
        confirmLabel="Delete"
      />
    </div>
  );
}

export default function WebEndpoints() {
  return <WebEndpointsContent />;
}
