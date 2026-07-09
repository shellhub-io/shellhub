import { useState, FormEvent } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { PlusIcon, ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import {
  useAddMember,
  useCreateActivationToken,
} from "@/hooks/useMemberMutations";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import CopyButton from "@/components/common/CopyButton";
import { EMAIL_REGEX } from "@/utils/validation";
import { RoleSelector } from "./constants";
import { type AssignableRole, buildActivationLink } from "./helpers";
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
  const createToken = useCreateActivationToken();
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<AssignableRole>("operator");
  const [name, setName] = useState("");
  const [username, setUsername] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [emailError, setEmailError] = useState("");
  const [error, setError] = useState("");
  // When we provision a brand-new account, we surface its activation link here
  // instead of closing, so the admin can copy it in one step.
  const [activationLink, setActivationLink] = useState("");
  // Enterprise: a namespace admin who isn't a system admin can't provision an
  // account directly — the add is enqueued for a system admin to approve. We
  // confirm that here instead of closing silently.
  const [requestSubmitted, setRequestSubmitted] = useState(false);

  useResetOnOpen(open, () => {
    setEmail("");
    setRole("operator");
    setName("");
    setUsername("");
    setEmailError("");
    setError("");
    setActivationLink("");
    setRequestSubmitted(false);
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
      const trimmedName = name.trim();
      const trimmedUsername = username.trim();
      const namespace = await addMember.mutateAsync({
        path: { tenant: tenantId },
        body: {
          email: trimmedEmail,
          role,
          // Only sent when provisioning a brand-new account; ignored by the
          // server when the email already resolves to an existing user.
          ...(trimmedName ? { name: trimmedName } : {}),
          ...(trimmedUsername ? { username: trimmedUsername } : {}),
        },
      });

      const memberRow = namespace?.members?.find(
        (m) => m.email?.toLowerCase() === trimmedEmail.toLowerCase(),
      );

      // A non-admin provisioned the account: it exists but is inert until a system
      // admin approves it, and only an admin can mint its link. Confirm and stop.
      if (memberRow?.awaiting_approval) {
        setRequestSubmitted(true);
        return;
      }

      // An admin-provisioned account comes back not-confirmed and auto-approved;
      // mint its activation link right away so there's no second step.
      if (memberRow?.account_status === "not-confirmed" && memberRow.id) {
        try {
          const data = await createToken.mutateAsync({
            path: { id: memberRow.id },
          });
          setActivationLink(
            buildActivationLink(memberRow.id, data?.token ?? ""),
          );
          return;
        } catch {
          // The member exists; a link can still be copied from the members list.
          onClose();
          return;
        }
      }

      // Existing account added directly, or nothing more to do.
      onClose();
    } catch {
      setError(
        "Failed to add member. If the person has no account yet, fill in the name and username to create one.",
      );
    } finally {
      setSubmitting(false);
    }
  };

  if (requestSubmitted) {
    return (
      <Drawer
        open={open}
        onClose={onClose}
        title="Request Submitted"
        footer={
          <Button variant="primary" onClick={onClose}>
            Done
          </Button>
        }
      >
        <div className="space-y-4">
          <p className="text-sm text-text-muted">
            The account for{" "}
            <span className="font-medium text-text-primary">
              {trimmedEmail}
            </span>{" "}
            was created and is awaiting a system administrator's approval. It
            shows as "Awaiting approval" in the members list.
          </p>
          <p className="text-2xs text-text-muted">
            Once approved, you'll be able to copy the activation link from the
            members list and share it with them.
          </p>
        </div>
      </Drawer>
    );
  }

  if (activationLink) {
    return (
      <Drawer
        open={open}
        onClose={onClose}
        title="Member Provisioned"
        footer={
          <Button variant="primary" onClick={onClose}>
            Done
          </Button>
        }
      >
        <div className="space-y-4">
          <p className="text-sm text-text-muted">
            The account was created. Share this one-time activation link with{" "}
            <span className="font-medium text-text-primary">
              {trimmedEmail}
            </span>{" "}
            so they can set a password.
          </p>
          <div className="flex items-start gap-2 p-3 bg-accent-yellow/8 border border-accent-yellow/20 rounded-lg">
            <ExclamationTriangleIcon
              className="w-4 h-4 text-accent-yellow shrink-0 mt-0.5"
              strokeWidth={2}
            />
            <p className="text-2xs text-accent-yellow">
              Copy it now — it will not be shown again. You can generate a new
              one later from the members list.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex-1">
              <InputField
                id="add-member-activation-link"
                label="Activation link"
                hideLabel
                readOnly
                value={activationLink}
                onChange={() => {}}
                variant="mono"
              />
            </div>
            <CopyButton text={activationLink} size="md" showLabel />
          </div>
        </div>
      </Drawer>
    );
  }

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
          hint="An existing account is added directly; a new email is provisioned below"
          error={emailError}
        />
        <RoleSelector value={role} onChange={setRole} />

        <div className="space-y-5 border-t border-border pt-5">
          <p className="text-2xs text-text-muted">
            New account (optional) — fill these in to provision an account when
            the email has none yet. The person sets their password through an
            activation link.
          </p>
          <InputField
            id="add-member-name"
            label="Name"
            value={name}
            onChange={setName}
            placeholder="John Doe"
          />
          <InputField
            id="add-member-username"
            label="Username"
            value={username}
            onChange={setUsername}
            placeholder="john_doe"
          />
        </div>

        {error && <p className="text-2xs text-accent-red">{error}</p>}
      </form>
    </Drawer>
  );
}

export default AddMemberDrawer;
