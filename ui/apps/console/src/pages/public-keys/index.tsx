import { useState } from "react";
import { usePublicKeys } from "@/hooks/usePublicKeys";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { useDeletePublicKey } from "@/hooks/usePublicKeyMutations";
import PageHeader from "@/components/common/PageHeader";
import EmptyState from "@/components/common/EmptyState";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import CopyButton from "@/components/common/CopyButton";
import DataTable, { type Column } from "@/components/common/DataTable";
import SearchField from "@/components/common/fields/SearchField";
import KeyDrawer from "./KeyDrawer";
import { formatDate } from "@/utils/date";
import RestrictedAction from "@/components/common/RestrictedAction";
import {
  KeyIcon,
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
import { PublicKeyResponse as PublicKey } from "@/client";
import { Button, IconButton } from "@shellhub/design-system/primitives";

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
          const label =
            typeof tag === "string" ? tag : (tag as { name: string }).name;
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
        {isAllUsers ? (
          <UsersIcon className="w-3 h-3 shrink-0" strokeWidth={2} />
        ) : (
          <UserIcon className="w-3 h-3 shrink-0" strokeWidth={2} />
        )}
        {username}
      </span>
      <span className="text-text-muted/40 text-xs">{"\u2192"}</span>
      {deviceNode}
    </div>
  );
}

/* ── page ────────────────────────────────────────── */

const SEARCH_DEBOUNCE_MS = 300;

type PublicKeysParams = {
  page: number;
  search: string;
};

const DEFAULTS: PublicKeysParams = {
  page: 1,
  search: "",
};

export default function PublicKeys() {
  const { params, setPage, setSearch } = usePaginatedListState<PublicKeysParams>({
    defaults: DEFAULTS,
  });

  const debouncedSearch = useDebouncedValue(params.search, SEARCH_DEBOUNCE_MS);
  const { publicKeys, totalCount, isLoading } = usePublicKeys({
    page: params.page,
    search: debouncedSearch,
  });
  const deleteKey = useDeletePublicKey();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<PublicKey | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{
    fingerprint: string;
    name: string;
  } | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const closeDelete = () => {
    setDeleteError(null);
    setDeleteTarget(null);
  };

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    setDeleteError(null);
    try {
      await deleteKey.mutateAsync({
        path: { fingerprint: deleteTarget.fingerprint },
      });
      if (publicKeys.length === 1 && params.page > 1) setPage(params.page - 1);
      closeDelete();
    } catch (err) {
      setDeleteError(
        err instanceof Error ? err.message : "Failed to delete public key.",
      );
    }
  };

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
            <IconButton
              variant="primary"
              title="Edit"
              aria-label={`Edit ${pk.name}`}
              onClick={() => openEdit(pk)}
            >
              <PencilSquareIcon className="w-4 h-4" strokeWidth={2} />
            </IconButton>
          </RestrictedAction>
          <RestrictedAction action="publicKey:remove">
            <IconButton
              variant="danger"
              title="Delete"
              aria-label={`Delete ${pk.name}`}
              onClick={() =>
                setDeleteTarget({
                  fingerprint: pk.fingerprint,
                  name: pk.name,
                })
              }
            >
              <TrashIcon className="w-4 h-4" strokeWidth={2} />
            </IconButton>
          </RestrictedAction>
        </div>
      ),
    },
  ];

  /* Full-page onboarding empty state (no keys at all) */
  if (!isLoading && publicKeys.length === 0 && !debouncedSearch) {
    return (
      <>
        <EmptyState
          icon={<KeyIcon className="w-8 h-8" />}
          overline="SSH Authentication"
          title="Public Keys"
          description="Set up SSH public keys to enable secure, passwordless authentication to your devices. Manage access by user, hostname, or device tags."
          features={[
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
          ]}
          footnote="Supports RSA, DSA, ECDSA, and ED25519 key types."
        >
          <RestrictedAction action="publicKey:create">
            <Button
              size="lg"
              onClick={openNew}
              icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
            >
              Add your first key
            </Button>
          </RestrictedAction>
        </EmptyState>

        <KeyDrawer
          open={drawerOpen}
          editKey={editTarget}
          onClose={closeDrawer}
        />
      </>
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
          <Button
            onClick={openNew}
            icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Add Public Key
          </Button>
        </RestrictedAction>
      </PageHeader>

      <SearchField
        className="mb-4"
        value={params.search}
        onChange={(next) => setSearch(next)}
        placeholder="Search by name or fingerprint..."
        aria-label="Search public keys by name or fingerprint"
      />

      <DataTable
        columns={columns}
        data={publicKeys}
        rowKey={(pk) => pk.fingerprint}
        isLoading={isLoading}
        loadingMessage="Loading public keys..."
        page={params.page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="key"
        onPageChange={setPage}
        emptyMessage={
          debouncedSearch
            ? `No keys matching \u201C${debouncedSearch}\u201D`
            : "No public keys found"
        }
      />

      <KeyDrawer open={drawerOpen} editKey={editTarget} onClose={closeDrawer} />

      <ConfirmDialog
        open={!!deleteTarget}
        onClose={closeDelete}
        onConfirm={confirmDelete}
        title="Delete Public Key"
        description={
          <>
            Are you sure you want to delete{" "}
            <span className="font-medium text-text-primary">
              {deleteTarget?.name}
            </span>
            ? This action cannot be undone.
          </>
        }
        confirmLabel="Delete"
      >
        {deleteError && (
          <p className="text-xs text-accent-red">{deleteError}</p>
        )}
      </ConfirmDialog>
    </div>
  );
}
