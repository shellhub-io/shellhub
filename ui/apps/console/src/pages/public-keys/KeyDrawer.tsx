import { useMemo } from "react";
import { useWatch } from "react-hook-form";
import { isSdkError } from "@/api/errors";
import {
  UserGroupIcon,
  UserIcon,
  TagIcon,
  ClipboardDocumentListIcon,
} from "@heroicons/react/24/outline";
import { DevicesIcon } from "@shellhub/design-system/primitives";
import {
  useCreatePublicKey,
  useUpdatePublicKey,
} from "@/hooks/usePublicKeyMutations";
import type { PublicKeyResponse } from "@/client";
import RadioCard from "@/components/common/fields/RadioCard";
import FormDrawer from "@/components/common/FormDrawer";
import {
  FormInputField,
  FormRadioGroupField,
  FormTagsSelector,
} from "@/components/common/fields/rhf";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import KeyDataInput from "./KeyDataInput";
import {
  DEFAULT_VALUES,
  buildKeyBody,
  buildKeyDefaults,
  keySchema,
  type KeyFormValues,
} from "./keySchema";

export default function KeyDrawer({
  open,
  editKey,
  onClose,
}: {
  open: boolean;
  editKey: PublicKeyResponse | null;
  onClose: () => void;
}) {
  const createKey = useCreatePublicKey();
  const updateKey = useUpdatePublicKey();
  const isEdit = !!editKey;

  const schema = useMemo(() => keySchema(isEdit ? "edit" : "create"), [isEdit]);

  const form = useDrawerForm(
    open,
    schema,
    editKey ? buildKeyDefaults(editKey) : DEFAULT_VALUES,
  );
  const { control, setError, setValue, getValues } = form;

  const usernameOption = useWatch({ control, name: "usernameOption" });
  const filterOption = useWatch({ control, name: "filterOption" });

  const handleFileName = (filename: string) => {
    if (!getValues("name").trim())
      setValue("name", filename || "Imported Public Key", {
        shouldValidate: true,
      });
  };

  const onSubmit = async (values: KeyFormValues) => {
    try {
      if (isEdit && editKey) {
        const body = buildKeyBody(values);
        await updateKey.mutateAsync({
          path: { fingerprint: editKey.fingerprint },
          body: {
            name: body.name,
            username: body.username,
            filter: body.filter,
          },
        });
      } else {
        await createKey.mutateAsync({ body: buildKeyBody(values) });
      }
      onClose();
    } catch (err: unknown) {
      if (!isEdit && isSdkError(err) && err.status === 409) {
        setError("data", { message: "This public key already exists." });
      } else {
        setError("root", {
          message:
            err instanceof Error
              ? err.message
              : `Failed to ${isEdit ? "update" : "create"} public key`,
        });
      }
    }
  };

  return (
    <FormDrawer
      form={form}
      onSubmit={onSubmit}
      open={open}
      onClose={onClose}
      title={isEdit ? "Edit Public Key" : "New Public Key"}
      submitLabel={isEdit ? "Save Changes" : "Create Key"}
    >
      <FormInputField
        name="name"
        control={control}
        id="public-key-name"
        label="Name"
        placeholder="Name used to identify the public key"
      />

      <div>
        <FormRadioGroupField<KeyFormValues, "all" | "username">
          name="usernameOption"
          control={control}
          label="Username access"
        >
          <RadioCard
            value="all"
            icon={<UserGroupIcon className="w-4 h-4" />}
            label="Allow any user"
            description="The key will work for all usernames on the device."
          />
          <RadioCard
            value="username"
            icon={<UserIcon className="w-4 h-4" />}
            label="Restrict by username"
            description="Only allow connections matching a username pattern."
          />
        </FormRadioGroupField>
        {usernameOption === "username" && (
          <div className="mt-2">
            <FormInputField
              name="username"
              control={control}
              id="public-key-username-pattern"
              label="Username pattern"
              hideLabel
              placeholder="e.g. root"
              variant="mono"
            />
          </div>
        )}
      </div>

      <div>
        <FormRadioGroupField<KeyFormValues, "all" | "hostname" | "tags">
          name="filterOption"
          control={control}
          label="Device access"
        >
          <RadioCard
            value="all"
            icon={<DevicesIcon className="w-4 h-4" />}
            label="All devices"
            description="The key will be accepted by any device in the namespace."
          />
          <RadioCard
            value="hostname"
            icon={<ClipboardDocumentListIcon className="w-4 h-4" />}
            label="Filter by hostname"
            description="Restrict access using a regexp pattern for hostname."
          />
          <RadioCard
            value="tags"
            icon={<TagIcon className="w-4 h-4" />}
            label="Filter by tags"
            description="Restrict access to devices matching specific tags."
          />
        </FormRadioGroupField>
        {filterOption === "hostname" && (
          <div className="mt-2">
            <FormInputField
              name="hostname"
              control={control}
              id="public-key-hostname-pattern"
              label="Hostname pattern"
              hideLabel
              placeholder="e.g. .*"
              variant="mono"
            />
          </div>
        )}
        {filterOption === "tags" && (
          <div className="mt-2">
            <FormTagsSelector
              name="tags"
              control={control}
              id="public-key-filter-tags"
              label="Filter by tags"
            />
          </div>
        )}
      </div>

      <KeyDataInput
        name="data"
        control={control}
        disabled={isEdit}
        onFileName={handleFileName}
      />
    </FormDrawer>
  );
}
