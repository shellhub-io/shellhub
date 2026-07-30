// Encodings a device may emit. The browser decodes them, so the terminal shows
// what the device wrote rather than the gateway's interpretation of it.
export const TERMINAL_ENCODINGS = [
  { value: "utf-8", label: "Unicode (UTF-8)" },
  { value: "gbk", label: "Chinese Simplified (GBK)" },
  { value: "gb18030", label: "Chinese Simplified (GB18030)" },
  { value: "big5", label: "Chinese Traditional (Big5)" },
  { value: "shift_jis", label: "Japanese (Shift_JIS)" },
  { value: "euc-jp", label: "Japanese (EUC-JP)" },
  { value: "euc-kr", label: "Korean (EUC-KR)" },
  { value: "windows-1251", label: "Cyrillic (Windows-1251)" },
  { value: "windows-1252", label: "Western (Windows-1252)" },
  { value: "iso-8859-15", label: "Western (ISO-8859-15)" },
] as const;

export type TerminalEncoding = (typeof TERMINAL_ENCODINGS)[number]["value"];

export const DEFAULT_TERMINAL_ENCODING: TerminalEncoding = "utf-8";

export const isTerminalEncoding = (value: string | null): value is TerminalEncoding =>
  TERMINAL_ENCODINGS.some((encoding) => encoding.value === value);

export interface TerminalOutputDecoder {
  decode(bytes: Uint8Array): string;
}

/**
 * Creates a decoder for one terminal session's output.
 *
 * The decoder is stateful across calls: a multi-byte character split across
 * WebSocket frames is held until the bytes completing it arrive, for every
 * encoding rather than just UTF-8. An encoding the browser does not know falls
 * back to UTF-8 instead of throwing.
 */
export const createOutputDecoder = (
  encoding: string = DEFAULT_TERMINAL_ENCODING,
): TerminalOutputDecoder => {
  let decoder: TextDecoder;

  try {
    decoder = new TextDecoder(encoding, { fatal: false });
  } catch {
    decoder = new TextDecoder(DEFAULT_TERMINAL_ENCODING, { fatal: false });
  }

  return {
    decode: (bytes: Uint8Array) => decoder.decode(bytes, { stream: true }),
  };
};
