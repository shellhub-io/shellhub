import {
  CheckCircleIcon,
  BookOpenIcon,
  ChatBubbleLeftRightIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace } from "@/hooks/useNamespaces";
import type { NormalizedDevice } from "@/hooks/useDevices";
import CopyButton from "@/components/common/CopyButton";
import { Button } from "@shellhub/design-system/primitives";

interface WizardStep4CompleteProps {
  device: NormalizedDevice | null;
}

export default function WizardStep4Complete({
  device,
}: WizardStep4CompleteProps) {
  const username = useAuthStore((s) => s.username);
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: ns } = useNamespace(tenantId);
  const namespace = ns?.name;
  const hostname = window.location.hostname;

  const sshCmd =
    device && username && namespace
      ? `ssh ${username}@${namespace}.${device.name}@${hostname}`
      : null;

  return (
    <div className="py-2 flex flex-col gap-6">
      {/* Success card */}
      <div className="bg-accent-green/8 border border-accent-green/20 rounded-xl px-5 py-6 flex flex-col items-center text-center gap-3">
        <div className="w-12 h-12 rounded-full bg-accent-green/15 border border-accent-green/25 flex items-center justify-center">
          <CheckCircleIcon
            className="w-6 h-6 text-accent-green"
            strokeWidth={1.5}
          />
        </div>
        <div>
          <p className="text-base font-mono font-bold text-text-primary">
            {device?.name ?? "Your device"} is online.
          </p>
          <p className="text-xs text-text-muted mt-1">
            Accepted and ready for SSH connections.
          </p>
        </div>
      </div>

      {/* SSH command */}
      {sshCmd && (
        <div>
          <h3 className="text-xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2">
            Connect via SSH
          </h3>
          <div className="relative group">
            <div className="bg-background border border-border rounded-xl p-4 pr-12 font-mono text-xs text-text-secondary leading-relaxed break-all">
              <span className="text-primary/50 select-none">$ </span>
              {sshCmd}
            </div>
            <div className="absolute top-3 right-3">
              <CopyButton text={sshCmd} size="md" />
            </div>
          </div>
        </div>
      )}

      {/* Resource links */}
      <nav
        aria-label="Resources"
        className="flex flex-wrap items-center justify-center gap-4"
      >
        <Button
          variant="secondary"
          as="a"
          size="sm"
          href="https://docs.shellhub.io"
          target="_blank"
          rel="noopener noreferrer"
          icon={<BookOpenIcon className="w-4 h-4" />}
        >
          Documentation
        </Button>
        <Button
          variant="secondary"
          as="a"
          size="sm"
          href="https://gitter.im/shellhub-io/community"
          target="_blank"
          rel="noopener noreferrer"
          icon={<ChatBubbleLeftRightIcon className="w-4 h-4" />}
        >
          Community
        </Button>
      </nav>
    </div>
  );
}
