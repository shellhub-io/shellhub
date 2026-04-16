import { useState, type FormEvent } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useUpdateUser } from "@/hooks/useAdminUserMutations";
import { useAuthStore } from "@/stores/authStore";
import { isSdkError } from "@/api/errors";
import Drawer from "@/components/common/Drawer";
import { LABEL, INPUT } from "@/utils/styles";
import PasswordInput from "./PasswordInput";
import NamespaceLimitFields from "./NamespaceLimitFields";
import type { UserStatus } from "./UserStatusChip";

export interface EditableUser {
  id: string;
  name: string;
  username: string;
  email: string;
  admin?: boolean;
  max_namespaces?: number;
  status?: UserStatus;
}

interface EditUserDrawerProps {
  open: boolean;
  onClose: () => void;
  user: EditableUser | null;
}

export default function EditUserDrawer({
  open,
  onClose,
  user,
}: EditUserDrawerProps) {
  const updateUser = useUpdateUser();
  const currentUsername = useAuthStore((s) => s.username);

  const [name, setName] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmed, setConfirmed] = useState(false);
  const [admin, setAdmin] = useState(false);
  const [limitEnabled, setLimitEnabled] = useState(false);
  const [limitDisabled, setLimitDisabled] = useState(false);
  const [maxNamespaces, setMaxNamespaces] = useState(1);
  const [error, setError] = useState("");

  useResetOnOpen(open, () => {
    setName(user?.name ?? "");
    setUsername(user?.username ?? "");
    setEmail(user?.email ?? "");
    setPassword("");
    setConfirmed(user?.status === "confirmed");
    setAdmin(user?.admin ?? false);

    const maxNs = user?.max_namespaces;
    if (maxNs !== undefined && maxNs >= 0) {
      setLimitEnabled(true);
      setLimitDisabled(maxNs === 0);
      setMaxNamespaces(maxNs || 1);
    } else {
      setLimitEnabled(false);
      setLimitDisabled(false);
      setMaxNamespaces(1);
    }

    setError("");
  });

  const isConfirmed = user?.status === "confirmed";
  const canChangeStatus = !isConfirmed;
  const isSelf = user?.username === currentUsername;

  const computeMaxNamespaces = (): number | undefined => {
    if (!limitEnabled) {
      const orig = user?.max_namespaces;
      return orig !== undefined && orig < 0 ? orig : undefined;
    }
    if (limitDisabled) return 0;
    return maxNamespaces;
  };

  const canSubmit = name.trim() && username.trim() && email.trim();

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (!canSubmit || !user) return;
    setError("");
    try {
      await updateUser.mutateAsync({
        path: { id: user.id },
        body: {
          name: name.trim(),
          username: username.trim(),
          email: email.trim(),
          password,
          confirmed,
          admin,
          max_namespaces: computeMaxNamespaces(),
        },
      });
      onClose();
    } catch (err) {
      if (isSdkError(err) && err.status === 409) {
        setError("A user with this email or username already exists.");
      } else {
        setError("Failed to update user. Please try again.");
      }
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Edit User"
      subtitle={
        user ? <span className="font-mono">{user.username}</span> : undefined
      }
      footer={(
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={() => void handleSubmit()}
            disabled={!canSubmit || updateUser.isPending}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {updateUser.isPending && (
              <span
                aria-hidden="true"
                className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
              />
            )}
            Save Changes
          </button>
        </>
      )}
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        {/* Name */}
        <div>
          <label className={LABEL} htmlFor="edit-user-name">
            Name
          </label>
          <input
            id="edit-user-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus={open}
            className={INPUT}
          />
        </div>

        {/* Username */}
        <div>
          <label className={LABEL} htmlFor="edit-user-username">
            Username
          </label>
          <input
            id="edit-user-username"
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className={INPUT}
          />
          <p className="text-2xs text-text-muted mt-1.5">
            3-30 characters, letters, numbers, hyphens, dots, underscores, @
          </p>
        </div>

        {/* Email */}
        <div>
          <label className={LABEL} htmlFor="edit-user-email">
            Email
          </label>
          <input
            id="edit-user-email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className={INPUT}
          />
        </div>

        {/* Password */}
        <PasswordInput
          id="edit-user-password"
          value={password}
          onChange={setPassword}
          placeholder="Leave blank to keep current"
          hint="Leave blank to keep the current password"
        />

        {/* Namespace Limit */}
        <NamespaceLimitFields
          idPrefix="edit-user"
          limitEnabled={limitEnabled}
          onLimitEnabledChange={setLimitEnabled}
          limitDisabled={limitDisabled}
          onLimitDisabledChange={setLimitDisabled}
          maxNamespaces={maxNamespaces}
          onMaxNamespacesChange={setMaxNamespaces}
        />

        {/* Confirmed */}
        <label
          className={`flex items-center gap-2 ${canChangeStatus ? "cursor-pointer" : "cursor-not-allowed opacity-dim"}`}
          title={
            isConfirmed
              ? "Cannot remove confirmation from a confirmed user"
              : undefined
          }
        >
          <input
            type="checkbox"
            checked={confirmed}
            onChange={(e) => setConfirmed(e.target.checked)}
            disabled={!canChangeStatus}
            className="w-4 h-4 rounded border-border bg-card text-primary focus:ring-primary/20 disabled:opacity-dim"
          />
          <span className="text-sm text-text-primary">Confirmed</span>
        </label>

        {/* Admin */}
        <label
          className={`flex items-center gap-2 ${isSelf && admin ? "cursor-not-allowed opacity-dim" : "cursor-pointer"}`}
          title={
            isSelf && admin
              ? "Cannot remove your own admin privilege"
              : undefined
          }
        >
          <input
            type="checkbox"
            checked={admin}
            onChange={(e) => setAdmin(e.target.checked)}
            disabled={isSelf && admin}
            className="w-4 h-4 rounded border-border bg-card text-primary focus:ring-primary/20 disabled:opacity-dim"
          />
          <span className="text-sm text-text-primary">Admin user</span>
        </label>

        {/* Error */}
        {error && (
          <p role="alert" className="text-2xs text-accent-red">
            {error}
          </p>
        )}
      </form>
    </Drawer>
  );
}
