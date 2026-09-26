import { ReactNode, useState } from "react";
import {
  ArrowRightStartOnRectangleIcon,
  CheckIcon,
  PencilSquareIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useNamespace, type Namespace } from "@/hooks/useNamespaces";
import { useEditNamespace } from "@/hooks/useNamespaceMutations";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useStats } from "@/hooks/useStats";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { useAuthStore } from "@/stores/authStore";
import { formatDateShort } from "@/utils/date";
import { getInitials } from "@/utils/string";
import { RoleBadge } from "@/pages/team/constants";
import { roleSummary } from "@/pages/team/helpers";
import CopyButton from "@/components/common/CopyButton";
import FormModal from "@/components/common/FormModal";
import FormInputField from "@/components/common/fields/rhf/FormInputField";
import PageLoader from "@/components/common/PageLoader";
import SettingsSection from "@/components/settings/SettingsSection";
import {
  NAMESPACE_NAME_HINT,
  NAMESPACE_NAME_MAX_LENGTH,
} from "@/utils/validation";
import {
  namespaceRenameSchema,
  buildNamespaceRenameDefaults,
  type NamespaceRenameFormValues,
} from "./namespaceRenameSchema";
import { DeleteNamespaceZone, LeaveDialog } from "./NamespaceDangerZone";

function RenameModal({
  open,
  onClose,
  currentName,
  tenantId,
}: {
  open: boolean;
  onClose: () => void;
  currentName: string;
  tenantId: string;
}) {
  const editNs = useEditNamespace();

  const form = useDrawerForm(
    open,
    namespaceRenameSchema,
    buildNamespaceRenameDefaults(currentName),
  );
  const { control, setValue, setError, clearErrors } = form;

  const onValid = async (values: NamespaceRenameFormValues) => {
    clearErrors("root");
    try {
      await editNs.mutateAsync({
        path: { tenant: tenantId },
        body: { name: values.name },
      });
      onClose();
    } catch {
      setError("root", {
        message:
          "Couldn't rename the namespace. The name may already be taken.",
      });
    }
  };

  return (
    <FormModal
      size="sm"
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      icon={<PencilSquareIcon />}
      title="Rename namespace"
      description="Every device's SSHID in this namespace changes with it."
      submitLabel="Save"
      requireDirty
      submitIcon={<CheckIcon className="w-4 h-4" strokeWidth={2} />}
    >
      <FormInputField
        name="name"
        control={control}
        id="edit-ns-name"
        label="Namespace name"
        placeholder="my-namespace"
        hint={NAMESPACE_NAME_HINT}
        maxLength={NAMESPACE_NAME_MAX_LENGTH}
        onValueChange={(v) => {
          setValue("name", v.toLowerCase(), {
            shouldDirty: true,
            shouldValidate: true,
          });
          clearErrors("root");
        }}
      />
    </FormModal>
  );
}

function Stat({ label, value, hint }: { label: string; value: ReactNode; hint?: ReactNode }) {
  return (
    <div role="group" aria-label={label} className="min-w-0 px-5 py-3.5">
      <p className="text-2xs font-mono uppercase tracking-label text-text-muted">
        {label}
      </p>
      <p className="mt-1 text-sm font-medium text-text-primary truncate">
        {value}
        {hint && (
          <span className="ml-1.5 text-xs font-normal text-text-muted">
            {hint}
          </span>
        )}
      </p>
    </div>
  );
}

function NamespaceCard({
  ns,
  onRename,
  onLeave,
}: {
  ns: Namespace;
  onRename?: () => void;
  onLeave?: () => void;
}) {
  const userId = useAuthStore((st) => st.userId);
  const role = useAuthStore((st) => st.role);
  const { stats } = useStats();
  const owner =
    ns.owner === userId
      ? "you"
      : (ns.members.find((m) => m.id === ns.owner)?.email ?? "another user");
  const team = ns.type === "team";
  const unlimited = ns.max_devices < 0;

  return (
    <div className="rounded-xl border border-border bg-card overflow-hidden">
      <div className="grid grid-cols-[auto_1fr] sm:grid-cols-[auto_1fr_auto] items-center gap-x-4 gap-y-3 px-5 py-5">
        <span
          aria-hidden="true"
          className="grid place-items-center w-12 h-12 shrink-0 rounded-xl border border-primary/25 bg-primary/10 text-base font-semibold text-primary"
        >
          {getInitials(ns.name)}
        </span>
        <div className="min-w-0">
          <p className="flex items-center gap-2.5 min-w-0">
            <span className="text-lg font-semibold font-mono text-text-primary truncate">
              {ns.name}
            </span>
            <span
              className={cn(
                "inline-flex items-center px-2 py-0.5 text-2xs font-medium rounded-md border",
                team
                  ? "bg-primary/10 text-primary border-primary/20"
                  : "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20",
              )}
            >
              {team ? "Team" : "Personal"}
            </span>
          </p>
          <p className="mt-1 flex flex-wrap items-center gap-x-1.5 gap-y-1 text-xs text-text-muted">
            {role && (
              <span title={roleSummary(role)} className="inline-flex">
                <RoleBadge role={role} />
              </span>
            )}
            {role !== "owner" && <span>owned by {owner} ·</span>}
            <span>since {formatDateShort(ns.created_at)}</span>
          </p>
        </div>
        {(onRename || onLeave) && (
          <div className="col-start-2 sm:col-start-3 sm:row-start-1 justify-self-start flex items-center gap-2">
            {onRename && (
              <Button size="sm" variant="secondary" onClick={onRename}>
                Rename
              </Button>
            )}
            {onLeave && (
              <Button
                size="sm"
                variant="dangerSoft"
                icon={<ArrowRightStartOnRectangleIcon className="w-4 h-4" />}
                onClick={onLeave}
              >
                Leave
              </Button>
            )}
          </div>
        )}
      </div>

      <div className="grid grid-cols-3 divide-x divide-border border-t border-border">
        <Stat
          label="Devices"
          value={ns.devices_accepted_count}
          hint={unlimited ? "no limit" : `of ${ns.max_devices}`}
        />
        <Stat label="Online" value={stats?.online_devices ?? "–"} />
        <Stat label="Members" value={ns.members.length} />
      </div>

      <div
        role="group"
        aria-label="Tenant ID"
        className="flex flex-wrap items-center gap-x-3 gap-y-1 pl-5 pr-2 py-2 border-t border-border bg-surface/50"
      >
        <span className="text-2xs font-mono uppercase tracking-label text-text-muted">
          Tenant ID
        </span>
        <span className="flex items-center gap-1 min-w-0">
          <code className="min-w-0 text-[11px] sm:text-xs font-mono text-accent-cyan select-all break-all">
            {ns.tenant_id}
          </code>
          <CopyButton text={ns.tenant_id} size="md" />
        </span>
      </div>
    </div>
  );
}

/**
 * The general settings: the namespace at a glance, with the role held in it and its tenant ID, and
 * the way out of it: deleting it for the owner, leaving it for everyone else.
 */
export default function GeneralSettings() {
  const { tenant: tenantId } = useAuthStore();
  const { namespace: ns } = useNamespace(tenantId ?? "");
  const canRename = useHasPermission("namespace:rename");
  const canDelete = useHasPermission("namespace:delete");
  const [renameOpen, setRenameOpen] = useState(false);
  const [leaveOpen, setLeaveOpen] = useState(false);

  if (!ns) return <PageLoader label="Loading settings" padding="lg" />;

  return (
    <SettingsSection
      title="General"
      description="The namespace's name and the identifiers that refer to it."
    >
      <NamespaceCard
        ns={ns}
        onRename={canRename ? () => setRenameOpen(true) : undefined}
        onLeave={canDelete ? undefined : () => setLeaveOpen(true)}
      />

      <DeleteNamespaceZone ns={ns} />

      {leaveOpen && (
        <LeaveDialog
          tenantId={ns.tenant_id}
          onClose={() => setLeaveOpen(false)}
        />
      )}
      <RenameModal
        open={renameOpen}
        onClose={() => setRenameOpen(false)}
        currentName={ns.name}
        tenantId={ns.tenant_id}
      />
    </SettingsSection>
  );
}
