import { useId, useState } from "react";
import { TrashIcon, VideoCameraIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import SettingsSection from "@/components/settings/SettingsSection";
import SettingsField from "@/components/settings/SettingsField";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import RadioSegment from "@/components/common/fields/RadioSegment";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import { useRecordingsStore } from "@/stores/recordingsStore";
import { useLocalRecordings } from "@/hooks/useLocalRecordings";
import { isRecordingSupported } from "@/utils/recordings";
import { formatBytes } from "@/utils/bytes";

const RETENTIONS = [
  { value: "forever", label: "Forever", days: null },
  { value: "7", label: "7 days", days: 7 },
  { value: "30", label: "30 days", days: 30 },
  { value: "90", label: "90 days", days: 90 },
] as const;

type Retention = (typeof RETENTIONS)[number]["value"];

function retentionOf(days: number | null): Retention {
  return RETENTIONS.find((r) => r.days === days)?.value ?? "forever";
}

/**
 * The sessions this browser recorded in its own storage: how much room they take, wiping them, and
 * how long they are kept. Recordings older than the retention are dropped the next time the list
 * is read.
 */
export default function RecordingsPreferences() {
  const recordings = useLocalRecordings();
  const retentionDays = useRecordingsStore((s) => s.retentionDays);
  const setRetentionDays = useRecordingsStore((s) => s.setRetentionDays);
  const clearAll = useRecordingsStore((s) => s.clearAll);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const retentionId = useId();
  const supported = isRecordingSupported();

  const count = recordings.length;
  const used = recordings.reduce((total, r) => total + r.size, 0);

  return (
    <SettingsSection
      title="Local recordings"
      description="Sessions recorded in this browser. They stay in its storage and are never uploaded."
    >
      <div className="flex items-center gap-4 px-5 py-4 rounded-xl border border-border bg-card">
        <span className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shrink-0">
          <VideoCameraIcon className="w-5 h-5" />
        </span>
        <div className="min-w-0 flex-1">
          {supported ? (
            <>
              <p className="text-sm font-medium text-text-primary">
                {count} {count === 1 ? "recording" : "recordings"}
              </p>
              <p className="mt-0.5 text-xs text-text-muted font-mono">
                {formatBytes(used)} in this browser
              </p>
            </>
          ) : (
            <>
              <p className="text-sm font-medium text-text-primary">
                Not available here
              </p>
              <p className="mt-0.5 text-xs text-text-muted">
                This browser can't keep recordings of its own.
              </p>
            </>
          )}
        </div>
        {supported && (
          <Button
            size="sm"
            variant="dangerSoft"
            disabled={count === 0}
            onClick={() => setConfirmOpen(true)}
          >
            <TrashIcon className="w-4 h-4" />
            Delete all
          </Button>
        )}
      </div>

      {supported && (
        <SettingsField
          stacked
          titleId={retentionId}
          title="Keep recordings"
          description="Older recordings are deleted the next time the list loads."
        >
          <RadioGroupField
            labelledBy={retentionId}
            value={retentionOf(retentionDays)}
            onChange={(value: Retention) =>
              setRetentionDays(
                RETENTIONS.find((r) => r.value === value)?.days ?? null,
              )
            }
            containerClassName="flex gap-1 p-1 max-w-md rounded-lg border border-border bg-card whitespace-nowrap"
          >
            {RETENTIONS.map(({ value, label }) => (
              <RadioSegment key={value} value={value} label={label} />
            ))}
          </RadioGroupField>
        </SettingsField>
      )}

      <ConfirmDialog
        open={confirmOpen}
        onClose={() => setConfirmOpen(false)}
        onConfirm={async () => {
          await clearAll();
          setConfirmOpen(false);
        }}
        icon={<TrashIcon className="w-5 h-5" />}
        variant="danger"
        title="Delete all recordings in this browser?"
        description={`${count === 1 ? "The recording" : `All ${count} recordings`} kept in this browser will be deleted. Recordings on the server are not affected.`}
        confirmLabel="Delete all"
      />
    </SettingsSection>
  );
}
