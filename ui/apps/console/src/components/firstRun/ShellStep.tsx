import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Button } from "@shellhub/design-system/primitives";
import CopyButton from "@/components/common/CopyButton";
import ConnectModal from "@/components/ConnectModal";
import RestrictedAction from "@/components/common/RestrictedAction";
import { useTerminalStore } from "@/stores/terminalStore";
import { useDevice } from "@/hooks/useDevice";
import { useSshEndpoint } from "@/hooks/useSshEndpoint";
import {
  DEFAULT_LOGIN,
  buildSshid,
  sshCommand,
  sshPortFlag,
} from "@/utils/sshid";
import DeviceCard from "./DeviceCard";
import type { PairedDevice } from "./useFirstDevice";

const ONLINE_POLL_MS = 3000;

interface ShellStepProps {
  device: PairedDevice;
  namespace: string;
  signsInByKey: boolean;
}

/**
 * The last step: the device just paired, watched until its agent connects, and the two ways to
 * open a shell on it, in the browser or from the user's own terminal. The browser terminal lives
 * in the console's frame, so once one is open the user is taken to the device's page, where it
 * shows.
 */
export default function ShellStep({
  device,
  namespace,
  signsInByKey,
}: ShellStepProps) {
  const navigate = useNavigate();
  const endpoint = useSshEndpoint();
  const [login, setLogin] = useState(DEFAULT_LOGIN);
  const [connecting, setConnecting] = useState(false);
  const { device: details } = useDevice(device.uid, {
    refetchInterval: (current) => (current?.online ? false : ONLINE_POLL_MS),
  });
  const online = details?.online ?? false;
  const name = details?.name ?? device.name;

  const effectiveLogin = login || DEFAULT_LOGIN;
  const command = sshCommand(effectiveLogin, namespace, name, endpoint);

  const closeConnect = () => {
    setConnecting(false);
    const opened = useTerminalStore
      .getState()
      .sessions.some((s) => s.deviceUid === device.uid);
    if (opened) void navigate(`/devices/${device.uid}`);
  };

  return (
    <div className="flex flex-col gap-4">
      <DeviceCard
        device={details ?? { name: device.name }}
        status={
          online ? (
            <span className="inline-flex items-center gap-1.5 text-2xs font-mono text-accent-green">
              <span className="w-1.5 h-1.5 rounded-full bg-accent-green" />
              online
            </span>
          ) : (
            <span className="text-2xs font-mono text-text-muted">
              connecting...
            </span>
          )
        }
      />

      <div className="rounded-xl border border-border divide-y divide-border">
        <div className="flex flex-col sm:flex-row sm:items-center gap-3 p-4">
          <div className="flex-1 min-w-0">
            <p className="text-sm font-semibold text-text-primary">
              In the browser
            </p>
            <p className="text-xs text-text-muted">
              A terminal right here, as the login you pick.
            </p>
          </div>
          <RestrictedAction action="device:connect">
            <Button onClick={() => setConnecting(true)}>Open terminal</Button>
          </RestrictedAction>
        </div>
        <div className="p-4">
          <div className="flex items-start gap-3 mb-3">
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-text-primary">
                From your terminal
              </p>
              <p className="text-xs text-text-muted">
                {signsInByKey
                  ? "Uses your SSH key. The first time, it shows a link to approve it."
                  : "The device's own users and passwords apply."}{" "}
                Click the login to change it.
              </p>
            </div>
            <CopyButton text={command} showLabel />
          </div>
          <p className="bg-background border border-border rounded-lg px-3.5 py-2.5 font-mono text-xs text-accent-cyan whitespace-nowrap overflow-x-auto">
            ssh {sshPortFlag(endpoint)}
            <input
              aria-label="Login on the device"
              value={login}
              onChange={(e) => setLogin(e.target.value.trim())}
              spellCheck={false}
              autoComplete="off"
              style={{ width: `${Math.max(effectiveLogin.length, 1)}ch` }}
              className="bg-transparent text-accent-yellow border-b border-dashed border-accent-yellow/60 outline-none focus:border-accent-yellow p-0"
            />
            @{namespace}.{name}@{endpoint.host}
          </p>
        </div>
      </div>

      <p className="text-xs text-text-muted">
        Next:{" "}
        <Link
          to="/access-policies"
          className="text-primary hover:text-primary-400"
        >
          decide who can get in
        </Link>{" "}
        ·{" "}
        <Link to="/team" className="text-primary hover:text-primary-400">
          invite your team to {namespace}
        </Link>
      </p>

      <ConnectModal
        open={connecting}
        onClose={closeConnect}
        deviceUid={device.uid}
        deviceName={name}
        sshid={buildSshid(namespace, name)}
      />
    </div>
  );
}
