import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { isSdkError } from "@/api/errors";
import { KeyIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { apiKeyCreateMutation } from "@/client";
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
import { KEY_NAME_HINT, KEY_NAME_MAX_LENGTH } from "@/utils/validation";
import { FormRoleSelector } from "./constants";
import {
  generateKeySchema,
  GENERATE_KEY_DEFAULTS,
  buildGenerateKeyBody,
  type GenerateKeyFormValues,
} from "./schemas";
import { EXPIRY_OPTIONS } from "./helpers";

/**
 * Creates an API key and shows it. This is the only time the secret is readable, so the modal
 * has to present it before it can be closed.
 */
function GenerateKeyModal({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const invalidateKeys = useInvalidateByIds("apiKeyList");
  const createKey = useMutation({
    ...apiKeyCreateMutation(),
    onSuccess: invalidateKeys,
  });
  const form = useDrawerForm(open, generateKeySchema, GENERATE_KEY_DEFAULTS);
  const {
    control,
    handleSubmit,
    setError,
    clearErrors,
    formState: { isValid, isSubmitting, isDirty, errors },
  } = form;

  const [generatedKey, setGeneratedKey] = useState("");

  useResetOnOpen(open, () => setGeneratedKey(""));

  const onValid = async (values: GenerateKeyFormValues) => {
    clearErrors("root");
    try {
      const result = await createKey.mutateAsync({
        body: buildGenerateKeyBody(values),
      });
      setGeneratedKey(result.key);
    } catch (err) {
      if (isSdkError(err) && err.status === 400) {
        setError("name", { message: KEY_NAME_HINT });
      } else {
        setError("root", {
          message: "Failed to generate API key. The name may already exist.",
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
        title="Generate API key"
        description="An API key lets scripts and integrations call the API as this namespace."
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
                Generate Key
              </Button>
            </>
          )
        }
      >
        {generatedKey ? (
          <div className="space-y-5">
            <SecretReveal
              title="API Key Generated"
              note="Copy this key now. You won't be able to see it again."
              label="Your API Key"
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
              id="generate-key-name"
              label="Name"
              placeholder="e.g. ci-pipeline"
              maxLength={KEY_NAME_MAX_LENGTH}
            />
            <FormRoleSelector name="role" control={control} />
            <FormRadioGroupField
              name="expiresIn"
              control={control}
              label="Expiration"
              containerClassName="flex flex-wrap gap-1.5"
            >
              {EXPIRY_OPTIONS.map((opt) => (
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

export default GenerateKeyModal;
