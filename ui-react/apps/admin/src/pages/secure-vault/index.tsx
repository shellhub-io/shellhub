import { useEffect, useState } from "react";
import {
  ShieldCheckIcon,
  LockClosedIcon,
  KeyIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  PencilSquareIcon,
  TrashIcon,
  FingerPrintIcon,
} from "@heroicons/react/24/outline";
import { useVaultStore } from "@/stores/vaultStore";
import PageHeader from "@/components/common/PageHeader";
import CopyButton from "@/components/common/CopyButton";
import VaultSetupDialog from "@/components/vault/VaultSetupDialog";
import VaultUnlockDialog from "@/components/vault/VaultUnlockDialog";
import VaultLockedBanner from "@/components/vault/VaultLockedBanner";
import VaultSettingsSection from "@/components/vault/VaultSettingsSection";
import KeyDrawer from "./KeyDrawer";
import KeyDeleteDialog from "./KeyDeleteDialog";
import { formatDate } from "@/utils/date";
import { TH } from "@/utils/styles";
import type { VaultKeyEntry } from "@/types/vault";

function KeyRow({
  entry,
  onEdit,
  onDelete,
}: {
  entry: VaultKeyEntry;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <tr className="group border-b border-border/60 hover:bg-hover-subtle transition-colors">
      <td className="px-4 py-3.5">
        <div className="flex items-center gap-2">
          <KeyIcon className="w-4 h-4 text-text-muted shrink-0" />
          <span className="text-sm font-medium text-text-primary">
            {entry.name}
          </span>
        </div>
      </td>
      <td className="px-4 py-3.5">
        <div className="flex items-center gap-1">
          <code
            className="text-2xs font-mono text-text-muted truncate max-w-[200px]"
            title={entry.fingerprint}
          >
            {entry.fingerprint}
          </code>
          <CopyButton text={entry.fingerprint} />
        </div>
      </td>
      <td className="px-4 py-3.5">
        {entry.hasPassphrase ? (
          <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-accent-yellow/10 text-accent-yellow text-2xs font-mono rounded">
            <LockClosedIcon className="w-2.5 h-2.5" strokeWidth={2} />
            Protected
          </span>
        ) : (
          <span className="text-2xs font-mono text-text-muted">None</span>
        )}
      </td>
      <td className="px-4 py-3.5">
        <span className="text-xs font-mono text-text-muted">
          {formatDate(entry.createdAt)}
        </span>
      </td>
      <td className="px-4 py-3.5 text-right">
        <div className="flex items-center justify-end gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            onClick={onEdit}
            title="Edit"
            className="p-1.5 rounded-md text-text-muted hover:text-primary hover:bg-primary/10 transition-all"
          >
            <PencilSquareIcon className="w-4 h-4" strokeWidth={2} />
          </button>
          <button
            onClick={onDelete}
            title="Delete"
            className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/10 transition-all"
          >
            <TrashIcon className="w-4 h-4" strokeWidth={2} />
          </button>
        </div>
      </td>
    </tr>
  );
}

export default function SecureVault() {
  const { status, keys, refreshStatus } = useVaultStore();
  const [setupOpen, setSetupOpen] = useState(false);
  const [unlockOpen, setUnlockOpen] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<VaultKeyEntry | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<VaultKeyEntry | null>(null);
  const [search, setSearch] = useState("");

  useEffect(() => {
    refreshStatus();
  }, [refreshStatus]);

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
        <div className="relative -m-8 flex-1 flex flex-col overflow-hidden">
          <div className="absolute inset-0 pointer-events-none">
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
                  <ShieldCheckIcon className="w-8 h-8 text-primary" />
                </div>
                <span className="inline-block text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
                  Encrypted Key Storage
                </span>
                <h1 className="text-3xl font-bold text-text-primary mb-3">
                  Secure Vault
                </h1>
                <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
                  Store and encrypt your SSH private keys with a master password.
                  Your keys never leave your browser and are protected at rest.
                </p>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-10">
                {[
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
                <button
                  onClick={() => setSetupOpen(true)}
                  className="inline-flex items-center gap-2 px-6 py-3 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all shadow-lg shadow-primary/20"
                >
                  <ShieldCheckIcon className="w-4 h-4" strokeWidth={2} />
                  Set Up Secure Vault
                </button>
              </div>
            </div>
          </div>
        </div>
        <VaultSetupDialog
          open={setupOpen}
          onClose={() => setSetupOpen(false)}
        />
      </>
    );
  }

  return (
    <div>
      {status === "locked" && (
        <div className="mb-6 animate-fade-in">
          <VaultLockedBanner onUnlock={() => setUnlockOpen(true)} />
        </div>
      )}

      {status === "unlocked" && (
        <>
          {keys.length === 0 ? (
            <div className="animate-fade-in">
              <PageHeader
                variant="decorated"
                icon={<ShieldCheckIcon className="w-6 h-6" />}
                overline="Security"
                title="Secure Vault"
                description="Your vault is set up. Add your first SSH private key to get started."
              >
                <button
                  onClick={openNew}
                  className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
                >
                  <PlusIcon className="w-4 h-4" strokeWidth={2} />
                  Add Private Key
                </button>
              </PageHeader>
              <div className="py-16 text-center">
                <KeyIcon className="w-10 h-10 text-text-muted/40 mx-auto mb-3" />
                <p className="text-sm text-text-muted">
                  No keys stored yet. Add your first SSH private key.
                </p>
              </div>
            </div>
          ) : (
            <>
              <PageHeader
                variant="decorated"
                icon={<ShieldCheckIcon className="w-6 h-6" />}
                overline="Security"
                title="Secure Vault"
                description="Manage your encrypted SSH private keys."
              >
                <button
                  onClick={openNew}
                  className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
                >
                  <PlusIcon className="w-4 h-4" strokeWidth={2} />
                  Add Private Key
                </button>
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

              {filtered.length === 0 ? (
                <div className="py-12 text-center animate-fade-in">
                  <p className="text-sm text-text-muted">
                    No keys matching &ldquo;{search}&rdquo;
                  </p>
                </div>
              ) : (
                <div className="bg-card border border-border rounded-lg overflow-hidden animate-fade-in">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b border-border bg-surface/50">
                        <th className={TH}>Name</th>
                        <th className={TH}>Fingerprint</th>
                        <th className={TH}>Passphrase</th>
                        <th className={TH}>Added</th>
                        <th className="px-4 py-3 w-16" />
                      </tr>
                    </thead>
                    <tbody>
                      {filtered.map((entry) => (
                        <KeyRow
                          key={entry.id}
                          entry={entry}
                          onEdit={() => openEdit(entry)}
                          onDelete={() => setDeleteTarget(entry)}
                        />
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </>
          )}
        </>
      )}

      <VaultSettingsSection />

      <KeyDrawer
        open={drawerOpen}
        editKey={editTarget}
        onClose={closeDrawer}
      />
      <KeyDeleteDialog
        open={!!deleteTarget}
        entry={deleteTarget}
        onClose={() => setDeleteTarget(null)}
      />
      <VaultUnlockDialog
        open={unlockOpen}
        onClose={() => setUnlockOpen(false)}
      />
    </div>
  );
}
