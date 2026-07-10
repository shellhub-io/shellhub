import { useState, FormEvent } from "react";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { Card, Button } from "@shellhub/design-system/primitives";
import {
  InformationCircleIcon,
  CheckCircleIcon,
} from "@heroicons/react/24/outline";
import { useGenerateInvitationLink } from "@/hooks/useInvitationMutations";
import Drawer from "@/components/common/Drawer";
import CopyButton from "@/components/common/CopyButton";
import InputField from "@/components/common/fields/InputField";
import { getConfig } from "@/env";
import { RoleSelector } from "./constants";
import { type AssignableRole } from "./helpers";
import { LABEL } from "@/utils/styles";
import { EMAIL_REGEX } from "@/utils/validation";

interface AddMemberDrawerProps {
  open: boolean;
  onClose: () => void;
  tenantId: string;
}

function AddMemberDrawer({ open, onClose, tenantId }: AddMemberDrawerProps) {
  const generateLink = useGenerateInvitationLink();
  // Cloud also emails the invitee (it has SMTP); the admin gets the link back either way.
  const emailDelivery = getConfig().cloud;
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<AssignableRole>("operator");
  const [submitting, setSubmitting] = useState(false);
  const [emailError, setEmailError] = useState("");
  const [error, setError] = useState("");
  const [generatedLink, setGeneratedLink] = useState("");
  // Enterprise adds an existing account directly (no link). We show a confirmation
  // instead of the link card in that case.
  const [addedDirectly, setAddedDirectly] = useState(false);

  useResetOnOpen(open, () => {
    setEmail("");
    setRole("operator");
    setSubmitting(false);
    setEmailError("");
    setError("");
    setGeneratedLink("");
    setAddedDirectly(false);
  });

  const handleClose = () => {
    onClose();
  };

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
      const result = await generateLink.mutateAsync({
        path: { tenant: tenantId },
        body: { email: trimmedEmail, role },
      });
      // An empty link means the account existed and was added directly (enterprise).
      const link = result.link ?? "";
      if (link) setGeneratedLink(link);
      else setAddedDirectly(true);
    } catch (err) {
      if (isSdkError(err)) {
        switch (err.status) {
          case 400:
            setEmailError("Invalid email or role.");
            break;
          case 403:
            setError("You don't have permission to invite members.");
            break;
          case 404:
            setEmailError("No account exists for this email.");
            break;
          case 409:
            setEmailError(
              "This user is already a member or has a pending invitation.",
            );
            break;
          default:
            setError("Failed to send invitation. Please try again.");
        }
      } else {
        setError("Failed to send invitation. Please try again.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  const done = !!generatedLink || addedDirectly;

  return (
    <Drawer
      open={open}
      onClose={handleClose}
      title={
        addedDirectly ? "Member Added" : done ? "Invitation Link" : "Add Member"
      }
      subtitle={
        done ? <span className="font-mono">{trimmedEmail}</span> : undefined
      }
      footer={
        done ? (
          <Button variant="primary" onClick={handleClose}>
            Done
          </Button>
        ) : (
          <>
            <Button variant="ghost" onClick={handleClose}>
              Cancel
            </Button>
            {/* The label stays neutral: we don't know upfront whether this becomes a
                direct add or an invitation — the result screen reveals which. */}
            <Button
              variant="primary"
              onClick={() => void handleSubmit()}
              disabled={!emailValid || submitting}
              loading={submitting}
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
            <span id="invitation-link-label" className={LABEL}>
              Invitation link
            </span>
            <Card
              aria-labelledby="invitation-link-label"
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
        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
          <InputField
            id="invitation-email"
            label="Email"
            type="email"
            value={email}
            onChange={(v) => {
              setEmail(v);
              if (emailError) setEmailError("");
            }}
            placeholder="user@example.com"
            error={emailError || undefined}
          />

          <RoleSelector value={role} onChange={setRole} />

          {error && (
            <div
              role="alert"
              className="flex items-start gap-2 bg-accent-red/[0.06] border border-accent-red/20 rounded-lg px-3 py-2.5 text-xs text-accent-red"
            >
              <InformationCircleIcon className="w-4 h-4 shrink-0 mt-px" />
              <span>{error}</span>
            </div>
          )}
        </form>
      )}
    </Drawer>
  );
}

export default AddMemberDrawer;
