import { useEffect, useState, FormEvent } from "react";
import { Link } from "react-router-dom";
import {
  ServerStackIcon,
  PlusIcon,
  ChevronDoubleRightIcon,
  TrashIcon,
  PencilSquareIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import PageHeader from "@/components/common/PageHeader";
import Drawer from "@/components/common/Drawer";
import Alert from "@/components/common/Alert";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import FieldLabel from "@/components/common/fields/FieldLabel";
import DevicePicker from "@/components/common/DevicePicker";
import DeviceChip from "@/components/common/DeviceChip";
import OnlineDot from "@/components/common/OnlineDot";
import ConnectDrawer from "@/components/ConnectDrawer";
import { INPUT } from "@/utils/styles";
import { buildSshid } from "@/utils/sshid";
import { useConnections, useConnectionStatus } from "@/hooks/useConnections";
import { isSdkError } from "@/api/errors";
import { useDevice } from "@/hooks/useDevice";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import {
  useCreateConnection,
  useUpdateConnection,
  useDeleteConnection,
} from "@/hooks/useConnectionMutations";
import { useTerminalStore } from "@/stores/terminalStore";
import type { Connection } from "@/types/connection";

type Kind = "direct" | "device";

function ConnectionFormDrawer({
  open,
  editing,
  onClose,
}: {
  open: boolean;
  editing: Connection | null;
  onClose: () => void;
}) {
  const [label, setLabel] = useState("");
  const [kind, setKind] = useState<Kind>("direct");
  const [host, setHost] = useState("");
  const [port, setPort] = useState("22");
  const [deviceUid, setDeviceUid] = useState("");
  const [deviceName, setDeviceName] = useState("");
  const [username, setUsername] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [unreachable, setUnreachable] = useState(false);

  const createMutation = useCreateConnection();
  const updateMutation = useUpdateConnection();

  useEffect(() => {
    if (!open) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setError(null);
    setUnreachable(false);
    if (editing) {
      setLabel(editing.label);
      setKind((editing.kind as Kind) || "direct");
      setHost(editing.host);
      setPort(String(editing.port || 22));
      setDeviceUid(editing.device_uid);
      setDeviceName("");
      setUsername(editing.username);
    } else {
      setLabel("");
      setKind("direct");
      setHost("");
      setPort("22");
      setDeviceUid("");
      setDeviceName("");
      setUsername("");
    }
  }, [open, editing]);

  const portNum = Number(port);
  const portValid =
    Number.isInteger(portNum) && portNum >= 1 && portNum <= 65535;

  const pending = createMutation.isPending || updateMutation.isPending;
  const canSave =
    label.trim().length > 0 &&
    (kind === "direct"
      ? host.trim().length > 0 && portValid
      : deviceUid.length > 0) &&
    !pending;

  // The backend probes a direct target at save time and returns 422 if it's
  // unreachable. We surface that here with a NAT/agent hint and a "Save anyway"
  // (which re-submits with force).
  const save = (force: boolean) => {
    setError(null);

    const body = {
      label: label.trim(),
      username: username.trim(),
      kind,
      host: kind === "direct" ? host.trim() : undefined,
      port: kind === "direct" ? portNum : undefined,
      device_uid: kind === "device" ? deviceUid : undefined,
      force,
    };

    if (editing) {
      updateMutation.mutate(
        { id: editing.id, body },
        {
          onSuccess: () => onClose(),
          onError: () =>
            setError("Failed to save connection. Check the fields."),
        },
      );

      return;
    }

    createMutation.mutate(body, {
      onSuccess: () => onClose(),
      onError: (err) => {
        if (isSdkError(err) && err.status === 422) {
          setUnreachable(true);
        } else {
          setError("Failed to save connection. Check the fields.");
        }
      },
    });
  };

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!canSave) return;
    setUnreachable(false);
    save(false);
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={editing ? "Edit Connection" : "Add Connection"}
      subtitle="A saved way to reach an SSH target"
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            form="connection-form"
            disabled={!canSave}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            <PlusIcon className="w-4 h-4" strokeWidth={2} />
            {pending ? "Saving..." : editing ? "Save" : "Add"}
          </button>
        </>
      }
    >
      <form id="connection-form" onSubmit={handleSubmit} className="space-y-5">
        <InputField
          id="connection-label"
          label="Label"
          value={label}
          onChange={setLabel}
          placeholder="e.g. db-primary"
          autoFocus={open}
        />

        <div>
          <FieldLabel htmlFor="connection-kind">Type</FieldLabel>
          <select
            id="connection-kind"
            value={kind}
            onChange={(e) => setKind(e.target.value as Kind)}
            className={INPUT}
          >
            <option value="direct">Direct host</option>
            <option value="device">Device</option>
          </select>
          <p className="text-2xs text-text-muted mt-2">
            {kind === "direct"
              ? "Connect straight to a host by address. No agent required."
              : "Connect to a device running the ShellHub agent."}
          </p>
        </div>

        {kind === "direct" ? (
          <>
            <InputField
              id="connection-host"
              label="Hostname or IP"
              value={host}
              onChange={setHost}
              placeholder="e.g. 10.0.0.5 or db.internal"
            />
            <InputField
              id="connection-port"
              label="Port"
              value={port}
              onChange={setPort}
              placeholder="22"
            />
          </>
        ) : (
          <div>
            <FieldLabel htmlFor="connection-device">Device</FieldLabel>
            <DevicePicker
              value={deviceUid}
              valueLabel={deviceName}
              onChange={(uid, name) => {
                setDeviceUid(uid);
                setDeviceName(name);
              }}
            />
          </div>
        )}

        <InputField
          id="connection-username"
          label="Username (optional)"
          value={username}
          onChange={setUsername}
          placeholder="e.g. root"
        />

        {unreachable && (
          <div className="rounded-lg border border-accent-yellow/30 bg-accent-yellow/10 p-3 space-y-2">
            <p className="text-xs text-accent-yellow flex items-center gap-1.5 font-medium">
              <ExclamationTriangleIcon className="w-4 h-4 shrink-0" />
              Couldn't reach {host.trim()}:{portNum}
            </p>
            <p className="text-2xs text-text-secondary">
              If it's behind NAT or a firewall, install the ShellHub agent on it
              to reach it through the gateway. No public IP or open ports
              needed.
            </p>
            <div className="flex items-center gap-3 pt-0.5">
              <Link
                to="/devices/add"
                className="text-2xs text-primary hover:underline inline-flex items-center gap-1 font-medium"
              >
                Install the agent
                <ChevronDoubleRightIcon className="w-3 h-3" strokeWidth={2} />
              </Link>
              <button
                type="button"
                onClick={() => save(true)}
                disabled={pending}
                className="text-2xs text-text-secondary hover:text-text-primary underline disabled:opacity-dim"
              >
                Save anyway
              </button>
            </div>
          </div>
        )}

        {error && <Alert variant="error">{error}</Alert>}
      </form>
    </Drawer>
  );
}

function ConnectConnectionDrawer({
  connection,
  onClose,
}: {
  connection: Connection | null;
  onClose: () => void;
}) {
  const open = !!connection;
  const openTerminal = useTerminalStore((s) => s.open);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  useEffect(() => {
    if (connection) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setUsername(connection.username ?? "");
      setPassword("");
    }
  }, [connection]);

  const canConnect =
    !!connection && username.trim().length > 0 && password.length > 0;

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!connection || !canConnect) return;

    if (connection.kind === "device") {
      openTerminal({
        kind: "device",
        deviceUid: connection.device_uid,
        deviceName: connection.label,
        username: username.trim(),
        password,
      });
    } else {
      openTerminal({
        kind: "connect",
        deviceUid: connection.id,
        deviceName: connection.label,
        host: connection.host,
        port: connection.port,
        username: username.trim(),
        password,
      });
    }
    onClose();
  };

  const target =
    connection?.kind === "device"
      ? `device: ${connection?.device_uid?.slice(0, 12) ?? ""}`
      : `${connection?.host}:${connection?.port}`;

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Connect"
      subtitle={<span className="font-mono">{connection?.label ?? ""}</span>}
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            form="connect-connection-form"
            disabled={!canConnect}
            className="px-5 py-2.5 bg-accent-green/90 hover:bg-accent-green text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            <ChevronDoubleRightIcon className="w-4 h-4" strokeWidth={2} />
            Connect
          </button>
        </>
      }
    >
      <form
        id="connect-connection-form"
        onSubmit={handleSubmit}
        className="space-y-5"
      >
        <div className="bg-card border border-border rounded-lg p-3.5">
          <span className="text-2xs text-text-muted uppercase tracking-wider">
            Target
          </span>
          <code className="block text-xs font-mono text-text-secondary mt-1">
            {target}
          </code>
        </div>
        <InputField
          id="connect-username"
          label="Username"
          value={username}
          onChange={setUsername}
          placeholder="e.g. root"
          autoFocus={open}
        />
        <PasswordField
          id="connect-password"
          label="Password"
          value={password}
          onChange={setPassword}
          placeholder="Enter password"
        />
      </form>
    </Drawer>
  );
}

function ConnectionRow({
  connection,
  onConnect,
  onEdit,
  onDelete,
}: {
  connection: Connection;
  onConnect: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const c = connection;
  const { online } = useConnectionStatus(c.id);
  const { device } = useDevice(c.device_uid);

  return (
    <tr className="border-b border-border last:border-0 hover:bg-hover-subtle">
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
      <td className="px-4 py-3 font-medium text-text-primary">{c.label}</td>
      <td className="px-4 py-3">
        {c.kind === "device" ? (
          <DeviceChip
            uid={c.device_uid}
            name={device?.name ?? `${c.device_uid.slice(0, 8)}…`}
            osId={device?.info?.id}
            onClick={(e) => e.stopPropagation()}
          />
        ) : (
          <span className="inline-flex items-center gap-1.5 font-mono text-xs text-text-secondary">
            <ServerStackIcon className="w-3.5 h-3.5 text-text-muted shrink-0" />
            {c.host}:{c.port}
          </span>
        )}
      </td>
      <td className="px-4 py-3 text-text-secondary">{c.username || "—"}</td>
      <td className="px-4 py-3">
        <div className="flex items-center justify-end gap-1.5">
          {online === false ? (
            <span className="text-2xs text-text-muted/50 font-mono mr-1">
              {c.kind === "direct" ? "Unreachable" : "Offline"}
            </span>
          ) : (
            <button
              onClick={onConnect}
              className="inline-flex items-center gap-1 px-2.5 py-1 bg-accent-green/10 text-accent-green text-2xs font-semibold rounded-md hover:bg-accent-green/20 border border-accent-green/20 transition-all"
            >
              <ChevronDoubleRightIcon className="w-3 h-3" strokeWidth={2} />
              Connect
            </button>
          )}
          <button
            onClick={onEdit}
            className="inline-flex items-center px-2 py-1 text-2xs text-text-secondary hover:bg-hover-subtle rounded-md border border-transparent hover:border-border transition-all"
            aria-label="Edit connection"
          >
            <PencilSquareIcon className="w-3.5 h-3.5" strokeWidth={2} />
          </button>
          <button
            onClick={onDelete}
            className="inline-flex items-center px-2 py-1 text-2xs text-accent-red hover:bg-accent-red/10 rounded-md border border-transparent hover:border-accent-red/20 transition-all"
            aria-label="Delete connection"
          >
            <TrashIcon className="w-3.5 h-3.5" strokeWidth={2} />
          </button>
        </div>
      </td>
    </tr>
  );
}

export default function Connections() {
  const { connections, isLoading, error } = useConnections();
  const deleteMutation = useDeleteConnection();
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Connection | null>(null);
  const [connectTarget, setConnectTarget] = useState<Connection | null>(null);

  // Device connections reuse the rich device Connect drawer (vault key / password
  // + challenge-response). Direct connections use the lightweight one.
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace } = useNamespace(tenantId);
  const nsName = namespace?.name ?? "";
  const deviceTarget = connectTarget?.kind === "device" ? connectTarget : null;
  const directTarget = connectTarget?.kind === "direct" ? connectTarget : null;
  const { device: targetDevice } = useDevice(deviceTarget?.device_uid ?? "");
  const targetDeviceName = targetDevice?.name ?? deviceTarget?.label ?? "";

  const openAdd = () => {
    setEditing(null);
    setFormOpen(true);
  };

  const openEdit = (c: Connection) => {
    setEditing(c);
    setFormOpen(true);
  };

  return (
    <div>
      <PageHeader
        icon={<ServerStackIcon className="w-6 h-6" />}
        overline="Connection Management"
        title="Connections"
        description="Saved SSH connections to your devices and external hosts."
      >
        <button
          type="button"
          onClick={openAdd}
          className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200"
        >
          <PlusIcon className="w-4 h-4" strokeWidth={2} />
          Add Connection
        </button>
      </PageHeader>

      {error && (
        <Alert variant="error" className="mb-4">
          {error.message}
        </Alert>
      )}

      <div className="bg-surface border border-border rounded-lg overflow-hidden">
        {isLoading ? (
          <div className="p-8 text-center text-xs font-mono text-text-muted">
            Loading connections...
          </div>
        ) : connections.length === 0 ? (
          <div className="p-10 text-center">
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
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-text-secondary">
                <th className="px-4 py-3 w-10" />
                <th className="text-left font-medium px-4 py-3">Label</th>
                <th className="text-left font-medium px-4 py-3">Target</th>
                <th className="text-left font-medium px-4 py-3">Username</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody>
              {connections.map((c) => (
                <ConnectionRow
                  key={c.id}
                  connection={c}
                  onConnect={() => setConnectTarget(c)}
                  onEdit={() => openEdit(c)}
                  onDelete={() => deleteMutation.mutate(c.id)}
                />
              ))}
            </tbody>
          </table>
        )}
      </div>

      <ConnectionFormDrawer
        open={formOpen}
        editing={editing}
        onClose={() => setFormOpen(false)}
      />
      <ConnectConnectionDrawer
        connection={directTarget}
        onClose={() => setConnectTarget(null)}
      />
      <ConnectDrawer
        open={!!deviceTarget}
        onClose={() => setConnectTarget(null)}
        deviceUid={deviceTarget?.device_uid ?? ""}
        deviceName={targetDeviceName}
        sshid={
          nsName && targetDeviceName
            ? buildSshid(nsName, targetDeviceName)
            : (deviceTarget?.device_uid ?? "")
        }
      />
    </div>
  );
}
