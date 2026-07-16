import { type KeyboardEvent, useRef } from "react";
import InputField from "@/components/common/fields/InputField";
import KeyFileInput from "@/components/common/fields/KeyFileInput";
import { LABEL } from "@/utils/styles";
import { MODE_INFO } from "./constants";

export type InstallKeyMode = "automatic" | "manual" | "webhook" | "allowlist";

// The selector's options are derived from the shared MODE_INFO so mode icon/label/description live in
// one place (also used by the list's Enrollment cell).
const OPTIONS = (
  ["automatic", "manual", "webhook", "allowlist"] as InstallKeyMode[]
).map((value) => ({ value, ...MODE_INFO[value] }));

/**
 * The enrollment policy for a key: a list of selectable cards (title + description) across the four
 * modes, revealing the mode-specific config (webhook endpoint + secret, or the MAC allowlist) beneath.
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
}: {
  // Unique per drawer: Create and Edit are both mounted at once, so shared field ids would cross-wire
  // their labels (a click in one drawer targeting the other's inert input).
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
}) {
  const clamp = (n: number, min: number, max: number) =>
    Math.min(max, Math.max(min, n));

  // Roving-tabindex radiogroup: Tab lands on the selected option, then arrows move the selection (and
  // focus) between the four, wrapping around — the keyboard behavior a radiogroup is expected to have.
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
                <div className="space-y-3 border-t border-primary/20 bg-card/40 px-3.5 py-3">
                  <InputField
                    id={`${idPrefix}-webhook-url`}
                    label="Webhook URL"
                    value={webhookUrl}
                    onChange={onWebhookUrlChange}
                    placeholder="https://register.example.com/hook"
                    hint="Called with a signed payload at registration. http or https."
                  />
                  <InputField
                    id={`${idPrefix}-webhook-secret`}
                    label="Signing secret"
                    type="password"
                    value={webhookSecret}
                    onChange={onWebhookSecretChange}
                    hint="Signs the request as the X-ShellHub-Signature header (HMAC-SHA256), so your endpoint can verify it came from ShellHub."
                  />
                  <div className="grid grid-cols-2 gap-3">
                    <InputField
                      id={`${idPrefix}-webhook-timeout`}
                      label="Reply timeout (s)"
                      type="number"
                      value={String(webhookTimeout || 5)}
                      onChange={(v) =>
                        onWebhookTimeoutChange(
                          clamp(parseInt(v, 10) || 5, 1, 15),
                        )
                      }
                      hint="How long ShellHub waits for your endpoint to answer, in seconds (1–15)."
                    />
                    <InputField
                      id={`${idPrefix}-webhook-window`}
                      label="Callback window (h)"
                      type="number"
                      value={String(
                        Math.round((webhookCallbackTtl || 3600) / 3600),
                      )}
                      onChange={(v) =>
                        onWebhookCallbackTtlChange(
                          clamp(parseInt(v, 10) || 1, 1, 24) * 3600,
                        )
                      }
                      hint="If your endpoint replies later instead of right away, how long it has to call back (up to 24h)."
                    />
                  </div>
                </div>
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
