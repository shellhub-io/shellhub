import { useEffect, useState } from "react";
import {
  useParams,
  useNavigate,
  useSearchParams,
  Navigate,
} from "react-router-dom";
import Breadcrumb from "@/components/common/Breadcrumb";
import {
  TagIcon,
  XMarkIcon,
  PlusIcon,
  PencilSquareIcon,
  CheckIcon,
  TrashIcon,
  InformationCircleIcon,
  ServerIcon,
  ClockIcon,
  CubeIcon,
  ChevronDoubleRightIcon,
} from "@heroicons/react/24/outline";
import { isSdkError } from "../api/errors";
import { useContainer } from "../hooks/useContainer";
import {
  useRenameContainer,
  useAddContainerTag,
  useRemoveContainerTag,
} from "../hooks/useContainerMutations";
import { normalizeContainer } from "../hooks/useContainers";
import { useNamespace } from "../hooks/useNamespaces";
import { useAuthStore } from "../stores/authStore";
import { useTerminalStore } from "../stores/terminalStore";
import { useContainerActions } from "../hooks/useContainerActions";
import ContainerActionsPortal from "./containers/ContainerActionsPortal";
import ConnectDrawer from "../components/ConnectDrawer";
import CopyButton from "../components/common/CopyButton";
import { formatDateFull, formatRelative } from "../utils/date";
import { buildSshid } from "../utils/sshid";
import { useHasPermission } from "../hooks/useHasPermission";
import RestrictedAction from "../components/common/RestrictedAction";
import PageLoader from "@/components/common/PageLoader";
import { Button, Card, IconButton } from "@shellhub/design-system/primitives";

/* ─── Shared styles ─── */
const LABEL =
  "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";
const VALUE = "text-sm text-text-primary font-medium mt-0.5";

/* ─── Info Row ─── */
function InfoItem({
  label,
  value,
  mono,
  copyable,
  truncate,
}: {
  label: string;
  value: string;
  mono?: boolean;
  copyable?: boolean;
  truncate?: number;
}) {
  const display = truncate && value ? value.slice(0, truncate) : value;

  return (
    <div>
      <dt className={LABEL}>{label}</dt>
      <dd className="flex items-center gap-1 mt-0.5">
        <span
          className={`text-sm text-text-primary ${mono ? "font-mono text-xs" : "font-medium"}`}
        >
          {display || "—"}
        </span>
        {copyable && value && <CopyButton text={value} />}
      </dd>
    </div>
  );
}

/* ─── Tags Section ─── */
function TagsSection({ uid, tags }: { uid: string; tags: string[] }) {
  const addTagMutation = useAddContainerTag();
  const removeTagMutation = useRemoveContainerTag();
  const canEditTags = useHasPermission("tag:edit");
  const [input, setInput] = useState("");
  const [adding, setAdding] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleAdd = async () => {
    const tag = input.trim();
    if (!tag) return;
    setError(null);

    if (tags.includes(tag)) {
      setError("This tag is already added.");
      return;
    }
    if (tags.length >= 3) return;
    if (tag.length < 3) {
      setError("Tag must be at least 3 characters.");
      return;
    }
    if (tag.length > 255) {
      setError("Tag must be at most 255 characters.");
      return;
    }
    if (!/^[a-zA-Z0-9]+$/.test(tag)) {
      setError("Tag must contain only letters and numbers.");
      return;
    }

    setAdding(true);
    try {
      await addTagMutation.mutateAsync({ path: { uid, name: tag } });
      setInput("");
    } catch (e) {
      const status = isSdkError(e) ? e.status : undefined;
      if (status === 403) setError("You don't have permission to add tags.");
      else if (status === 400) setError(`"${tag}" is not a valid tag name.`);
      else setError("Failed to add tag.");
    }
    setAdding(false);
  };

  const handleRemove = async (tag: string) => {
    setError(null);
    try {
      await removeTagMutation.mutateAsync({ path: { uid, name: tag } });
    } catch (e) {
      const status = isSdkError(e) ? e.status : undefined;
      if (status === 403) setError("You don't have permission to remove tags.");
      else setError(`Failed to remove "${tag}".`);
    }
  };

  return (
    <div>
      <h3 className={LABEL + " mb-2"}>Tags</h3>
      <div className="flex flex-wrap items-center gap-2">
        {tags.map((tag) => (
          <span
            key={tag}
            className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-primary/10 text-primary text-xs rounded-md font-medium"
          >
            <TagIcon className="w-3 h-3" strokeWidth={2} />
            {tag}
            {canEditTags && (
              <button
                type="button"
                onClick={() => void handleRemove(tag)}
                aria-label={`Remove tag ${tag}`}
                className="hover:text-white transition-colors"
              >
                <XMarkIcon className="w-3 h-3" strokeWidth={2} />
              </button>
            )}
          </span>
        ))}
        {canEditTags && tags.length < 3 && (
          <div className="flex items-center gap-1.5">
            <input
              type="text"
              value={input}
              onChange={(e) => {
                setInput(e.target.value);
                setError(null);
              }}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  void handleAdd();
                }
              }}
              placeholder="Add tag..."
              aria-label="Add tag"
              className="w-28 px-2.5 py-1 bg-card border border-border rounded-md text-xs text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/40 transition-all"
            />
            <IconButton
              variant="primary"
              size="sm"
              disabled={adding || !input.trim()}
              aria-label="Add tag"
              onClick={() => void handleAdd()}
            >
              <PlusIcon className="w-4 h-4" strokeWidth={2} />
            </IconButton>
          </div>
        )}
      </div>
      {tags.length >= 3 && (
        <p className="text-2xs text-text-muted mt-1.5">
          Maximum of 3 tags reached.
        </p>
      )}
      {error && (
        <p role="alert" className="text-2xs text-accent-red mt-1.5">
          {error}
        </p>
      )}
    </div>
  );
}

/* ─── Rename Inline ─── */
function RenameSection({
  uid,
  currentName,
}: {
  uid: string;
  currentName: string;
}) {
  const renameMutation = useRenameContainer();
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(currentName);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSave = async () => {
    if (!name.trim() || name.trim() === currentName) {
      setEditing(false);
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await renameMutation.mutateAsync({
        path: { uid },
        body: { name: name.trim() },
      });
      setEditing(false);
    } catch (e) {
      const status = isSdkError(e) ? e.status : undefined;
      if (status === 409)
        setError("A container with that name already exists.");
      else if (status === 400) setError("Invalid container name.");
      else setError("Failed to rename container.");
    }
    setSaving(false);
  };

  if (!editing) {
    return (
      <div className="flex items-center gap-2">
        <h1 className="text-2xl font-bold text-text-primary">{currentName}</h1>
        <IconButton
          variant="primary"
          title="Rename container"
          aria-label="Rename container"
          onClick={() => {
            setName(currentName);
            setEditing(true);
          }}
        >
          <PencilSquareIcon className="w-4 h-4" />
        </IconButton>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center gap-2">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") void handleSave();
            if (e.key === "Escape") setEditing(false);
          }}
          aria-label="Container name"
          className="text-2xl font-bold text-text-primary bg-transparent border-b-2 border-primary/50 focus:outline-none focus:border-primary w-full max-w-md"
        />
        <IconButton
          variant="primary"
          aria-label="Save name"
          disabled={saving}
          onClick={() => void handleSave()}
        >
          <CheckIcon className="w-4 h-4" strokeWidth={2} />
        </IconButton>
        <IconButton
          aria-label="Cancel rename"
          onClick={() => setEditing(false)}
        >
          <XMarkIcon className="w-4 h-4" strokeWidth={2} />
        </IconButton>
      </div>
      {error && (
        <p role="alert" className="text-2xs text-accent-red mt-1">
          {error}
        </p>
      )}
    </div>
  );
}

/* ─── Page ─── */
export default function ContainerDetails() {
  const { uid } = useParams<{ uid: string }>();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const { container, isLoading, error } = useContainer(uid ?? "");

  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const existingSession = useTerminalStore((s) =>
    s.sessions.find((sess) => sess.deviceUid === uid),
  );
  const restoreTerminal = useTerminalStore((s) => s.restore);
  const [connectOpen, setConnectOpen] = useState(false);

  const containerActions = useContainerActions({
    // ContainerDetails never showed the billing warning on master — a 402 on
    // accept surfaces as the dialog's inline error instead. Keep that behavior.
    enableBillingWarning: false,
    onSuccess: (action) => {
      if (action === "remove") void navigate("/containers");
    },
  });

  const shouldAutoConnect =
    searchParams.get("connect") === "true" &&
    container?.online &&
    !existingSession;
  const [autoConnectDone, setAutoConnectDone] = useState(false);
  if (shouldAutoConnect && !autoConnectDone) {
    setAutoConnectDone(true);
    setConnectOpen(true);
  }
  if (!shouldAutoConnect && autoConnectDone) {
    setAutoConnectDone(false);
  }

  useEffect(() => {
    if (
      searchParams.get("connect") === "true" &&
      container?.online &&
      existingSession
    ) {
      restoreTerminal(existingSession.id);
    }
  }, [searchParams, container, existingSession, restoreTerminal]);

  if (!uid) return <Navigate to="/containers" replace />;

  if (isLoading) {
    return <PageLoader label="Loading container details" />;
  }

  if (error || !container) {
    return <Navigate to="/containers" replace />;
  }

  const nsName = currentNamespace?.name ?? "";
  const sshid = nsName ? buildSshid(nsName, container.name) : container.uid;

  const tags = normalizeContainer(container).tags;

  return (
    <div className="animate-fade-in">
      <Breadcrumb
        items={[
          { label: "Containers", to: "/containers" },
          { label: container.name },
        ]}
      />

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4 mb-8">
        <div className="flex items-start gap-4">
          <div className="relative shrink-0">
            <div className="w-14 h-14 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center">
              <CubeIcon className="w-7 h-7 text-primary" />
            </div>
            <span
              className={`absolute -bottom-1 -right-1 w-4 h-4 rounded-full border-2 border-background ${
                container.online
                  ? "bg-accent-green shadow-[0_0_8px_rgba(130,165,104,0.5)]"
                  : "bg-text-muted/40"
              }`}
            />
          </div>

          <div>
            <RenameSection uid={container.uid} currentName={container.name} />
            <div className="flex items-center gap-2 mt-1.5">
              <span
                className={`inline-flex items-center gap-1 px-2 py-0.5 text-2xs font-semibold rounded-md ${
                  container.online
                    ? "bg-accent-green/10 text-accent-green border border-accent-green/20"
                    : "bg-text-muted/10 text-text-muted border border-border"
                }`}
              >
                <span
                  className={`w-1.5 h-1.5 rounded-full ${container.online ? "bg-accent-green" : "bg-text-muted/60"}`}
                />
                {container.online ? "Online" : "Offline"}
              </span>
              <span
                className={`inline-flex items-center px-2 py-0.5 text-2xs font-medium rounded-md ${
                  container.status === "accepted"
                    ? "bg-accent-green/10 text-accent-green"
                    : container.status === "pending"
                      ? "bg-accent-yellow/10 text-accent-yellow"
                      : "bg-accent-red/10 text-accent-red"
                }`}
              >
                {container.status
                  ? container.status.charAt(0).toUpperCase() +
                    container.status.slice(1)
                  : "Unknown"}
              </span>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 shrink-0">
          {container.status === "accepted" && (
            <>
              <RestrictedAction action="device:connect">
                <Button
                  variant="success"
                  onClick={() => {
                    if (existingSession) {
                      restoreTerminal(existingSession.id);
                    } else {
                      setConnectOpen(true);
                    }
                  }}
                  disabled={!container.online}
                  icon={
                    <ChevronDoubleRightIcon
                      className="w-4 h-4"
                      strokeWidth={2}
                    />
                  }
                >
                  Connect
                </Button>
              </RestrictedAction>
              <RestrictedAction action="device:remove">
                <IconButton
                  variant="danger"
                  size="lg"
                  title="Remove container"
                  aria-label="Remove container"
                  className="border border-border"
                  onClick={() =>
                    containerActions.requestAction(container, "remove")
                  }
                >
                  <TrashIcon className="w-4 h-4" />
                </IconButton>
              </RestrictedAction>
            </>
          )}
          {container.status === "pending" && (
            <>
              <RestrictedAction action="device:accept">
                <Button
                  variant="success"
                  onClick={() =>
                    containerActions.requestAction(container, "accept")
                  }
                >
                  Accept
                </Button>
              </RestrictedAction>
              <RestrictedAction action="device:reject">
                <Button
                  variant="warning"
                  onClick={() =>
                    containerActions.requestAction(container, "reject")
                  }
                >
                  Reject
                </Button>
              </RestrictedAction>
            </>
          )}
          {container.status === "rejected" && (
            <>
              <RestrictedAction action="device:accept">
                <Button
                  variant="success"
                  onClick={() =>
                    containerActions.requestAction(container, "accept")
                  }
                >
                  Accept
                </Button>
              </RestrictedAction>
              <RestrictedAction action="device:remove">
                <Button
                  variant="destructive"
                  onClick={() =>
                    containerActions.requestAction(container, "remove")
                  }
                >
                  Remove
                </Button>
              </RestrictedAction>
            </>
          )}
        </div>
      </div>

      {/* SSHID Banner */}
      {container.status === "accepted" && (
        <Card className="p-4 mb-6 flex items-center justify-between gap-4">
          <div>
            <p className={LABEL}>SSHID</p>
            <code className="text-sm font-mono text-accent-cyan mt-0.5 block">
              {sshid}
            </code>
          </div>
          <CopyButton text={sshid} />
        </Card>
      )}

      {/* Info Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
        <Card className="p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <InformationCircleIcon className="w-4 h-4 text-primary" />
            Identity
          </h3>
          <dl className="space-y-3">
            <InfoItem
              label="UID"
              value={container.uid}
              mono
              copyable
              truncate={8}
            />
            <InfoItem
              label="MAC Address"
              value={container.identity?.mac ?? ""}
              mono
              copyable
            />
            <InfoItem
              label="Remote Address"
              value={container.remote_addr ?? ""}
              mono
            />
          </dl>
        </Card>

        <Card className="p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <ServerIcon className="w-4 h-4 text-primary" />
            Container
          </h3>
          <dl className="space-y-3">
            <InfoItem
              label="Image"
              value={container.info?.pretty_name ?? ""}
              mono
            />
            <InfoItem
              label="Architecture"
              value={container.info?.arch ?? ""}
              mono
            />
            <InfoItem
              label="Agent Version"
              value={container.info?.version ?? ""}
              mono
            />
          </dl>
        </Card>

        <Card className="p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <ClockIcon className="w-4 h-4 text-primary" />
            Timeline
          </h3>
          <dl className="space-y-3">
            <div>
              <dt className={LABEL}>Created</dt>
              <dd className={VALUE}>{formatDateFull(container.created_at)}</dd>
            </div>
            <div>
              <dt className={LABEL}>Last Seen</dt>
              <dd className="flex items-center gap-2 mt-0.5">
                <span className="text-sm text-text-primary font-medium">
                  {formatRelative(container.last_seen)}
                </span>
                <span className="text-2xs text-text-muted">
                  {formatDateFull(container.last_seen)}
                </span>
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Status Updated</dt>
              <dd className={VALUE}>
                {formatDateFull(container.status_update_at ?? "")}
              </dd>
            </div>
          </dl>
        </Card>
      </div>

      {/* Tags */}
      <Card className="p-5 mb-6">
        <TagsSection uid={container.uid} tags={tags} />
      </Card>

      {/* Connect Drawer */}
      <ConnectDrawer
        open={connectOpen}
        onClose={() => setConnectOpen(false)}
        deviceUid={container.uid}
        deviceName={container.name}
        sshid={sshid}
      />

      {/* Action Dialog */}
      <ContainerActionsPortal controller={containerActions} />
    </div>
  );
}
