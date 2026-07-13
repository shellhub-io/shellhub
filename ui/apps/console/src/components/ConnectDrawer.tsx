import { useEffect, useReducer, useState, FormEvent } from "react";
import {
  LockClosedIcon,
  KeyIcon,
  ChevronDoubleRightIcon,
  ShieldCheckIcon,
  ExclamationCircleIcon,
  VideoCameraIcon,
  CheckIcon,
} from "@heroicons/react/24/outline";
import { useTerminalStore } from "../stores/terminalStore";
import type { TerminalSession } from "../stores/terminalStore";
import { useVaultStore } from "../stores/vaultStore";
import { useAuthStore } from "../stores/authStore";
import { useNamespace } from "../hooks/useNamespaces";
import { getFingerprint, validatePrivateKey } from "../utils/sshKeys";
import { isRecordingSupported } from "../utils/recordings";
import CopyButton from "./common/CopyButton";
import Drawer from "./common/Drawer";
import VaultLockedBanner from "./vault/VaultLockedBanner";
import VaultUnlockDialog from "./vault/VaultUnlockDialog";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import FieldLabel from "@/components/common/fields/FieldLabel";
import RadioCard from "@/components/common/fields/RadioCard";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import RadioSegment from "@/components/common/fields/RadioSegment";
import { INPUT, LABEL } from "../utils/styles";
import { cn } from "@shellhub/design-system/cn";
import { Card, Button } from "@shellhub/design-system/primitives";
import type { VaultKeyEntry } from "../types/vault";

interface Props {
  open: boolean;
  onClose: () => void;
  deviceUid: string;
  deviceName: string;
  sshid: string;
}

interface FormState {
  username: string;
  authMethod: "password" | "key";
  password: string;
  keySource: "vault" | "manual";
  selectedKeyId: string;
  privateKey: string;
  manualKeyValid: boolean;
  manualKeyEncrypted: boolean;
  passphrase: string;
  keyError: string | null;
  recordSession: boolean;
}

type FormAction =
  | { type: "reset" }
  | { type: "setUsername"; value: string }
  | { type: "setAuthMethod"; value: "password" | "key" }
  | { type: "setPassword"; value: string }
  | { type: "setKeySource"; value: "vault" | "manual" }
  | { type: "setSelectedKeyId"; value: string }
  | { type: "setManualKey"; value: string; valid: boolean; encrypted: boolean }
  | { type: "setPassphrase"; value: string }
  | { type: "setKeyError"; value: string | null }
  | { type: "setRecordSession"; value: boolean };

const initialState: FormState = {
  username: "",
  authMethod: "password",
  password: "",
  keySource: "vault",
  selectedKeyId: "",
  privateKey: "",
  manualKeyValid: false,
  manualKeyEncrypted: false,
  passphrase: "",
  keyError: null,
  recordSession: true,
};

function formReducer(state: FormState, action: FormAction): FormState {
  switch (action.type) {
    case "reset":
      return initialState;
    case "setUsername":
      return { ...state, username: action.value };
    case "setAuthMethod":
      return { ...state, authMethod: action.value };
    case "setPassword":
      return { ...state, password: action.value };
    case "setKeySource":
      return { ...state, keySource: action.value };
    case "setSelectedKeyId":
      return { ...state, selectedKeyId: action.value, passphrase: "" };
    case "setManualKey":
      return {
        ...state,
        privateKey: action.value,
        manualKeyValid: action.valid,
        manualKeyEncrypted: action.encrypted,
        passphrase: "",
      };
    case "setPassphrase":
      return { ...state, passphrase: action.value };
    case "setKeyError":
      return { ...state, keyError: action.value };
    case "setRecordSession":
      return { ...state, recordSession: action.value };
  }
}

export default function ConnectDrawer({
  open,
  onClose,
  deviceUid,
  deviceName,
  sshid,
}: Props) {
  const openTerminal = useTerminalStore((s) => s.open);
  const vaultStatus = useVaultStore((s) => s.status);
  const vaultKeys = useVaultStore((s) => s.keys);
  const refreshVault = useVaultStore((s) => s.refreshStatus);

  const [state, dispatch] = useReducer(formReducer, initialState);
  const [unlockOpen, setUnlockOpen] = useState(false);
  const recordingSupported = isRecordingSupported();

  const tenant = useAuthStore((s) => s.tenant);
  const { namespace } = useNamespace(tenant ?? "");
  const namespaceRecords = namespace?.settings?.session_record ?? false;

  useEffect(() => {
    if (!open) return;
    dispatch({ type: "reset" });
    void refreshVault();
  }, [open, refreshVault]);

  const hasVaultKeys = vaultStatus === "unlocked" && vaultKeys.length > 0;
  const effectiveKeySource = hasVaultKeys ? state.keySource : "manual";

  const selectedVaultKey: VaultKeyEntry | undefined = hasVaultKeys
    ? vaultKeys.find((k) => k.id === state.selectedKeyId)
    : undefined;

  const canConnect =
    state.username.trim().length > 0 &&
    (state.authMethod === "password"
      ? state.password.trim().length > 0
      : effectiveKeySource === "vault"
        ? !!selectedVaultKey &&
          (!selectedVaultKey.hasPassphrase ||
            state.passphrase.trim().length > 0)
        : state.manualKeyValid &&
          (!state.manualKeyEncrypted || state.passphrase.trim().length > 0));

  const handleManualKeyChange = (pem: string) => {
    if (!pem.trim()) {
      dispatch({
        type: "setManualKey",
        value: pem,
        valid: false,
        encrypted: false,
      });
      return;
    }
    const result = validatePrivateKey(pem.trim());
    dispatch({
      type: "setManualKey",
      value: pem,
      valid: result.valid,
      encrypted: result.valid && result.encrypted,
    });
  };

  const handleConnect = (e: FormEvent) => {
    e.preventDefault();
    if (!canConnect) return;

    let params: Omit<TerminalSession, "id" | "state" | "connectionStatus">;
    if (state.authMethod === "password") {
      params = {
        deviceUid,
        deviceName,
        username: state.username.trim(),
        password: state.password,
      };
    } else {
      const key =
        effectiveKeySource === "vault" && selectedVaultKey
          ? selectedVaultKey.data
          : state.privateKey.trim();
      const phrase =
        effectiveKeySource === "vault" && selectedVaultKey
          ? selectedVaultKey.hasPassphrase
            ? state.passphrase
            : undefined
          : state.manualKeyEncrypted
            ? state.passphrase
            : undefined;

      let fingerprint: string;
      try {
        fingerprint = getFingerprint(key, phrase);
      } catch {
        dispatch({
          type: "setKeyError",
          value: "Failed to read private key. Check the key or passphrase.",
        });
        return;
      }
      if (
        effectiveKeySource === "vault" &&
        selectedVaultKey &&
        fingerprint !== selectedVaultKey.fingerprint
      ) {
        dispatch({
          type: "setKeyError",
          value:
            "Key data appears corrupted. Try re-importing the key into the vault.",
        });
        return;
      }
      dispatch({ type: "setKeyError", value: null });

      params = {
        deviceUid,
        deviceName,
        username: state.username.trim(),
        password: "",
        fingerprint,
        privateKey: key,
        passphrase: phrase,
      };
    }

    // Opt-in recording. Captured client-side to OPFS — no picker, no upload.
    // Skip when the namespace already records server-side.
    if (!namespaceRecords && state.recordSession && isRecordingSupported()) {
      params = { ...params, record: true };
    }

    openTerminal(params);
    onClose();
  };

  return (
    <>
      <VaultUnlockDialog
        open={unlockOpen}
        onClose={() => setUnlockOpen(false)}
      />
      <Drawer
        open={open}
        onClose={onClose}
        title="Connect"
        subtitle={<span className="font-mono">{deviceName}</span>}
        footer={
          <>
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button
              variant="success"
              type="submit"
              form={`connect-form-${deviceUid}`}
              disabled={!canConnect}
              icon={
                <ChevronDoubleRightIcon className="w-4 h-4" strokeWidth={2} />
              }
            >
              Connect
            </Button>
          </>
        }
      >
        <form
          id={`connect-form-${deviceUid}`}
          onSubmit={handleConnect}
          className="space-y-5"
        >
          {/* SSHID helper */}
          <Card className="rounded-lg p-3.5">
            <span className={LABEL}>Connect via terminal</span>
            <div className="flex items-center gap-2">
              <code className="text-xs font-mono flex-1 truncate">
                <span className="text-accent-cyan">ssh </span>
                {state.username.trim() ? (
                  <span className="text-accent-cyan">
                    {state.username.trim()}@{sshid}
                  </span>
                ) : (
                  <>
                    <span className="text-text-muted italic">
                      &lt;username&gt;
                    </span>
                    <span className="text-accent-cyan">@{sshid}</span>
                  </>
                )}
              </code>
              <CopyButton
                text={
                  state.username.trim()
                    ? `ssh ${state.username.trim()}@${sshid}`
                    : `ssh <username>@${sshid}`
                }
              />
            </div>
            {state.username.trim() ? (
              <p className="text-2xs text-accent-green mt-2">
                Command ready — copy and run in your terminal.
              </p>
            ) : (
              <p className="text-2xs text-text-muted mt-2">
                Enter your device OS username below to complete this command.
              </p>
            )}
          </Card>

          <div className="flex items-center gap-3">
            <div className="flex-1 h-px bg-border" />
            <span className="text-2xs text-text-secondary font-medium uppercase tracking-wider">
              or connect via web
            </span>
            <div className="flex-1 h-px bg-border" />
          </div>

          <InputField
            id="connect-username"
            label="Username"
            value={state.username}
            onChange={(v) => dispatch({ type: "setUsername", value: v })}
            placeholder="e.g. root"
          />

          {/* Auth Method */}
          <RadioGroupField
            label="Authentication"
            value={state.authMethod}
            onChange={(v) => dispatch({ type: "setAuthMethod", value: v })}
          >
            <RadioCard
              value="password"
              icon={<LockClosedIcon className="w-4 h-4" />}
              label="Password"
              description="Authenticate with your device password."
            />
            <RadioCard
              value="key"
              icon={<KeyIcon className="w-4 h-4" />}
              label="Private Key"
              description="Authenticate using your SSH private key."
            />
          </RadioGroupField>

          {/* Password field */}
          {state.authMethod === "password" && (
            <PasswordField
              id="connect-password"
              label="Password"
              autoComplete="current-password"
              value={state.password}
              onChange={(v) => dispatch({ type: "setPassword", value: v })}
              placeholder="Enter device password"
            />
          )}

          {/* Private Key fields */}
          {state.authMethod === "key" && (
            <>
              {/* Vault locked warning */}
              {vaultStatus === "locked" && (
                <VaultLockedBanner onUnlock={() => setUnlockOpen(true)} />
              )}

              {/* Key source toggle (only if vault has keys) */}
              {hasVaultKeys && (
                <RadioGroupField
                  label="Key Source"
                  value={state.keySource}
                  onChange={(value) =>
                    dispatch({ type: "setKeySource", value })
                  }
                  containerClassName="flex gap-1 p-0.5 bg-card border border-border rounded-lg"
                >
                  <RadioSegment
                    value="vault"
                    label="Vault"
                    icon={<ShieldCheckIcon className="w-3.5 h-3.5" />}
                  />
                  <RadioSegment
                    value="manual"
                    label="Manual"
                    icon={<KeyIcon className="w-3.5 h-3.5" />}
                  />
                </RadioGroupField>
              )}

              {/* Vault key selector */}
              {effectiveKeySource === "vault" ? (
                <>
                  <div>
                    <FieldLabel htmlFor="connect-vault-key">
                      Select Key
                    </FieldLabel>
                    <select
                      id="connect-vault-key"
                      value={state.selectedKeyId}
                      onChange={(e) =>
                        dispatch({
                          type: "setSelectedKeyId",
                          value: e.target.value,
                        })
                      }
                      className={INPUT}
                    >
                      <option value="">Choose a key...</option>
                      {vaultKeys.map((k) => (
                        <option key={k.id} value={k.id}>
                          {k.name}
                        </option>
                      ))}
                    </select>
                  </div>
                  {selectedVaultKey?.hasPassphrase && (
                    <PasswordField
                      id="connect-vault-passphrase"
                      label="Passphrase"
                      value={state.passphrase}
                      onChange={(v) =>
                        dispatch({ type: "setPassphrase", value: v })
                      }
                      placeholder="Key passphrase"
                      suppressPasswordManager
                    />
                  )}
                </>
              ) : (
                <>
                  {/* Manual key input */}
                  <div>
                    <FieldLabel htmlFor="connect-manual-private-key">
                      Private Key
                    </FieldLabel>
                    <textarea
                      id="connect-manual-private-key"
                      value={state.privateKey}
                      onChange={(e) => handleManualKeyChange(e.target.value)}
                      placeholder={"-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
                      rows={5}
                      className={cn(INPUT, "font-mono text-xs resize-none")}
                    />
                  </div>
                  {state.manualKeyEncrypted && (
                    <PasswordField
                      id="connect-manual-passphrase"
                      label="Passphrase"
                      value={state.passphrase}
                      onChange={(v) =>
                        dispatch({ type: "setPassphrase", value: v })
                      }
                      placeholder="Enter passphrase for encrypted key"
                      suppressPasswordManager
                      hint="This key is encrypted and requires a passphrase."
                    />
                  )}
                </>
              )}
            </>
          )}

          {state.keyError && (
            <p className="text-2xs text-accent-red flex items-center gap-1">
              <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" />
              {state.keyError}
            </p>
          )}

          {namespaceRecords && (
            <div className="w-full px-3.5 py-3 rounded-lg border border-border bg-card text-left">
              <div className="flex items-start gap-3">
                <span
                  aria-hidden="true"
                  className="mt-0.5 shrink-0 text-text-secondary"
                >
                  <VideoCameraIcon className="w-5 h-5" />
                </span>
                <div className="min-w-0 flex-1">
                  <span className="flex items-center gap-1.5 text-sm font-medium text-text-primary">
                    <span className="inline-flex items-center gap-1 text-accent-red">
                      <span className="w-1.5 h-1.5 rounded-full bg-accent-red animate-pulse-subtle" />
                      <span className="text-[10px] font-bold tracking-wide">
                        REC
                      </span>
                    </span>
                    Session recording is on
                  </span>
                  <span className="block text-2xs text-text-muted mt-0.5">
                    Recorded on the server by your namespace's policy.
                  </span>
                </div>
              </div>
            </div>
          )}

          {!namespaceRecords && recordingSupported && (
            <label
              className={cn("flex items-start gap-3 w-full px-3.5 py-3 rounded-lg border text-left transition-all cursor-pointer focus-within:ring-2 focus-within:ring-primary/40", state.recordSession ? "bg-primary/[0.06] border-primary/30 ring-1 ring-primary/10" : "bg-card border-border hover:border-border-light hover:bg-hover-subtle")}
            >
              <input
                type="checkbox"
                checked={state.recordSession}
                onChange={(e) =>
                  dispatch({
                    type: "setRecordSession",
                    value: e.target.checked,
                  })
                }
                className="sr-only"
              />
              <span
                aria-hidden="true"
                className={cn("mt-0.5 shrink-0 transition-colors", state.recordSession ? "text-primary" : "text-text-muted")}
              >
                <VideoCameraIcon className="w-4 h-4" />
              </span>
              <div className="min-w-0 flex-1">
                <span className="block text-sm font-medium text-text-primary">
                  Record this session
                </span>
                <span className="block text-2xs text-text-muted mt-0.5">
                  Save this session in your browser to replay it locally later.
                </span>
              </div>
              <span
                aria-hidden="true"
                className={cn("mt-0.5 shrink-0 w-4 h-4 rounded border flex items-center justify-center transition-all", state.recordSession ? "bg-primary border-primary text-white" : "border-text-muted/40")}
              >
                {state.recordSession && (
                  <CheckIcon className="w-3 h-3" strokeWidth={3} />
                )}
              </span>
            </label>
          )}
        </form>
      </Drawer>
    </>
  );
}
