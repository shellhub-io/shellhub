import { format } from "date-fns";
import { generateRandomUUID } from "@/utils/random-uuid";

// Client-side web-terminal session recording (Community Edition). Captures the
// terminal output to an asciinema v2 (.cast) file streamed into the browser's
// OPFS (Origin Private File System) — no server storage, no folder picker,
// bounded memory. Recordings are listed in an in-app manager and exported via a
// normal browser download.

/** OPFS subdirectory holding `<id>.cast` payloads and `<id>.json` sidecars. */
const DIR = "session-recordings";

export interface RecordingMeta {
  /** OPFS basename (uuid). */
  id: string;
  /** Friendly filename used for the browser download. */
  filename: string;
  deviceName: string;
  /** Device UID — used to pair a local recording with its server session. */
  deviceUid: string;
  /** SSH username the session connected with. */
  username: string;
  /**
   * Server session UID, when known. Lets a local recording dedupe exactly
   * against its server-side counterpart. Absent until the server reports it.
   */
  sessionUid?: string;
  /** Initial terminal dimensions (asciicast header). */
  width: number;
  height: number;
  /** Recording length in seconds. */
  durationSec: number;
  /** Session start, epoch milliseconds. */
  createdAt: number;
  /** Size of the .cast payload in bytes (filled when listing). */
  size: number;
}

function headerLine(cols: number, rows: number): string {
  return `${JSON.stringify({
    version: 2,
    width: cols,
    height: rows,
    timestamp: Math.floor(Date.now() / 1000),
  })}\n`;
}

function outputLine(elapsed: number, text: string): string {
  return `${JSON.stringify([elapsed, "o", text])}\n`;
}

function resizeLine(elapsed: number, cols: number, rows: number): string {
  return `${JSON.stringify([elapsed, "r", `${cols}x${rows}`])}\n`;
}

/** True when the browser can stream recordings to OPFS (the only backend). */
export function isRecordingSupported(): boolean {
  return (
    typeof navigator !== "undefined" &&
    typeof navigator.storage?.getDirectory === "function" &&
    typeof FileSystemFileHandle !== "undefined" &&
    typeof FileSystemFileHandle.prototype.createWritable === "function"
  );
}

// OPFS is scoped per browser origin, not per authenticated user, so recordings
// are namespaced under the signed-in user's id: a second user on the same
// profile never sees or can replay another user's recordings, even within a
// shared namespace. Set by authStore as the session changes.
let userScope: string | null = null;

export function setRecordingsScope(userId: string | null): void {
  userScope = userId;
}

async function recordingsDir(): Promise<FileSystemDirectoryHandle> {
  if (!userScope) throw new Error("session recording: no user scope set");
  const root = await navigator.storage.getDirectory();
  const base = await root.getDirectoryHandle(DIR, { create: true });
  return base.getDirectoryHandle(userScope, { create: true });
}

/** Build a filesystem-safe `.cast` download name for a device. */
export function castFilename(deviceName: string): string {
  const slug = (deviceName || "session")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return `shellhub-${slug || "session"}-${format(new Date(), "yyyyMMdd-HHmmss")}.cast`;
}

/** Trigger a normal browser download (lands in the default Downloads dir). */
export function downloadCast(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

/**
 * Streams a terminal recording to an OPFS file. Created via `create()` (async,
 * opens the writable); `start()` writes the header; output/resize append event
 * lines; `finish()` closes the file and writes the sidecar metadata.
 */
export class OpfsCastRecorder {
  private startMs = 0;
  private started = false;
  private count = 0;
  private failed = false;
  private chain: Promise<void> = Promise.resolve();
  private cols = 80;
  private rows = 24;
  private lastElapsed = 0;

  private constructor(
    private readonly id: string,
    // Captured at create() so finalize/cleanup survive a userScope change (e.g.
    // logout nulls the scope before the unmount runs finish()).
    private readonly dir: FileSystemDirectoryHandle,
    private readonly writable: FileSystemWritableFileStream,
    private readonly deviceName: string,
    private readonly deviceUid: string,
    private readonly username: string,
    private sessionUid?: string,
  ) {}

  static async create(
    deviceName: string,
    deviceUid: string,
    username: string,
    sessionUid?: string,
  ): Promise<OpfsCastRecorder> {
    const dir = await recordingsDir();
    const id = generateRandomUUID();
    const handle = await dir.getFileHandle(`${id}.cast`, { create: true });
    const writable = await handle.createWritable();
    return new OpfsCastRecorder(
      id,
      dir,
      writable,
      deviceName,
      deviceUid,
      username,
      sessionUid,
    );
  }

  /** Record the server session UID once the server reports it. */
  setSessionUid(uid: string): void {
    this.sessionUid = uid;
  }

  start(cols: number, rows: number): void {
    if (this.started) return;
    this.started = true;
    this.startMs = Date.now();
    this.cols = cols;
    this.rows = rows;
    this.write(headerLine(cols, rows));
  }

  recordOutput(text: string): void {
    if (!this.started || this.failed) return;
    this.count += 1;
    this.lastElapsed = this.elapsed();
    this.write(outputLine(this.lastElapsed, text));
  }

  recordResize(cols: number, rows: number): void {
    if (!this.started || this.failed) return;
    this.count += 1;
    this.lastElapsed = this.elapsed();
    this.write(resizeLine(this.lastElapsed, cols, rows));
  }

  get eventCount(): number {
    return this.count;
  }

  /** Close the file and persist sidecar metadata. Returns null if empty. */
  async finish(): Promise<RecordingMeta | null> {
    try {
      await this.chain;
      await this.writable.close();
    } catch (err) {
      console.error("session recording: failed to close file", err);
      return null;
    }
    if (this.count === 0) {
      await this.removeFiles();
      return null;
    }
    const meta: RecordingMeta = {
      id: this.id,
      filename: castFilename(this.deviceName),
      deviceName: this.deviceName,
      deviceUid: this.deviceUid,
      username: this.username,
      sessionUid: this.sessionUid,
      width: this.cols,
      height: this.rows,
      durationSec: this.lastElapsed,
      createdAt: this.startMs,
      size: 0,
    };
    try {
      const sidecar = await this.dir.getFileHandle(`${this.id}.json`, {
        create: true,
      });
      const w = await sidecar.createWritable();
      await w.write(JSON.stringify(meta));
      await w.close();
    } catch (err) {
      console.error("session recording: failed to write metadata", err);
    }
    return meta;
  }

  /** Drop the file without keeping it (throwaway StrictMode remount, etc.). */
  async discard(): Promise<void> {
    this.failed = true;
    try {
      await this.writable.abort();
    } catch {
      // best-effort
    }
    await this.removeFiles();
  }

  private elapsed(): number {
    return (Date.now() - this.startMs) / 1000;
  }

  private write(line: string): void {
    this.chain = this.chain
      .then(() => this.writable.write(line))
      .catch((err) => {
        if (!this.failed) {
          this.failed = true;
          console.error("session recording: write failed", err);
        }
      });
  }

  private async removeFiles(): Promise<void> {
    try {
      await this.dir.removeEntry(`${this.id}.cast`).catch(() => undefined);
      await this.dir.removeEntry(`${this.id}.json`).catch(() => undefined);
    } catch {
      // best-effort
    }
  }
}

/** List stored recordings, newest first. Empty when OPFS is unsupported. */
export async function listRecordings(): Promise<RecordingMeta[]> {
  if (!isRecordingSupported() || !userScope) return [];
  const dir = await recordingsDir();
  const metas: RecordingMeta[] = [];
  // `entries()` is an async iterator over [name, handle] pairs.
  const entries = (
    dir as unknown as {
      entries(): AsyncIterableIterator<[string, FileSystemHandle]>;
    }
  ).entries();
  for await (const [name, handle] of entries) {
    if (!name.endsWith(".json") || handle.kind !== "file") continue;
    try {
      const metaFile = await (handle as FileSystemFileHandle).getFile();
      const meta = JSON.parse(await metaFile.text()) as RecordingMeta;
      const castHandle = await dir.getFileHandle(`${meta.id}.cast`);
      meta.size = (await castHandle.getFile()).size;
      metas.push(meta);
    } catch {
      // Skip orphaned/corrupt sidecars (e.g. .cast was removed).
    }
  }
  return metas.sort((a, b) => b.createdAt - a.createdAt);
}

/** Read a recording's bytes and trigger a browser download. */
export async function downloadRecording(meta: RecordingMeta): Promise<void> {
  const dir = await recordingsDir();
  const handle = await dir.getFileHandle(`${meta.id}.cast`);
  const file = await handle.getFile();
  downloadCast(
    new Blob([await file.arrayBuffer()], { type: "application/x-asciicast" }),
    meta.filename,
  );
}

/** Read a recording's .cast content as a string (for inline playback). */
export async function readRecording(meta: RecordingMeta): Promise<string> {
  const dir = await recordingsDir();
  const handle = await dir.getFileHandle(`${meta.id}.cast`);
  return (await handle.getFile()).text();
}

/** Delete a recording (payload + sidecar). */
export async function deleteRecording(id: string): Promise<void> {
  const dir = await recordingsDir();
  await dir.removeEntry(`${id}.cast`).catch(() => undefined);
  await dir.removeEntry(`${id}.json`).catch(() => undefined);
}

/** Delete every stored recording (payloads + sidecars). */
export async function clearRecordings(): Promise<void> {
  if (!isRecordingSupported() || !userScope) return;
  const dir = await recordingsDir();
  const names: string[] = [];
  const entries = (
    dir as unknown as {
      entries(): AsyncIterableIterator<[string, FileSystemHandle]>;
    }
  ).entries();
  for await (const [name, handle] of entries) {
    if (handle.kind === "file") names.push(name);
  }
  await Promise.all(
    names.map((n) => dir.removeEntry(n).catch(() => undefined)),
  );
}

/** Delete recordings whose start time is older than `maxAgeDays`. */
export async function pruneRecordings(maxAgeDays: number): Promise<void> {
  if (!isRecordingSupported() || !userScope || maxAgeDays <= 0) return;
  const cutoff = Date.now() - maxAgeDays * 86_400_000;
  const metas = await listRecordings();
  await Promise.all(
    metas.filter((m) => m.createdAt < cutoff).map((m) => deleteRecording(m.id)),
  );
}
