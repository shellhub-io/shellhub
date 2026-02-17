import {
  useState,
  useRef,
  useCallback,
  useEffect,
  DragEvent,
  ClipboardEvent,
} from "react";
import { CheckCircleIcon, ArrowUpTrayIcon } from "@heroicons/react/24/outline";
import { isPublicKeyValid } from "../../utils/sshKeys";
import { LABEL, INPUT_MONO } from "../../utils/styles";

/* --- File / Text Key Input --- */
function KeyDataInput({
  value,
  onChange,
  error,
  disabled,
  onFileName,
}: {
  value: string;
  onChange: (v: string) => void;
  error?: string;
  disabled?: boolean;
  onFileName?: (name: string) => void;
}) {
  const [mode, setMode] = useState<"file" | "text">(disabled ? "text" : "file");
  const [dragging, setDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const processFile = (file: File) => {
    if (file.size > 512 * 1024) return;
    const reader = new FileReader();
    reader.onload = () => {
      const text = reader.result as string;
      onChange(text);
      if (onFileName) {
        const base = file.name.replace(/\.[^.]+$/, "");
        onFileName(base);
      }
    };
    reader.readAsText(file);
  };

  const handleDrop = (e: DragEvent) => {
    e.preventDefault();
    setDragging(false);
    const file = e.dataTransfer?.files?.[0];
    if (file) processFile(file);
  };

  const handlePaste = useCallback(
    (e: ClipboardEvent | Event) => {
      if (disabled) return;
      const ce = e as ClipboardEvent;
      const text = ce.clipboardData?.getData("text");
      if (text && isPublicKeyValid(text)) {
        e.preventDefault();
        onChange(text);
      }
    },
    [disabled, onChange],
  );

  useEffect(() => {
    if (disabled) return;
    const handler = (e: Event) => handlePaste(e);
    document.addEventListener("paste", handler);
    return () => document.removeEventListener("paste", handler);
  }, [disabled, handlePaste]);

  return (
    <div>
      <div className="flex items-center justify-between mb-1.5">
        <label className={LABEL + " !mb-0"}>Public key data</label>
        {!disabled && (
          <div className="flex gap-1">
            <button
              type="button"
              onClick={() => setMode("file")}
              className={`px-2 py-0.5 rounded text-2xs font-medium transition-all ${mode === "file" ? "bg-primary/10 text-primary" : "text-text-muted hover:text-text-secondary"}`}
            >
              File
            </button>
            <button
              type="button"
              onClick={() => setMode("text")}
              className={`px-2 py-0.5 rounded text-2xs font-medium transition-all ${mode === "text" ? "bg-primary/10 text-primary" : "text-text-muted hover:text-text-secondary"}`}
            >
              Text
            </button>
          </div>
        )}
      </div>

      {mode === "file" && !disabled ? (
        <div
          onDragOver={(e) => {
            e.preventDefault();
            setDragging(true);
          }}
          onDragLeave={() => setDragging(false)}
          onDrop={handleDrop}
          onClick={() => fileInputRef.current?.click()}
          className={`flex flex-col items-center justify-center gap-2 px-4 py-6 border-2 border-dashed rounded-lg cursor-pointer transition-all ${
            dragging
              ? "border-primary bg-primary/5"
              : value
                ? "border-accent-green/30 bg-accent-green/5"
                : `border-border hover:border-primary/30 ${error ? "border-accent-red/30" : ""}`
          }`}
        >
          <input
            ref={fileInputRef}
            type="file"
            accept=".pub,.pem,.key,.txt"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) processFile(file);
            }}
          />
          {value ? (
            <>
              <CheckCircleIcon className="w-5 h-5 text-accent-green" />
              <span className="text-xs text-accent-green font-medium">
                Key loaded
              </span>
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  onChange("");
                }}
                className="text-2xs text-text-muted hover:text-text-primary transition-colors"
              >
                Clear
              </button>
            </>
          ) : (
            <>
              <ArrowUpTrayIcon className="w-5 h-5 text-text-muted" />
              <span className="text-xs text-text-secondary">
                Drop <span className="font-mono">.pub</span> file, paste, or
                browse
              </span>
            </>
          )}
        </div>
      ) : (
        <textarea
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={disabled ? "" : "ssh-rsa AAAAB3NzaC1yc2E..."}
          rows={3}
          disabled={disabled}
          className={`${INPUT_MONO} resize-none ${disabled ? "opacity-50 cursor-not-allowed" : ""}`}
        />
      )}
      {!disabled && (
        <p className="mt-1 text-2xs text-text-muted">
          RSA, DSA, ECDSA, ED25519 — PEM and OpenSSH formats.
        </p>
      )}
      {disabled && (
        <p className="mt-1 text-2xs text-text-muted">
          Public key data cannot be modified after creation.
        </p>
      )}
      {error && <p className="mt-1 text-2xs text-accent-red">{error}</p>}
    </div>
  );
}

export default KeyDataInput;
