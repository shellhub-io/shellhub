import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { KeyIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { createInstanceApiKeyMutation } from "@/client";
import SecretReveal from "@/components/common/SecretReveal";
import Modal from "@/components/common/Modal";
import {
  FormInputField,
  FormRadioGroupField,
} from "@/components/common/fields/rhf";
import FormRootError from "@/components/common/fields/FormRootError";
import RadioPill from "@/components/common/fields/RadioPill";
import { useDiscardGuard } from "@/hooks/useDiscardGuard";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { KEY_NAME_MAX_LENGTH } from "@/utils/validation";
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
  const invalidateKeys = useInvalidateByIds("listInstanceApiKeys");
  const createKey = useMutation({
    ...createInstanceApiKeyMutation(),
    onSuccess: invalidateKeys,
  });
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
        body: buildGenerateInstanceKeyBody(values),
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
            <SecretReveal
              title="Instance API Key Generated"
              note="Copy this key now. You won't be able to see it again."
              label="Your Instance API Key"
              value={generatedKey}
            />
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
              maxLength={KEY_NAME_MAX_LENGTH}
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
