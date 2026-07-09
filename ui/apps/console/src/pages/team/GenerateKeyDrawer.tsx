import { useState } from "react";
import { isSdkError } from "@/api/errors";
import { KeyIcon, CheckIcon } from "@heroicons/react/24/outline";
import { Card, Button } from "@shellhub/design-system/primitives";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useCreateApiKey } from "@/hooks/useApiKeyMutations";
import CopyButton from "@/components/common/CopyButton";
import Drawer from "@/components/common/Drawer";
import {
  FormInputField,
  FormRadioGroupField,
} from "@/components/common/fields/rhf";
import FormRootError from "@/components/common/fields/FormRootError";
import RadioPill from "@/components/common/fields/RadioPill";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { FormRoleSelector } from "./constants";
import {
  generateKeySchema,
  GENERATE_KEY_DEFAULTS,
  buildGenerateKeyBody,
  type GenerateKeyFormValues,
} from "./schemas";
import { EXPIRY_OPTIONS } from "./helpers";
import { LABEL } from "@/utils/styles";

function GenerateKeyDrawer({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const createKey = useCreateApiKey();
  const form = useDrawerForm(open, generateKeySchema, GENERATE_KEY_DEFAULTS);
  const {
    control,
    handleSubmit,
    setError,
    clearErrors,
    formState: { isValid, isSubmitting, errors },
  } = form;

  const [generatedKey, setGeneratedKey] = useState("");

  useResetOnOpen(open, () => setGeneratedKey(""));

  const onValid = async (values: GenerateKeyFormValues) => {
    clearErrors("root");
    try {
      const result = await createKey.mutateAsync({
        body: buildGenerateKeyBody(values),
      });
      setGeneratedKey(result.id);
    } catch (err) {
      if (isSdkError(err) && err.status === 400) {
        setError("name", {
          message:
            "Name must be 3–20 characters: letters, numbers, - and _ only.",
        });
      } else {
        setError("root", {
          message: "Failed to generate API key. The name may already exist.",
        });
      }
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Generate API Key"
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
            maxLength={20}
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
    </Drawer>
  );
}

export default GenerateKeyDrawer;
