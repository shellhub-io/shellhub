import { useState } from "react";
import { usePublicKeys, type PublicKey } from "@/hooks/usePublicKeys";
import { useDeletePublicKey } from "@/hooks/usePublicKeyMutations";
import PageHeader from "@/components/common/PageHeader";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import CopyButton from "@/components/common/CopyButton";
import DataTable, { type Column } from "@/components/common/DataTable";
import KeyDrawer from "./KeyDrawer";
import { formatDate } from "@/utils/date";
import RestrictedAction from "@/components/common/RestrictedAction";
import {
  KeyIcon,
  MagnifyingGlassIcon,
  PlusIcon,
  ShieldCheckIcon,
  TagIcon,
  UsersIcon,
  UserIcon,
  GlobeAltIcon,
  ServerIcon,
  PencilSquareIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";

const PER_PAGE = 10;

/* ── scope cell ──────────────────────────────────── */

function ScopeCell({ pk }: { pk: PublicKey }) {
  const isAllUsers = pk.username === ".*" || !pk.username;
  const username = isAllUsers ? "All users" : pk.username;

  let deviceNode: React.ReactNode;
  if (pk.filter.tags && pk.filter.tags.length > 0) {
    deviceNode = (
      <span className="inline-flex items-center gap-1.5 flex-wrap">
        {pk.filter.tags.map((tag) => {
          const label
            = typeof tag === "string" ? tag : (tag as { name: string }).name;
          return (
            <span
              key={label}
              className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-primary/10 text-primary text-2xs font-mono rounded"
            >
              <TagIcon className="w-2.5 h-2.5 shrink-0" strokeWidth={2} />
              {label}
            </span>
          );
        })}
      </span>
    );
  } else if (pk.filter.hostname && pk.filter.hostname !== ".*") {
    deviceNode = (
      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-accent-yellow/10 text-accent-yellow text-2xs font-mono rounded">
        <ServerIcon className="w-2.5 h-2.5 shrink-0" strokeWidth={2} />
        {pk.filter.hostname}
      </span>
    );
  } else {
    deviceNode = (
      <span className="inline-flex items-center gap-1 text-2xs font-mono text-text-muted">
        <GlobeAltIcon className="w-2.5 h-2.5 shrink-0" strokeWidth={2} />
        All devices
      </span>
    );
  }

  return (
    <div className="flex items-center gap-2 flex-wrap">
      <span
        className={`inline-flex items-center gap-1 text-xs font-mono ${isAllUsers ? "text-text-muted" : "text-text-secondary"}`}
      >
        {isAllUsers
          ? <UsersIcon className="w-3 h-3 shrink-0" strokeWidth={2} />
          : <UserIcon className="w-3 h-3 shrink-0" strokeWidth={2} />}
        {username}
      </span>
      <span className="text-text-muted/40 text-xs">{"\u2192"}</span>
      {deviceNode}
    </div>
  );
}

/* ── page ────────────────────────────────────────── */

export default function PublicKeys() {
  const [page, setPage] = useState(1);
  const { publicKeys, totalCount, isLoading } = usePublicKeys({ page });
  const deleteKey = useDeletePublicKey();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<PublicKey | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{
    fingerprint: string;
    name: string;
  } | null>(null);
  const [search, setSearch] = useState("");

  const openNew = () => {
    setEditTarget(null);
    setDrawerOpen(true);
  };
  const openEdit = (key: PublicKey) => {
    setEditTarget(key);
    setDrawerOpen(true);
  };
  const closeDrawer = () => {
    setDrawerOpen(false);
    setEditTarget(null);
  };

  const totalPages = Math.ceil(totalCount / PER_PAGE);
  const filtered = search
    ? publicKeys.filter(
      (k) =>
        k.name.toLowerCase().includes(search.toLowerCase())
        || k.fingerprint.toLowerCase().includes(search.toLowerCase()),
    )
    : publicKeys;

  const columns: Column<PublicKey>[] = [
    {
      key: "name",
      header: "Name",
      render: (pk) => (
        <span className="text-sm font-medium text-text-primary">{pk.name}</span>
      ),
    },
    {
      key: "scope",
      header: "Scope",
      render: (pk) => <ScopeCell pk={pk} />,
    },
    {
      key: "fingerprint",
      header: "Fingerprint",
      render: (pk) => (
        <div className="flex items-center gap-1">
          <code
            className="text-2xs font-mono text-text-muted truncate max-w-[200px]"
            title={pk.fingerprint}
          >
            {pk.fingerprint}
          </code>
          <CopyButton text={pk.fingerprint} />
        </div>
      ),
    },
    {
      key: "added",
      header: "Added",
      render: (pk) => (
        <span className="text-xs font-mono text-text-muted">
          {formatDate(pk.created_at)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "",
      headerClassName: "w-16",
      render: (pk) => (
        <div className="flex items-center justify-end gap-0.5">
          <RestrictedAction action="publicKey:edit">
            <button
              onClick={() => openEdit(pk)}
              title="Edit"
              className="p-1.5 rounded-md text-text-muted hover:text-primary hover:bg-primary/10 transition-all"
            >
              <PencilSquareIcon className="w-4 h-4" strokeWidth={2} />
            </button>
          </RestrictedAction>
          <RestrictedAction action="publicKey:remove">
            <button
              onClick={() =>
                setDeleteTarget({
                  fingerprint: pk.fingerprint,
                  name: pk.name,
                })}
              title="Delete"
              className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/10 transition-all"
            >
              <TrashIcon className="w-4 h-4" strokeWidth={2} />
            </button>
          </RestrictedAction>
        </div>
      ),
    },
  ];

  /* Full-page onboarding empty state (no keys at all) */
  if (!isLoading && publicKeys.length === 0) {
    return (
      <div>
        <div className="relative -mx-8 -mt-8 min-h-[calc(100vh-3.5rem)] flex flex-col">
          <div className="absolute inset-0 overflow-hidden pointer-events-none">
            <div className="absolute -top-32 left-1/3 w-[500px] h-[500px] bg-primary/5 rounded-full blur-[120px] animate-pulse-subtle" />
            <div
              className="absolute bottom-0 right-1/4 w-[400px] h-[400px] bg-accent-blue/5 rounded-full blur-[100px] animate-pulse-subtle"
              style={{ animationDelay: "1s" }}
            />
            <div className="absolute inset-0 grid-bg opacity-30" />
          </div>
          <div className="flex-1 flex items-center justify-center px-8 py-12">
            <div className="w-full max-w-2xl animate-fade-in">
              <div className="text-center mb-10">
                <div className="w-16 h-16 rounded-2xl bg-primary/10 border border-primary/20 flex items-center justify-center mx-auto mb-6">
                  <KeyIcon className="w-8 h-8 text-primary" />
                </div>
                <span className="inline-block text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
                  SSH Authentication
                </span>
                <h1 className="text-3xl font-bold text-text-primary mb-3">
                  Public Keys
                </h1>
                <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
                  Set up SSH public keys to enable secure, passwordless
                  authentication to your devices. Manage access by user,
                  hostname, or device tags.
                </p>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-10">
                {[
                  {
                    icon: <ShieldCheckIcon className="w-5 h-5" />,
                    title: "Passwordless Access",
                    description:
                      "Authenticate via SSH keys instead of passwords for stronger security.",
                  },
                  {
                    icon: <UsersIcon className="w-5 h-5" />,
                    title: "User Control",
                    description:
                      "Restrict which usernames can connect with each public key.",
                  },
                  {
                    icon: <TagIcon className="w-5 h-5" />,
                    title: "Device Filtering",
                    description:
                      "Scope keys to specific devices using hostname patterns or tags.",
                  },
                ].map((h, idx) => (
                  <div
                    key={h.title}
                    className="bg-card/60 border border-border rounded-xl p-5 text-center animate-slide-up"
                    style={{ animationDelay: `${150 + idx * 100}ms` }}
                  >
                    <div className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center mx-auto mb-3 text-primary">
                      {h.icon}
                    </div>
                    <h3 className="text-sm font-semibold text-text-primary mb-1">
                      {h.title}
                    </h3>
                    <p className="text-xs text-text-muted leading-relaxed">
                      {h.description}
                    </p>
                  </div>
                ))}
              </div>
              <div
                className="text-center animate-slide-up"
                style={{ animationDelay: "450ms" }}
              >
                <RestrictedAction action="publicKey:create">
                  <button
                    onClick={openNew}
                    className="inline-flex items-center gap-2 px-6 py-3 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all shadow-lg shadow-primary/20"
                  >
                    <PlusIcon className="w-4 h-4" strokeWidth={2} />
                    Add your first key
                  </button>
                </RestrictedAction>
                <p className="mt-4 text-2xs text-text-muted">
                  Supports RSA, DSA, ECDSA, and ED25519 key types.
                </p>
              </div>
            </div>
          </div>
        </div>

        <KeyDrawer open={drawerOpen} editKey={editTarget} onClose={closeDrawer} />
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        icon={<KeyIcon className="w-6 h-6" />}
        overline="Security"
        title="Public Keys"
        description="Manage SSH public keys for passwordless authentication to your devices."
      >
        <RestrictedAction action="publicKey:create">
          <button
            onClick={openNew}
            className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
          >
            <PlusIcon className="w-4 h-4" strokeWidth={2} />
            Add Public Key
          </button>
        </RestrictedAction>
      </PageHeader>

      <div className="mb-4 animate-fade-in">
        <div className="relative max-w-sm">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted pointer-events-none" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by name or fingerprint..."
            className="w-full pl-9 pr-3.5 py-2 bg-card border border-border rounded-lg text-sm text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
          />
        </div>
      </div>

      <DataTable
        columns={columns}
        data={filtered}
        rowKey={(pk) => pk.fingerprint}
        isLoading={isLoading}
        loadingMessage="Loading public keys..."
        page={page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="key"
        onPageChange={setPage}
        emptyMessage={search ? `No keys matching \u201C${search}\u201D` : "No public keys found"}
      />

      <KeyDrawer open={drawerOpen} editKey={editTarget} onClose={closeDrawer} />

      <ConfirmDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={async () => {
          await deleteKey.mutateAsync({ path: { fingerprint: deleteTarget!.fingerprint } });
          if (publicKeys.length === 1 && page > 1) setPage(page - 1);
          setDeleteTarget(null);
        }}
        title="Delete Public Key"
        description={(
          <>
            Are you sure you want to delete
            {" "}
            <span className="font-medium text-text-primary">
              {deleteTarget?.name}
            </span>
            ? This action cannot be undone.
          </>
        )}
        confirmLabel="Delete"
      />
    </div>
  );
}
