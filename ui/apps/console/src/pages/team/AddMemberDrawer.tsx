import { useState, FormEvent } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { PlusIcon } from "@heroicons/react/24/outline";
import { useAddMember } from "@/hooks/useMemberMutations";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import { EMAIL_REGEX } from "@/utils/validation";
import { RoleSelector } from "./constants";
import { type AssignableRole } from "./helpers";
import { Button } from "@shellhub/design-system/primitives";

/* --- Add Member Drawer --- */

function AddMemberDrawer({
  open,
  onClose,
  tenantId,
}: {
  open: boolean;
  onClose: () => void;
  tenantId: string;
}) {
  const addMember = useAddMember();
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<AssignableRole>("operator");
  const [submitting, setSubmitting] = useState(false);
  const [emailError, setEmailError] = useState("");
  const [error, setError] = useState("");

  useResetOnOpen(open, () => {
    setEmail("");
    setRole("operator");
    setEmailError("");
    setError("");
  });

  const trimmedEmail = email.trim();
  const emailValid = EMAIL_REGEX.test(trimmedEmail);

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (!emailValid) {
      setEmailError("Enter a valid email address.");
      return;
    }
    setSubmitting(true);
    setEmailError("");
    setError("");
    try {
      await addMember.mutateAsync({
        path: { tenant: tenantId },
        body: { email: trimmedEmail, role },
      });
      onClose();
    } catch {
      setError("Failed to add member. Check the email and try again.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Add Member"
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={() => void handleSubmit()}
            disabled={!emailValid || submitting}
            loading={submitting}
            icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Add Member
          </Button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        <InputField
          id="add-member-email"
          label="Email"
          type="email"
          value={email}
          onChange={(email) => {
            setEmail(email);
            if (emailError) setEmailError("");
          }}
          placeholder="user@example.com"
          hint="Must have an existing ShellHub account"

          error={emailError}
        />
        <RoleSelector value={role} onChange={setRole} />
        {error && <p className="text-2xs text-accent-red">{error}</p>}
      </form>
    </Drawer>
  );
}

export default AddMemberDrawer;
