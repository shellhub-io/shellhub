import { useEffect, useReducer, FormEvent } from "react";
import {
  LockClosedIcon,
  KeyIcon,
  ChevronDoubleRightIcon,
  ShieldCheckIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { useTerminalStore } from "../stores/terminalStore";
import { useVaultStore } from "../stores/vaultStore";
import { getFingerprint, validatePrivateKey } from "../utils/ssh-keys";
import CopyButton from "./common/CopyButton";
import Drawer from "./common/Drawer";
import { LABEL, INPUT } from "../utils/styles";
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
  | { type: "setKeyError"; value: string | null };

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

  useEffect(() => {
    if (!open) return;
    dispatch({ type: "reset" });
    refreshVault();
  }, [open, refreshVault]);

  const hasVaultKeys = vaultStatus === "unlocked" && vaultKeys.length > 0;
  const effectiveKeySource = hasVaultKeys ? state.keySource : "manual";

  const selectedVaultKey: VaultKeyEntry | undefined = hasVaultKeys
    ? vaultKeys.find((k) => k.id === state.selectedKeyId)
    : undefined;

  const canConnect =
    state.username.trim() &&
    (state.authMethod === "password"
      ? state.password.trim()
      : effectiveKeySource === "vault"
        ? !!selectedVaultKey && (!selectedVaultKey.hasPassphrase || state.passphrase.trim())
        : state.manualKeyValid && (!state.manualKeyEncrypted || state.passphrase.trim()));

  const handleManualKeyChange = (pem: string) => {
    if (!pem.trim()) {
      dispatch({ type: "setManualKey", value: pem, valid: false, encrypted: false });
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

    if (state.authMethod === "password") {
      openTerminal({
        deviceUid,
        deviceName,
        username: state.username.trim(),
        password: state.password,
      });
    } else {
      const key = effectiveKeySource === "vault" && selectedVaultKey
        ? selectedVaultKey.data
        : state.privateKey.trim();
      const phrase = effectiveKeySource === "vault" && selectedVaultKey
        ? (selectedVaultKey.hasPassphrase ? state.passphrase : undefined)
        : (state.manualKeyEncrypted ? state.passphrase : undefined);

      let fingerprint: string;
      try {
        fingerprint = getFingerprint(key, phrase);
      } catch {
        dispatch({ type: "setKeyError", value: "Failed to read private key. Check the key or passphrase." });
        return;
      }
      dispatch({ type: "setKeyError", value: null });

      openTerminal({
        deviceUid,
        deviceName,
        username: state.username.trim(),
        password: "",
        fingerprint,
        privateKey: key,
        passphrase: phrase,
      });
    }
    onClose();
  };

  return (
    <>
      <Drawer
        open={open}
        onClose={onClose}
        title="Connect"
        subtitle={<span className="font-mono">{deviceName}</span>}
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
              type="submit"
              onClick={handleConnect}
              disabled={!canConnect}
              className="px-5 py-2.5 bg-accent-green/90 hover:bg-accent-green text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
            >
              <ChevronDoubleRightIcon className="w-4 h-4" strokeWidth={2} />
              Connect
            </button>
          </>
        }
      >
        <form onSubmit={handleConnect} className="space-y-5">
          {/* SSHID helper */}
          <div className="bg-card border border-border rounded-lg p-3.5">
            <p className={LABEL}>Connect via terminal</p>
            <div className="flex items-center gap-2">
              <code className="text-xs font-mono text-accent-cyan flex-1 truncate">
                ssh {sshid}
              </code>
              <CopyButton text={`ssh ${sshid}`} />
            </div>
          </div>

          <div className="flex items-center gap-3">
            <div className="flex-1 h-px bg-border" />
            <span className="text-2xs text-text-muted font-mono uppercase tracking-wider">
              or connect via web
            </span>
            <div className="flex-1 h-px bg-border" />
          </div>

          {/* Username */}
          <div>
            <label className={LABEL}>Username</label>
            <input
              type="text"
              value={state.username}
              onChange={(e) => dispatch({ type: "setUsername", value: e.target.value })}
              placeholder="e.g. root"
              autoFocus={open}
              className={INPUT}
            />
          </div>

          {/* Auth Method */}
          <div>
            <label className={LABEL}>Authentication</label>
            <div className="space-y-2">
              <button
                type="button"
                onClick={() => dispatch({ type: "setAuthMethod", value: "password" })}
                className={`flex items-start gap-3 w-full px-3.5 py-3 rounded-lg border text-left transition-all ${
                  state.authMethod === "password"
                    ? "bg-primary/[0.06] border-primary/30 ring-1 ring-primary/10"
                    : "bg-card border-border hover:border-border-light hover:bg-hover-subtle"
                }`}
              >
                <div
                  className={`mt-0.5 shrink-0 w-4 h-4 rounded-full border-2 flex items-center justify-center transition-all ${state.authMethod === "password" ? "border-primary" : "border-text-muted/40"}`}
                >
                  {state.authMethod === "password" && (
                    <div className="w-2 h-2 rounded-full bg-primary" />
                  )}
                </div>
                <div className="flex items-start gap-2.5 min-w-0">
                  <span
                    className={`mt-0.5 shrink-0 transition-colors ${state.authMethod === "password" ? "text-primary" : "text-text-muted"}`}
                  >
                    <LockClosedIcon className="w-4 h-4" />
                  </span>
                  <div className="min-w-0">
                    <span
                      className={`block text-sm font-medium transition-colors ${state.authMethod === "password" ? "text-text-primary" : "text-text-secondary"}`}
                    >
                      Password
                    </span>
                    <span className="block text-2xs text-text-muted mt-0.5">
                      Authenticate with your device password.
                    </span>
                  </div>
                </div>
              </button>
              <button
                type="button"
                onClick={() => dispatch({ type: "setAuthMethod", value: "key" })}
                className={`flex items-start gap-3 w-full px-3.5 py-3 rounded-lg border text-left transition-all ${
                  state.authMethod === "key"
                    ? "bg-primary/[0.06] border-primary/30 ring-1 ring-primary/10"
                    : "bg-card border-border hover:border-border-light hover:bg-hover-subtle"
                }`}
              >
                <div
                  className={`mt-0.5 shrink-0 w-4 h-4 rounded-full border-2 flex items-center justify-center transition-all ${state.authMethod === "key" ? "border-primary" : "border-text-muted/40"}`}
                >
                  {state.authMethod === "key" && (
                    <div className="w-2 h-2 rounded-full bg-primary" />
                  )}
                </div>
                <div className="flex items-start gap-2.5 min-w-0">
                  <span
                    className={`mt-0.5 shrink-0 transition-colors ${state.authMethod === "key" ? "text-primary" : "text-text-muted"}`}
                  >
                    <KeyIcon className="w-4 h-4" />
                  </span>
                  <div className="min-w-0">
                    <span
                      className={`block text-sm font-medium transition-colors ${state.authMethod === "key" ? "text-text-primary" : "text-text-secondary"}`}
                    >
                      Private Key
                    </span>
                    <span className="block text-2xs text-text-muted mt-0.5">
                      Authenticate using your SSH private key.
                    </span>
                  </div>
                </div>
              </button>
            </div>
          </div>

          {/* Password field */}
          {state.authMethod === "password" && (
            <div>
              <label className={LABEL}>Password</label>
              <input
                type="password"
                value={state.password}
                onChange={(e) => dispatch({ type: "setPassword", value: e.target.value })}
                placeholder="Enter device password"
                className={INPUT}
              />
            </div>
          )}

          {/* Private Key fields */}
          {state.authMethod === "key" && (
            <>
              {/* Key source toggle (only if vault has keys) */}
              {hasVaultKeys && (
                <div>
                  <label className={LABEL}>Key Source</label>
                  <div className="flex gap-1 p-0.5 bg-card border border-border rounded-lg">
                    <button
                      type="button"
                      onClick={() => dispatch({ type: "setKeySource", value: "vault" })}
                      className={`flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium transition-all ${
                        state.keySource === "vault"
                          ? "bg-primary/10 text-primary border border-primary/20"
                          : "text-text-secondary hover:text-text-primary"
                      }`}
                    >
                      <ShieldCheckIcon className="w-3.5 h-3.5" />
                      Vault
                    </button>
                    <button
                      type="button"
                      onClick={() => dispatch({ type: "setKeySource", value: "manual" })}
                      className={`flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium transition-all ${
                        state.keySource === "manual"
                          ? "bg-primary/10 text-primary border border-primary/20"
                          : "text-text-secondary hover:text-text-primary"
                      }`}
                    >
                      <KeyIcon className="w-3.5 h-3.5" />
                      Manual
                    </button>
                  </div>
                </div>
              )}

              {/* Vault key selector */}
              {effectiveKeySource === "vault" ? (
                <>
                  <div>
                    <label className={LABEL}>Select Key</label>
                    <select
                      value={state.selectedKeyId}
                      onChange={(e) => dispatch({ type: "setSelectedKeyId", value: e.target.value })}
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
                    <div>
                      <label className={LABEL}>Passphrase</label>
                      <input
                        type="password"
                        value={state.passphrase}
                        onChange={(e) => dispatch({ type: "setPassphrase", value: e.target.value })}
                        placeholder="Key passphrase"
                        className={INPUT}
                      />
                    </div>
                  )}
                </>
              ) : (
                <>
                  {/* Manual key input */}
                  <div>
                    <label className={LABEL}>Private Key</label>
                    <textarea
                      value={state.privateKey}
                      onChange={(e) => handleManualKeyChange(e.target.value)}
                      placeholder={"-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
                      rows={5}
                      className={`${INPUT} font-mono text-xs resize-none`}
                    />
                  </div>
                  {state.manualKeyEncrypted && (
                    <div>
                      <label className={LABEL}>Passphrase</label>
                      <input
                        type="password"
                        autoComplete="off"
                        value={state.passphrase}
                        onChange={(e) => dispatch({ type: "setPassphrase", value: e.target.value })}
                        placeholder="Enter passphrase for encrypted key"
                        className={INPUT}
                      />
                      <p className="text-2xs text-text-muted mt-1.5">
                        This key is encrypted and requires a passphrase.
                      </p>
                    </div>
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
        </form>
      </Drawer>

    </>
  );
}
