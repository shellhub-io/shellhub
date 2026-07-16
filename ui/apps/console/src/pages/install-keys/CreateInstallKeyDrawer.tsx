import { useState, FormEvent } from "react";
import { CheckIcon, TicketIcon } from "@heroicons/react/24/outline";
import { Button, Card, WindowChrome } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useCreateInstallKey } from "@/hooks/useInstallKeyMutations";
import { useAuthStore } from "@/stores/authStore";
import { buildInstallCommand } from "@/utils/installCommand";
import CopyButton from "@/components/common/CopyButton";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import TagsSelector from "@/components/common/fields/TagsSelector";
import EphemeralField from "./EphemeralField";
import ExpirationField from "./ExpirationField";
import ModeField, { type InstallKeyMode } from "./ModeField";
import UsageLimitField from "./UsageLimitField";
import {
  defaultExpiry,
  parseAllowedMacs,
  validateModeConfig,
  validateName,
} from "./helpers";
import { LABEL } from "@/utils/styles";

/* --- Create Install Key Drawer --- */

function CreateInstallKeyDrawer({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const createKey = useCreateInstallKey();
  const tenant = useAuthStore((s) => s.tenant);
  const [name, setName] = useState("");
  const [mode, setMode] = useState<InstallKeyMode>("automatic");
  const [webhookUrl, setWebhookUrl] = useState("");
  const [webhookSecret, setWebhookSecret] = useState("");
  const [allowedMacs, setAllowedMacs] = useState("");
  const [webhookTimeout, setWebhookTimeout] = useState(5);
  const [webhookCallbackTtl, setWebhookCallbackTtl] = useState(3600);
  const [usageLimit, setUsageLimit] = useState(1);
  const [ephemeral, setEphemeral] = useState(false);
  const [ephemeralTimeout, setEphemeralTimeout] = useState(10);
  const [tags, setTags] = useState<string[]>([]);
  const [expiresAt, setExpiresAt] = useState<string | null>(defaultExpiry());
  const [submitting, setSubmitting] = useState(false);
  const [nameError, setNameError] = useState("");
  const [error, setError] = useState("");
  const [generatedKey, setGeneratedKey] = useState("");

  const macList = parseAllowedMacs(allowedMacs);
  const modeError = validateModeConfig(
    mode,
    webhookUrl,
    webhookSecret,
    macList,
  );

  useResetOnOpen(open, () => {
    setName("");
    setMode("automatic");
    setWebhookUrl("");
    setWebhookSecret("");
    setAllowedMacs("");
    setWebhookTimeout(5);
    setWebhookCallbackTtl(3600);
    setUsageLimit(1);
    setEphemeral(false);
    setEphemeralTimeout(10);
    setTags([]);
    setExpiresAt(defaultExpiry());
    setNameError("");
    setError("");
    setGeneratedKey("");
  });

  const handleNameChange = (value: string) => {
    setName(value);
    if (nameError) setNameError(validateName(value.trim()));
  };

  const handleNameBlur = () => {
    if (name) setNameError(validateName(name.trim()));
  };

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    const validationError = validateName(name.trim());
    if (validationError) {
      setNameError(validationError);
      return;
    }
    if (modeError) {
      setError(modeError);
      return;
    }
    setSubmitting(true);
    setNameError("");
    setError("");
    try {
      const result = await createKey.mutateAsync({
        body: {
          name: name.trim(),
          mode,
          ...(mode === "webhook"
            ? {
                webhook_url: webhookUrl.trim(),
                webhook_secret: webhookSecret,
                webhook_timeout: webhookTimeout,
                webhook_callback_ttl: webhookCallbackTtl,
              }
            : {}),
          ...(mode === "allowlist" ? { allowed_macs: macList } : {}),
          expires_at: expiresAt,
          // Reusability is derived server-side from this: 1 is single-use,
          // N (>=2) enrolls N devices, 0 is unlimited (reusable forever).
          usage_limit: usageLimit,
          ephemeral,
          // Only meaningful for ephemeral keys; already clamped to 1-10 by the field.
          ...(ephemeral ? { ephemeral_timeout: ephemeralTimeout } : {}),
          tags,
        },
      });
      setGeneratedKey(result.key);
    } catch (err) {
      if (isSdkError(err) && err.status === 409) {
        setNameError("A key with that name already exists.");
      } else if (isSdkError(err) && err.status === 400) {
        setNameError(
          "Name must be 3–20 characters: letters, numbers, - and _ only.",
        );
      } else {
        setError(
          "Couldn't create the Install Key. Please review the values and try again.",
        );
      }
    } finally {
      setSubmitting(false);
    }
  };

  const installCommand = buildInstallCommand(
    `TENANT_ID=${tenant} INSTALL_KEY=${generatedKey}`,
    window.location.origin,
  );

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Create Install Key"
      footer={
        generatedKey ? (
          <Button variant="primary" onClick={onClose}>
            Done
          </Button>
        ) : (
          <>
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={() => void handleSubmit()}
              disabled={
                submitting || !!nameError || !!modeError || !name.trim()
              }
              loading={submitting}
              icon={<TicketIcon className="w-4 h-4" strokeWidth={2} />}
            >
              Create Install Key
            </Button>
          </>
        )
      }
    >
      {generatedKey ? (
        <div className="space-y-5">
          <div className="flex items-start gap-3 bg-accent-green/[0.06] border border-accent-green/20 rounded-xl px-4 py-3.5">
            <CheckIcon className="w-5 h-5 text-accent-green shrink-0 mt-0.5" />
            <div>
              <p className="text-sm font-medium text-text-primary">
                Install Key Created
              </p>
              <p className="text-2xs text-text-muted mt-0.5">
                Copy it now. You can also reveal it again later from the list.
              </p>
            </div>
          </div>
          <div>
            <span id="generated-install-key-label" className={LABEL}>
              Your Install Key
            </span>
            <Card
              aria-labelledby="generated-install-key-label"
              className="rounded-lg px-3.5 py-2.5 flex items-center gap-2"
            >
              <code className="flex-1 text-xs font-mono text-accent-cyan break-all select-all">
                {generatedKey}
              </code>
              <CopyButton text={generatedKey} size="md" />
            </Card>
          </div>
          <div>
            <span className={LABEL}>Install with this key</span>
            <WindowChrome
              variant="terminal"
              size="sm"
              titleBarSlot={<CopyButton text={installCommand} showLabel />}
            >
              <pre className="text-accent-cyan whitespace-pre overflow-x-auto">
                <span className="text-text-muted select-none">$ </span>
                {installCommand}
              </pre>
            </WindowChrome>
          </div>
        </div>
      ) : (
        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
          <InputField
            id="create-install-key-name"
            label="Name"
            value={name}
            onChange={handleNameChange}
            onBlur={handleNameBlur}
            placeholder="e.g. fleet-provisioning"
            error={nameError || undefined}
            maxLength={20}
          />
          <ModeField
            idPrefix="create-install-key"
            mode={mode}
            onModeChange={setMode}
            webhookUrl={webhookUrl}
            onWebhookUrlChange={setWebhookUrl}
            webhookSecret={webhookSecret}
            onWebhookSecretChange={setWebhookSecret}
            allowedMacs={allowedMacs}
            onAllowedMacsChange={setAllowedMacs}
            webhookTimeout={webhookTimeout}
            onWebhookTimeoutChange={setWebhookTimeout}
            webhookCallbackTtl={webhookCallbackTtl}
            onWebhookCallbackTtlChange={setWebhookCallbackTtl}
          />
          <ExpirationField value={expiresAt} onChange={setExpiresAt} />
          <UsageLimitField value={usageLimit} onChange={setUsageLimit} />
          <EphemeralField
            id="create-install-key-ephemeral"
            enabled={ephemeral}
            onEnabledChange={setEphemeral}
            timeout={ephemeralTimeout}
            onTimeoutChange={setEphemeralTimeout}
          />
          <TagsSelector
            id="create-install-key-tags"
            label="Tags"
            selected={tags}
            onChange={setTags}
            hint="Tags applied to every device registered with this key."
          />
          {error && <p className="text-2xs text-accent-red">{error}</p>}
        </form>
      )}
    </Drawer>
  );
}

export default CreateInstallKeyDrawer;
