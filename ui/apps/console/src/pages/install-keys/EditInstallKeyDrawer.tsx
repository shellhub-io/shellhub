import { useState, type FormEvent } from "react";
import { CheckIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useUpdateInstallKey } from "@/hooks/useInstallKeyMutations";
import { type InstallKey, type InstallKeyUpdate } from "@/client";
import { isSystemKey } from "./helpers";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import TagsSelector from "@/components/common/fields/TagsSelector";
import EphemeralField from "./EphemeralField";
import ExpirationField from "./ExpirationField";
import ModeField, { type InstallKeyMode } from "./ModeField";
import UsageLimitField from "./UsageLimitField";
import { parseAllowedMacs, validateModeConfig, validateName } from "./helpers";

function EditInstallKeyDrawer({
  installKey,
  onClose,
}: {
  installKey: InstallKey | null;
  onClose: () => void;
}) {
  const updateKey = useUpdateInstallKey();
  const open = installKey !== null;
  // The legacy/system key is edit-restricted: only its enrollment mode is shown and sent.
  const isSystem = installKey ? isSystemKey(installKey) : false;
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
  const [expiresAt, setExpiresAt] = useState<string | null>(null);
  const [tags, setTags] = useState<string[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [nameError, setNameError] = useState("");
  const [error, setError] = useState("");

  const macList = parseAllowedMacs(allowedMacs);
  // A key already in webhook mode may keep its stored (write-only) secret by leaving the field blank; a
  // secret is only required when switching into webhook mode.
  const alreadyWebhook = installKey?.mode === "webhook";
  const modeError = validateModeConfig(
    mode,
    webhookUrl,
    webhookSecret,
    macList,
    {
      secretOptional: alreadyWebhook,
    },
  );

  // The usage limit can't be lowered below the devices that already enrolled: those enrollments
  // happened and the counter can't be walked back (the backend enforces this too). Unlimited (0) is
  // always allowed. Validated live so the message points at the right field instead of a stray 400.
  const usedTimes = installKey?.used_times ?? 0;
  const usageLimitError =
    usageLimit !== 0 && usageLimit < usedTimes
      ? `Limit can't be below the ${usedTimes} device${usedTimes === 1 ? "" : "s"} already registered with this key.`
      : "";

  useResetOnOpen(open, () => {
    setName(installKey?.name ?? "");
    setMode((installKey?.mode as InstallKeyMode) ?? "automatic");
    setWebhookUrl(installKey?.webhook_url ?? "");
    setWebhookSecret("");
    setAllowedMacs((installKey?.allowed_macs ?? []).join("\n"));
    setWebhookTimeout(installKey?.webhook_timeout || 5);
    setWebhookCallbackTtl(installKey?.webhook_callback_ttl || 3600);
    setUsageLimit(installKey?.usage_limit ?? 1);
    setEphemeral(installKey?.ephemeral ?? false);
    setEphemeralTimeout(installKey?.ephemeral_timeout || 10);
    setExpiresAt(installKey?.expires_at ?? null);
    setTags(installKey?.tags ?? []);
    setNameError("");
    setError("");
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
    if (!installKey) return;
    // The legacy/system key only exposes its mode; its name and other fields are fixed and must not
    // be sent (the API rejects changing them).
    if (!isSystem) {
      const validationError = validateName(name.trim());
      if (validationError) {
        setNameError(validationError);
        return;
      }
      // Block the impossible limit here so the message lands on the field, not as a bare 400.
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
      const modeBody: InstallKeyUpdate = {
        mode,
        // The webhook secret is only sent when the user typed one, so an unchanged webhook key keeps
        // its stored secret.
        ...(mode === "webhook"
          ? {
              webhook_url: webhookUrl.trim(),
              webhook_timeout: webhookTimeout,
              webhook_callback_ttl: webhookCallbackTtl,
              ...(webhookSecret ? { webhook_secret: webhookSecret } : {}),
            }
          : {}),
        ...(mode === "allowlist" ? { allowed_macs: macList } : {}),
      };
      const body: InstallKeyUpdate = isSystem
        ? modeBody
        : {
            ...modeBody,
            name: name.trim(),
            usage_limit: usageLimit,
            expires_at: expiresAt,
            tags,
            ephemeral,
            // Only meaningful for ephemeral keys; already clamped to 1-10 by the field.
            ...(ephemeral ? { ephemeral_timeout: ephemeralTimeout } : {}),
          };

      await updateKey.mutateAsync({ path: { key: installKey.name }, body });
      onClose();
    } catch (err) {
      // 409 is the rename collision (the one error the status alone pins to a field). Every other
      // 400 is bodyless, so it can't be attributed to a specific field; name, limit and expiry are
      // all validated client-side above, so a 400 here is an unexpected edge, shown at form level.
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
    <Drawer
      open={open}
      onClose={onClose}
      title="Edit Install Key"
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
            register with only a tenant ID (no Install Key). Its mode sets what
            happens to every such keyless device.
          </p>
        )}
        {!isSystem && (
          <InputField
            id="edit-install-key-name"
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
          idPrefix="edit-install-key"
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
        {!isSystem && (
          <>
            <ExpirationField value={expiresAt} onChange={setExpiresAt} />
            <div className="space-y-1.5">
              <UsageLimitField value={usageLimit} onChange={setUsageLimit} />
              {usageLimitError && (
                <p className="text-2xs text-accent-red">{usageLimitError}</p>
              )}
            </div>
            <EphemeralField
              id="edit-install-key-ephemeral"
              enabled={ephemeral}
              onEnabledChange={setEphemeral}
              timeout={ephemeralTimeout}
              onTimeoutChange={setEphemeralTimeout}
            />
            <TagsSelector
              id="edit-install-key-tags"
              label="Tags"
              selected={tags}
              onChange={setTags}
              hint="Tags applied to every device registered with this key."
            />
          </>
        )}
        {error && <p className="text-2xs text-accent-red">{error}</p>}
      </form>
    </Drawer>
  );
}

export default EditInstallKeyDrawer;
