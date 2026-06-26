import { useEffect, useReducer, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import {
  LockClosedIcon,
  KeyIcon,
  ChevronDoubleRightIcon,
  ChevronUpDownIcon,
  ShieldCheckIcon,
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
  ArrowTopRightOnSquareIcon,
  ServerStackIcon,
  UserIcon,
  UsersIcon,
} from "@heroicons/react/24/outline";
import { useTerminalStore } from "../stores/terminalStore";
import { useVaultStore } from "../stores/vaultStore";
import { useClickOutside } from "../hooks/useClickOutside";
import {
  getFingerprint,
  getPublicKey,
  validatePrivateKey,
} from "../utils/ssh-keys";
import {
  useCreateConnection,
  useUpdateConnection,
} from "../hooks/useConnectionMutations";
import {
  useCreateTeamConnection,
  useUpdateTeamConnection,
  useUpdateTeamConnectionPrefs,
} from "../hooks/useTeamConnectionMutations";
import { useTeamConnectionPrefs } from "../hooks/useTeamConnections";
import {
  scanHostKey,
  getHostKey,
  acceptHostKey,
  type HostKeyScanResult,
} from "@/api/hostKeys";
import CopyButton from "./common/CopyButton";
import Drawer from "./common/Drawer";
import Alert from "./common/Alert";
import DevicePicker from "./common/DevicePicker";
import VaultLockedBanner from "./vault/VaultLockedBanner";
import VaultUnlockDialog from "./vault/VaultUnlockDialog";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import CheckboxField from "@/components/common/fields/CheckboxField";
import FieldLabel from "@/components/common/fields/FieldLabel";
import RadioCard from "@/components/common/fields/RadioCard";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import RadioSegment from "@/components/common/fields/RadioSegment";
import KeyFileInput from "@/components/common/fields/KeyFileInput";
import PremiumUpsell from "@/components/common/PremiumUpsell";
import { getConfig } from "@/env";
import { INPUT, LABEL } from "../utils/styles";
import { Card, Button } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { connectionDirty } from "@/utils/connectionDirty";
import type { VaultKeyEntry } from "../types/vault";
import type { Connection, TeamConnection } from "@/client";

// The single drawer for reaching an SSH target. Depending on props it can:
//   - Connect to a fixed device target (device/container pages).
//   - Connect to a saved connection (connections page rows).
//   - Create or edit a saved connection.
//
// Scope decides where the auth lives. A "personal" connection belongs to the
// caller, so its auth (username + key) is stored on the connection record. A
// "team" connection is shared with the namespace, so the target is shared but
// the auth is the caller's own per-user pref. Team is an Enterprise/Cloud
// capability; only the secret never leaves the browser in either case.
type Scope = "personal" | "team";

interface Props {
  open: boolean;
  onClose: () => void;
  deviceUid?: string;
  deviceName?: string;
  sshid?: string;
  connection?: Connection | null;
  // Scope of the connection being edited/connected. Ignored on create, where
  // the Personal/Team toggle drives it.
  scope?: Scope;
  // Whether the caller may create a team connection (Enterprise/Cloud + operator+).
  canCreateTeam?: boolean;
  editable?: boolean;
  onSaved?: () => void;
}

type Kind = "external" | "device";

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

// Map a team mutation result into the shared Connection shape the drawer's
// callbacks consume (auth stays empty; it lives in per-user prefs).
function teamResultToConnection(t: TeamConnection): Connection {
  return {
    id: t.id,
    tenant_id: t.tenant_id,
    owner_id: t.created_by,
    label: t.label,
    kind: t.kind,
    host: t.host ?? "",
    port: t.port || 22,
    device_uid: t.device_uid,
    username: "",
    auth_method: "",
    key_fingerprint: "",
    created_at: t.created_at,
    updated_at: t.updated_at,
  };
}

function VaultKeySelect({
  keys,
  value,
  onChange,
}: {
  keys: VaultKeyEntry[];
  value: string;
  onChange: (id: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  useClickOutside(ref, () => setOpen(false));

  const selected = keys.find((k) => k.id === value);

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className={`${INPUT} flex items-center justify-between text-left`}
      >
        <span
          className={
            selected ? "text-text-primary truncate" : "text-text-muted"
          }
        >
          {selected ? selected.name : "Choose a key…"}
        </span>
        <ChevronUpDownIcon className="w-4 h-4 text-text-muted shrink-0" />
      </button>

      {open && (
        <div className="absolute z-20 mt-1 w-full bg-card border border-border rounded-lg shadow-lg overflow-hidden">
          <div className="max-h-60 overflow-y-auto">
            {keys.length === 0 ? (
              <div className="px-3 py-3 text-xs font-mono text-text-muted">
                No keys in vault
              </div>
            ) : (
              keys.map((k) => (
                <button
                  key={k.id}
                  type="button"
                  onClick={() => {
                    onChange(k.id);
                    setOpen(false);
                  }}
                  className="w-full flex items-center gap-2 px-3 py-2 text-left text-sm hover:bg-hover-subtle transition-colors"
                >
                  <KeyIcon className="w-3.5 h-3.5 text-text-muted shrink-0" />
                  <span className="flex-1 truncate text-text-primary">
                    {k.name}
                  </span>
                  {k.algorithm && (
                    <span className="text-2xs text-text-muted shrink-0">
                      {k.algorithm}
                    </span>
                  )}
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}

export default function ConnectDrawer({
  open,
  onClose,
  deviceUid = "",
  deviceName = "",
  sshid = "",
  connection = null,
  scope = "personal",
  canCreateTeam = false,
  editable = false,
  onSaved,
}: Props) {
  const queryClient = useQueryClient();
  const openTerminal = useTerminalStore((s) => s.open);
  const vaultStatus = useVaultStore((s) => s.status);
  const vaultKeys = useVaultStore((s) => s.keys);
  const refreshVault = useVaultStore((s) => s.refreshStatus);

  const createMutation = useCreateConnection();
  const updateMutation = useUpdateConnection();
  const createTeamMutation = useCreateTeamConnection();
  const updateTeamMutation = useUpdateTeamConnection();
  const teamPrefsMutation = useUpdateTeamConnectionPrefs();

  const isCreate = editable && !connection;
  const isEdit = editable && !!connection;
  const isConnect = !editable;

  // Whether the edition has team connections at all (Cloud/Enterprise). On
  // Community it doesn't, so the Team scope is shown as a locked Pro upsell.
  const cfg = getConfig();
  const teamEdition = !!cfg.cloud || !!cfg.enterprise;

  const [state, dispatch] = useReducer(formReducer, initialState);
  const [unlockOpen, setUnlockOpen] = useState(false);

  // Target editor state (only used when `editable`).
  const [tScope, setTScope] = useState<Scope>("personal");
  const [tLabel, setTLabel] = useState("");
  const [tKind, setTKind] = useState<Kind>("external");
  const [tHost, setTHost] = useState("");
  const [tPort, setTPort] = useState("22");
  const [tDeviceUid, setTDeviceUid] = useState("");
  const [tDeviceName, setTDeviceName] = useState("");
  const [unreachable, setUnreachable] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saveOnConnect, setSaveOnConnect] = useState(false);

  // Host-key (TOFU) confirmation state for external connects.
  const [hostKeyResult, setHostKeyResult] = useState<HostKeyScanResult | null>(
    null,
  );
  const [hostKeyBusy, setHostKeyBusy] = useState(false);
  const [hostKeyError, setHostKeyError] = useState<string | null>(null);
  // Stash the resolved key across the host-key confirmation step.
  const pendingKeyRef = useRef<{
    key: string;
    phrase: string | undefined;
    fingerprint: string;
    publicKey: string | undefined;
  } | null>(null);

  // Effective scope: the toggle drives create; the prop fixes edit/connect.
  const effectiveScope: Scope = isCreate ? tScope : scope;
  const isTeam = effectiveScope === "team";
  // API scope param: "team" maps to the shared "namespace" record.
  const apiScope = isTeam ? "namespace" : "personal";

  // A team connection's auth is the caller's own per-user pref, fetched
  // separately for connect and edit (create has no connection yet).
  const { prefs: teamPrefs } = useTeamConnectionPrefs(
    !isCreate && scope === "team" && connection ? connection.id : undefined,
  );

  useEffect(() => {
    if (!open) return;
    dispatch({ type: "reset" });
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setUnreachable(false);
    setSaveError(null);
    // Create defaults to saving the new connection (a "New connection" the user
    // can still opt out of); connect starts unchecked.
    setSaveOnConnect(isCreate);
    setHostKeyResult(null);
    setHostKeyBusy(false);
    setHostKeyError(null);
    setTScope(isCreate ? "personal" : scope);

    if (connection) {
      setTLabel(connection.label);
      setTKind(connection.kind === "device" ? "device" : "external");
      setTHost(connection.host);
      setTPort(String(connection.port || 22));
      setTDeviceUid(connection.device_uid);
      setTDeviceName(connection.label);

      // Personal auth lives on the record; team auth is prefilled from the
      // per-user prefs effect below.
      if (scope === "personal") {
        dispatch({ type: "setUsername", value: connection.username ?? "" });
        if (
          connection.auth_method === "password" ||
          connection.auth_method === "key"
        ) {
          dispatch({ type: "setAuthMethod", value: connection.auth_method });
        }
      }
    } else {
      setTLabel("");
      setTKind("external");
      setTHost("");
      setTPort("22");
      setTDeviceUid("");
      setTDeviceName("");
    }

    void refreshVault();
  }, [open, connection, scope, isCreate, refreshVault]);

  // Prefill from the caller's team prefs (team connect/edit), once they load.
  useEffect(() => {
    if (!open || isCreate || scope !== "team" || !teamPrefs) return;
    if (teamPrefs.username)
      dispatch({ type: "setUsername", value: teamPrefs.username });
    if (
      teamPrefs.auth_method === "password" ||
      teamPrefs.auth_method === "key"
    ) {
      dispatch({ type: "setAuthMethod", value: teamPrefs.auth_method });
    }
  }, [open, isCreate, scope, teamPrefs]);

  const hasVaultKeys = vaultStatus === "unlocked" && vaultKeys.length > 0;
  // The vault is offerable as a key source when it holds keys or is just locked
  // (unlocking reveals them). Keeping "vault" selectable while locked lets the
  // locked notice live inside the Vault tab instead of floating above the toggle.
  const vaultPresent = vaultStatus === "locked" || hasVaultKeys;
  const effectiveKeySource = vaultPresent ? state.keySource : "manual";

  const selectedVaultKey: VaultKeyEntry | undefined = hasVaultKeys
    ? vaultKeys.find((k) => k.id === state.selectedKeyId)
    : undefined;

  // Which key the saved auth pref points at (team: per-user prefs; personal: row).
  const preferredKeyFingerprint = isTeam
    ? (teamPrefs?.key_fingerprint ?? "")
    : (connection?.key_fingerprint ?? "");

  useEffect(() => {
    if (!open || !preferredKeyFingerprint || !hasVaultKeys) return;
    const match = vaultKeys.find(
      (k) => k.fingerprint === preferredKeyFingerprint,
    );
    if (match) dispatch({ type: "setSelectedKeyId", value: match.id });
  }, [open, preferredKeyFingerprint, hasVaultKeys, vaultKeys]);

  const preferredKeyAvailable =
    vaultStatus === "unlocked" &&
    vaultKeys.some((k) => k.fingerprint === preferredKeyFingerprint);
  const preferredKeyMissing =
    state.authMethod === "key" &&
    preferredKeyFingerprint !== "" &&
    vaultStatus !== "locked" &&
    !preferredKeyAvailable;

  // The connection authenticates with a key that lives in the vault, but the
  // vault is locked and the user hasn't pasted a one-off key instead. Connecting
  // should prompt an unlock rather than fail.
  const needsVaultUnlock =
    state.authMethod === "key" &&
    vaultStatus === "locked" &&
    preferredKeyFingerprint !== "" &&
    !state.privateKey.trim();

  const targetKind: Kind = editable
    ? tKind
    : connection
      ? connection.kind === "device"
        ? "device"
        : "external"
      : "device";
  const isExternal = targetKind === "external";
  const targetHost = editable ? tHost.trim() : (connection?.host ?? "");
  const portNum = editable ? Number(tPort) : (connection?.port ?? 22);
  const portValid =
    Number.isInteger(portNum) && portNum >= 1 && portNum <= 65535;
  const targetDeviceUid = editable
    ? tDeviceUid
    : (connection?.device_uid ?? deviceUid);
  const targetDeviceName = editable
    ? tDeviceName
    : (connection?.label ?? deviceName);
  const targetLabel = editable
    ? tLabel.trim()
    : (connection?.label ?? deviceName);

  // Whether the editable target differs from the saved record. A team edit may
  // be opened by a non-manager just to set their own auth; when the shared
  // target is untouched we skip the operator+-only update entirely.
  const targetChanged =
    !connection ||
    targetLabel !== connection.label ||
    targetKind !== (connection.kind === "device" ? "device" : "external") ||
    (isExternal
      ? targetHost !== connection.host || portNum !== connection.port
      : targetDeviceUid !== connection.device_uid);

  const showConnect = !isEdit;
  // Create folds save into the Connect action via the "Save connection"
  // checkbox; only edit keeps a dedicated Save button.
  const showSave = isEdit;
  // Auth is always editable. Personal auth lives on the record; team auth is the
  // caller's own per-user pref, seeded on save (create/edit) or on connect.
  const showAuth = true;
  const pending =
    createMutation.isPending ||
    updateMutation.isPending ||
    createTeamMutation.isPending ||
    updateTeamMutation.isPending ||
    teamPrefsMutation.isPending;

  const targetValid = editable
    ? targetLabel.length > 0 &&
      (isExternal
        ? targetHost.length > 0 && portValid
        : targetDeviceUid.length > 0)
    : true;

  const authValid =
    !showAuth ||
    (state.username.trim().length > 0 &&
      (state.authMethod === "password"
        ? state.password.trim().length > 0
        : effectiveKeySource === "vault"
          ? !!selectedVaultKey &&
            (!selectedVaultKey.hasPassphrase ||
              state.passphrase.trim().length > 0)
          : state.manualKeyValid &&
            (!state.manualKeyEncrypted || state.passphrase.trim().length > 0)));

  // A locked vault still lets the user click Connect: the click prompts an
  // unlock (handled in doConnect) instead of being blocked by the missing key.
  const canConnect = targetValid && (authValid || needsVaultUnlock) && !pending;
  const canSave = showSave && targetValid && !pending;

  const currentKeyFp =
    state.authMethod === "key"
      ? effectiveKeySource === "vault"
        ? (selectedVaultKey?.fingerprint ?? "")
        : state.manualKeyValid
          ? "manual"
          : ""
      : "";
  const authRecord = isTeam
    ? {
        username: teamPrefs?.username ?? "",
        auth_method: teamPrefs?.auth_method ?? "",
        key_fingerprint: teamPrefs?.key_fingerprint ?? "",
      }
    : {
        username: connection?.username ?? "",
        auth_method: connection?.auth_method ?? "",
        key_fingerprint: connection?.key_fingerprint ?? "",
      };
  const dirty =
    isConnect &&
    !!connection &&
    connectionDirty(
      {
        username: state.username,
        authMethod: state.authMethod,
        keyFingerprint: currentKeyFp,
      },
      authRecord,
    );

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

  // Best-effort current auth preference to persist (never the secret).
  const currentAuthPref = (): {
    auth_method: string;
    key_fingerprint: string;
  } => {
    if (state.authMethod === "password") {
      return { auth_method: "password", key_fingerprint: "" };
    }
    if (effectiveKeySource === "vault" && selectedVaultKey) {
      return {
        auth_method: "key",
        key_fingerprint: selectedVaultKey.fingerprint,
      };
    }
    try {
      const phrase = state.manualKeyEncrypted ? state.passphrase : undefined;
      return {
        auth_method: "key",
        key_fingerprint: getFingerprint(state.privateKey.trim(), phrase),
      };
    } catch {
      return { auth_method: "key", key_fingerprint: "" };
    }
  };

  const targetBody = () => ({
    label: targetLabel,
    kind: targetKind,
    host: isExternal ? targetHost : undefined,
    port: isExternal ? portNum : undefined,
    device_uid: isExternal ? undefined : targetDeviceUid,
  });

  // Persist the caller's team auth pref (username + key pointer, never secret).
  const seedTeamPrefs = (id: string, onDone?: () => void) => {
    const pref = currentAuthPref();
    teamPrefsMutation.mutate(
      {
        id,
        body: {
          username: state.username.trim(),
          auth_method: pref.auth_method,
          key_fingerprint: pref.key_fingerprint,
        },
      },
      { onSuccess: () => onDone?.() },
    );
  };

  // Create or update the connection target. Personal connections carry the
  // caller's auth on the record; team connections store only the shared target
  // (the per-user auth is seeded separately). Runs onDone with the record.
  const persistTarget = (force: boolean, onDone?: (c: Connection) => void) => {
    setSaveError(null);
    const onError = (err?: unknown) => {
      // Editing a team target needs operator+; tell the user that apart from a
      // generic failure (they reach this drawer to set their own auth too).
      if (isSdkError(err) && err.status === 403) {
        setSaveError("You don't have permission to change this connection.");

        return;
      }

      setSaveError("Failed to save connection. Check the fields.");
    };

    if (isTeam) {
      const body = targetBody();
      if (connection) {
        // The shared target is operator+-only; if it's unchanged (e.g. a member
        // just setting their own auth) keep the record and skip the update.
        if (!targetChanged) {
          onDone?.(connection);
          return;
        }
        updateTeamMutation.mutate(
          { id: connection.id, body },
          {
            onSuccess: (c) => {
              onSaved?.();
              onDone?.(teamResultToConnection(c));
            },
            onError,
          },
        );
        return;
      }

      createTeamMutation.mutate(body, {
        onSuccess: (c) => {
          onSaved?.();
          onDone?.(teamResultToConnection(c));
        },
        onError,
      });
      return;
    }

    const pref = currentAuthPref();
    const body = {
      ...targetBody(),
      username: state.username.trim(),
      auth_method: pref.auth_method,
      key_fingerprint: pref.key_fingerprint,
      force,
    };

    if (connection) {
      updateMutation.mutate(
        { id: connection.id, body },
        {
          onSuccess: (c) => {
            onSaved?.();
            onDone?.(c);
          },
          onError,
        },
      );
      return;
    }

    createMutation.mutate(body, {
      onSuccess: (c) => {
        onSaved?.();
        onDone?.(c);
      },
      onError: (err) => {
        if (isSdkError(err) && err.status === 422) {
          // 422 distinguishes a blocked address (policy) from an unreachable host
          // (the NAT/firewall + install-agent funnel).
          if ((err as { error?: string }).error === "blocked") {
            setSaveError(
              `${targetHost}:${portNum} isn't a permitted connection target.`,
            );
          } else {
            setUnreachable(true);
          }
        } else {
          onError();
        }
      },
    });
  };

  // Save (bookmark). Creating a team connection also seeds the creator's prefs.
  const save = (force: boolean) => {
    setUnreachable(false);
    persistTarget(force, (saved) => {
      // Team save (create or edit) also persists the caller's own auth pref,
      // but only when they actually provided a username.
      if (isTeam && state.username.trim()) {
        seedTeamPrefs(saved.id, () => onClose());
      } else {
        onClose();
      }
    });
  };

  const handleSave = () => {
    if (!canSave) return;
    save(false);
  };

  const resolveKey = () => {
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
      return null;
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
      return null;
    }

    let publicKey: string | undefined;
    if (isExternal) {
      try {
        publicKey = getPublicKey(key, phrase);
      } catch {
        dispatch({
          type: "setKeyError",
          value: "Failed to read private key. Check the key or passphrase.",
        });
        return null;
      }
    }
    dispatch({ type: "setKeyError", value: null });
    return { key, phrase, fingerprint, publicKey };
  };

  const openTerminalFor = (
    connId: string,
    keyArg: ReturnType<typeof resolveKey>,
    hostKey: string,
  ) => {
    const base = isExternal
      ? {
          kind: "connect" as const,
          deviceUid: connId,
          deviceName: targetLabel || targetHost,
          host: targetHost,
          port: portNum,
          knownHostKey: hostKey,
        }
      : { deviceUid: targetDeviceUid, deviceName: targetDeviceName };

    if (!keyArg) {
      openTerminal({
        ...base,
        username: state.username.trim(),
        password: state.password,
      });
      return;
    }

    openTerminal({
      ...base,
      username: state.username.trim(),
      password: "",
      fingerprint: keyArg.fingerprint,
      privateKey: keyArg.key,
      passphrase: keyArg.phrase,
      publicKey: keyArg.publicKey,
    });
  };

  // Open the terminal (creating/seeding as needed) once the host key is settled.
  const finishConnect = (
    keyArg: ReturnType<typeof resolveKey>,
    hostKey: string,
  ) => {
    // Create mode: minting the record IS the save (forced, since the user chose
    // to connect). Team also seeds the creator's prefs before opening.
    if (isCreate) {
      // The "Save connection" checkbox decides whether to bookmark the target.
      // Unchecked is a one-off, ephemeral connect with no record persisted.
      if (!saveOnConnect) {
        const ephemeralId = isExternal
          ? `external:${targetHost}:${portNum}`
          : targetDeviceUid;
        openTerminalFor(ephemeralId, keyArg, hostKey);
        onClose();
        return;
      }

      persistTarget(true, (created) => {
        if (isTeam && state.username.trim()) {
          seedTeamPrefs(created.id, () => {
            openTerminalFor(created.id, keyArg, hostKey);
            onClose();
          });
        } else {
          openTerminalFor(created.id, keyArg, hostKey);
          onClose();
        }
      });
      return;
    }

    openTerminalFor(connection?.id ?? targetDeviceUid, keyArg, hostKey);
    // Connect mode: persist the tweak only if "Save changes" is ticked. Team
    // writes per-user prefs; personal updates the connection's auth.
    if (connection && dirty && saveOnConnect) {
      if (isTeam) {
        seedTeamPrefs(connection.id);
      } else {
        persistTarget(false);
      }
    }
    onClose();
  };

  const doConnect = async () => {
    if (!canConnect || hostKeyBusy) return;

    // Key auth whose key lives in the locked vault: prompt an unlock instead of
    // attempting the connect. Once unlocked, the saved key resolves and the user
    // can connect.
    if (needsVaultUnlock) {
      setUnlockOpen(true);
      return;
    }

    let keyArg: ReturnType<typeof resolveKey> = null;
    if (state.authMethod === "key") {
      keyArg = resolveKey();
      if (!keyArg) return;
    }

    // Device targets go through the agent tunnel; no host key to verify.
    if (!isExternal) {
      finishConnect(keyArg, "");

      return;
    }

    // External: verify the host key (TOFU). If already trusted, pass the stored
    // key straight through (the server re-checks). On first use, scan and ask.
    setHostKeyError(null);
    setHostKeyBusy(true);
    try {
      const stored = await getHostKey(targetHost, portNum, apiScope);
      if (stored) {
        setHostKeyBusy(false);
        finishConnect(keyArg, stored.public_key);

        return;
      }

      const scan = await scanHostKey(targetHost, portNum, apiScope);
      setHostKeyBusy(false);
      if (scan.status === "trusted") {
        finishConnect(keyArg, scan.public_key);

        return;
      }

      pendingKeyRef.current = keyArg;
      setHostKeyResult(scan);
    } catch (err) {
      setHostKeyBusy(false);
      if (
        isSdkError(err) &&
        err.status === 422 &&
        (err as { error?: string }).error === "blocked"
      ) {
        setHostKeyError(
          `${targetHost}:${portNum} isn't a permitted connection target.`,
        );
      } else {
        setHostKeyError(
          "Couldn't read the host key. Check that the host is reachable.",
        );
      }
    }
  };

  // Accept the scanned host key (TOFU) and continue connecting.
  const acceptAndConnect = async () => {
    if (!hostKeyResult) return;

    setHostKeyBusy(true);
    try {
      await acceptHostKey({
        host: targetHost,
        port: portNum,
        scope: apiScope,
        key_type: hostKeyResult.key_type,
        public_key: hostKeyResult.public_key,
        fingerprint: hostKeyResult.fingerprint,
      });
      // The accept goes through the raw API (not useAcceptHostKey), so invalidate
      // the cache the host-key modal reads or it shows stale "no key stored".
      void queryClient.invalidateQueries({
        queryKey: ["host-key", apiScope, targetHost, portNum],
      });

      const publicKey = hostKeyResult.public_key;
      const keyArg = pendingKeyRef.current;
      setHostKeyBusy(false);
      setHostKeyResult(null);
      finishConnect(keyArg, publicKey);
    } catch {
      setHostKeyBusy(false);
      setHostKeyError("Failed to save the host key. Try again.");
    }
  };

  const formId = `connect-form-${connection?.id ?? deviceUid ?? "new"}`;

  const title = isCreate
    ? "New connection"
    : isEdit
      ? isTeam
        ? "Edit team connection"
        : "Edit connection"
      : `Connect to ${targetLabel || targetDeviceName}`;

  return (
    <>
      <VaultUnlockDialog
        open={unlockOpen}
        onClose={() => setUnlockOpen(false)}
      />
      <Drawer
        open={open}
        onClose={onClose}
        title={title}
        subtitle={
          editable ? "A device or external host you reach over SSH" : undefined
        }
        footer={
          <>
            {(isCreate || (isConnect && dirty)) && (
              <div className="mr-auto">
                <CheckboxField
                  id="connect-save-changes"
                  label={isCreate ? "Save connection" : "Save changes"}
                  checked={saveOnConnect}
                  onChange={setSaveOnConnect}
                />
              </div>
            )}
            <Button variant="ghost" type="button" onClick={onClose}>
              Cancel
            </Button>
            {showSave && (
              <Button
                variant="primary"
                type="button"
                onClick={handleSave}
                disabled={!canSave}
              >
                {pending ? "Saving…" : "Save"}
              </Button>
            )}
            {showConnect && (
              <Button
                variant="success"
                type="button"
                onClick={() => void doConnect()}
                disabled={!canConnect || hostKeyBusy}
                icon={
                  <ChevronDoubleRightIcon className="w-4 h-4" strokeWidth={2} />
                }
              >
                {hostKeyBusy ? "Verifying…" : "Connect"}
              </Button>
            )}
          </>
        }
      >
        <form
          id={formId}
          onSubmit={(e) => {
            e.preventDefault();
            if (isEdit) handleSave();
            else void doConnect();
          }}
          className="space-y-5"
        >
          <button type="submit" className="hidden" aria-hidden tabIndex={-1} />

          {editable ? (
            <>
              <InputField
                id="connect-label"
                label="Label"
                value={tLabel}
                onChange={setTLabel}
                placeholder="e.g. db-primary"
                // eslint-disable-next-line jsx-a11y/no-autofocus -- move focus into the freshly opened drawer
                autoFocus={open}
              />

              {/* Scope: personal (yours) vs team (shared). Team is a Cloud/
                  Enterprise feature, so Community offers Personal and pitches
                  Team as an upsell card. Only on create. */}
              {isCreate && canCreateTeam && (
                <RadioGroupField
                  label="Scope"
                  value={tScope}
                  onChange={setTScope}
                >
                  <RadioCard
                    value="personal"
                    icon={<UserIcon className="w-4 h-4" />}
                    label="Personal"
                    description="Only you can see and use this connection."
                  />
                  <RadioCard
                    value="team"
                    icon={<UsersIcon className="w-4 h-4" />}
                    label="Team"
                    description="Shared with everyone in the namespace. Each member uses their own auth."
                  />
                </RadioGroupField>
              )}
              {isCreate && !teamEdition && (
                <div className="space-y-2">
                  <RadioGroupField
                    label="Scope"
                    value={tScope}
                    onChange={setTScope}
                  >
                    <RadioCard
                      value="personal"
                      icon={<UserIcon className="w-4 h-4" />}
                      label="Personal"
                      description="Only you can see and use this connection."
                    />
                  </RadioGroupField>
                  <PremiumUpsell
                    icon={<UsersIcon className="w-5 h-5" />}
                    title="Team"
                    description="Share one target with your whole team, each member with their own auth."
                  />
                </div>
              )}

              {/* Type is fixed after creation: external host and device are
                  distinct targets. On edit the selector is shown disabled so the
                  type stays visible but can't change (the backend rejects it too). */}
              <fieldset
                disabled={isEdit}
                className={
                  isEdit ? "opacity-60 pointer-events-none" : undefined
                }
              >
                <RadioGroupField label="Type" value={tKind} onChange={setTKind}>
                  <RadioCard
                    value="external"
                    icon={<ArrowTopRightOnSquareIcon className="w-4 h-4" />}
                    label="External host"
                    description="Connect straight to a host by address. No agent required."
                  />
                  <RadioCard
                    value="device"
                    icon={<ServerStackIcon className="w-4 h-4" />}
                    label="Device"
                    description="Connect to a device running the ShellHub agent."
                  />
                </RadioGroupField>
              </fieldset>
              {tKind === "external" ? (
                <>
                  <InputField
                    id="connect-host"
                    label="Hostname or IP"
                    value={tHost}
                    onChange={setTHost}
                    placeholder="e.g. 10.0.0.5 or db.internal"
                  />
                  <InputField
                    id="connect-port"
                    label="Port"
                    value={tPort}
                    onChange={setTPort}
                    placeholder="22"
                  />
                </>
              ) : (
                <div>
                  <FieldLabel htmlFor="connect-device">Device</FieldLabel>
                  <DevicePicker
                    value={tDeviceUid}
                    valueLabel={tDeviceName}
                    onChange={(uid, name) => {
                      setTDeviceUid(uid);
                      setTDeviceName(name);
                    }}
                  />
                </div>
              )}
            </>
          ) : isExternal ? (
            <Card className="rounded-lg p-3.5">
              <span className={LABEL}>Target</span>
              <code className="block text-xs font-mono text-text-secondary mt-1">
                {targetHost}:{portNum}
              </code>
            </Card>
          ) : (
            <>
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
                    Enter your device OS username below to complete this
                    command.
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
            </>
          )}

          {showAuth && (
            <>
              <InputField
                id="connect-username"
                label={isCreate ? "Username (optional)" : "Username"}
                value={state.username}
                onChange={(v) => dispatch({ type: "setUsername", value: v })}
                placeholder="e.g. root"
              />

              <RadioGroupField
                label="Authentication"
                value={state.authMethod}
                onChange={(v) => dispatch({ type: "setAuthMethod", value: v })}
              >
                <RadioCard
                  value="password"
                  icon={<LockClosedIcon className="w-4 h-4" />}
                  label="Password"
                  description="Authenticate with a password."
                />
                <RadioCard
                  value="key"
                  icon={<KeyIcon className="w-4 h-4" />}
                  label="Private Key"
                  description="Authenticate using your SSH private key."
                />
              </RadioGroupField>

              {state.authMethod === "password" && (
                <PasswordField
                  id="connect-password"
                  label="Password"
                  autoComplete="current-password"
                  value={state.password}
                  onChange={(v) => dispatch({ type: "setPassword", value: v })}
                  placeholder="Enter password"
                />
              )}

              {state.authMethod === "key" && (
                <>
                  {preferredKeyMissing && (
                    <div className="rounded-lg border border-accent-yellow/30 bg-accent-yellow/10 p-3 space-y-1.5">
                      <p className="text-xs text-accent-yellow flex items-start gap-1.5 font-medium">
                        <ExclamationTriangleIcon className="w-4 h-4 shrink-0 mt-px" />
                        <span>
                          {vaultStatus === "uninitialized"
                            ? "You don't have a vault yet, so this connection's saved key isn't available."
                            : "This connection's saved key isn't in your vault."}
                        </span>
                      </p>
                      <p className="text-2xs text-text-secondary">
                        Set up your vault to store the key, or paste a key
                        manually below for a one-off connect.
                      </p>
                      <Link
                        to="/secure-vault"
                        className="text-2xs text-primary hover:underline inline-flex items-center gap-1 font-medium"
                      >
                        {vaultStatus === "uninitialized"
                          ? "Set up your vault"
                          : "Open your vault"}
                        <ChevronDoubleRightIcon
                          className="w-3 h-3"
                          strokeWidth={2}
                        />
                      </Link>
                    </div>
                  )}

                  {vaultPresent && (
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

                  {effectiveKeySource === "vault" ? (
                    vaultStatus === "locked" ? (
                      <VaultLockedBanner onUnlock={() => setUnlockOpen(true)} />
                    ) : (
                      <>
                        <div>
                          <FieldLabel htmlFor="connect-vault-key">
                            Select Key
                          </FieldLabel>
                          <VaultKeySelect
                            keys={vaultKeys}
                            value={state.selectedKeyId}
                            onChange={(id) =>
                              dispatch({ type: "setSelectedKeyId", value: id })
                            }
                          />
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
                    )
                  ) : (
                    <>
                      <KeyFileInput
                        id="connect-manual-private-key"
                        label="Private Key"
                        value={state.privateKey}
                        onChange={handleManualKeyChange}
                        validate={(text) =>
                          validatePrivateKey(text.trim()).valid
                        }
                        accept=".pem,.key,.txt"
                        placeholder={"-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
                        rows={5}
                        hint="RSA, ECDSA, ED25519 — PEM and OpenSSH formats. The key stays in your browser."
                      />
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
            </>
          )}

          {unreachable && (
            <div className="rounded-lg border border-accent-yellow/30 bg-accent-yellow/10 p-3 space-y-2">
              <p className="text-xs text-accent-yellow flex items-center gap-1.5 font-medium">
                <ExclamationTriangleIcon className="w-4 h-4 shrink-0" />
                Couldn't reach {targetHost}:{portNum}
              </p>
              <p className="text-2xs text-text-secondary">
                If it's behind NAT or a firewall, install the ShellHub agent on
                it to reach it through the gateway. No public IP or open ports
                needed.
              </p>
              <div className="flex items-center gap-3 pt-0.5">
                <Link
                  to="/devices/add"
                  className="text-2xs text-primary hover:underline inline-flex items-center gap-1 font-medium"
                >
                  Install the agent
                  <ChevronDoubleRightIcon className="w-3 h-3" strokeWidth={2} />
                </Link>
                <button
                  type="button"
                  onClick={() => save(true)}
                  disabled={pending}
                  className="text-2xs text-text-secondary hover:text-text-primary underline disabled:opacity-dim"
                >
                  Save anyway
                </button>
              </div>
            </div>
          )}

          {/* Host key confirmation (TOFU) for an external target's first connect,
              or when the stored key changed. */}
          {hostKeyResult && (
            <div
              className={`rounded-lg border p-3 space-y-2 ${
                hostKeyResult.status === "changed"
                  ? "border-accent-red/40 bg-accent-red/10"
                  : "border-accent-yellow/30 bg-accent-yellow/10"
              }`}
            >
              <p
                className={`text-xs flex items-center gap-1.5 font-medium ${
                  hostKeyResult.status === "changed"
                    ? "text-accent-red"
                    : "text-accent-yellow"
                }`}
              >
                {hostKeyResult.status === "changed" ? (
                  <ExclamationTriangleIcon className="w-4 h-4 shrink-0" />
                ) : (
                  <ShieldCheckIcon className="w-4 h-4 shrink-0" />
                )}
                {hostKeyResult.status === "changed"
                  ? "Host key changed — possible man-in-the-middle"
                  : `First connection to ${targetHost}:${portNum}`}
              </p>
              <p className="text-2xs text-text-secondary">
                Verify this host key fingerprint out of band before trusting it.
                {hostKeyResult.status === "changed" &&
                  hostKeyResult.stored &&
                  ` It was ${hostKeyResult.stored.fingerprint}.`}
              </p>
              <div className="flex items-center gap-2">
                <code className="font-mono text-2xs break-all text-text-primary">
                  {hostKeyResult.key_type} {hostKeyResult.fingerprint}
                </code>
                <CopyButton text={hostKeyResult.fingerprint} />
              </div>
              <div className="flex items-center gap-3 pt-0.5">
                <button
                  type="button"
                  onClick={() => void acceptAndConnect()}
                  disabled={hostKeyBusy}
                  className="text-2xs font-semibold text-primary hover:underline disabled:opacity-dim"
                >
                  {hostKeyBusy
                    ? "Saving…"
                    : hostKeyResult.status === "changed"
                      ? "Accept new key and connect"
                      : "Accept and connect"}
                </button>
                <button
                  type="button"
                  onClick={() => setHostKeyResult(null)}
                  disabled={hostKeyBusy}
                  className="text-2xs text-text-secondary hover:text-text-primary"
                >
                  Cancel
                </button>
              </div>
            </div>
          )}

          {hostKeyError && <Alert variant="error">{hostKeyError}</Alert>}

          {saveError && <Alert variant="error">{saveError}</Alert>}
        </form>
      </Drawer>
    </>
  );
}
