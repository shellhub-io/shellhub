import { useState } from "react";
import {
  ShareIcon,
  EyeIcon,
  ClipboardDocumentIcon,
  CheckIcon,
  ArrowTopRightOnSquareIcon,
  XCircleIcon,
} from "@heroicons/react/24/outline";
import { useShares, useRevokeShare } from "@/hooks/useShares";
import type { Share } from "@/types/share";
import PageHeader from "@/components/common/PageHeader";
import DataTable, { type Column } from "@/components/common/DataTable";
import DeviceChip from "@/components/common/DeviceChip";
import Spinner from "@/components/common/Spinner";
import { formatDuration } from "@/utils/date";

function CopyLinkButton({ url }: { url: string }) {
  const [copied, setCopied] = useState(false);

  const handleClick = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard unavailable — ignore.
    }
  };

  return (
    <button
      onClick={(e) => void handleClick(e)}
      title="Copy link"
      className={`inline-flex items-center justify-center gap-1 min-w-[5.5rem] px-2.5 py-1 text-2xs font-semibold rounded-md border transition-all ${
        copied
          ? "bg-accent-green/10 text-accent-green border-accent-green/20"
          : "bg-primary/10 text-primary border-primary/20 hover:bg-primary/20"
      }`}
    >
      {copied ? (
        <CheckIcon className="w-3 h-3" />
      ) : (
        <ClipboardDocumentIcon className="w-3 h-3" />
      )}
      {copied ? "Copied" : "Copy link"}
    </button>
  );
}

function EndButton({ onEnd }: { onEnd: () => Promise<unknown> }) {
  const [ending, setEnding] = useState(false);

  const handleClick = async (e: React.MouseEvent) => {
    e.stopPropagation();
    setEnding(true);
    try {
      await onEnd();
    } finally {
      setEnding(false);
    }
  };

  return (
    <button
      onClick={(e) => void handleClick(e)}
      disabled={ending}
      title="End share"
      className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/10 transition-colors disabled:opacity-dim"
    >
      {ending ? (
        <Spinner size="xs" tone="onSurface" />
      ) : (
        <XCircleIcon className="w-4 h-4" strokeWidth={2} />
      )}
    </button>
  );
}

export default function SharedTerminals() {
  const { shares, isLoading } = useShares();
  const revokeShare = useRevokeShare();

  const columns: Column<Share>[] = [
    {
      key: "name",
      header: "Name",
      render: (s) => (
        <div className="flex items-center gap-2.5">
          <span
            className="h-2 w-2 shrink-0 rounded-full bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]"
            title="Live"
          />
          {s.name ? (
            <span className="text-sm text-text-primary">{s.name}</span>
          ) : (
            <span className="text-sm italic text-text-muted">Untitled</span>
          )}
        </div>
      ),
    },
    {
      key: "command",
      header: "Command",
      render: (s) =>
        s.command ? (
          <code className="text-xs font-mono text-text-secondary bg-surface px-1.5 py-0.5 rounded">
            {s.command}
          </code>
        ) : (
          <span className="text-xs font-mono text-text-muted">login shell</span>
        ),
    },
    {
      key: "mode",
      header: "Mode",
      render: (s) =>
        s.writable ? (
          <span className="inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border border-primary/30 bg-primary/10 text-primary">
            Collaborative
          </span>
        ) : (
          <span className="inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border border-border text-text-muted">
            Read-only
          </span>
        ),
    },
    {
      key: "device",
      header: "Device",
      render: (s) => (
        <DeviceChip
          uid={s.device_uid}
          name={s.device_name || s.device_uid.substring(0, 8)}
          online={s.device_online}
          osId={s.device_os}
          onClick={(e) => e.stopPropagation()}
        />
      ),
    },
    {
      key: "viewers",
      header: "Viewers",
      render: (s) => (
        <span
          className={`inline-flex items-center gap-1.5 text-xs font-mono tabular-nums ${
            s.viewers > 0 ? "text-text-primary" : "text-text-muted"
          }`}
        >
          <EyeIcon className="w-3.5 h-3.5 text-text-muted" />
          {s.viewers}
        </span>
      ),
    },
    {
      key: "duration",
      header: "Duration",
      render: (s) => {
        const noLimit = new Date(s.expires_at).getFullYear() < 2000;
        return (
          <div className="flex flex-col">
            <span className="text-xs font-mono text-text-secondary tabular-nums">
              {formatDuration(s.created_at, s.created_at, true)}
            </span>
            <span className="text-2xs text-text-muted/70">
              {noLimit
                ? "no limit"
                : `${formatDuration(new Date().toISOString(), s.expires_at, false)} left`}
            </span>
          </div>
        );
      },
    },
    {
      key: "actions",
      header: "",
      headerClassName: "w-52",
      render: (s) => (
        <div className="flex items-center justify-end gap-1">
          <CopyLinkButton url={s.url} />
          <a
            href={s.url}
            target="_blank"
            rel="noreferrer"
            onClick={(e) => e.stopPropagation()}
            title="Open in new tab"
            className="inline-flex items-center gap-1 px-2.5 py-1 text-text-muted text-2xs font-semibold rounded-md hover:text-text-primary hover:bg-surface border border-border transition-all"
          >
            <ArrowTopRightOnSquareIcon className="w-3 h-3" />
            Open
          </a>
          <EndButton onEnd={() => revokeShare.mutateAsync(s.token)} />
        </div>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<ShareIcon className="w-6 h-6" />}
        overline="Terminals"
        title="Shared Terminals"
        description="Live terminals shared from your devices and the number of viewers watching each one."
      />

      <DataTable
        columns={columns}
        data={shares}
        rowKey={(s) => s.token}
        isLoading={isLoading}
        itemLabel="shared terminal"
        emptyState={
          <div className="text-center">
            <ShareIcon
              className="mx-auto mb-3 h-10 w-10 text-text-muted/30"
              strokeWidth={1}
            />
            <p className="text-xs font-mono text-text-muted">
              No active shared terminals
            </p>
            <p className="mt-2 text-2xs text-text-muted/70">
              Run{" "}
              <code className="rounded bg-surface px-1 py-0.5 text-primary">
                shellhub-agent share
              </code>{" "}
              on a device to start one
            </p>
          </div>
        }
      />
    </div>
  );
}
