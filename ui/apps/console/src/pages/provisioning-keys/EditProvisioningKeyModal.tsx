import { useState, type FormEvent } from "react";
import { CheckIcon } from "@heroicons/react/24/outline";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useUpdateProvisioningKey } from "@/hooks/useProvisioningKeyMutations";
import { type ProvisioningKey, type ProvisioningKeyUpdate } from "@/client";
import {
  getRemainingDays,
  isSystemKey,
  keyExpiryUpdatePayload,
} from "./helpers";
import Modal from "@/components/common/Modal";
import InputField from "@/components/common/fields/InputField";
import TagsSelector from "@/components/common/fields/TagsSelector";
import EphemeralField from "./EphemeralField";
import ExpirationField from "./ExpirationField";
import ModeField, { type ProvisioningKeyMode } from "./ModeField";
import UsageLimitField from "./UsageLimitField";
import {
  parseAllowedIdentities,
  validateModeConfig,
  validateName,
} from "./helpers";

/**
 * Edits a provisioning key's name and limits. The key itself is not re-issued.
 */
function EditProvisioningKeyModal({
  provisioningKey,
  onClose,
}: {
  provisioningKey: ProvisioningKey | null;
  onClose: () => void;
}) {
  const updateKey = useUpdateProvisioningKey();
  const open = provisioningKey !== null;
  const isSystem = provisioningKey ? isSystemKey(provisioningKey) : false;
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
  const [expiresIn, setExpiresIn] = useState("-1");
  const [expiryTouched, setExpiryTouched] = useState(false);
  const [isExpired, setIsExpired] = useState(false);
  const [tags, setTags] = useState<string[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [nameError, setNameError] = useState("");
  const [error, setError] = useState("");

  const identityList = parseAllowedIdentities(allowedIdentities);
  const alreadyWebhook = provisioningKey?.mode === "webhook";
  const modeError = validateModeConfig(
    mode,
    webhookUrl,
    webhookSecret,
    identityList,
    {
      secretOptional: alreadyWebhook,
      webhookTimeout,
      webhookCallbackTtl,
    },
  );

  const usedTimes = provisioningKey?.used_times ?? 0;
  const usageLimitError =
    usageLimit !== 0 && usageLimit < usedTimes
      ? `Limit can't be below the ${usedTimes} device${usedTimes === 1 ? "" : "s"} already registered with this key.`
      : "";

  useResetOnOpen(open, () => {
    setName(provisioningKey?.name ?? "");
    setMode((provisioningKey?.mode as ProvisioningKeyMode) ?? "automatic");
    setWebhookUrl(provisioningKey?.webhook_url ?? "");
    setWebhookSecret("");
    setAllowedIdentities(
      (provisioningKey?.allowed_identities ?? []).join("\n"),
    );
    setWebhookTimeout(provisioningKey?.webhook_timeout || 5);
    setWebhookCallbackTtl(provisioningKey?.webhook_callback_ttl || 3600);
    setUsageLimit(provisioningKey?.usage_limit ?? 0);
    setEphemeral(provisioningKey?.ephemeral ?? false);
    setEphemeralTimeout(provisioningKey?.ephemeral_timeout || 10);
    const { days, expired } = getRemainingDays(provisioningKey?.expires_at);
    setExpiresIn(days);
    setExpiryTouched(false);
    setIsExpired(expired);
    setTags(provisioningKey?.tags ?? []);
    setNameError("");
    setError("");
  });

  const handleExpiresInChange = (value: string) => {
    setExpiresIn(value);
    setExpiryTouched(true);
    setIsExpired(false);
  };

  const handleNameChange = (value: string) => {
    setName(value);
    if (nameError) setNameError(validateName(value.trim()));
  };

  const handleNameBlur = () => {
    if (name) setNameError(validateName(name.trim()));
  };

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (!provisioningKey) return;
    if (!isSystem) {
      const validationError = validateName(name.trim());
      if (validationError) {
        setNameError(validationError);
        return;
      }
      if (usageLimitError) {
        return;
      }
    }
    if (modeError) {
      setError(modeError);
      return;
    }
    setSubmitting(true);
    setNameError("");
    setError("");
    try {
      const modeBody: ProvisioningKeyUpdate = {
        mode,
        ...(mode === "webhook"
          ? {
              webhook_url: webhookUrl.trim(),
              webhook_timeout: webhookTimeout,
              webhook_callback_ttl: webhookCallbackTtl,
              ...(webhookSecret ? { webhook_secret: webhookSecret } : {}),
            }
          : {}),
        ...(mode === "allowlist" ? { allowed_identities: identityList } : {}),
      };
      const body: ProvisioningKeyUpdate = isSystem
        ? modeBody
        : {
            ...modeBody,
            name: name.trim(),
            usage_limit: usageLimit,
            ...(expiryTouched ? keyExpiryUpdatePayload(expiresIn) : {}),
            tags,
            ephemeral,
            ...(ephemeral ? { ephemeral_timeout: ephemeralTimeout } : {}),
          };

      await updateKey.mutateAsync({
        path: { key: provisioningKey.name },
        body,
      });
      onClose();
    } catch (err) {
      if (isSdkError(err) && err.status === 409) {
        setNameError("A key with that name already exists.");
      } else {
        setError(
          "Couldn't save the changes. Please review the values and try again.",
        );
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Edit Provisioning Key"
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={() => void handleSubmit()}
            disabled={
              submitting ||
              !!modeError ||
              (!isSystem && (!!nameError || !!usageLimitError || !name.trim()))
            }
            loading={submitting}
            icon={<CheckIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Save
          </Button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        {isSystem && (
          <p className="text-2xs text-text-muted">
            This is the namespace's default key: it accepts devices that
            register with only a tenant ID (no Provisioning Key). Its mode sets
            what happens to every such keyless device.
          </p>
        )}
        {!isSystem && (
          <InputField
            id="edit-provisioning-key-name"
            label="Name"
            value={name}
            onChange={handleNameChange}
            onBlur={handleNameBlur}
            placeholder="e.g. fleet-provisioning"
            error={nameError || undefined}
            maxLength={20}
          />
        )}
        <ModeField
          idPrefix="edit-provisioning-key"
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
          isEditing={alreadyWebhook}
          panelKey={provisioningKey?.name}
        />
        {!isSystem && (
          <>
            <div>
              {isExpired && (
                <Callout variant="warning" className="mb-3">
                  This key has expired. Set a new expiration to resume
                  registrations.
                </Callout>
              )}
              <ExpirationField
                expiresIn={expiresIn}
                onExpiresInChange={handleExpiresInChange}
              />
            </div>
            <div className="space-y-1.5">
              <UsageLimitField value={usageLimit} onChange={setUsageLimit} />
              {usageLimitError && (
                <p className="text-2xs text-accent-red">{usageLimitError}</p>
              )}
            </div>
            <EphemeralField
              id="edit-provisioning-key-ephemeral"
              enabled={ephemeral}
              onEnabledChange={setEphemeral}
              timeout={ephemeralTimeout}
              onTimeoutChange={setEphemeralTimeout}
            />
            <TagsSelector
              id="edit-provisioning-key-tags"
              label="Tags"
              selected={tags}
              onChange={setTags}
              hint="Tags applied to every device registered with this key."
            />
          </>
        )}
        {error && <p className="text-2xs text-accent-red">{error}</p>}
      </form>
    </Modal>
  );
}

export default EditProvisioningKeyModal;
