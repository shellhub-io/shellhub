import { type KeyboardEvent, useRef, useState } from "react";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import NumericInput from "@/components/common/fields/NumericInput";
import KeyFileInput from "@/components/common/fields/KeyFileInput";
import { LABEL } from "@/utils/styles";
import { MODE_INFO } from "./constants";
import {
  TIMEOUT_MIN,
  TIMEOUT_MAX,
  WINDOW_MIN_H,
  WINDOW_MAX_H,
  isWebhookUrl,
} from "./helpers";

/**
 * How an install key enrols a device: automatically, waiting for approval, by calling out to a
 * webhook, or only for devices on an allowlist.
 */
export type InstallKeyMode = "automatic" | "manual" | "webhook" | "allowlist";

const OPTIONS = (
  ["automatic", "manual", "webhook", "allowlist"] as InstallKeyMode[]
).map((value) => ({ value, ...MODE_INFO[value] }));

function WebhookPanel({
  idPrefix,
  webhookUrl,
  onWebhookUrlChange,
  webhookSecret,
  onWebhookSecretChange,
  webhookTimeout,
  onWebhookTimeoutChange,
  webhookCallbackTtl,
  onWebhookCallbackTtlChange,
  isEditing,
}: {
  idPrefix: string;
  webhookUrl: string;
  onWebhookUrlChange: (value: string) => void;
  webhookSecret: string;
  onWebhookSecretChange: (value: string) => void;
  webhookTimeout: number;
  onWebhookTimeoutChange: (value: number) => void;
  webhookCallbackTtl: number;
  onWebhookCallbackTtlChange: (value: number) => void;
  isEditing?: boolean;
}) {
  const [urlError, setUrlError] = useState("");
  const [timeoutStr, setTimeoutStr] = useState(String(webhookTimeout || 5));
  const [windowStr, setWindowStr] = useState(
    String(Math.round((webhookCallbackTtl || 3600) / 3600)),
  );

  const handleUrlChange = (value: string) => {
    onWebhookUrlChange(value);
    if (urlError && isWebhookUrl(value)) setUrlError("");
  };

  const handleUrlBlur = () => {
    if (webhookUrl.trim() && !isWebhookUrl(webhookUrl)) {
      setUrlError("Webhook URL must be a valid http or https URL.");
    }
  };

  const clamp = (n: number, min: number, max: number) =>
    Math.min(max, Math.max(min, n));

  const rangeError = (raw: string, min: number, max: number) => {
    const n = parseInt(raw, 10);
    if (!raw || Number.isNaN(n) || n < min || n > max)
      return `Must be ${min}–${max}.`;
    return undefined;
  };

  const timeoutError = rangeError(timeoutStr, TIMEOUT_MIN, TIMEOUT_MAX);
  const windowError = rangeError(windowStr, WINDOW_MIN_H, WINDOW_MAX_H);

  const handleTimeoutChange = (raw: string) => {
    setTimeoutStr(raw);
    const n = parseInt(raw, 10);
    onWebhookTimeoutChange(!raw || Number.isNaN(n) ? 0 : n);
  };

  const handleTimeoutBlur = () => {
    const n = clamp(
      parseInt(timeoutStr, 10) || TIMEOUT_MIN,
      TIMEOUT_MIN,
      TIMEOUT_MAX,
    );
    setTimeoutStr(String(n));
    onWebhookTimeoutChange(n);
  };

  const handleWindowChange = (raw: string) => {
    setWindowStr(raw);
    const n = parseInt(raw, 10);
    onWebhookCallbackTtlChange(!raw || Number.isNaN(n) ? 0 : n * 3600);
  };

  const handleWindowBlur = () => {
    const n = clamp(
      parseInt(windowStr, 10) || WINDOW_MIN_H,
      WINDOW_MIN_H,
      WINDOW_MAX_H,
    );
    setWindowStr(String(n));
    onWebhookCallbackTtlChange(n * 3600);
  };

  const secretHint = isEditing
    ? "Leave blank to keep the current secret. Signs the request as the X-ShellHub-Signature header (HMAC-SHA256)."
    : "Signs the request as the X-ShellHub-Signature header (HMAC-SHA256), so your endpoint can verify it came from ShellHub.";

  return (
    <div className="space-y-3 border-t border-primary/20 bg-card/40 px-3.5 py-3">
      <InputField
        id={`${idPrefix}-webhook-url`}
        label="Webhook URL"
        value={webhookUrl}
        onChange={handleUrlChange}
        onBlur={handleUrlBlur}
        placeholder="https://register.example.com/hook"
        hint="Called with a signed payload at registration. http or https."
        error={urlError || undefined}
        autoComplete="webhook"
      />
      <PasswordField
        id={`${idPrefix}-webhook-secret`}
        label="Signing secret"
        value={webhookSecret}
        suppressPasswordManager
        onChange={onWebhookSecretChange}
        hint={secretHint}
      />
      <div className="flex flex-wrap gap-3">
        <div className="flex-1 min-w-[10rem]">
          <NumericInput
            id={`${idPrefix}-webhook-timeout`}
            label="Reply timeout (s)"
            value={timeoutStr}
            onChange={handleTimeoutChange}
            onBlur={handleTimeoutBlur}
            hint="How long ShellHub waits for your endpoint to answer, in seconds (1–15)."
            error={timeoutError}
          />
        </div>
        <div className="flex-1 min-w-[10rem]">
          <NumericInput
            id={`${idPrefix}-webhook-window`}
            label="Callback window (h)"
            value={windowStr}
            onChange={handleWindowChange}
            onBlur={handleWindowBlur}
            hint="If your endpoint replies later instead of right away, how long it has to call back (up to 24h)."
            error={windowError}
          />
        </div>
      </div>
    </div>
  );
}

/**
 * Picks an install key's enrolment mode. The options are built from the shared MODE_INFO, so the
 * icon, label and description here are the same ones the list's Enrollment cell shows.
 */
export default function ModeField({
  idPrefix,
  mode,
  onModeChange,
  webhookUrl,
  onWebhookUrlChange,
  webhookSecret,
  onWebhookSecretChange,
  allowedMacs,
  onAllowedMacsChange,
  webhookTimeout,
  onWebhookTimeoutChange,
  webhookCallbackTtl,
  onWebhookCallbackTtlChange,
  isEditing,
  panelKey,
}: {
  idPrefix: string;
  mode: InstallKeyMode;
  onModeChange: (mode: InstallKeyMode) => void;
  webhookUrl: string;
  onWebhookUrlChange: (value: string) => void;
  webhookSecret: string;
  onWebhookSecretChange: (value: string) => void;
  allowedMacs: string;
  onAllowedMacsChange: (value: string) => void;
  webhookTimeout: number;
  onWebhookTimeoutChange: (value: number) => void;
  webhookCallbackTtl: number;
  onWebhookCallbackTtlChange: (value: number) => void;
  isEditing?: boolean;
  panelKey?: string;
}) {
  const radioRefs = useRef<Record<string, HTMLButtonElement | null>>({});
  const onRadioKeyDown = (event: KeyboardEvent, index: number) => {
    const forward = event.key === "ArrowDown" || event.key === "ArrowRight";
    const backward = event.key === "ArrowUp" || event.key === "ArrowLeft";
    if (!forward && !backward) return;

    event.preventDefault();
    const next =
      OPTIONS[(index + (forward ? 1 : -1) + OPTIONS.length) % OPTIONS.length]
        .value;
    onModeChange(next);
    radioRefs.current[next]?.focus();
  };

  return (
    <div className="space-y-3">
      <span className={LABEL}>Registration mode</span>

      <div
        role="radiogroup"
        aria-label="Registration mode"
        className="space-y-2"
      >
        {OPTIONS.map((option, index) => {
          const Icon = option.icon;
          const selected = option.value === mode;
          return (
            <div
              key={option.value}
              className={`overflow-hidden rounded-xl border transition-colors ${
                selected
                  ? "border-primary/50 bg-primary/[0.04]"
                  : "border-border hover:border-border-strong"
              }`}
            >
              <button
                type="button"
                role="radio"
                aria-checked={selected}
                tabIndex={selected ? 0 : -1}
                ref={(el) => {
                  radioRefs.current[option.value] = el;
                }}
                onClick={() => onModeChange(option.value)}
                onKeyDown={(e) => onRadioKeyDown(e, index)}
                className="flex w-full items-start gap-3 px-3.5 py-3 text-left"
              >
                <Icon
                  className={`mt-0.5 h-5 w-5 shrink-0 ${
                    selected ? "text-primary" : "text-text-secondary"
                  }`}
                  strokeWidth={1.8}
                />
                <div className="min-w-0 flex-1">
                  <span
                    className={`block text-xs font-medium ${
                      selected ? "text-text-primary" : "text-text-secondary"
                    }`}
                  >
                    {option.label}
                  </span>
                  <span className="mt-0.5 block text-2xs text-text-muted">
                    {option.description}
                  </span>
                </div>
                <span
                  className={`mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full border ${
                    selected ? "border-primary" : "border-border-strong"
                  }`}
                >
                  {selected && (
                    <span className="h-2 w-2 rounded-full bg-primary" />
                  )}
                </span>
              </button>

              {selected && option.value === "webhook" && (
                <WebhookPanel
                  key={panelKey}
                  idPrefix={idPrefix}
                  webhookUrl={webhookUrl}
                  onWebhookUrlChange={onWebhookUrlChange}
                  webhookSecret={webhookSecret}
                  onWebhookSecretChange={onWebhookSecretChange}
                  webhookTimeout={webhookTimeout}
                  onWebhookTimeoutChange={onWebhookTimeoutChange}
                  webhookCallbackTtl={webhookCallbackTtl}
                  onWebhookCallbackTtlChange={onWebhookCallbackTtlChange}
                  isEditing={isEditing}
                />
              )}

              {selected && option.value === "allowlist" && (
                <div className="border-t border-primary/20 bg-card/40 px-3.5 py-3">
                  <KeyFileInput
                    id={`${idPrefix}-allowed-macs`}
                    label="Allowed MACs"
                    value={allowedMacs}
                    onChange={onAllowedMacsChange}
                    validate={(text) => text.trim().length > 0}
                    accept=".txt,.csv,text/plain"
                    maxBytes={2 * 1024 * 1024}
                    rows={4}
                    placeholder={"aa:bb:cc:dd:ee:ff\n11:22:33:44:55:66"}
                    emptyLabel="Drop a MAC list, paste, or browse"
                    loadedLabel="MAC list loaded"
                    hint="One MAC per line. A convenience filter, not a security boundary: MACs can be spoofed. Max 2 MB."
                  />
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
