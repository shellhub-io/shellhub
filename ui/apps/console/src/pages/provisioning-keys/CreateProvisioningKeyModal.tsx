import { useMutation } from "@tanstack/react-query";
import { useState, FormEvent } from "react";
import { TicketIcon } from "@heroicons/react/24/outline";
import { Button, WindowChrome } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { buildInstallCommand } from "@/utils/installCommand";
import CopyButton from "@/components/common/CopyButton";
import SecretReveal from "@/components/common/SecretReveal";
import Modal from "@/components/common/Modal";
import InputField from "@/components/common/fields/InputField";
import TagsSelector from "@/components/common/fields/TagsSelector";
import { provisioningKeyCreateMutation } from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { LABEL } from "@/utils/styles";
import {
  KEY_NAME_HINT,
  KEY_NAME_MAX_LENGTH,
  validateKeyName,
} from "@/utils/validation";
import EphemeralField from "./EphemeralField";
import ExpirationField from "./ExpirationField";
import ModeField, { type ProvisioningKeyMode } from "./ModeField";
import UsageLimitField from "./UsageLimitField";
import {
  keyExpiryPayload,
  parseAllowedIdentities,
  validateModeConfig,
} from "./helpers";

/**
 * Creates a provisioning key and shows it with its install command. onCreated receives the new
 * key's trimmed name, so the caller can select it in its list.
 */
function CreateProvisioningKeyModal({
  open,
  onClose,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  onCreated?: (name: string) => void;
}) {
  const invalidateKeys = useInvalidateByIds("provisioningKeyList");
  const createKey = useMutation({
    ...provisioningKeyCreateMutation(),
    onSuccess: invalidateKeys,
  });
  const [name, setName] = useState("");
  const [mode, setMode] = useState<ProvisioningKeyMode>("automatic");
  const [webhookUrl, setWebhookUrl] = useState("");
  const [webhookSecret, setWebhookSecret] = useState("");
  const [allowedIdentities, setAllowedIdentities] = useState("");
  const [webhookTimeout, setWebhookTimeout] = useState(5);
  const [webhookCallbackTtl, setWebhookCallbackTtl] = useState(3600);
  const [usageLimit, setUsageLimit] = useState(0);
  const [ephemeral, setEphemeral] = useState(false);
  const [ephemeralTimeout, setEphemeralTimeout] = useState(10);
  const [tags, setTags] = useState<string[]>([]);
  const [expiresIn, setExpiresIn] = useState("30");
  const [submitting, setSubmitting] = useState(false);
  const [nameError, setNameError] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [generatedKey, setGeneratedKey] = useState("");

  const identityList = parseAllowedIdentities(allowedIdentities);
  const modeError = validateModeConfig(
    mode,
    webhookUrl,
    webhookSecret,
    identityList,
    { webhookTimeout, webhookCallbackTtl },
  );

  useResetOnOpen(open, () => {
    setName("");
    setMode("automatic");
    setWebhookUrl("");
    setWebhookSecret("");
    setAllowedIdentities("");
    setWebhookTimeout(5);
    setWebhookCallbackTtl(3600);
    setUsageLimit(0);
    setEphemeral(false);
    setEphemeralTimeout(10);
    setTags([]);
    setExpiresIn("30");
    setNameError(null);
    setError("");
    setGeneratedKey("");
  });

  const handleNameChange = (value: string) => {
    setName(value);
    if (nameError) setNameError(validateKeyName(value.trim()));
  };

  const handleNameBlur = () => {
    if (name) setNameError(validateKeyName(name.trim()));
  };

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    const validationError = validateKeyName(name.trim());
    if (validationError) {
      setNameError(validationError);
      return;
    }
    if (modeError) {
      setError(modeError);
      return;
    }
    setSubmitting(true);
    setNameError(null);
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
          ...(mode === "allowlist" ? { allowed_identities: identityList } : {}),
          ...keyExpiryPayload(expiresIn),
          usage_limit: usageLimit,
          ephemeral,
          ...(ephemeral ? { ephemeral_timeout: ephemeralTimeout } : {}),
          tags,
        },
      });
      setGeneratedKey(result.key);
      onCreated?.(name.trim());
    } catch (err) {
      if (isSdkError(err) && err.status === 409) {
        setNameError("A key with that name already exists.");
      } else if (isSdkError(err) && err.status === 400) {
        setNameError(KEY_NAME_HINT);
      } else {
        setError(
          "Couldn't create the Provisioning Key. Please review the values and try again.",
        );
      }
    } finally {
      setSubmitting(false);
    }
  };

  const installCommand = buildInstallCommand(
    `PROVISIONING_KEY=${generatedKey}`,
    window.location.origin,
  );

  return (
    <Modal
      open={open}
      onClose={onClose}
      icon={<TicketIcon />}
      title="Create provisioning key"
      description="Devices installed with this key register in this namespace on their own."
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
              Create Provisioning Key
            </Button>
          </>
        )
      }
    >
      {generatedKey ? (
        <div className="space-y-5">
          <SecretReveal
            title="Provisioning Key Created"
            note="Copy it now. You can also reveal it again later from the list."
            label="Your Provisioning Key"
            value={generatedKey}
          />
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
            id="create-provisioning-key-name"
            label="Name"
            value={name}
            onChange={handleNameChange}
            onBlur={handleNameBlur}
            placeholder="e.g. fleet-provisioning"
            error={nameError || undefined}
            maxLength={KEY_NAME_MAX_LENGTH}
          />
          <ModeField
            idPrefix="create-provisioning-key"
            mode={mode}
            onModeChange={setMode}
            webhookUrl={webhookUrl}
            onWebhookUrlChange={setWebhookUrl}
            webhookSecret={webhookSecret}
            onWebhookSecretChange={setWebhookSecret}
            allowedIdentities={allowedIdentities}
            onAllowedIdentitiesChange={setAllowedIdentities}
            webhookTimeout={webhookTimeout}
            onWebhookTimeoutChange={setWebhookTimeout}
            webhookCallbackTtl={webhookCallbackTtl}
            onWebhookCallbackTtlChange={setWebhookCallbackTtl}
          />
          <ExpirationField
            expiresIn={expiresIn}
            onExpiresInChange={setExpiresIn}
          />
          <UsageLimitField value={usageLimit} onChange={setUsageLimit} />
          <EphemeralField
            id="create-provisioning-key-ephemeral"
            enabled={ephemeral}
            onEnabledChange={setEphemeral}
            timeout={ephemeralTimeout}
            onTimeoutChange={setEphemeralTimeout}
          />
          <TagsSelector
            id="create-provisioning-key-tags"
            label="Tags"
            selected={tags}
            onChange={setTags}
            hint="Tags applied to every device registered with this key."
          />
          {error && <p className="text-2xs text-accent-red">{error}</p>}
        </form>
      )}
    </Modal>
  );
}

export default CreateProvisioningKeyModal;
