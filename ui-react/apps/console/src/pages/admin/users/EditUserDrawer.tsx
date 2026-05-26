import { useState, type FormEvent } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useUpdateUser } from "@/hooks/useAdminUserMutations";
import { useAuthStore } from "@/stores/authStore";
import { isSdkError } from "@/api/errors";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import CheckboxField from "@/components/common/fields/CheckboxField";
import NamespaceLimitFields from "./NamespaceLimitFields";
import { isMaxNamespacesValid } from "@/utils/validation";
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
  const [maxNamespaces, setMaxNamespaces] = useState("1");
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
      setMaxNamespaces(String(maxNs || 1));
    } else {
      setLimitEnabled(false);
      setLimitDisabled(false);
      setMaxNamespaces("1");
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
    return parseInt(maxNamespaces, 10);
  };

  const canSubmit =
    name.trim() &&
    username.trim() &&
    email.trim() &&
    isMaxNamespacesValid(limitEnabled, limitDisabled, maxNamespaces);

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
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            type="button"
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
      }
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        <InputField
          id="edit-user-name"
          label="Name"
          value={name}
          onChange={setName}
          autoFocus={open}
        />

        <InputField
          id="edit-user-username"
          label="Username"
          value={username}
          onChange={setUsername}
          hint="3-30 characters, letters, numbers, hyphens, dots, underscores, @"
        />

        <InputField
          id="edit-user-email"
          label="Email"
          type="email"
          value={email}
          onChange={setEmail}
        />

        <PasswordField
          id="edit-user-password"
          label="Password"
          value={password}
          onChange={setPassword}
          placeholder="Leave blank to keep current"
          hint="Leave blank to keep the current password"
          suppressPasswordManager
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
        <CheckboxField
          id="edit-user-confirmed"
          label="Confirmed"
          checked={confirmed}
          onChange={setConfirmed}
          disabled={!canChangeStatus}
          title={
            isConfirmed
              ? "Cannot remove confirmation from a confirmed user"
              : undefined
          }
        />

        {/* Admin */}
        <CheckboxField
          id="edit-user-admin"
          label="Admin user"
          checked={admin}
          onChange={setAdmin}
          disabled={isSelf && admin}
          title={
            isSelf && admin
              ? "Cannot remove your own admin privilege"
              : undefined
          }
        />

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
