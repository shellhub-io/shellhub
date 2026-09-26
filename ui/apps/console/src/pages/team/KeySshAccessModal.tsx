import { useState } from "react";
import {
  ArrowRightIcon,
  BoltIcon,
  ClockIcon,
  FingerPrintIcon,
  KeyIcon,
  PlusIcon,
  ShieldCheckIcon,
  TrashIcon,
  CommandLineIcon,
  NoSymbolIcon,
} from "@heroicons/react/24/outline";
import { Link } from "react-router-dom";
import {
  Button,
  IconBadge,
  IconButton,
} from "@shellhub/design-system/primitives";
import Modal from "@/components/common/Modal";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import IdentityModal from "@/pages/ssh-identities/IdentityModal";
import { useApiKeySshIdentities } from "@/hooks/useApiKeySshIdentities";
import { useDeleteSSHIdentity } from "@/hooks/useSSHIdentityMutations";
import { useAccessPolicies } from "@/hooks/useAccessPolicies";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { formatDateShort } from "@/utils/date";
import { type ApiKey, type SshIdentity } from "@/client";
import ObjectName from "@/components/common/ObjectName";

const SECTION =
  "text-2xs font-semibold tracking-wide text-text-muted uppercase";

function Missing({
  icon,
  title,
  children,
  action,
}: {
  icon: React.ReactNode;
  title: string;
  children: React.ReactNode;
  action: React.ReactNode;
}) {
  return (
    <div className="flex flex-col items-center text-center gap-3 py-7 px-4 border border-dashed border-border rounded-xl">
      <IconBadge color="neutral" size="md">
        {icon}
      </IconBadge>
      <div className="space-y-1">
        <p className="text-sm text-text-primary">{title}</p>
        <p className="text-xs text-text-muted max-w-[17rem]">{children}</p>
      </div>
      {action}
    </div>
  );
}

function Trait({
  icon,
  children,
}: {
  icon: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <span className="inline-flex items-center gap-1 text-2xs text-text-muted">
      {icon}
      {children}
    </span>
  );
}

/**
 * SSH access for one API key: the credentials it connects with, and the policies that say
 * where. Both halves are here because either one alone reaches nothing, and until now they
 * lived on two pages that never mentioned each other.
 */
function KeySshAccessModal({
  open,
  apiKey,
  onClose,
}: {
  open: boolean;
  apiKey: ApiKey | null;
  onClose: () => void;
}) {
  const keyName = apiKey?.name ?? "";
  const { identities, isLoading } = useApiKeySshIdentities(keyName, open);
  const { policies } = useAccessPolicies();
  const { tenant } = useAuthStore();
  const { namespace } = useNamespace(tenant ?? "");
  const isIdentityMode = namespace?.settings?.ssh_access_mode === "identity";
  const deleteIdentity = useDeleteSSHIdentity();
  const invalidate = useInvalidateByIds("listApiKeySshIdentities");
  const [addOpen, setAddOpen] = useState(false);
  const [revokeTarget, setRevokeTarget] = useState<SshIdentity | null>(null);

  const reaching = policies.filter(
    (p) => p.subject.type === "api-key" && p.subject.value === apiKey?.id,
  );

  const closeAdd = () => {
    setAddOpen(false);
    void invalidate();
  };

  const confirmRevoke = async () => {
    if (!revokeTarget) return;
    await deleteIdentity.mutateAsync({ path: { id: revokeTarget.id } });
    setRevokeTarget(null);
    void invalidate();
  };

  return (
    <>
      <Modal
        size="lg"
        open={open}
        onClose={onClose}
        icon={<CommandLineIcon />}
        title="SSH access"
        description={
          <>
            The keys and policies that let <ObjectName>{keyName}</ObjectName>{" "}
            connect over SSH.
          </>
        }
        footer={
          <Button variant="primary" onClick={onClose}>
            Done
          </Button>
        }
      >
        <div className="space-y-6">
          <section className="space-y-2.5">
            <div className="flex items-center justify-between">
              <span className={SECTION}>Credentials</span>
              {identities.length > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setAddOpen(true)}
                  icon={<PlusIcon className="w-3.5 h-3.5" strokeWidth={2} />}
                >
                  Add
                </Button>
              )}
            </div>

            {isLoading ? (
              <div className="h-16 rounded-xl bg-surface animate-pulse" />
            ) : identities.length === 0 ? (
              <Missing
                icon={<KeyIcon className="w-5 h-5" strokeWidth={1.5} />}
                title="Nothing to connect with"
                action={
                  <Button
                    variant="primary"
                    onClick={() => setAddOpen(true)}
                    icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
                  >
                    Add SSH key
                  </Button>
                }
              >
                Upload the public key the automation connects with. The private
                half never leaves it.
              </Missing>
            ) : (
              <ul className="space-y-2">
                {identities.map((identity) => (
                  <li
                    key={identity.id}
                    className="group flex items-start gap-3 bg-card border border-border rounded-xl px-3.5 py-3 hover:border-border-light transition-colors"
                  >
                    <IconBadge color="primary" size="sm">
                      <KeyIcon className="w-4 h-4" strokeWidth={1.5} />
                    </IconBadge>

                    <div className="min-w-0 flex-1 space-y-1.5">
                      <p className="text-sm font-medium text-text-primary truncate">
                        {identity.name}
                      </p>
                      <p className="flex items-center gap-1 text-2xs font-mono text-text-muted truncate">
                        <FingerPrintIcon className="w-3 h-3 shrink-0" />
                        {identity.fingerprint}
                      </p>
                      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
                        {identity.single_use && (
                          <Trait icon={<BoltIcon className="w-3 h-3" />}>
                            single use
                          </Trait>
                        )}
                        <Trait icon={<ClockIcon className="w-3 h-3" />}>
                          {identity.expires_at
                            ? `expires ${formatDateShort(identity.expires_at)}`
                            : "never expires"}
                        </Trait>
                        {identity.last_used_at && (
                          <Trait icon={<ArrowRightIcon className="w-3 h-3" />}>
                            used {formatDateShort(identity.last_used_at)}
                          </Trait>
                        )}
                      </div>
                    </div>

                    <IconButton
                      variant="danger"
                      title="Revoke"
                      aria-label={`Revoke ${identity.name}`}
                      onClick={() => setRevokeTarget(identity)}
                    >
                      <TrashIcon className="w-4 h-4" />
                    </IconButton>
                  </li>
                ))}
              </ul>
            )}
          </section>

          <section className="space-y-2.5">
            <span className={SECTION}>Where it reaches</span>

            {reaching.length === 0 ? (
              <p className="text-xs text-text-muted">
                No policy names this key yet.
                {isIdentityMode && (
                  <>
                    {" "}
                    <Link
                      to="/access-policies"
                      className="text-primary hover:underline"
                    >
                      Write one
                    </Link>
                    .
                  </>
                )}
              </p>
            ) : (
              <ul className="space-y-1.5">
                {reaching.map((policy) => (
                  <li
                    key={policy.id}
                    className="flex items-center gap-2.5 bg-card border border-border rounded-lg px-3 py-2"
                  >
                    <ShieldCheckIcon
                      className={
                        policy.action === "allow"
                          ? "w-4 h-4 text-accent-green shrink-0"
                          : "w-4 h-4 text-accent-red shrink-0"
                      }
                      strokeWidth={1.5}
                    />
                    <span className="text-xs text-text-primary truncate flex-1">
                      {policy.name}
                    </span>
                    <span className="text-2xs text-text-muted whitespace-nowrap">
                      {policy.logins.join(", ") || "any login"}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </section>
        </div>
      </Modal>

      <IdentityModal
        open={addOpen}
        editIdentity={null}
        apiKeyName={keyName}
        onClose={closeAdd}
      />

      <ConfirmDialog
        open={!!revokeTarget}
        icon={<NoSymbolIcon />}
        title="Revoke key"
        description={
          <>
            <ObjectName>{keyName}</ObjectName> can no longer connect with it.
            Anything using it fails on its next run.
          </>
        }
        confirmLabel="Revoke key"
        onConfirm={() => void confirmRevoke()}
        onClose={() => setRevokeTarget(null)}
      />
    </>
  );
}

export default KeySshAccessModal;
