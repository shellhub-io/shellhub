import { useState } from "react";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useWatch } from "react-hook-form";
import { Card, Button } from "@shellhub/design-system/primitives";
import { CheckCircleIcon } from "@heroicons/react/24/outline";
import { useGenerateInvitationLink } from "@/hooks/useInvitationMutations";
import Drawer from "@/components/common/Drawer";
import CopyButton from "@/components/common/CopyButton";
import { FormInputField } from "@/components/common/fields/rhf";
import FormRootError from "@/components/common/fields/FormRootError";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { getConfig } from "@/env";
import { FormRoleSelector } from "./constants";
import {
  addMemberSchema,
  ADD_MEMBER_DEFAULTS,
  buildAddMemberBody,
  type AddMemberFormValues,
} from "./schemas";
import { LABEL } from "@/utils/styles";

interface AddMemberDrawerProps {
  open: boolean;
  onClose: () => void;
  tenantId: string;
}

function AddMemberDrawer({ open, onClose, tenantId }: AddMemberDrawerProps) {
  const generateLink = useGenerateInvitationLink();
  const emailDelivery = getConfig().cloud;
  const form = useDrawerForm(open, addMemberSchema, ADD_MEMBER_DEFAULTS);
  const {
    control,
    handleSubmit,
    setError,
    clearErrors,
    formState: { isValid, isSubmitting, errors },
  } = form;

  const email = useWatch({ control, name: "email" });
  const role = useWatch({ control, name: "role" });
  const trimmedEmail = (email ?? "").trim();

  const [generatedLink, setGeneratedLink] = useState("");
  const [addedDirectly, setAddedDirectly] = useState(false);

  useResetOnOpen(open, () => {
    setGeneratedLink("");
    setAddedDirectly(false);
  });

  const onValid = async (values: AddMemberFormValues) => {
    clearErrors("root");
    try {
      const result = await generateLink.mutateAsync({
        path: { tenant: tenantId },
        body: buildAddMemberBody(values),
      });
      const link = result.link ?? "";
      if (link) setGeneratedLink(link);
      else setAddedDirectly(true);
    } catch (err) {
      const defaultMessage = "Failed to send invitation. Please try again.";

      if (!isSdkError(err)) {
        setError("root", { message: defaultMessage });
        return;
      }

      const sdkErrorHandlers: Partial<Record<number, { name: "email" | "root"; message: string }>> = {
        400: { name: "email", message: "Invalid email or role." },
        403: { name: "root", message: "You don't have permission to invite members." },
        404: { name: "email", message: "No account exists for this email." },
        409: { name: "email", message: "This user is already a member or has a pending invitation." },
      };

      const sdkError = sdkErrorHandlers[err.status] ?? {
        name: "root",
        message: defaultMessage,
      };

      setError(sdkError.name, { message: sdkError.message });
    }
  };

  const done = !!generatedLink || addedDirectly;

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={
        addedDirectly ? "Member Added" : done ? "Invitation Link" : "Add Member"
      }
      subtitle={
        done ? <span className="font-mono">{trimmedEmail}</span> : undefined
      }
      footer={
        done ? (
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
            >
              Add Member
            </Button>
          </>
        )
      }
    >
      {addedDirectly ? (
        <Card className="rounded-lg px-3.5 py-3 flex items-center gap-3">
          <CheckCircleIcon className="w-5 h-5 text-accent-green shrink-0" />
          <p className="text-xs text-text-secondary leading-relaxed">
            This person already has a ShellHub account, so we added them to the
            namespace as{" "}
            <span className="font-medium text-text-primary">{role}</span> right
            away.
          </p>
        </Card>
      ) : done ? (
        <div className="space-y-3">
          <div>
            <span id="add-member-link-label" className={LABEL}>
              Invitation link
            </span>
            <Card
              aria-labelledby="add-member-link-label"
              className="rounded-lg px-3.5 py-2.5 flex items-center gap-2"
            >
              <code className="flex-1 text-xs font-mono text-accent-cyan break-all select-all">
                {generatedLink}
              </code>
              <CopyButton text={generatedLink} size="md" showLabel />
            </Card>
          </div>
          <p className="text-2xs text-text-muted leading-relaxed">
            {emailDelivery ? (
              <>
                We emailed the invitation to{" "}
                <span className="font-mono text-text-secondary">
                  {trimmedEmail}
                </span>
                . Share this link too if you'd rather send it yourself. It works
                only for this address and expires in 7 days.
              </>
            ) : (
              <>
                Send it to{" "}
                <span className="font-mono text-text-secondary">
                  {trimmedEmail}
                </span>{" "}
                to join the namespace. The link works only for this address and
                expires in 7 days.
              </>
            )}
          </p>
        </div>
      ) : (
        <form
          onSubmit={(e) => void handleSubmit(onValid)(e)}
          className="space-y-5"
        >
          <FormInputField
            name="email"
            control={control}
            id="add-member-email"
            label="Email"
            type="email"
            placeholder="user@example.com"
          />

          <FormRoleSelector name="role" control={control} />

          <FormRootError message={errors.root?.message} />
        </form>
      )}
    </Drawer>
  );
}

export default AddMemberDrawer;
