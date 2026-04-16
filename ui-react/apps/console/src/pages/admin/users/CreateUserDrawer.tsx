import { useState, type FormEvent } from "react";
import { PlusIcon } from "@heroicons/react/24/outline";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useCreateUser } from "@/hooks/useAdminUserMutations";
import { isSdkError } from "@/api/errors";
import Drawer from "@/components/common/Drawer";
import { LABEL, INPUT } from "@/utils/styles";
import PasswordInput from "./PasswordInput";
import NamespaceLimitFields from "./NamespaceLimitFields";

interface CreateUserDrawerProps {
  open: boolean;
  onClose: () => void;
}

export default function CreateUserDrawer({
  open,
  onClose,
}: CreateUserDrawerProps) {
  const createUser = useCreateUser();

  const [name, setName] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [admin, setAdmin] = useState(false);
  const [limitEnabled, setLimitEnabled] = useState(false);
  const [limitDisabled, setLimitDisabled] = useState(false);
  const [maxNamespaces, setMaxNamespaces] = useState(1);
  const [error, setError] = useState("");

  useResetOnOpen(open, () => {
    setName("");
    setUsername("");
    setEmail("");
    setPassword("");
    setAdmin(false);
    setLimitEnabled(false);
    setLimitDisabled(false);
    setMaxNamespaces(1);
    setError("");
  });

  const computeMaxNamespaces = (): number | undefined => {
    if (!limitEnabled) return undefined;
    if (limitDisabled) return 0;
    return maxNamespaces;
  };

  const canSubmit
    = name.trim() && username.trim() && email.trim() && password.trim();

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (!canSubmit) return;
    setError("");
    try {
      await createUser.mutateAsync({
        body: {
          name: name.trim(),
          username: username.trim(),
          email: email.trim(),
          password,
          admin,
          max_namespaces: computeMaxNamespaces(),
        },
      });
      onClose();
    } catch (err) {
      if (isSdkError(err) && err.status === 409) {
        setError("A user with this email or username already exists.");
      } else {
        setError("Failed to create user. Please try again.");
      }
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Create User"
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
            disabled={!canSubmit || createUser.isPending}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {createUser.isPending ? (
              <span
                aria-hidden="true"
                className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
              />
            ) : (
              <PlusIcon className="w-4 h-4" strokeWidth={2} />
            )}
            Create User
          </button>
        </>
      )}
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        {/* Name */}
        <div>
          <label className={LABEL} htmlFor="create-user-name">
            Name
          </label>
          <input
            id="create-user-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="John Doe"
            autoFocus={open}
            className={INPUT}
          />
        </div>

        {/* Username */}
        <div>
          <label className={LABEL} htmlFor="create-user-username">
            Username
          </label>
          <input
            id="create-user-username"
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="johndoe"
            className={INPUT}
          />
          <p className="text-2xs text-text-muted mt-1.5">
            3-30 characters, letters, numbers, hyphens, dots, underscores, @
          </p>
        </div>

        {/* Email */}
        <div>
          <label className={LABEL} htmlFor="create-user-email">
            Email
          </label>
          <input
            id="create-user-email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="john@example.com"
            className={INPUT}
          />
        </div>

        {/* Password */}
        <PasswordInput
          id="create-user-password"
          value={password}
          onChange={setPassword}
          placeholder="Enter password"
          hint="5-30 characters"
        />

        {/* Namespace Limit */}
        <NamespaceLimitFields
          idPrefix="create-user"
          limitEnabled={limitEnabled}
          onLimitEnabledChange={setLimitEnabled}
          limitDisabled={limitDisabled}
          onLimitDisabledChange={setLimitDisabled}
          maxNamespaces={maxNamespaces}
          onMaxNamespacesChange={setMaxNamespaces}
        />

        {/* Admin */}
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={admin}
            onChange={(e) => setAdmin(e.target.checked)}
            className="w-4 h-4 rounded border-border bg-card text-primary focus:ring-primary/20"
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
