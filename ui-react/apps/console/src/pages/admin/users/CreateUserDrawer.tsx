import { useState, type FormEvent } from "react";
import { PlusIcon } from "@heroicons/react/24/outline";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useCreateUser } from "@/hooks/useAdminUserMutations";
import { isSdkError } from "@/api/errors";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import CheckboxField from "@/components/common/fields/CheckboxField";
import NamespaceLimitFields from "./NamespaceLimitFields";
import { isMaxNamespacesValid } from "@/utils/validation";

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
  const [maxNamespaces, setMaxNamespaces] = useState("1");
  const [error, setError] = useState("");

  useResetOnOpen(open, () => {
    setName("");
    setUsername("");
    setEmail("");
    setPassword("");
    setAdmin(false);
    setLimitEnabled(false);
    setLimitDisabled(false);
    setMaxNamespaces("1");
    setError("");
  });

  const computeMaxNamespaces = (): number | undefined => {
    if (!limitEnabled) return undefined;
    if (limitDisabled) return 0;
    return parseInt(maxNamespaces, 10);
  };

  const canSubmit =
    name.trim() &&
    username.trim() &&
    email.trim() &&
    password.trim() &&
    isMaxNamespacesValid(limitEnabled, limitDisabled, maxNamespaces);

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
      }
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        <InputField
          id="create-user-name"
          label="Name"
          value={name}
          onChange={setName}
          placeholder="John Doe"
          autoFocus={open}
        />

        <InputField
          id="create-user-username"
          label="Username"
          value={username}
          onChange={setUsername}
          placeholder="johndoe"
          hint="3-30 characters, letters, numbers, hyphens, dots, underscores, @"
        />

        <InputField
          id="create-user-email"
          label="Email"
          type="email"
          value={email}
          onChange={setEmail}
          placeholder="john@example.com"
        />

        <PasswordField
          id="create-user-password"
          label="Password"
          value={password}
          onChange={setPassword}
          placeholder="Enter password"
          hint="5-30 characters"
          suppressPasswordManager
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
        <CheckboxField
          id="create-user-admin"
          label="Admin user"
          checked={admin}
          onChange={setAdmin}
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
