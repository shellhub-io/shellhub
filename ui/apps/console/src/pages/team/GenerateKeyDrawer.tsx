import { useState, FormEvent } from "react";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { KeyIcon, CheckIcon } from "@heroicons/react/24/outline";
import { Card } from "@shellhub/design-system/primitives";
import { useCreateApiKey } from "@/hooks/useApiKeyMutations";
import { type ApiKeyCreate } from "@/client";
import CopyButton from "@/components/common/CopyButton";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import RadioPill from "@/components/common/fields/RadioPill";
import { RoleSelector } from "./constants";
import { EXPIRY_OPTIONS, type AssignableRole } from "./helpers";
import { LABEL } from "@/utils/styles";
import Spinner from "@/components/common/Spinner";

function validateName(value: string): string {
  if (value.length < 3) return "Name must be at least 3 characters.";
  if (value.length > 20) return "Name must be at most 20 characters.";
  if (!/^[a-zA-Z0-9_-]+$/.test(value))
    return "Name can only contain letters, numbers, - and _.";
  return "";
}

/* --- Generate API Key Drawer --- */

function GenerateKeyDrawer({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const createKey = useCreateApiKey();
  const [name, setName] = useState("");
  const [role, setRole] = useState<AssignableRole>("administrator");
  const [expiresIn, setExpiresIn] = useState<ApiKeyCreate["expires_at"]>(30);
  const [submitting, setSubmitting] = useState(false);
  const [nameError, setNameError] = useState("");
  const [error, setError] = useState("");
  const [generatedKey, setGeneratedKey] = useState("");

  useResetOnOpen(open, () => {
    setName("");
    setRole("administrator");
    setExpiresIn(30);
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
    setSubmitting(true);
    setNameError("");
    setError("");
    try {
      const result = await createKey.mutateAsync({
        body: { name: name.trim(), role, expires_at: expiresIn },
      });
      setGeneratedKey(result.id);
    } catch (err) {
      if (isSdkError(err) && err.status === 400) {
        setNameError(
          "Name must be 3–20 characters: letters, numbers, - and _ only.",
        );
      } else {
        setError("Failed to generate API key. The name may already exist.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Generate API Key"
      footer={
        generatedKey ? (
          <button
            type="button"
            onClick={onClose}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
          >
            Done
          </button>
        ) : (
          <>
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={() => void handleSubmit()}
              disabled={submitting || !!nameError || !name.trim()}
              className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
            >
              {submitting ? (
                <Spinner size="md" tone="onPrimary" />
              ) : (
                <KeyIcon className="w-4 h-4" strokeWidth={2} />
              )}
              Generate Key
            </button>
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
                API Key Generated
              </p>
              <p className="text-2xs text-text-muted mt-0.5">
                Copy this key now. You won't be able to see it again.
              </p>
            </div>
          </div>
          <div>
            <span id="generated-api-key-label" className={LABEL}>
              Your API Key
            </span>
            <Card
              aria-labelledby="generated-api-key-label"
              className="rounded-lg px-3.5 py-2.5 flex items-center gap-2"
            >
              <code className="flex-1 text-xs font-mono text-accent-cyan break-all select-all">
                {generatedKey}
              </code>
              <CopyButton text={generatedKey} size="md" />
            </Card>
          </div>
        </div>
      ) : (
        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
          <InputField
            id="generate-key-name"
            label="Name"
            value={name}
            onChange={handleNameChange}
            onBlur={handleNameBlur}
            placeholder="e.g. ci-pipeline"
            error={nameError || undefined}
            maxLength={20}
            autoFocus={open}
          />
          <RoleSelector value={role} onChange={setRole} />
          <RadioGroupField
            label="Expiration"
            value={String(expiresIn)}
            onChange={(v) =>
              setExpiresIn(Number(v) as ApiKeyCreate["expires_at"])
            }
            containerClassName="flex flex-wrap gap-1.5"
          >
            {EXPIRY_OPTIONS.map((opt) => (
              <RadioPill
                key={opt.value}
                value={String(opt.value)}
                label={opt.label}
              />
            ))}
          </RadioGroupField>
          {error && <p className="text-2xs text-accent-red">{error}</p>}
        </form>
      )}
    </Drawer>
  );
}

export default GenerateKeyDrawer;
