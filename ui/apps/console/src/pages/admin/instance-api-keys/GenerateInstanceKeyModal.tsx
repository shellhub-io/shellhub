import { useState } from "react";
import { KeyIcon, CheckIcon } from "@heroicons/react/24/outline";
import { Card, Button } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useCreateInstanceApiKey } from "@/hooks/useInstanceApiKeyMutations";
import CopyButton from "@/components/common/CopyButton";
import Modal from "@/components/common/Modal";
import {
  FormInputField,
  FormRadioGroupField,
} from "@/components/common/fields/rhf";
import FormRootError from "@/components/common/fields/FormRootError";
import RadioPill from "@/components/common/fields/RadioPill";
import { useDiscardGuard } from "@/hooks/useDiscardGuard";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { LABEL } from "@/utils/styles";
import {
  generateInstanceKeySchema,
  GENERATE_INSTANCE_KEY_DEFAULTS,
  buildGenerateInstanceKeyBody,
  INSTANCE_KEY_EXPIRY_OPTIONS,
  type GenerateInstanceKeyFormValues,
} from "./schemas";

/**
 * Creates an instance API key and shows it. This is the only time the secret is readable, so the
 * modal has to present it before it can be closed.
 */
function GenerateInstanceKeyModal({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const createKey = useCreateInstanceApiKey();
  const form = useDrawerForm(
    open,
    generateInstanceKeySchema,
    GENERATE_INSTANCE_KEY_DEFAULTS,
  );
  const {
    control,
    handleSubmit,
    setError,
    clearErrors,
    formState: { isValid, isSubmitting, isDirty, errors },
  } = form;

  const [generatedKey, setGeneratedKey] = useState("");

  useResetOnOpen(open, () => setGeneratedKey(""));

  const onValid = async (values: GenerateInstanceKeyFormValues) => {
    clearErrors("root");
    try {
      const result = await createKey.mutateAsync({
        body: buildGenerateInstanceKeyBody({
          name: values.name,
          expiresAt: values.expiresAt,
        }),
      });
      setGeneratedKey(result.id);
    } catch (err) {
      if (isSdkError(err) && err.status === 409) {
        setError("name", { message: "That name is already taken." });
      } else {
        setError("root", {
          message: "Failed to generate the instance API key.",
        });
      }
    }
  };

  const { requestClose, prompt } = useDiscardGuard({
    open,
    dirty: isDirty && !isSubmitting && !generatedKey,
    onClose,
  });

  return (
    <>
      <Modal
        open={open}
        onClose={requestClose}
        icon={<KeyIcon />}
        title="Generate instance API key"
        description="Instance API keys let automation act as an instance administrator."
        footer={
          generatedKey ? (
            <Button variant="primary" onClick={onClose}>
              Done
            </Button>
          ) : (
            <>
              <Button variant="ghost" onClick={requestClose}>
                Cancel
              </Button>
              <Button
                variant="primary"
                onClick={() => void handleSubmit(onValid)()}
                disabled={!isValid || isSubmitting}
                loading={isSubmitting}
                icon={<KeyIcon className="w-4 h-4" strokeWidth={2} />}
              >
                Generate
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
                  Instance API Key Generated
                </p>
                <p className="text-2xs text-text-muted mt-0.5">
                  Copy this key now. You won't be able to see it again.
                </p>
              </div>
            </div>
            <div>
              <span id="generated-instance-api-key-label" className={LABEL}>
                Your Instance API Key
              </span>
              <Card
                aria-labelledby="generated-instance-api-key-label"
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
          <form
            onSubmit={(e) => void handleSubmit(onValid)(e)}
            className="space-y-5"
          >
            <FormInputField
              name="name"
              control={control}
              id="generate-instance-key-name"
              label="Name"
              placeholder="e.g. billing-export"
              maxLength={20}
            />
            <FormRadioGroupField
              name="expiresAt"
              control={control}
              label="Expiration"
              containerClassName="flex flex-wrap gap-1.5"
            >
              {INSTANCE_KEY_EXPIRY_OPTIONS.map((opt) => (
                <RadioPill
                  key={opt.value}
                  value={String(opt.value)}
                  label={opt.label}
                />
              ))}
            </FormRadioGroupField>
            <FormRootError message={errors.root?.message} />
          </form>
        )}
      </Modal>
      {prompt}
    </>
  );
}

export default GenerateInstanceKeyModal;
