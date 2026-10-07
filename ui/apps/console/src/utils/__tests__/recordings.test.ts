import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  OpfsCastRecorder,
  readRecording,
  setRecordingsScope,
} from "@/utils/recordings";

class FakeDirectory {
  readonly files = new Map<string, string>();
  readonly directories = new Map<string, FakeDirectory>();

  getDirectoryHandle(name: string) {
    const directory = this.directories.get(name) ?? new FakeDirectory();
    this.directories.set(name, directory);
    return Promise.resolve(directory);
  }

  getFileHandle(name: string) {
    const files = this.files;
    if (!files.has(name)) files.set(name, "");
    return Promise.resolve({
      getFile: () =>
        Promise.resolve({ text: () => Promise.resolve(files.get(name) ?? "") }),
      createWritable: () =>
        Promise.resolve({
          write: (chunk: string) => {
            files.set(name, `${files.get(name) ?? ""}${chunk}`);
            return Promise.resolve();
          },
          close: () => Promise.resolve(),
          abort: () => Promise.resolve(),
        }),
    });
  }

  removeEntry(name: string) {
    this.files.delete(name);
    return Promise.resolve();
  }
}

let storage: FakeDirectory;

function countCasts(directory: FakeDirectory): number {
  let casts = [...directory.files.keys()].filter((name) =>
    name.endsWith(".cast"),
  ).length;
  for (const child of directory.directories.values()) {
    casts += countCasts(child);
  }
  return casts;
}

async function startRecording() {
  const recorder = await OpfsCastRecorder.create(
    "device",
    "device-uid",
    "root",
  );
  recorder.start(80, 24);
  return recorder;
}

describe("OpfsCastRecorder", () => {
  beforeEach(() => {
    storage = new FakeDirectory();
    vi.stubGlobal("navigator", {
      storage: { getDirectory: () => Promise.resolve(storage) },
    });
    setRecordingsScope("user-1");
  });

  afterEach(() => {
    setRecordingsScope(null);
    vi.unstubAllGlobals();
  });

  it("keeps a recording of a session that printed output", async () => {
    const recorder = await startRecording();
    recorder.recordOutput("hello");

    const meta = await recorder.finish();

    if (!meta) throw new Error("expected the recording to be kept");
    const [, event] = (await readRecording(meta))
      .trim()
      .split("\n")
      .map((line) => JSON.parse(line) as unknown);
    expect(event).toEqual([expect.any(Number), "o", "hello"]);
  });

  it("drops a recording of a session that only resized", async () => {
    const recorder = await startRecording();
    recorder.recordResize(132, 50);

    expect(await recorder.finish()).toBeNull();
    expect(countCasts(storage)).toBe(0);
  });
});
