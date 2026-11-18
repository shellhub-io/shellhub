import { Card } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import CopyButton from "@/components/common/CopyButton";
import { useCurrentNamespace } from "@/hooks/useNamespaceRequests";
import { buildSshid } from "@/utils/sshid";
import { LABEL_BASE } from "@/utils/styles";

interface SshidCardProps {
  uid: string;
  name: string;
  className?: string;
}

/**
 * The SSHID of a device or container in the active namespace, with a copy button. Shows the uid
 * until the namespace's name has loaded.
 */
export default function SshidCard({ uid, name, className }: SshidCardProps) {
  const { namespace } = useCurrentNamespace();
  const sshid = namespace?.name ? buildSshid(namespace.name, name) : uid;

  return (
    <Card
      className={cn("flex items-center justify-between gap-4 p-4", className)}
    >
      <div className="min-w-0">
        <p className={LABEL_BASE}>SSHID</p>
        <code className="mt-0.5 block break-all font-mono text-sm text-accent-cyan">
          {sshid}
        </code>
      </div>
      <CopyButton text={sshid} />
    </Card>
  );
}
