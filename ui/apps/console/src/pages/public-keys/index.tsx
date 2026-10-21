import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { paginatedQueryOptions } from "@/api/pagination";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { toBase64Json } from "@/utils/encoding";
import { usePagedList } from "@/hooks/usePagedList";
import PageHeader from "@/components/common/PageHeader";
import EmptyState from "@/components/common/EmptyState";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import CopyButton from "@/components/common/CopyButton";
import DataTable, { type Column } from "@/components/common/DataTable";
import SearchField from "@/components/common/fields/SearchField";
import KeyModal from "./KeyModal";
import { formatRelative } from "@/utils/date";
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
import {
  deletePublicKeyMutation,
  getPublicKeys,
  getPublicKeysQueryKey,
  type PublicKeyResponse as PublicKey,
} from "@/client";
import { Button, Callout, IconButton } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { apiErrorMessage } from "@/api/errors";
import { PER_PAGE } from "@/utils/pagination";
import { useNavSectionTitle } from "@/components/layout/navSections";
import ObjectName from "@/components/common/ObjectName";

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
        className={cn(
          "inline-flex items-center gap-1 text-xs font-mono",
          isAllUsers ? "text-text-muted" : "text-text-secondary",
        )}
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

type PublicKeysParams = {
  page: number;
  search: string;
};

function nameOrFingerprintFilter(search: string): string {
  return toBase64Json([
    { type: "operator", params: { name: "or" } },
    {
      type: "property",
      params: { name: "name", operator: "contains", value: search },
    },
    { type: "operator", params: { name: "or" } },
    {
      type: "property",
      params: { name: "fingerprint", operator: "contains", value: search },
    },
  ]);
}

const DEFAULTS: PublicKeysParams = {
  page: 1,
  search: "",
};

/**
 * The public keys list: the keys that may authenticate to this namespace's devices.
 */
export default function PublicKeys() {
  const sectionTitle = useNavSectionTitle("/sshkeys/public-keys");
  const {
    params, appliedSearch: searchTerm, setSearch, rows: publicKeys, totalCount, isLoading, error,
    pagination,
  } = usePagedList({
    defaults: DEFAULTS,
    query: ({ page, search }: PublicKeysParams) =>
      paginatedQueryOptions(getPublicKeys, getPublicKeysQueryKey, {
        query: {
          page,
          per_page: PER_PAGE,
          ...(search ? { filter: nameOrFingerprintFilter(search) } : {}),
        },
      }),
  });
  const invalidateKeys = useInvalidateByIds("getPublicKeys");
  const deleteKey = useMutation({
    ...deletePublicKeyMutation(),
    onSuccess: invalidateKeys,
  });
  const [modalOpen, setModalOpen] = useState(false);
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
      closeDelete();
    } catch (err) {
      setDeleteError(
        err instanceof Error ? err.message : "Failed to delete public key.",
      );
    }
  };

  const openNew = () => {
    setEditTarget(null);
    setModalOpen(true);
  };
  const openEdit = (key: PublicKey) => {
    setEditTarget(key);
    setModalOpen(true);
  };
  const closeModal = () => {
    setModalOpen(false);
    setEditTarget(null);
  };

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
          {formatRelative(pk.created_at)}
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

  if (!isLoading && !error && totalCount === 0 && !searchTerm) {
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

        <KeyModal
          open={modalOpen}
          editKey={editTarget}
          onClose={closeModal}
        />
      </>
    );
  }

  return (
    <div>
      <PageHeader
        icon={<KeyIcon className="w-6 h-6" />}
        overline={sectionTitle}
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

      {error && (
        <Callout variant="error" className="mb-4">
          {apiErrorMessage(error)}
        </Callout>
      )}

      <DataTable
        columns={columns}
        data={publicKeys}
        rowKey={(pk) => pk.fingerprint}
        isLoading={isLoading}
        loadingMessage="Loading public keys..."
        {...pagination}
        itemLabel="key"
        emptyMessage={
          searchTerm
            ? `No keys matching \u201C${searchTerm}\u201D`
            : "No public keys found"
        }
      />

      <KeyModal open={modalOpen} editKey={editTarget} onClose={closeModal} />

      <ConfirmDialog
        open={!!deleteTarget}
        onClose={closeDelete}
        onConfirm={confirmDelete}
        icon={<TrashIcon />}
        title="Delete public key"
        description={
          <>
            Whoever holds <ObjectName>{deleteTarget?.name}</ObjectName> can no
            longer connect with it. This can't be undone.
          </>
        }
        confirmLabel="Delete key"
      >
        {deleteError && (
          <p className="text-xs text-accent-red">{deleteError}</p>
        )}
      </ConfirmDialog>
    </div>
  );
}
