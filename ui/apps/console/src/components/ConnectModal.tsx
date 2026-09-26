import { useEffect, useId, useReducer, useState, FormEvent } from "react";
import { useQueryClient } from "@tanstack/react-query";
import {
  LockClosedIcon,
  KeyIcon,
  ShieldCheckIcon,
  FingerPrintIcon,
  ArrowTopRightOnSquareIcon,
  ExclamationCircleIcon,
  CommandLineIcon,
} from "@heroicons/react/24/outline";
import { useTerminalStore } from "../stores/terminalStore";
import type { TerminalSession } from "../stores/terminalStore";
import { useVaultStore } from "../stores/vaultStore";
import { useAuthStore } from "../stores/authStore";
import { useNamespace } from "../hooks/useNamespaces";
import { useCreateSSHIdentity } from "../hooks/useSSHIdentityMutations";
import { getFingerprint, validatePrivateKey } from "../utils/sshKeys";
import {
  ensureBrowserKey,
  persistBrowserKey,
  browserLabel,
  type BrowserKey,
} from "../utils/browserKey";
import {
  BROWSER_KEY_QUERY_KEY,
  useBrowserKeyFingerprint,
} from "@/hooks/useBrowserKey";
import { isRecordingSupported } from "../utils/recordings";
import { isAlreadyEnrolled } from "../utils/sshIdentity";
import { listSshIdentitiesOptions } from "../client";
import BrowserEnrollDialog from "./terminal/BrowserEnrollDialog";
import CopyButton from "./common/CopyButton";
import Modal from "./common/Modal";
import RecBadge from "@/components/sessions/RecBadge";
import VaultLockedBanner from "./vault/VaultLockedBanner";
import VaultUnlockDialog from "./vault/VaultUnlockDialog";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import FieldLabel from "@/components/common/fields/FieldLabel";
import FieldHint from "@/components/common/fields/FieldHint";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import RadioSegment from "@/components/common/fields/RadioSegment";
import { INPUT } from "../utils/styles";
import { cn } from "@shellhub/design-system/cn";
import { Button, Callout, StatusDot } from "@shellhub/design-system/primitives";
import DistroIcon from "@/components/common/DistroIcon";
import { useDevice } from "@/hooks/useDevice";
import { useSSHIdentities } from "@/hooks/useSSHIdentities";
import { DEFAULT_LOGIN, parseSshid, sshUrl } from "@/utils/sshid";
import type { VaultKeyEntry } from "../types/vault";

interface Props {
  open: boolean;
  onClose: () => void;
  deviceUid: string;
  deviceName: string;
  sshid: string;
}

type ConnectParams = Omit<TerminalSession, "id" | "state" | "connectionStatus">;

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
  keyUnavailable: string | null;
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
  | { type: "setKeyUnavailable"; value: string | null }
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
  keyUnavailable: null,
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
    case "setKeyUnavailable":
      return { ...state, keyUnavailable: action.value };
    case "setRecordSession":
      return { ...state, recordSession: action.value };
  }
}

function ContextDivider() {
  return (
    <span aria-hidden="true" className="w-px h-3 shrink-0 bg-border-light" />
  );
}

function DeviceContext({
  uid,
  name,
  sshid,
}: {
  uid: string;
  name: string;
  sshid: string;
}) {
  const { device } = useDevice(uid);
  const online = device?.online ?? false;
  const parts = parseSshid(sshid);

  return (
    <span className="mt-1.5 flex items-center justify-between gap-3 text-xs">
      <span className="flex items-center gap-2 min-w-0 shrink">
        <StatusDot online={online} className="shrink-0" />
        <span title={name} className="font-medium text-text-primary truncate">
          {name}
        </span>
      </span>
      {device?.info?.pretty_name && (
        <>
          <ContextDivider />
          <span className="inline-flex items-center gap-1.5 min-w-24 shrink-[3] text-text-muted">
            <DistroIcon
              id={device.info.id ?? ""}
              className="shrink-0 text-sm leading-none"
            />
            <span title={device.info.pretty_name} className="truncate">
              {device.info.pretty_name}
            </span>
          </span>
        </>
      )}
      <ContextDivider />
      <span className="flex items-center gap-0.5 min-w-0 shrink-[2]">
        <code
          title={sshid}
          className="flex min-w-0 font-mono text-2xs text-accent-cyan"
        >
          {parts?.namespace && (
            <span className="min-w-0 shrink-[1000] truncate">
              {parts.namespace}.
            </span>
          )}
          <span className="min-w-0 truncate">{parts?.device ?? sshid}</span>
        </code>
        <CopyButton text={sshid} className="-ml-0.5" />
      </span>
    </span>
  );
}

function shortFingerprint(fingerprint: string) {
  return fingerprint.length > 20
    ? `${fingerprint.slice(0, 11)}…${fingerprint.slice(-4)}`
    : fingerprint;
}

function IdentityCard({
  name,
  email,
  keyName,
  fingerprint,
}: {
  name: string;
  email: string | null;
  keyName: string;
  fingerprint: string | null;
}) {
  const labelId = useId();
  const hintId = useId();

  return (
    <div>
      <FieldLabel id={labelId}>Identity</FieldLabel>
      <div
        role="group"
        aria-labelledby={labelId}
        aria-describedby={fingerprint ? undefined : hintId}
        className="flex items-center gap-3.5 p-3.5 rounded-xl border border-border bg-card"
      >
        <span
          aria-hidden="true"
          className="grid place-items-center w-11 h-11 shrink-0 rounded-full border border-primary/20 bg-primary/15 font-mono text-base font-bold text-primary uppercase"
        >
          {name.charAt(0) || "?"}
        </span>
        <div className="min-w-0 flex-1">
          <div className="text-sm font-semibold text-text-primary truncate">
            {name}
          </div>
          {email && (
            <div className="mt-0.5 text-2xs text-text-muted truncate">
              {email}
            </div>
          )}
          <div className="mt-1 flex items-center gap-1.5 min-w-0 text-2xs text-text-secondary">
            <FingerPrintIcon
              aria-hidden="true"
              className="w-3.5 h-3.5 shrink-0 text-primary"
            />
            <span className="truncate">{keyName}</span>
            <code
              title={fingerprint ?? undefined}
              className="font-mono text-text-muted truncate"
            >
              {fingerprint ? shortFingerprint(fingerprint) : "new key"}
            </code>
          </div>
        </div>
      </div>
      {!fingerprint && (
        <FieldHint id={hintId}>
          This browser makes a key and adds it to your SSH identities when you
          open the shell.
        </FieldHint>
      )}
    </div>
  );
}

/**
 * The connect dialog for a device: it opens a shell in the browser, or hands the login to the
 * user's own SSH client through an ssh:// link. It is the one place a user copies an SSHID from,
 * so the SSHID it shows and copies has to be the one that works.
 */
export default function ConnectModal({
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
  const recordStateId = useId();
  const [unlockOpen, setUnlockOpen] = useState(false);
  const [pendingEnroll, setPendingEnroll] = useState<{
    key: BrowserKey;
    scope: string;
    params: ConnectParams;
  } | null>(null);
  const recordingSupported = isRecordingSupported();

  const tenant = useAuthStore((s) => s.tenant);
  const userId = useAuthStore((s) => s.userId);
  const accountName = useAuthStore((s) => s.name || s.username || s.email);
  const accountEmail = useAuthStore((s) => s.email);
  const createIdentity = useCreateSSHIdentity();
  const queryClient = useQueryClient();
  const { namespace } = useNamespace(tenant ?? "");
  const namespaceRecords = namespace?.settings?.session_record ?? false;

  const identityMode = namespace?.settings?.ssh_access_mode === "identity";
  const browserKeyFingerprint = useBrowserKeyFingerprint();
  const { identities } = useSSHIdentities({ enabled: open && identityMode });
  const browserIdentity = identities.find(
    (i) => i.source === "browser" && i.fingerprint === browserKeyFingerprint,
  );

  useEffect(() => {
    if (!open) return;
    dispatch({ type: "reset" });
    void refreshVault();
  }, [open, refreshVault]);

  const username = state.username.trim() || DEFAULT_LOGIN;
  const externalUrl = sshUrl(sshid, username);

  const hasVaultKeys = vaultStatus === "unlocked" && vaultKeys.length > 0;
  const effectiveKeySource = hasVaultKeys ? state.keySource : "manual";

  const selectedVaultKey: VaultKeyEntry | undefined = hasVaultKeys
    ? vaultKeys.find((k) => k.id === state.selectedKeyId)
    : undefined;

  const canConnect =
    identityMode ||
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

  const finalizeConnect = (params: ConnectParams) => {
    const withRecord =
      !namespaceRecords && state.recordSession && isRecordingSupported()
        ? { ...params, record: true }
        : params;
    openTerminal({ ...withRecord, tenant: tenant ?? undefined });
    onClose();
  };

  const attachKey = (
    params: ConnectParams,
    key: BrowserKey,
  ): ConnectParams => ({
    ...params,
    fingerprint: key.fingerprint,
    publicKeyLine: key.publicKeyLine,
    browserKey: key.privateKey,
  });

  const enrollAndConnect = async (
    key: BrowserKey,
    scope: string,
    params: ConnectParams,
    name: string,
  ) => {
    try {
      await createIdentity.mutateAsync({
        body: { name, data: key.publicKeyLine, source: "browser" },
      });
    } catch (err: unknown) {
      if (!isAlreadyEnrolled(err)) throw err;
    }
    await persistBrowserKey(scope, key);
    await queryClient.invalidateQueries({ queryKey: BROWSER_KEY_QUERY_KEY });
    finalizeConnect(attachKey(params, key));
  };

  const connect = async () => {
    if (identityMode) {
      const params: ConnectParams = {
        deviceUid,
        deviceName,
        username,
        password: "",
      };

      const scope = userId && tenant ? `${userId}:${tenant}` : null;
      const key = scope ? await ensureBrowserKey(scope) : null;

      if (!key || !scope) {
        dispatch({
          type: "setKeyUnavailable",
          value:
            "This browser can't hold an SSH key, so it can't open the shell here. It needs Ed25519 in WebCrypto and IndexedDB, which private windows and older browsers often block. Open in external terminal still works.",
        });
        return;
      }

      let registered: boolean;
      try {
        const identities = await queryClient.fetchQuery(
          listSshIdentitiesOptions({}),
        );
        registered = identities.some((i) => i.fingerprint === key.fingerprint);
      } catch {
        dispatch({
          type: "setKeyUnavailable",
          value:
            "Couldn't check your SSH identities, so we can't tell whether this browser's key is still valid. Try again in a moment.",
        });
        return;
      }

      if (registered) {
        finalizeConnect(attachKey(params, key));
        return;
      }

      setPendingEnroll({ key, scope, params });
      return;
    }

    let params: ConnectParams;
    if (state.authMethod === "password") {
      params = {
        deviceUid,
        deviceName,
        username,
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
          value:
            "Couldn't read the private key. Check the key and its passphrase.",
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
        username,
        password: "",
        fingerprint,
        privateKey: key,
        passphrase: phrase,
      };
    }

    finalizeConnect(params);
  };

  const handleConnect = (e: FormEvent) => {
    e.preventDefault();
    if (canConnect) void connect();
  };

  return (
    <>
      <VaultUnlockDialog
        open={unlockOpen}
        onClose={() => setUnlockOpen(false)}
      />
      <BrowserEnrollDialog
        open={pendingEnroll !== null}
        defaultName={browserLabel()}
        onClose={() => setPendingEnroll(null)}
        onConfirm={async (name) => {
          if (!pendingEnroll) return;
          await enrollAndConnect(
            pendingEnroll.key,
            pendingEnroll.scope,
            pendingEnroll.params,
            name,
          );
          setPendingEnroll(null);
        }}
      />
      <Modal
        open={open}
        onClose={onClose}
        icon={<CommandLineIcon />}
        title="Connect"
        description={
          <>
            Open a shell in the browser, or in your own SSH client.
            <DeviceContext uid={deviceUid} name={deviceName} sshid={sshid} />
          </>
        }
        footerStart={
          externalUrl && (
            <a
              href={externalUrl}
              title="Open the shell in your own SSH client"
              className="inline-flex items-center gap-1.5 hover:text-text-primary transition-colors"
            >
              <ArrowTopRightOnSquareIcon className="w-3.5 h-3.5" />
              Open in external terminal
            </a>
          )
        }
        footer={
          <>
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button
              variant="primary"
              type="submit"
              form={`connect-form-${deviceUid}`}
              disabled={!canConnect}
            >
              Open in browser
            </Button>
          </>
        }
      >
        <form
          id={`connect-form-${deviceUid}`}
          onSubmit={handleConnect}
          className="space-y-5"
        >
          <InputField
            id="connect-username"
            label="Login"
            data-autofocus
            variant="mono"
            autoComplete="off"
            spellCheck={false}
            value={state.username}
            onChange={(v) => dispatch({ type: "setUsername", value: v })}
            placeholder={DEFAULT_LOGIN}
          />

          {identityMode &&
            (state.keyUnavailable ? (
              <Callout variant="error">{state.keyUnavailable}</Callout>
            ) : (
              <IdentityCard
                name={accountName ?? ""}
                email={accountEmail}
                keyName={browserIdentity?.name ?? browserLabel()}
                fingerprint={browserIdentity?.fingerprint ?? null}
              />
            ))}

          {!identityMode && (
            <>
              <RadioGroupField
                label="Authentication"
                value={state.authMethod}
                onChange={(v) => dispatch({ type: "setAuthMethod", value: v })}
                containerClassName="flex gap-1 p-0.5 bg-card border border-border rounded-lg"
              >
                <RadioSegment
                  value="password"
                  label="Password"
                  icon={<LockClosedIcon className="w-3.5 h-3.5" />}
                />
                <RadioSegment
                  value="key"
                  label="Private key"
                  icon={<KeyIcon className="w-3.5 h-3.5" />}
                />
              </RadioGroupField>
              {state.authMethod === "password" && (
                <PasswordField
                  id="connect-password"
                  label="Password"
                  autoComplete="current-password"
                  value={state.password}
                  onChange={(v) => dispatch({ type: "setPassword", value: v })}
                  placeholder="The login's password on the device"
                />
              )}

              {state.authMethod === "key" && (
                <>
                  {vaultStatus === "locked" && (
                    <VaultLockedBanner onUnlock={() => setUnlockOpen(true)} />
                  )}

                  {hasVaultKeys && (
                    <RadioGroupField
                      label="Key source"
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
                  {effectiveKeySource === "vault" ? (
                    <>
                      <div>
                        <FieldLabel htmlFor="connect-vault-key">Key</FieldLabel>
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
                          <option value="">Choose a key…</option>
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
                      <div>
                        <FieldLabel htmlFor="connect-manual-private-key">
                          Private key
                        </FieldLabel>
                        <textarea
                          id="connect-manual-private-key"
                          value={state.privateKey}
                          onChange={(e) =>
                            handleManualKeyChange(e.target.value)
                          }
                          placeholder={
                            "-----BEGIN OPENSSH PRIVATE KEY-----\n..."
                          }
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
                          placeholder="Key passphrase"
                          suppressPasswordManager
                          hint="This key is encrypted."
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
            </>
          )}

          {namespaceRecords && (
            <div className="flex items-center gap-2.5 text-xs text-text-secondary">
              <RecBadge on />
              <span>
                This session will be recorded and stored on the server by
                namespace policy.
              </span>
            </div>
          )}

          {!namespaceRecords && recordingSupported && (
            <div className="flex items-center gap-2.5 text-xs text-text-secondary">
              <RecBadge
                on={state.recordSession}
                onToggle={() =>
                  dispatch({
                    type: "setRecordSession",
                    value: !state.recordSession,
                  })
                }
                label="Record this session in this browser"
                describedBy={recordStateId}
              />
              <span id={recordStateId}>
                {state.recordSession
                  ? "This session will be recorded and stored in this browser."
                  : "This session won't be recorded. Press REC to record it and store it in this browser."}
              </span>
            </div>
          )}
        </form>
      </Modal>
    </>
  );
}
