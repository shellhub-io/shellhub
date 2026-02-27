import { CheckCircleIcon, ArrowUpTrayIcon } from "@heroicons/react/24/outline";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { useKeyFileInput } from "@/hooks/useKeyFileInput";
import { INPUT_MONO, LABEL } from "@/utils/styles";

interface KeyFileInputProps {
  label: string;
  value: string;
  onChange: (v: string) => void;
  validate: (text: string) => boolean;
  onFileName?: (name: string) => void;
  disabled?: boolean;
  error?: string | null;
  accept?: string;
  placeholder?: string;
  rows?: number;
  hint?: string;
  disabledHint?: string;
  loadedLabel?: string;
  emptyLabel?: string;
  id?: string;
}

export default function KeyFileInput({
  label,
  value,
  onChange,
  validate,
  onFileName,
  disabled,
  error,
  accept = ".pub,.pem,.key,.txt",
  placeholder = "",
  rows = 3,
  hint,
  disabledHint,
  loadedLabel = "Key loaded",
  emptyLabel = "Drop key file, paste, or browse",
  id,
}: KeyFileInputProps) {
  const {
    fileInputRef,
    dragging,
    inputMode,
    setInputMode,
    handleDrop,
    handleFileInputChange,
    setDragging,
  } = useKeyFileInput({ validate, onChange, onFileName, disabled });

  return (
    <div>
      <div className="flex items-center justify-between mb-1.5">
        <label htmlFor={id} className={LABEL + " !mb-0"}>
          {label}
        </label>
        {!disabled && (
          <div className="flex gap-1">
            <button
              type="button"
              onClick={() => setInputMode("file")}
              className={`px-2 py-0.5 rounded text-2xs font-medium transition-all ${inputMode === "file" ? "bg-primary/10 text-primary" : "text-text-muted hover:text-text-secondary"}`}
            >
              File
            </button>
            <button
              type="button"
              onClick={() => setInputMode("text")}
              className={`px-2 py-0.5 rounded text-2xs font-medium transition-all ${inputMode === "text" ? "bg-primary/10 text-primary" : "text-text-muted hover:text-text-secondary"}`}
            >
              Text
            </button>
          </div>
        )}
      </div>

      {inputMode === "file" && !disabled ? (
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
            accept={accept}
            className="hidden"
            onChange={handleFileInputChange}
          />
          {value ? (
            <>
              <CheckCircleIcon className="w-5 h-5 text-accent-green" />
              <span className="text-xs text-accent-green font-medium">
                {loadedLabel}
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
              <span className="text-xs text-text-secondary">{emptyLabel}</span>
            </>
          )}
        </div>
      ) : (
        <textarea
          id={id}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={disabled ? "" : placeholder}
          rows={rows}
          disabled={disabled}
          aria-invalid={!!error}
          aria-describedby={error && id ? `${id}-error` : undefined}
          className={`${INPUT_MONO} resize-none ${disabled ? "opacity-50 cursor-not-allowed" : ""}`}
        />
      )}

      {!disabled && hint && (
        <p className="mt-1 text-2xs text-text-muted">{hint}</p>
      )}
      {disabled && disabledHint && (
        <p className="mt-1 text-2xs text-text-muted">{disabledHint}</p>
      )}
      {error && (
        <p
          id={id ? `${id}-error` : undefined}
          className="text-2xs text-accent-red mt-1.5 flex items-center gap-1"
        >
          <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" />
          {error}
        </p>
      )}
    </div>
  );
}
