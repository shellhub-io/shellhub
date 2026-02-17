import { useEffect, useState } from "react";
import {
  PlusIcon,
  UserGroupIcon,
  PencilSquareIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { useMembersStore } from "../../stores/membersStore";
import { useAuthStore } from "../../stores/authStore";
import { type NamespaceMember } from "../../types/namespace";
import ConfirmDialog from "../../components/common/ConfirmDialog";
import { RoleBadge } from "./constants";
import { initials } from "./helpers";
import AddMemberDrawer from "./AddMemberDrawer";
import EditMemberDrawer from "./EditMemberDrawer";
import { TH } from "../../utils/styles";

/* --- Members Tab --- */

function MembersTab({ tenantId }: { tenantId: string }) {
  const {
    members,
    loading: membersLoading,
    fetch: fetchMembers,
    remove: removeMember,
  } = useMembersStore();
  const currentUserEmail = useAuthStore((s) => s.email);
  const [addOpen, setAddOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<NamespaceMember | null>(null);
  const [removeTarget, setRemoveTarget] = useState<NamespaceMember | null>(
    null,
  );

  useEffect(() => {
    fetchMembers(tenantId);
  }, [tenantId, fetchMembers]);

  // Filter out owner, sort alphabetically
  const sorted = members
    .filter((m) => m.role !== "owner")
    .sort((a, b) => a.email.localeCompare(b.email));

  return (
    <div className="animate-fade-in">
      <div className="flex items-center justify-between mb-5">
        <p className="text-sm text-text-muted">
          {sorted.length} member{sorted.length !== 1 ? "s" : ""}
        </p>
        <button
          onClick={() => setAddOpen(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
        >
          <PlusIcon className="w-4 h-4" strokeWidth={2} />
          Add Member
        </button>
      </div>

      {membersLoading ? (
        <div className="flex items-center justify-center py-16">
          <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
        </div>
      ) : sorted.length === 0 ? (
        <div className="text-center py-16">
          <UserGroupIcon
            className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
            strokeWidth={1}
          />
          <p className="text-sm text-text-muted">No members yet</p>
          <p className="text-2xs text-text-muted/60 mt-1">
            Add members to collaborate in this namespace
          </p>
        </div>
      ) : (
        <div className="bg-card border border-border rounded-xl overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border">
                <th className={TH}>Member</th>
                <th className={TH}>Role</th>
                <th className={`${TH} !text-right w-24`}>Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {sorted.map((m) => {
                const isSelf = m.email === currentUserEmail;
                return (
                  <tr
                    key={m.id}
                    className="group transition-colors hover:bg-hover-subtle"
                  >
                    <td className="px-4 py-3.5">
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
                    </td>
                    <td className="px-4 py-3.5">
                      <RoleBadge role={m.role} />
                    </td>
                    <td className="px-4 py-3.5 text-right">
                      {!isSelf && (
                        <div className="flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                          <button
                            onClick={() => setEditTarget(m)}
                            className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
                            title="Edit role"
                          >
                            <PencilSquareIcon className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => setRemoveTarget(m)}
                            className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/5 transition-colors"
                            title="Remove"
                          >
                            <TrashIcon className="w-4 h-4" />
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

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
        onClose={() => setRemoveTarget(null)}
        onConfirm={async () => {
          await removeMember(tenantId, removeTarget!.id);
          setRemoveTarget(null);
        }}
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
      />
    </div>
  );
}

export default MembersTab;
