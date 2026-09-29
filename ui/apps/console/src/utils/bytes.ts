/**
 * A size in bytes as a person reads it: whole kilobytes below a megabyte, whole megabytes above.
 */
export function formatBytes(bytes: number): string {
  return bytes >= 1024 * 1024
    ? `${Math.round(bytes / (1024 * 1024))} MB`
    : `${Math.round(bytes / 1024)} KB`;
}
