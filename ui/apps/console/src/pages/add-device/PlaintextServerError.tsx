import { Callout } from "@shellhub/design-system/primitives";

function isLoopbackHost(hostname: string): boolean {
  return (
    hostname === "localhost" ||
    hostname.endsWith(".localhost") ||
    hostname === "[::1]" ||
    /^127(\.\d{1,3}){3}$/.test(hostname)
  );
}

/**
 * Explains that devices installed from origin will not connect, because the agent refuses a plain
 * http:// server address unless its operator sets SHELLHUB_ALLOW_INSECURE_SERVER=true. Renders
 * nothing for an https origin or a loopback host, which the agent accepts.
 */
export default function PlaintextServerError({ origin }: { origin: string }) {
  const server = new URL(origin);
  if (server.protocol !== "http:" || isLoopbackHost(server.hostname)) {
    return null;
  }

  return (
    <Callout variant="error" className="mb-6">
      This ShellHub server is served over plain http://, which the agent
      refuses to connect to. Anyone on the network path between a device and
      the server could impersonate the server and get a root shell on the
      device. Serve ShellHub over https:// to add devices, or accept that risk
      by adding <code>SHELLHUB_ALLOW_INSECURE_SERVER=true</code> to the
      install command, before <code>sh</code>.
    </Callout>
  );
}
