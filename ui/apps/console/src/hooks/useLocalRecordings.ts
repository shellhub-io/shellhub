import { useEffect } from "react";
import { useRecordingsStore } from "@/stores/recordingsStore";
import { isRecordingSupported } from "@/utils/recordings";

/**
 * The recordings this browser holds, reloaded from its storage when the calling component mounts,
 * which also drops those past the retention. Empty in a browser that cannot record.
 */
export function useLocalRecordings() {
  const recordings = useRecordingsStore((s) => s.recordings);
  const refresh = useRecordingsStore((s) => s.refresh);

  useEffect(() => {
    if (isRecordingSupported()) void refresh();
  }, [refresh]);

  return recordings;
}
