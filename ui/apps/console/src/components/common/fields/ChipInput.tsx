import { useState, KeyboardEvent } from "react";
import { LABEL } from "@/utils/styles";
import FieldHint from "@/components/common/fields/FieldHint";
import { XMarkIcon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";

export default function ChipInput({
  id,
  label,
  placeholder,
  hint,
  values,
  onChange,
}: {
  id: string;
  label: string;
  placeholder: string;
  hint: string;
  values: string[];
  onChange: (next: string[]) => void;
}) {
  const [draft, setDraft] = useState("");

  const commit = (raw: string) => {
    const token = raw.trim();
    if (!token) {
      setDraft("");
      return;
    }
    if (!values.includes(token)) onChange([...values, token]);
    setDraft("");
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      commit(draft);
    } else if (e.key === "Backspace" && draft === "" && values.length > 0) {
      onChange(values.slice(0, -1));
    }
  };

  const hintId = `${id}-hint`;

  return (
    <div>
      <span className={LABEL}>{label}</span>
      <div className="flex flex-wrap gap-1.5 min-h-[42px] px-3 py-2 bg-card border border-border rounded-lg cursor-text transition-all focus-within:border-primary/50 focus-within:ring-1 focus-within:ring-primary/20">
        {values.map((v) => (
          <span
            key={v}
            className="inline-flex items-center gap-1 px-2 py-0.5 bg-primary/10 text-primary text-xs rounded-md font-medium"
          >
            {v}
            <IconButton
              size="sm"
              aria-label={`Remove ${v}`}
              onClick={(e) => {
                e.stopPropagation();
                onChange(values.filter((x) => x !== v));
              }}
            >
              <XMarkIcon className="w-3 h-3" strokeWidth={2} />
            </IconButton>
          </span>
        ))}
        <input
          id={id}
          type="text"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={handleKeyDown}
          onBlur={() => commit(draft)}
          placeholder={values.length === 0 ? placeholder : ""}
          aria-describedby={hintId}
          className="flex-1 min-w-[120px] bg-transparent text-sm text-text-primary placeholder:text-text-secondary outline-none"
        />
      </div>
      <FieldHint id={hintId}>{hint}</FieldHint>
    </div>
  );
}
