import {
  ShieldCheckIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import BaseDialog from "@/components/common/BaseDialog";
import CopyButton from "@/components/common/CopyButton";
import { Button } from "@shellhub/design-system/primitives";
import { useHostKey, useForgetHostKey } from "@/hooks/useHostKeys";
import type { HostKeyScope } from "@/api/hostKeys";

interface Props {
  open: boolean;
  onClose: () => void;
  host: string;
  port: number;
  scope: HostKeyScope;
  // Whether the caller may forget the key (owner for personal, operator+ for team).
  canForget: boolean;
}

function Row({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-start gap-3">
      <span className="w-20 shrink-0 text-2xs uppercase tracking-wider text-text-muted pt-0.5">
        {label}
      </span>
      <div className="flex-1 min-w-0 text-text-primary">{children}</div>
    </div>
  );
}

export default function HostKeyModal({
  open,
  onClose,
  host,
  port,
  scope,
  canForget,
}: Props) {
  const { knownHost, isLoading } = useHostKey(host, port, scope, open);
  const forget = useForgetHostKey();

  const titleId = "host-key-title";

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="md"
      aria-labelledby={titleId}
    >
      <div className="p-5 space-y-5">
        <div>
          <h2
            id={titleId}
            className="text-base font-semibold text-text-primary flex items-center gap-2"
          >
            <ShieldCheckIcon className="w-5 h-5 text-text-secondary" />
            Host key
          </h2>
          <p className="text-xs font-mono text-text-secondary mt-1">
            {host}:{port}
          </p>
        </div>

        {isLoading ? (
          <p className="text-xs font-mono text-text-muted">Loading…</p>
        ) : knownHost ? (
          <div className="space-y-3 text-sm">
            <Row label="Status">
              <span className="inline-flex items-center gap-1.5 text-accent-green">
                <ShieldCheckIcon className="w-4 h-4" />
                Trusted
              </span>
            </Row>
            <Row label="Type">
              <span className="font-mono text-xs">{knownHost.key_type}</span>
            </Row>
            <Row label="SHA256">
              <div className="flex items-center gap-2">
                <code className="font-mono text-xs break-all text-text-secondary">
                  {knownHost.fingerprint}
                </code>
                <CopyButton text={knownHost.fingerprint} />
              </div>
            </Row>
            <Row label="Added">
              <span className="text-xs text-text-secondary">
                {new Date(knownHost.created_at).toLocaleString()}
              </span>
            </Row>
          </div>
        ) : (
          <div className="rounded-lg border border-border bg-card p-3 flex items-start gap-2">
            <ExclamationTriangleIcon className="w-4 h-4 text-text-muted shrink-0 mt-px" />
            <p className="text-xs text-text-secondary">
              No host key stored yet. It will be verified the first time you
              connect.
            </p>
          </div>
        )}

        <div className="flex items-center justify-between pt-1">
          {knownHost && canForget ? (
            <button
              type="button"
              onClick={() =>
                forget.mutate({ host, port, scope }, { onSuccess: onClose })
              }
              disabled={forget.isPending}
              className="text-2xs font-semibold text-accent-red hover:underline disabled:opacity-dim"
            >
              {forget.isPending ? "Forgetting…" : "Forget key"}
            </button>
          ) : (
            <span />
          )}
          <Button variant="secondary" type="button" onClick={onClose}>
            Close
          </Button>
        </div>
      </div>
    </BaseDialog>
  );
}
