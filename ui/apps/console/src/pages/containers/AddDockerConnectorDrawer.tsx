import { CubeIcon } from "@heroicons/react/24/outline";
import Drawer from "@/components/common/Drawer";
import CopyButton from "@/components/common/CopyButton";
import { useAuthStore } from "@/stores/authStore";

function AddDockerConnectorDrawer({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const tenant = useAuthStore((s) => s.tenant) ?? "";
  const origin = window.location.origin;
  const command = `curl -sSf ${origin}/install.sh | TENANT_ID=${tenant} SERVER_ADDRESS=${origin} sh -s connector`;

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Registering a Docker Host"
      subtitle="Install the ShellHub Connector to add Docker containers"
      icon={<CubeIcon className="w-4 h-4 text-primary" />}
      width="md"
    >
      <div className="space-y-5">
        <p className="text-sm text-text-secondary leading-relaxed">
          In order to add Docker containers to ShellHub, you need to install the
          ShellHub Connector on the Docker host.
        </p>

        <p className="text-sm text-text-secondary leading-relaxed">
          The easiest way to install the ShellHub Connector is with our automatic
          one-line installation script, which connects to the Docker API and
          exposes the running containers within ShellHub.
        </p>

        <div>
          <p className="text-xs font-semibold text-text-primary mb-2">
            Run the following command on your Docker host:
          </p>
          <div className="bg-card border border-border rounded-xl p-4 relative group">
            <pre className="text-xs font-mono text-accent-cyan break-all whitespace-pre-wrap pr-8">
              {command}
            </pre>
            <div className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity">
              <CopyButton text={command} />
            </div>
          </div>
        </div>

        <div className="bg-primary/5 border border-primary/15 rounded-xl p-4">
          <p className="text-xs text-text-secondary leading-relaxed">
            <span className="font-semibold text-primary">Note:</span> Once the
            connector is running, any Docker containers on the host will
            automatically appear under the Pending tab. Accept them to allow
            SSH access.
          </p>
        </div>
      </div>
    </Drawer>
  );
}

export default AddDockerConnectorDrawer;
