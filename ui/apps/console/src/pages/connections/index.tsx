import { useMemo, useState } from "react";
import {
  ServerStackIcon,
  PlusIcon,
  ChevronDoubleRightIcon,
  TrashIcon,
  PencilSquareIcon,
  ArrowTopRightOnSquareIcon,
  UserIcon,
  UsersIcon,
  KeyIcon,
  LockClosedIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import PageHeader from "@/components/common/PageHeader";
import Alert from "@/components/common/Alert";
import DeviceChip from "@/components/common/DeviceChip";
import OnlineDot from "@/components/common/OnlineDot";
import ConnectDrawer from "@/components/ConnectDrawer";
import HostKeyModal from "@/components/HostKeyModal";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import Pagination from "@/components/common/Pagination";
import { buildSshid } from "@/utils/sshid";
import { useConnections, useConnectionStatus } from "@/hooks/useConnections";
import {
  useTeamConnections,
  useTeamConnectionStatus,
  useTeamConnectionPrefs,
} from "@/hooks/useTeamConnections";
import { useDevice } from "@/hooks/useDevice";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import { useDeleteConnection } from "@/hooks/useConnectionMutations";
import { useDeleteTeamConnection } from "@/hooks/useTeamConnectionMutations";
import { getConfig } from "@/env";
import type { Connection, TeamConnection } from "@/client";

type Scope = "personal" | "team";

// Both scopes map into this shape and render one table. authCell is a node
// because auth differs by scope: personal auth lives on the record, team auth is
// the caller's own per-user pref.
interface RowProps {
  scope: Scope;
  online: boolean | null;
  label: string;
  kind: string;
  host: string;
  port: number;
  deviceUid: string;
  authCell: React.ReactNode;
  canManage: boolean;
  onConnect: () => void;
  onEdit: () => void;
  onDelete: () => void;
  onShowHostKey?: () => void;
}

function ScopeBadge({ scope }: { scope: Scope }) {
  if (scope === "team") {
    return (
      <span className="inline-flex items-center gap-1 text-2xs font-medium text-text-secondary">
        <UsersIcon className="w-3 h-3" strokeWidth={2} />
        Team
      </span>
    );
  }

  return (
    <span className="inline-flex items-center gap-1 text-2xs font-medium text-text-muted">
      <UserIcon className="w-3 h-3" strokeWidth={2} />
      Personal
    </span>
  );
}

function RowView({
  scope,
  online,
  label,
  kind,
  host,
  port,
  deviceUid,
  authCell,
  canManage,
  onConnect,
  onEdit,
  onDelete,
  onShowHostKey,
  showScope,
}: RowProps & { showScope: boolean }) {
  const { device } = useDevice(deviceUid);

  return (
    <tr
      onClick={online === false ? undefined : onConnect}
      className={`border-b border-border last:border-0 hover:bg-hover-subtle ${
        online === false ? "" : "cursor-pointer"
      }`}
    >
      <td className="px-4 py-3 w-10">
        {online === null ? (
          <span
            className="inline-block w-2 h-2 rounded-full bg-text-muted/40 animate-pulse"
            title="Checking..."
          />
        ) : (
          <OnlineDot online={online} />
        )}
      </td>
      <td className="px-4 py-3 font-medium text-text-primary">{label}</td>
      <td className="px-4 py-3">
        {kind === "device" ? (
          <DeviceChip
            uid={deviceUid}
            name={device?.name ?? `${deviceUid.slice(0, 8)}…`}
            osId={device?.info?.id}
            onClick={(e) => e.stopPropagation()}
          />
        ) : (
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onShowHostKey?.();
            }}
            className="group/chip inline-flex items-center gap-1.5 px-2 py-1 bg-surface border border-border rounded-md text-xs font-mono font-medium text-text-secondary hover:text-text-primary hover:border-primary/40 hover:bg-primary/5 transition-all duration-150"
            title={`${host}:${port} · view host key`}
          >
            <ArrowTopRightOnSquareIcon
              className="w-3 h-3 text-text-muted group-hover/chip:text-primary shrink-0 transition-colors"
              strokeWidth={2}
            />
            <span className="truncate max-w-[32ch]">
              {host}:{port}
            </span>
          </button>
        )}
      </td>
      <td className="px-4 py-3">
        <div className="flex flex-col items-start gap-1">
          {showScope && <ScopeBadge scope={scope} />}
          <span className="text-text-secondary">{authCell}</span>
        </div>
      </td>
      <td className="px-4 py-3">
        <div className="flex items-center justify-end gap-1.5">
          {canManage && (
            <>
              <IconButton
                variant="primary"
                title="Edit"
                aria-label="Edit connection"
                onClick={(e) => {
                  e.stopPropagation();
                  onEdit();
                }}
              >
                <PencilSquareIcon className="w-4 h-4" strokeWidth={2} />
              </IconButton>
              <IconButton
                variant="danger"
                title="Delete"
                aria-label="Delete connection"
                onClick={(e) => {
                  e.stopPropagation();
                  onDelete();
                }}
              >
                <TrashIcon className="w-4 h-4" strokeWidth={2} />
              </IconButton>
            </>
          )}
          <div className="ml-2 w-24 flex justify-end">
            {online === false ? (
              <span className="text-2xs text-text-muted/50 font-mono">
                {kind === "external" ? "Unreachable" : "Offline"}
              </span>
            ) : (
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  onConnect();
                }}
                className="inline-flex items-center gap-1 px-2.5 py-1 bg-accent-green/10 text-accent-green text-2xs font-semibold rounded-md hover:bg-accent-green/20 border border-accent-green/20 transition-all"
              >
                <ChevronDoubleRightIcon className="w-3 h-3" strokeWidth={2} />
                Connect
              </button>
            )}
          </div>
        </div>
      </td>
    </tr>
  );
}

function AuthSummary({
  username,
  authMethod,
}: {
  username: string;
  authMethod: string;
}) {
  if (!username && !authMethod)
    return <span className="text-text-muted">—</span>;

  return (
    <span className="inline-flex items-center gap-1.5">
      {authMethod === "key" ? (
        <KeyIcon className="w-3.5 h-3.5 text-text-muted shrink-0" title="Key" />
      ) : authMethod === "password" ? (
        <LockClosedIcon
          className="w-3.5 h-3.5 text-text-muted shrink-0"
          title="Password"
        />
      ) : null}
      <span className="text-text-primary font-mono text-xs">
        {username || "—"}
      </span>
    </span>
  );
}

function PersonalRow({
  connection,
  showScope,
  onConnect,
  onEdit,
  onDelete,
  onShowHostKey,
}: {
  connection: Connection;
  showScope: boolean;
  onConnect: () => void;
  onEdit: () => void;
  onDelete: () => void;
  onShowHostKey: () => void;
}) {
  const { online } = useConnectionStatus(connection.id);

  return (
    <RowView
      scope="personal"
      online={online}
      label={connection.label}
      kind={connection.kind}
      host={connection.host}
      port={connection.port}
      deviceUid={connection.device_uid}
      authCell={
        <AuthSummary
          username={connection.username}
          authMethod={connection.auth_method}
        />
      }
      canManage
      showScope={showScope}
      onConnect={onConnect}
      onEdit={onEdit}
      onDelete={onDelete}
      onShowHostKey={onShowHostKey}
    />
  );
}

function TeamRow({
  connection,
  canManage,
  showScope,
  onConnect,
  onEdit,
  onDelete,
  onShowHostKey,
}: {
  connection: TeamConnection;
  canManage: boolean;
  showScope: boolean;
  onConnect: () => void;
  onEdit: () => void;
  onDelete: () => void;
  onShowHostKey: () => void;
}) {
  const { online } = useTeamConnectionStatus(connection.id);
  // The target is shared, the auth is the caller's own.
  const { prefs } = useTeamConnectionPrefs(connection.id);

  const authCell =
    prefs && (prefs.username || prefs.auth_method) ? (
      <AuthSummary username={prefs.username} authMethod={prefs.auth_method} />
    ) : (
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation();
          onEdit();
        }}
        className="inline-flex items-center gap-1 text-2xs font-semibold px-2 py-0.5 rounded-md text-accent-yellow bg-accent-yellow/10 border border-accent-yellow/30 hover:bg-accent-yellow/20"
        title="Set your username and key for this connection"
      >
        <ExclamationTriangleIcon className="w-3 h-3" strokeWidth={2} />
        Set up auth
      </button>
    );

  return (
    <RowView
      scope="team"
      online={online}
      label={connection.label}
      kind={connection.kind}
      host={connection.host}
      port={connection.port}
      deviceUid={connection.device_uid}
      authCell={authCell}
      canManage={canManage}
      showScope={showScope}
      onConnect={onConnect}
      onEdit={onEdit}
      onDelete={onDelete}
      onShowHostKey={onShowHostKey}
    />
  );
}

type DrawerState =
  | { mode: "create" }
  | { mode: "edit" | "connect"; scope: Scope; connection: Connection };

// Maps a team connection into the shared Connection shape. Auth fields stay
// empty: team auth is the caller's own per-user pref, resolved in the drawer,
// never on the record.
function teamToConnection(t: TeamConnection): Connection {
  return {
    id: t.id,
    tenant_id: t.tenant_id,
    owner_id: t.created_by,
    label: t.label,
    kind: t.kind,
    host: t.host ?? "",
    port: t.port || 22,
    device_uid: t.device_uid,
    username: "",
    auth_method: "",
    key_fingerprint: "",
    created_at: t.created_at,
    updated_at: t.updated_at,
  };
}

export default function Connections() {
  const [page, setPage] = useState(1);
  const perPage = 100;
  const {
    connections,
    totalCount: personalTotal,
    isLoading,
    error,
  } = useConnections({ page, perPage });
  const { teamConnections, totalCount: teamTotal } = useTeamConnections({
    page,
    perPage,
  });
  const deleteMutation = useDeleteConnection();
  const deleteTeamMutation = useDeleteTeamConnection();

  const role = useAuthStore((s) => s.role);
  const cfg = getConfig();
  const teamEnabled = !!cfg.cloud || !!cfg.enterprise;
  const canManageTeam = teamEnabled && !!role && role !== "observer";

  const [drawer, setDrawer] = useState<DrawerState | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<{
    scope: Scope;
    id: string;
    label: string;
  } | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [hostKey, setHostKey] = useState<{
    host: string;
    port: number;
    scope: "personal" | "namespace";
    canForget: boolean;
  } | null>(null);

  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace } = useNamespace(tenantId);
  const nsName = namespace?.name ?? "";

  const active = drawer && drawer.mode !== "create" ? drawer.connection : null;
  const activeScope =
    drawer && drawer.mode !== "create" ? drawer.scope : "personal";
  const activeDeviceUid = active?.kind === "device" ? active.device_uid : "";
  const { device: activeDevice } = useDevice(activeDeviceUid);
  const activeDeviceName = activeDevice?.name ?? active?.label ?? "";
  const editable = drawer?.mode === "create" || drawer?.mode === "edit";

  // Scope column always renders; in Community every row is "Personal" since team
  // connections only exist once the namespace is on a plan that has them.
  const showScope = true;
  const total = personalTotal + teamTotal;
  // Personal and team are separate paginated sources advanced by one control, so
  // the page count is the larger of the two.
  const totalPages = Math.max(
    Math.ceil(personalTotal / perPage),
    Math.ceil(teamTotal / perPage),
  );

  const onConnectPersonal = (c: Connection) =>
    setDrawer({ mode: "connect", scope: "personal", connection: c });
  const onEditPersonal = (c: Connection) =>
    setDrawer({ mode: "edit", scope: "personal", connection: c });
  const onConnectTeam = (t: TeamConnection) =>
    setDrawer({
      mode: "connect",
      scope: "team",
      connection: teamToConnection(t),
    });
  const onEditTeam = (t: TeamConnection) =>
    setDrawer({ mode: "edit", scope: "team", connection: teamToConnection(t) });

  const headers = useMemo(
    () => ["", "Label", "Target", showScope ? "Access" : "Your auth", ""],
    [showScope],
  );

  return (
    <div>
      <PageHeader
        icon={<ServerStackIcon className="w-6 h-6" />}
        overline="Connection Management"
        title="Connections"
        description="Reach your ShellHub devices and external SSH hosts from one place."
      >
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setDrawer({ mode: "create" })}
            className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200"
          >
            <PlusIcon className="w-4 h-4" strokeWidth={2} />
            New connection
          </button>
        </div>
      </PageHeader>

      {error && (
        <Alert variant="error" className="mb-4">
          {error.message}
        </Alert>
      )}

      {isLoading ? (
        <div className="bg-surface border border-border rounded-lg p-8 text-center text-xs font-mono text-text-muted">
          Loading connections...
        </div>
      ) : total === 0 ? (
        <div className="bg-surface border border-border rounded-lg p-10 text-center">
          <ServerStackIcon
            className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
            strokeWidth={1}
          />
          <p className="text-xs font-mono text-text-muted">
            No connections yet. Add one to reach a host or device from your
            browser.
          </p>
        </div>
      ) : (
        <div className="bg-surface border border-border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-text-secondary">
                {headers.map((h, i) => (
                  <th
                    key={i}
                    className={`px-4 py-3 ${
                      h ? "text-left font-medium" : i === 0 ? "w-10" : ""
                    }`}
                  >
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {connections.map((c) => (
                <PersonalRow
                  key={`p-${c.id}`}
                  connection={c}
                  showScope={showScope}
                  onConnect={() => onConnectPersonal(c)}
                  onEdit={() => onEditPersonal(c)}
                  onDelete={() =>
                    setConfirmDelete({
                      scope: "personal",
                      id: c.id,
                      label: c.label,
                    })
                  }
                  onShowHostKey={() =>
                    setHostKey({
                      host: c.host,
                      port: c.port,
                      scope: "personal",
                      canForget: true,
                    })
                  }
                />
              ))}
              {teamConnections.map((t) => (
                <TeamRow
                  key={`t-${t.id}`}
                  connection={t}
                  canManage={canManageTeam}
                  showScope={showScope}
                  onConnect={() => onConnectTeam(t)}
                  onEdit={() => onEditTeam(t)}
                  onDelete={() =>
                    setConfirmDelete({
                      scope: "team",
                      id: t.id,
                      label: t.label,
                    })
                  }
                  onShowHostKey={() =>
                    setHostKey({
                      host: t.host,
                      port: t.port,
                      scope: "namespace",
                      canForget: canManageTeam,
                    })
                  }
                />
              ))}
            </tbody>
          </table>
          <Pagination
            page={page}
            totalPages={totalPages}
            onPageChange={setPage}
          />
        </div>
      )}

      <ConnectDrawer
        open={!!drawer}
        onClose={() => setDrawer(null)}
        connection={active}
        scope={activeScope}
        canCreateTeam={canManageTeam}
        editable={editable}
        onSaved={() => setDrawer(null)}
        deviceUid={active?.device_uid ?? ""}
        deviceName={activeDeviceName}
        sshid={
          active?.kind === "device" && nsName && activeDeviceName
            ? buildSshid(nsName, activeDeviceName)
            : ""
        }
      />

      {hostKey && (
        <HostKeyModal
          open
          onClose={() => setHostKey(null)}
          host={hostKey.host}
          port={hostKey.port}
          scope={hostKey.scope}
          canForget={hostKey.canForget}
        />
      )}

      <ConfirmDialog
        open={!!confirmDelete}
        onClose={() => {
          setConfirmDelete(null);
          setDeleteError(null);
        }}
        title="Delete connection"
        description={
          confirmDelete ? (
            <>
              Delete{" "}
              <span className="font-medium text-text-primary">
                {confirmDelete.label}
              </span>
              ? This only removes the saved connection, not the device or host
              it points to.
            </>
          ) : (
            ""
          )
        }
        confirmLabel="Delete"
        variant="danger"
        errorMessage={deleteError}
        onConfirm={async () => {
          if (!confirmDelete) return;
          setDeleteError(null);
          try {
            if (confirmDelete.scope === "team") {
              await deleteTeamMutation.mutateAsync(confirmDelete.id);
            } else {
              await deleteMutation.mutateAsync(confirmDelete.id);
            }
            setConfirmDelete(null);
          } catch (e) {
            setDeleteError(
              e instanceof Error
                ? e.message
                : "Failed to delete the connection.",
            );
          }
        }}
      />
    </div>
  );
}
