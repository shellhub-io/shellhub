import { useState } from "react";
import {
  PlusIcon,
  UserGroupIcon,
  PencilSquareIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace, type NamespaceMember } from "@/hooks/useNamespaces";
import { useRemoveMember } from "@/hooks/useMemberMutations";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import DataTable, { type Column } from "@/components/common/DataTable";
import { RoleBadge } from "./constants";
import { initials } from "./helpers";
import AddMemberDrawer from "./AddMemberDrawer";
import EditMemberDrawer from "./EditMemberDrawer";
import RestrictedAction from "@/components/common/RestrictedAction";

function MembersTab({ tenantId }: { tenantId: string }) {
  const { namespace, isLoading: membersLoading } = useNamespace(tenantId);
  const removeMember = useRemoveMember();
  const currentUserEmail = useAuthStore((s) => s.email);
  const [addOpen, setAddOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<NamespaceMember | null>(null);
  const [removeTarget, setRemoveTarget] = useState<NamespaceMember | null>(
    null,
  );
  const [removeError, setRemoveError] = useState<string | null>(null);

  const closeRemove = () => {
    setRemoveError(null);
    setRemoveTarget(null);
  };

  const confirmRemove = async () => {
    if (!removeTarget) return;
    setRemoveError(null);
    try {
      await removeMember.mutateAsync({
        path: { tenant: tenantId, uid: removeTarget.id },
      });
      closeRemove();
    } catch (err) {
      setRemoveError(
        err instanceof Error ? err.message : "Failed to remove member.",
      );
    }
  };

  const members = (namespace?.members ?? []).filter(
    (m): m is NamespaceMember => !!m.id && !!m.role && !!m.email,
  );

  const sorted = members
    .filter((m) => m.role !== "owner")
    .sort((a, b) => a.email.localeCompare(b.email));

  const columns: Column<NamespaceMember>[] = [
    {
      key: "member",
      header: "Member",
      render: (m) => {
        const isSelf = m.email === currentUserEmail;
        return (
          <div className="flex items-center gap-3">
            <span className="w-8 h-8 rounded-lg flex items-center justify-center text-xs font-bold font-mono shrink-0 bg-card border border-border text-text-muted">
              {initials(m.email)}
            </span>
            <div>
              <span className="text-sm font-medium text-text-primary">
                {m.email}
              </span>
              {isSelf && (
                <span className="ml-2 text-2xs text-text-muted font-mono">
                  (you)
                </span>
              )}
            </div>
          </div>
        );
      },
    },
    {
      key: "role",
      header: "Role",
      render: (m) => <RoleBadge role={m.role} />,
    },
    {
      key: "actions",
      header: "Actions",
      headerClassName: "text-right",
      render: (m) => {
        const isSelf = m.email === currentUserEmail;
        if (isSelf) return null;
        return (
          <div className="flex items-center justify-end gap-1">
            <RestrictedAction action="namespace:editMember">
              <button
                onClick={() => setEditTarget(m)}
                className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
                title="Edit role"
              >
                <PencilSquareIcon className="w-4 h-4" />
              </button>
            </RestrictedAction>
            <RestrictedAction action="namespace:removeMember">
              <button
                onClick={() => setRemoveTarget(m)}
                className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/5 transition-colors"
                title="Remove"
              >
                <TrashIcon className="w-4 h-4" />
              </button>
            </RestrictedAction>
          </div>
        );
      },
    },
  ];

  return (
    <div className="animate-fade-in">
      <div className="flex items-center justify-between mb-5">
        <p className="text-sm text-text-muted">
          {sorted.length} member
          {sorted.length !== 1 ? "s" : ""}
        </p>
        <RestrictedAction action="namespace:addMember">
          <button
            onClick={() => setAddOpen(true)}
            className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
          >
            <PlusIcon className="w-4 h-4" strokeWidth={2} />
            Add Member
          </button>
        </RestrictedAction>
      </div>

      <DataTable
        columns={columns}
        data={sorted}
        rowKey={(m) => m.id}
        isLoading={membersLoading}
        loadingMessage="Loading members..."
        emptyState={
          <div className="text-center">
            <UserGroupIcon
              className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
              strokeWidth={1}
            />
            <p className="text-sm text-text-muted">No members yet</p>
            <p className="text-2xs text-text-muted/60 mt-1">
              Add members to collaborate in this namespace
            </p>
          </div>
        }
      />

      <AddMemberDrawer
        open={addOpen}
        onClose={() => setAddOpen(false)}
        tenantId={tenantId}
      />
      <EditMemberDrawer
        open={!!editTarget}
        onClose={() => setEditTarget(null)}
        tenantId={tenantId}
        member={editTarget}
      />
      <ConfirmDialog
        open={!!removeTarget}
        onClose={closeRemove}
        onConfirm={confirmRemove}
        title="Remove Member"
        description={
          <>
            Are you sure you want to remove{" "}
            <span className="font-medium text-text-primary">
              {removeTarget?.email}
            </span>{" "}
            from this namespace?
          </>
        }
        confirmLabel="Remove"
      >
        {removeError && (
          <p className="text-xs text-accent-red">{removeError}</p>
        )}
      </ConfirmDialog>
    </div>
  );
}

export default MembersTab;
