import { useEffect, useRef, useState } from "react";
import {
  ShieldCheckIcon,
  LockClosedIcon,
  KeyIcon,
  PlusIcon,
  PencilSquareIcon,
  TrashIcon,
  FingerPrintIcon,
} from "@heroicons/react/24/outline";
import { useVaultStore } from "@/stores/vaultStore";
import PageHeader from "@/components/common/PageHeader";
import EmptyState, {
  type EmptyStateFeature,
} from "@/components/common/EmptyState";
import CopyButton from "@/components/common/CopyButton";
import VaultSetupDialog from "@/components/vault/VaultSetupDialog";
import VaultUnlockDialog from "@/components/vault/VaultUnlockDialog";
import VaultSettingsSection from "@/components/vault/VaultSettingsSection";
import VaultSyncDialog from "@/components/vault/VaultSyncDialog";
import VaultSyncPromoDialog from "@/components/vault/VaultSyncPromoDialog";
import { useAuthStore } from "@/stores/authStore";
import {
  isVaultServerEnabled,
  isVaultSyncPromoDismissed,
} from "@/utils/vault-backend-factory";
import DataTable, { type Column } from "@/components/common/DataTable";
import SearchField from "@/components/common/fields/SearchField";
import KeyDrawer from "./KeyDrawer";
import KeyDeleteDialog from "./KeyDeleteDialog";
import { formatDate } from "@/utils/date";
import type { VaultKeyEntry } from "@/types/vault";
import { Button, IconButton } from "@shellhub/design-system/primitives";

const VAULT_FEATURES: EmptyStateFeature[] = [
  {
    icon: <LockClosedIcon className="w-5 h-5" />,
    title: "AES-256 Encryption",
    description:
      "Keys are encrypted with AES-256-GCM, derived from your master password.",
  },
  {
    icon: <FingerPrintIcon className="w-5 h-5" />,
    title: "Zero Knowledge",
    description:
      "Your master password is never stored — only you can unlock the vault.",
  },
  {
    icon: <KeyIcon className="w-5 h-5" />,
    title: "Quick Connect",
    description:
      "Select stored keys when connecting to devices — no more copy-pasting.",
  },
];

export default function SecureVault() {
  const status = useVaultStore((s) => s.status);
  const keys = useVaultStore((s) => s.keys);
  const refreshStatus = useVaultStore((s) => s.refreshStatus);
  const autoLockNonce = useVaultStore((s) => s.autoLockNonce);
  const storageMode = useVaultStore((s) => s.storageMode);
  const user = useAuthStore((s) => s.user);
  const tenant = useAuthStore((s) => s.tenant);
  const [setupOpen, setSetupOpen] = useState(false);
  const [unlockOpen, setUnlockOpen] = useState(false);
  const [syncOpen, setSyncOpen] = useState(false);
  const [promoOpen, setPromoOpen] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<VaultKeyEntry | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<VaultKeyEntry | null>(null);
  const [search, setSearch] = useState("");

  useEffect(() => {
    void refreshStatus();
  }, [refreshStatus]);

  const seenNonce = useRef(autoLockNonce);
  useEffect(() => {
    if (autoLockNonce !== seenNonce.current) {
      seenNonce.current = autoLockNonce;
      setUnlockOpen(true);
    }
  }, [autoLockNonce]);

  // Offer syncing a local vault to the ShellHub server right after the user
  // unlocks it: they just proved ownership and are looking at their keys, so
  // it's the natural moment to pitch portability. Fires only on the
  // locked -> unlocked transition (a real unlock), never on setup
  // (uninitialized -> unlocked, where the user just picked a location) nor on
  // a plain page open of an already-unlocked vault.
  const prevStatus = useRef(status);
  useEffect(() => {
    const cameFromLocked = prevStatus.current === "locked";
    prevStatus.current = status;

    if (
      cameFromLocked &&
      status === "unlocked" &&
      isVaultServerEnabled() &&
      storageMode === "local" &&
      !isVaultSyncPromoDismissed(user && tenant ? { user, tenant } : undefined)
    ) {
      setPromoOpen(true);
    }
  }, [status, storageMode, user, tenant]);

  const openNew = () => {
    setEditTarget(null);
    setDrawerOpen(true);
  };
  const openEdit = (entry: VaultKeyEntry) => {
    setEditTarget(entry);
    setDrawerOpen(true);
  };
  const closeDrawer = () => {
    setDrawerOpen(false);
    setEditTarget(null);
  };

  const filtered = search
    ? keys.filter(
        (k) =>
          k.name.toLowerCase().includes(search.toLowerCase()) ||
          k.fingerprint.toLowerCase().includes(search.toLowerCase()),
      )
    : keys;

  if (status === "uninitialized") {
    return (
      <>
        <EmptyState
          icon={<ShieldCheckIcon className="w-8 h-8" />}
          overline="Encrypted Key Storage"
          title="Secure Vault"
          description="Store and encrypt your SSH private keys with a master password. Your keys never leave your browser and are protected at rest."
          features={VAULT_FEATURES}
        >
          <Button
            size="lg"
            onClick={() => setSetupOpen(true)}
            icon={<ShieldCheckIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Set Up Secure Vault
          </Button>
        </EmptyState>
        <VaultSetupDialog
          open={setupOpen}
          onClose={() => setSetupOpen(false)}
        />
      </>
    );
  }

  if (status === "locked") {
    return (
      <>
        <EmptyState
          accent="yellow"
          icon={<LockClosedIcon className="w-8 h-8" />}
          overline="Vault Locked"
          title="Your vault is locked"
          description="Enter your master password to access your SSH keys and connect to devices."
          features={VAULT_FEATURES}
        >
          <Button
            size="lg"
            onClick={() => setUnlockOpen(true)}
            icon={<LockClosedIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Unlock Vault
          </Button>
        </EmptyState>
        <VaultUnlockDialog
          open={unlockOpen}
          onClose={() => setUnlockOpen(false)}
        />
      </>
    );
  }

  const columns: Column<VaultKeyEntry>[] = [
    {
      key: "name",
      header: "Name",
      render: (entry) => (
        <div className="flex items-center gap-2">
          <KeyIcon className="w-4 h-4 text-text-muted shrink-0" />
          <span className="text-sm font-medium text-text-primary">
            {entry.name}
          </span>
          {entry.hasPassphrase && (
            <LockClosedIcon
              className="w-3 h-3 text-accent-yellow shrink-0"
              strokeWidth={2}
              title="Encrypted"
            />
          )}
        </div>
      ),
    },
    {
      key: "fingerprint",
      header: "Fingerprint",
      render: (entry) => (
        <div className="flex items-center gap-1">
          <code
            className="text-2xs font-mono text-text-muted truncate max-w-[200px]"
            title={entry.fingerprint}
          >
            {entry.fingerprint}
          </code>
          <CopyButton text={entry.fingerprint} />
        </div>
      ),
    },
    {
      key: "algorithm",
      header: "Algorithm",
      render: (entry) => (
        <span className="text-xs font-mono text-text-secondary">
          {entry.algorithm ?? "\u2014"}
        </span>
      ),
    },
    {
      key: "added",
      header: "Added",
      render: (entry) => (
        <span className="text-xs font-mono text-text-muted">
          {formatDate(entry.createdAt)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "",
      headerClassName: "w-16",
      render: (entry) => (
        <div className="flex items-center justify-end gap-0.5">
          <IconButton
            variant="primary"
            title="Edit"
            aria-label={`Edit ${entry.name}`}
            onClick={() => openEdit(entry)}
          >
            <PencilSquareIcon className="w-4 h-4" strokeWidth={2} />
          </IconButton>
          <IconButton
            variant="danger"
            title="Delete"
            aria-label={`Delete ${entry.name}`}
            onClick={() => setDeleteTarget(entry)}
          >
            <TrashIcon className="w-4 h-4" strokeWidth={2} />
          </IconButton>
        </div>
      ),
    },
  ];

  return (
    <div>
      {status === "unlocked" && (
        <>
          <PageHeader
            icon={<ShieldCheckIcon className="w-6 h-6" />}
            overline="Security"
            title="Secure Vault"
            description="Manage your encrypted SSH private keys."
          >
            <Button
              onClick={openNew}
              icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
            >
              Add Private Key
            </Button>
          </PageHeader>

          <SearchField
            className="mb-4"
            value={search}
            onChange={setSearch}
            placeholder="Search by name or fingerprint..."
            aria-label="Search vault keys by name or fingerprint"
          />

          <DataTable
            columns={columns}
            data={filtered}
            rowKey={(entry) => entry.id}
            emptyState={
              <div className="text-center">
                <KeyIcon className="w-8 h-8 text-text-muted/30 mx-auto mb-2" />
                <p className="text-xs font-mono text-text-muted">
                  {search
                    ? `No keys matching "${search}"`
                    : "No keys yet. Add your first SSH private key."}
                </p>
              </div>
            }
          />
        </>
      )}

      <VaultSettingsSection />

      <KeyDrawer open={drawerOpen} editKey={editTarget} onClose={closeDrawer} />
      <KeyDeleteDialog
        open={!!deleteTarget}
        entry={deleteTarget}
        onClose={() => setDeleteTarget(null)}
      />
      <VaultSyncPromoDialog
        open={promoOpen}
        onClose={() => setPromoOpen(false)}
        onSync={() => setSyncOpen(true)}
      />
      <VaultSyncDialog
        open={syncOpen}
        onClose={() => setSyncOpen(false)}
        direction="to-server"
      />
    </div>
  );
}
