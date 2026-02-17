import { useState, FormEvent } from "react";
import {
  LockClosedIcon,
  KeyIcon,
  ChevronDoubleRightIcon,
} from "@heroicons/react/24/outline";
import { useTerminalStore } from "../stores/terminalStore";
import CopyButton from "./common/CopyButton";
import Drawer from "./common/Drawer";
import { LABEL, INPUT } from "../utils/styles";

interface Props {
  open: boolean;
  onClose: () => void;
  deviceUid: string;
  deviceName: string;
  sshid: string;
}

export default function ConnectDrawer({
  open,
  onClose,
  deviceUid,
  deviceName,
  sshid,
}: Props) {
  const openTerminal = useTerminalStore((s) => s.open);
  const [username, setUsername] = useState("");
  const [authMethod, setAuthMethod] = useState<"password" | "key">("password");
  const [password, setPassword] = useState("");
  const [privateKey, setPrivateKey] = useState("");
  const [passphrase, setPassphrase] = useState("");

  const [prevOpen, setPrevOpen] = useState(false);
  if (open !== prevOpen) {
    setPrevOpen(open);
    if (open) {
      setUsername("");
      setPassword("");
      setPrivateKey("");
      setPassphrase("");
      setAuthMethod("password");
    }
  }

  const canConnect =
    username.trim() &&
    (authMethod === "password" ? password.trim() : privateKey.trim());

  const handleConnect = (e: FormEvent) => {
    e.preventDefault();
    if (authMethod === "password") {
      openTerminal({
        deviceUid,
        deviceName,
        username: username.trim(),
        password,
      });
      onClose();
    }
  };

  return (
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
            value={username}
            onChange={(e) => setUsername(e.target.value)}
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
              onClick={() => setAuthMethod("password")}
              className={`flex items-start gap-3 w-full px-3.5 py-3 rounded-lg border text-left transition-all ${
                authMethod === "password"
                  ? "bg-primary/[0.06] border-primary/30 ring-1 ring-primary/10"
                  : "bg-card border-border hover:border-border-light hover:bg-hover-subtle"
              }`}
            >
              <div
                className={`mt-0.5 shrink-0 w-4 h-4 rounded-full border-2 flex items-center justify-center transition-all ${authMethod === "password" ? "border-primary" : "border-text-muted/40"}`}
              >
                {authMethod === "password" && (
                  <div className="w-2 h-2 rounded-full bg-primary" />
                )}
              </div>
              <div className="flex items-start gap-2.5 min-w-0">
                <span
                  className={`mt-0.5 shrink-0 transition-colors ${authMethod === "password" ? "text-primary" : "text-text-muted"}`}
                >
                  <LockClosedIcon className="w-4 h-4" />
                </span>
                <div className="min-w-0">
                  <span
                    className={`block text-sm font-medium transition-colors ${authMethod === "password" ? "text-text-primary" : "text-text-secondary"}`}
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
              onClick={() => setAuthMethod("key")}
              className={`flex items-start gap-3 w-full px-3.5 py-3 rounded-lg border text-left transition-all ${
                authMethod === "key"
                  ? "bg-primary/[0.06] border-primary/30 ring-1 ring-primary/10"
                  : "bg-card border-border hover:border-border-light hover:bg-hover-subtle"
              }`}
            >
              <div
                className={`mt-0.5 shrink-0 w-4 h-4 rounded-full border-2 flex items-center justify-center transition-all ${authMethod === "key" ? "border-primary" : "border-text-muted/40"}`}
              >
                {authMethod === "key" && (
                  <div className="w-2 h-2 rounded-full bg-primary" />
                )}
              </div>
              <div className="flex items-start gap-2.5 min-w-0">
                <span
                  className={`mt-0.5 shrink-0 transition-colors ${authMethod === "key" ? "text-primary" : "text-text-muted"}`}
                >
                  <KeyIcon className="w-4 h-4" />
                </span>
                <div className="min-w-0">
                  <span
                    className={`block text-sm font-medium transition-colors ${authMethod === "key" ? "text-text-primary" : "text-text-secondary"}`}
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
        {authMethod === "password" && (
          <div>
            <label className={LABEL}>Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter device password"
              className={INPUT}
            />
          </div>
        )}

        {/* Private Key fields */}
        {authMethod === "key" && (
          <>
            <div>
              <label className={LABEL}>Private Key</label>
              <textarea
                value={privateKey}
                onChange={(e) => setPrivateKey(e.target.value)}
                placeholder={"-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
                rows={5}
                className={`${INPUT} font-mono text-xs resize-none`}
              />
            </div>
            <div>
              <label className={LABEL}>
                Passphrase{" "}
                <span className="text-text-muted/50 normal-case tracking-normal">
                  (optional)
                </span>
              </label>
              <input
                type="password"
                value={passphrase}
                onChange={(e) => setPassphrase(e.target.value)}
                placeholder="Key passphrase"
                className={INPUT}
              />
            </div>
          </>
        )}
      </form>
    </Drawer>
  );
}
