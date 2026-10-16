import { Link } from "react-router-dom";
import { CubeIcon } from "@heroicons/react/24/outline";
import { Button, Card } from "@shellhub/design-system/primitives";
import Modal from "@/components/common/Modal";
import CopyButton from "@/components/common/CopyButton";
import { useNamespaceLegacy } from "@/hooks/useNamespaceLegacy";
import { useAuthStore } from "@/stores/authStore";
import { buildConnectorCommand } from "@/utils/installCommand";

/**
 * Explains how to attach a Docker host so its containers appear as devices. A legacy namespace
 * gets the command with its tenant; a new one enrolls containers with a provisioning key, which
 * is revealed under Add Device › Fleet by whoever may read it.
 */
function AddDockerConnectorModal({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const tenant = useAuthStore((s) => s.tenant) ?? "";
  const legacy = useNamespaceLegacy() === true;
  const command = buildConnectorCommand(
    `TENANT_ID=${tenant}`,
    window.location.origin,
  );

  return (
    <Modal
      open={open}
      onClose={onClose}
      icon={<CubeIcon />}
      title="Register a Docker host"
      description="Install the ShellHub Connector on the host to bring its containers into ShellHub."
    >
      <div className="space-y-5">
        <p className="text-sm text-text-secondary leading-relaxed">
          In order to add Docker containers to ShellHub, you need to install the
          ShellHub Connector on the Docker host.
        </p>

        {legacy ? (
          <>
            <p className="text-sm text-text-secondary leading-relaxed">
              The easiest way to install the ShellHub Connector is with our
              automatic one-line installation script, which connects to the
              Docker API and exposes the running containers within ShellHub.
            </p>

            <div>
              <p className="text-xs font-semibold text-text-primary mb-2">
                Run the following command on your Docker host:
              </p>
              <Card className="p-4 relative group">
                <pre className="text-xs font-mono text-accent-cyan break-all whitespace-pre-wrap pr-8">
                  {command}
                </pre>
                <div className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity">
                  <CopyButton text={command} />
                </div>
              </Card>
            </div>

            <div className="bg-primary/5 border border-primary/15 rounded-xl p-4">
              <p className="text-xs text-text-secondary leading-relaxed">
                <span className="font-semibold text-primary">Note:</span> Once
                the connector is running, any Docker containers on the host will
                wait as pending devices, linked from the Devices page. Accept
                them to allow SSH access.
              </p>
            </div>
          </>
        ) : (
          <>
            <p className="text-sm text-text-secondary leading-relaxed">
              The connector enrolls the host&apos;s containers with a
              provisioning key. Pick the key under Add Device › Fleet, choose
              the Docker connector method, and run the command it shows on the
              Docker host.
            </p>
            <Button as={Link} to="/devices/add/fleet" onClick={onClose}>
              Open Fleet
            </Button>
          </>
        )}
      </div>
    </Modal>
  );
}

export default AddDockerConnectorModal;
