import { useRef, useState, FormEvent } from "react";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import {
  EnvelopeIcon,
  LinkIcon,
  CheckCircleIcon,
  InformationCircleIcon,
} from "@heroicons/react/24/outline";
import {
  useGenerateInvitationLink,
  useSendInvitationEmail,
} from "@/hooks/useInvitationMutations";
import Drawer from "@/components/common/Drawer";
import CopyButton from "@/components/common/CopyButton";
import { LABEL, INPUT } from "@/utils/styles";
import { RoleSelector } from "./constants";
import { type AssignableRole } from "./helpers";

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

interface InvitationDrawerProps {
  open: boolean;
  onClose: () => void;
  tenantId: string;
  /** Fires when the drawer is closed *after* a successful invitation was
   *  created during this session. Typically used to navigate the user to the
   *  Invitations tab so they can see their freshly-created pending record. */
  onInvitationSent?: () => void;
}

function InvitationDrawer({
  open,
  onClose,
  tenantId,
  onInvitationSent,
}: InvitationDrawerProps) {
  const sendEmail = useSendInvitationEmail();
  const generateLink = useGenerateInvitationLink();
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<AssignableRole>("operator");
  const [wantLink, setWantLink] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [emailError, setEmailError] = useState("");
  const [error, setError] = useState("");
  const [generatedLink, setGeneratedLink] = useState("");

  // Ref (not state) so flipping it doesn't trigger re-renders and so the
  // onInvitationSent callback reads the latest value at close time.
  const sentRef = useRef(false);

  useResetOnOpen(open, () => {
    setEmail("");
    setRole("operator");
    setWantLink(false);
    setSubmitting(false);
    setEmailError("");
    setError("");
    setGeneratedLink("");
    sentRef.current = false;
  });

  const handleClose = () => {
    if (sentRef.current) onInvitationSent?.();
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
      const body = { email: trimmedEmail, role };
      let link = "";
      if (wantLink) {
        const result = await generateLink.mutateAsync({
          path: { tenant: tenantId },
          body,
        });
        link = result.link ?? "";
      } else {
        await sendEmail.mutateAsync({ path: { tenant: tenantId }, body });
      }
      sentRef.current = true;
      if (wantLink) setGeneratedLink(link);
      else handleClose();
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

  const done = !!generatedLink;

  return (
    <Drawer
      open={open}
      onClose={handleClose}
      title={done ? "Invitation Link" : "Invite Member"}
      subtitle={
        done ? <span className="font-mono">{trimmedEmail}</span> : undefined
      }
      footer={
        done ? (
          <button
            onClick={handleClose}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
          >
            Done
          </button>
        ) : (
          <>
            <button
              type="button"
              onClick={handleClose}
              className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={() => void handleSubmit()}
              disabled={!emailValid || submitting}
              className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
            >
              {submitting ? (
                <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              ) : wantLink ? (
                <LinkIcon className="w-4 h-4" strokeWidth={2} />
              ) : (
                <EnvelopeIcon className="w-4 h-4" strokeWidth={2} />
              )}
              {wantLink ? "Generate Link" : "Send Invite"}
            </button>
          </>
        )
      }
    >
      {done ? (
        <div className="space-y-5">
          <div className="flex items-start gap-3 bg-accent-green/[0.06] border border-accent-green/20 rounded-xl px-4 py-3.5">
            <CheckCircleIcon className="w-5 h-5 text-accent-green shrink-0 mt-0.5" />
            <div>
              <p className="text-sm font-medium text-text-primary">
                Invitation Ready
              </p>
              <p className="text-2xs text-text-muted mt-0.5 leading-relaxed">
                Share this link with the recipient. It is only valid for{" "}
                <span className="font-mono text-text-secondary">
                  {trimmedEmail}
                </span>{" "}
                and has a limited lifetime.
              </p>
            </div>
          </div>
          <div>
            <label className={LABEL}>Invitation link</label>
            <div className="flex items-center gap-2 bg-card border border-border rounded-lg px-3.5 py-2.5">
              <code className="flex-1 text-xs font-mono text-accent-cyan break-all select-all">
                {generatedLink}
              </code>
              <CopyButton text={generatedLink} size="md" />
            </div>
          </div>
        </div>
      ) : (
        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
          <div>
            <label className={LABEL} htmlFor="invitation-email">
              Email
            </label>
            <input
              id="invitation-email"
              type="email"
              value={email}
              onChange={(e) => {
                setEmail(e.target.value);
                if (emailError) setEmailError("");
              }}
              placeholder="user@example.com"
              autoFocus={open}
              className={`${INPUT} ${emailError ? "border-accent-red/60 focus:border-accent-red/60 focus:ring-accent-red/20" : ""}`}
              aria-invalid={!!emailError}
              aria-describedby={
                emailError ? "invitation-email-error" : undefined
              }
            />
            {emailError ? (
              <p
                id="invitation-email-error"
                className="mt-1.5 text-2xs text-accent-red"
              >
                {emailError}
              </p>
            ) : (
              <p className="mt-1.5 text-2xs text-text-muted">
                If no account matches this email, we'll send a sign-up link.
              </p>
            )}
          </div>

          <div>
            <label className={LABEL}>Role</label>
            <RoleSelector value={role} onChange={setRole} />
          </div>

          <label className="flex items-start gap-2.5 px-3 py-2.5 bg-card border border-border rounded-lg cursor-pointer hover:border-border-light transition-colors">
            <input
              type="checkbox"
              checked={wantLink}
              onChange={(e) => setWantLink(e.target.checked)}
              className="mt-0.5 h-3.5 w-3.5 rounded border-border bg-background text-primary focus:ring-1 focus:ring-primary/30 cursor-pointer"
            />
            <span className="flex-1">
              <span className="block text-xs font-medium text-text-secondary">
                Get the invite link instead of sending an email
              </span>
              <span className="block text-2xs text-text-muted mt-0.5">
                Useful when you want to share the invitation through another
                channel.
              </span>
            </span>
          </label>

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

export default InvitationDrawer;
