import { useState, useRef, type ComponentType, type SVGProps } from "react";
import {
  InformationCircleIcon,
  ExclamationTriangleIcon,
  ExclamationCircleIcon,
  CheckCircleIcon,
  XCircleIcon,
  GlobeAltIcon,
  FlagIcon,
  KeyIcon,
  ArrowUpTrayIcon,
  DocumentIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import PageHeader from "../../components/common/PageHeader";
import CopyButton from "../../components/common/CopyButton";
import { useAdminLicense } from "../../hooks/useAdminLicense";
import { useUploadLicense } from "../../hooks/useUploadLicense";
import {
  formatLicenseTimestamp,
  formatDeviceCount,
  formatRegions,
  getDisplayFeatures,
  validateLicenseFile,
  getLicenseAlertConfig,
} from "../../utils/license";
import type { GetLicenseResponse } from "../../client/types.gen";

type HeroIcon = ComponentType<SVGProps<SVGSVGElement>>;

/* ─── Status Alert ─── */

const alertStyles: Record<"info" | "warning" | "error", {
  container: string;
  Icon: HeroIcon;
  text: string;
}> = {
  info: { container: "bg-accent-blue/[0.06] border-accent-blue/10", Icon: InformationCircleIcon, text: "text-accent-blue" },
  warning: { container: "bg-accent-yellow/[0.06] border-accent-yellow/10", Icon: ExclamationTriangleIcon, text: "text-accent-yellow" },
  error: { container: "bg-accent-red/[0.06] border-accent-red/10", Icon: ExclamationCircleIcon, text: "text-accent-red" },
};

function LicenseStatusAlert({ license }: { license: GetLicenseResponse | null }) {
  const config = getLicenseAlertConfig(
    license
      ? { expired: license.expired, about_to_expire: license.about_to_expire, grace_period: license.grace_period }
      : null,
  );

  if (!config) return null;

  const s = alertStyles[config.variant];
  const isUrgent = config.variant === "error" || config.variant === "warning";

  return (
    <div
      role={isUrgent ? "alert" : "status"}
      aria-live={isUrgent ? "assertive" : "polite"}
      className={`flex items-center gap-3 px-4 py-3 rounded-lg border ${s.container}`}
    >
      <s.Icon className={`w-5 h-5 ${s.text} shrink-0`} />
      <p className={`text-sm font-medium ${s.text}`}>{config.message}</p>
    </div>
  );
}

/* ─── Detail Row ─── */

function DetailRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between py-2.5">
      <span className="text-sm text-text-muted">{label}</span>
      <span className="text-sm font-medium text-text-primary">{children}</span>
    </div>
  );
}

/* ─── License Details ─── */

function LicenseDetails({ license }: { license: GetLicenseResponse }) {
  const regionText = formatRegions(license.allowed_regions);
  const isGlobal = license.allowed_regions.length === 0;

  return (
    <section>
      <h2 className="text-xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3">
        License Information
      </h2>
      <div className="divide-y divide-border">
        <DetailRow label="Issued at">
          {formatLicenseTimestamp(license.issued_at)}
        </DetailRow>
        <DetailRow label="Starts at">
          {formatLicenseTimestamp(license.starts_at)}
        </DetailRow>
        <DetailRow label="Expires at">
          {formatLicenseTimestamp(license.expires_at)}
        </DetailRow>
        <DetailRow label="Allowed regions">
          <span className="inline-flex items-center gap-1.5">
            {isGlobal
              ? <GlobeAltIcon className="w-4 h-4 text-text-muted" />
              : <FlagIcon className="w-4 h-4 text-text-muted" />}
            {regionText}
          </span>
        </DetailRow>
      </div>
    </section>
  );
}

/* ─── License Owner ─── */

function LicenseOwner({ customer }: { customer: GetLicenseResponse["customer"] }) {
  return (
    <section>
      <h2 className="text-xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3">
        License Owner
      </h2>
      <div className="divide-y divide-border">
        <DetailRow label="ID">
          <span className="inline-flex items-center gap-1.5">
            <code className="text-xs font-mono text-accent-cyan">
              {customer.id ?? "\u2014"}
            </code>
            {customer.id && <CopyButton text={customer.id} />}
          </span>
        </DetailRow>
        <DetailRow label="Name">
          {customer.name ?? "\u2014"}
        </DetailRow>
        <DetailRow label="Email">
          {customer.email ?? "\u2014"}
        </DetailRow>
        <DetailRow label="Company">
          {customer.company ?? "\u2014"}
        </DetailRow>
      </div>
    </section>
  );
}

/* ─── License Features ─── */

function LicenseFeatures({ features }: { features: GetLicenseResponse["features"] }) {
  const display = getDisplayFeatures(features);

  return (
    <section>
      <h2 className="text-xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3">
        Features
      </h2>
      <div className="divide-y divide-border">
        {display.map((feature) => (
          <DetailRow key={feature.name} label={feature.label}>
            {feature.type === "number"
              ? (
                <span className="inline-flex items-center px-2 py-0.5 text-xs font-mono font-semibold bg-primary/10 text-primary border border-primary/20 rounded">
                  {formatDeviceCount(feature.value)}
                </span>
              )
              : feature.value
                ? (
                  <CheckCircleIcon
                    className="w-5 h-5 text-accent-green"
                    aria-label="Included"
                  />
                )
                : (
                  <XCircleIcon
                    className="w-5 h-5 text-accent-red"
                    aria-label="Not included"
                  />
                )}
          </DetailRow>
        ))}
      </div>
    </section>
  );
}

/* ─── License Upload ─── */

function LicenseUpload() {
  const upload = useUploadLicense();
  const [file, setFile] = useState<File | null>(null);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [feedback, setFeedback] = useState<{ type: "success" | "error"; message: string } | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const dragCounter = useRef(0);

  const selectFile = (selected: File | null) => {
    setFile(selected);
    setFeedback(null);
    setValidationError(selected ? validateLicenseFile(selected) : null);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selected = e.target.files?.[0];
    if (selected) selectFile(selected);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    dragCounter.current = 0;
    setIsDragging(false);
    const dropped = e.dataTransfer.files?.[0];
    if (dropped) selectFile(dropped);
  };

  const clearFile = () => {
    // Do NOT call selectFile here — that would reset feedback.
    // Only clear the file and its validation state.
    setFile(null);
    setValidationError(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const handleUpload = async () => {
    if (!file || validationError) return;
    setFeedback(null);
    try {
      await upload.mutateAsync({ body: { file } });
      setFeedback({ type: "success", message: "License uploaded successfully." });
      clearFile();
    } catch {
      setFeedback({ type: "error", message: "Failed to upload the license." });
    }
  };

  const canUpload = file && !validationError && !upload.isPending;

  return (
    <section aria-labelledby="upload-license-heading">
      <h2 id="upload-license-heading" className="text-xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3">
        Upload License
      </h2>
      <div className="space-y-3">
        <div>
          {/*
           * Drop zone — a <div role="button"> that calls fileInputRef.current.click()
           * programmatically. Avoids Chrome's label double-click behavior where
           * clicking a <label htmlFor="file-input"> re-fires after the OS dialog
           * closes, flashing the page and losing the selected file.
           * The hidden <input> has tabIndex={-1} so keyboard focus stays on the div.
           */}
          <div className="relative">
            {/* Hidden file input — not focusable, triggered only programmatically */}
            <input
              id="license-file"
              ref={fileInputRef}
              type="file"
              accept=".dat"
              tabIndex={-1}
              onChange={handleFileChange}
              aria-describedby={validationError ? "license-file-error" : undefined}
              aria-invalid={!!validationError || undefined}
              className="sr-only"
            />

            <div
              role="button"
              tabIndex={0}
              aria-label="Choose a .dat file or drag and drop"
              onClick={() => fileInputRef.current?.click()}
              onKeyDown={(e) => {
                if (e.key === "Enter" || e.key === " ") {
                  e.preventDefault();
                  fileInputRef.current?.click();
                }
              }}
              onDrop={handleDrop}
              onDragEnter={(e) => { e.preventDefault(); dragCounter.current++; setIsDragging(true); }}
              onDragOver={(e) => e.preventDefault()}
              onDragLeave={() => { if (--dragCounter.current === 0) setIsDragging(false); }}
              className={[
                "flex items-center gap-3 w-full px-3.5 py-2.5 rounded-lg border border-dashed cursor-pointer select-none transition-all duration-150",
                "focus-visible:ring-2 focus-visible:ring-primary/50 focus-visible:ring-offset-1 focus-visible:ring-offset-card focus-visible:outline-none",
                file ? "pr-9" : "",
                isDragging
                  ? "border-primary/50 bg-primary/[0.07]"
                  : file
                    ? "border-border-light bg-hover-subtle hover:bg-hover-medium"
                    : "border-border hover:border-border-light hover:bg-hover-subtle",
              ].join(" ")}
            >
              {file ? (
                <>
                  <DocumentIcon className="w-4 h-4 text-primary shrink-0" aria-hidden="true" />
                  <span className="text-sm text-text-primary font-medium truncate flex-1 min-w-0">
                    {file.name}
                  </span>
                  <span className="text-2xs font-mono text-text-muted shrink-0">
                    {(file.size / 1024).toFixed(1)} KB
                  </span>
                </>
              ) : (
                <>
                  <ArrowUpTrayIcon className="w-4 h-4 text-text-muted shrink-0" aria-hidden="true" />
                  <span className="text-sm">
                    <span className="text-text-secondary">Choose a .dat file</span>
                    <span className="text-text-muted hidden sm:inline"> or drag and drop</span>
                  </span>
                  <span className="ml-auto text-2xs font-mono text-text-muted/60 shrink-0">
                    under 32 KB
                  </span>
                </>
              )}
            </div>

            {/* Sibling of the drop zone — clicking it does NOT re-open the file picker */}
            {file && (
              <button
                type="button"
                onClick={clearFile}
                aria-label="Remove selected file"
                className="absolute right-2.5 top-1/2 -translate-y-1/2 p-0.5 rounded text-text-muted hover:text-text-primary hover:bg-hover-strong transition-colors"
              >
                <XMarkIcon className="w-3.5 h-3.5" aria-hidden="true" />
              </button>
            )}
          </div>

          {/* Persistent live region — always in DOM so SR catches text changes */}
          <div aria-live="polite" aria-atomic="true" className="min-h-[1rem] mt-1.5">
            {validationError && (
              <p id="license-file-error" className="text-2xs text-accent-red">
                {validationError}
              </p>
            )}
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => void handleUpload()}
            disabled={!canUpload}
            aria-busy={upload.isPending}
            aria-label={upload.isPending ? "Uploading license file" : "Upload license file"}
            className="inline-flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
          >
            {upload.isPending
              ? <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" aria-hidden="true" />
              : <ArrowUpTrayIcon className="w-4 h-4" aria-hidden="true" />}
            {upload.isPending ? "Uploading..." : "Upload"}
          </button>
          {feedback && (
            <p
              role={feedback.type === "error" ? "alert" : "status"}
              className={`text-sm font-medium ${
                feedback.type === "success" ? "text-accent-green" : "text-accent-red"
              }`}
            >
              {feedback.message}
            </p>
          )}
        </div>
      </div>
    </section>
  );
}

/* ─── Page ─── */

export default function AdminLicense() {
  const { data, isLoading, isError } = useAdminLicense();

  if (isLoading) {
    return (
      <div
        className="flex-1 flex items-center justify-center"
        role="status"
        aria-label="Loading license information"
      >
        <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center" role="alert">
          <ExclamationCircleIcon className="w-10 h-10 text-accent-red mx-auto mb-3" />
          <p className="text-sm font-medium text-text-primary">Failed to load license information</p>
          <p className="text-2xs text-text-muted mt-1">Please try again later.</p>
        </div>
      </div>
    );
  }

  const installedLicense = data && "grace_period" in data ? data : null;

  return (
    <div>
      <PageHeader
        icon={<KeyIcon className="w-6 h-6" />}
        overline="Admin Settings"
        title="License Details"
        description="Review the current license scope and upload a new file when your subscription changes."
      />

      <div className="space-y-6 animate-fade-in">
        <LicenseStatusAlert license={installedLicense} />

        {installedLicense && (
          <div className="bg-card border border-border rounded-xl p-6 space-y-6">
            <LicenseDetails license={installedLicense} />
            <hr className="border-border" />
            <LicenseOwner customer={installedLicense.customer} />
            <hr className="border-border" />
            <LicenseFeatures features={installedLicense.features} />
          </div>
        )}

        <div className="bg-card border border-border rounded-xl p-6">
          <LicenseUpload />
        </div>
      </div>
    </div>
  );
}
