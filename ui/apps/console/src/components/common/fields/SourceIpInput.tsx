import { useState, KeyboardEvent } from "react";
import {
  XMarkIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { LABEL } from "@/utils/styles";
import FieldHint from "@/components/common/fields/FieldHint";
import { parseSourceIp, sourceIpKind } from "@/utils/sourceIp";

const KIND_LABEL: Record<string, string> = {
  host: "host",
  private: "private",
  public: "public",
  any: "any",
  ipv6: "IPv6",
};

const SUGGESTIONS = [
  { value: "10.0.0.0/8", note: "private" },
  { value: "192.168.0.0/16", note: "private" },
  { value: "172.16.0.0/12", note: "private" },
];

// SourceIpInput is the access-policy Source IP field. Unlike a plain chip input it parses each
// entry live: a bare IP is normalized to a /32 host route, CIDRs are classified
// (private/public/size), an all-IPs entry is flagged, and nonsense is rejected before it can be
// added — so the user sees exactly what will be stored.
export default function SourceIpInput({
  id,
  label,
  hint,
  values,
  onChange,
}: {
  id: string;
  label: string;
  hint: string;
  values: string[];
  onChange: (next: string[]) => void;
}) {
  const [draft, setDraft] = useState("");
  const parsed = parseSourceIp(draft);
  const canAdd =
    parsed.status === "valid" ||
    parsed.status === "host" ||
    parsed.status === "any";

  const add = (cidr: string) => {
    if (!values.includes(cidr)) onChange([...values, cidr]);
    setDraft("");
  };

  const commitDraft = () => {
    if (canAdd) add(parsed.value);
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      commitDraft();
    } else if (e.key === "Backspace" && draft === "" && values.length > 0) {
      onChange(values.slice(0, -1));
    }
  };

  const hintId = `${id}-hint`;

  return (
    <div>
      <span className={LABEL}>{label}</span>
      <div className="flex flex-wrap gap-1.5 min-h-[42px] px-3 py-2 bg-card border border-border rounded-lg cursor-text transition-all focus-within:border-primary/50 focus-within:ring-1 focus-within:ring-primary/20">
        {values.map((v) => {
          const kind = sourceIpKind(v);
          const any = kind === "any";
          return (
            <span
              key={v}
              className={cn(
                "inline-flex items-center gap-1.5 pl-2 pr-1 py-0.5 text-xs rounded-md font-mono",
                any
                  ? "bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/40"
                  : "bg-primary/10 text-primary",
              )}
            >
              {v}
              {kind && (
                <span className="font-sans text-2xs text-text-muted bg-surface rounded px-1 py-px">
                  {KIND_LABEL[kind]}
                </span>
              )}
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
          );
        })}
        <input
          id={id}
          type="text"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={handleKeyDown}
          onBlur={commitDraft}
          placeholder={
            values.length === 0
              ? "type an IP or CIDR, e.g. 10.0.0.5 or 10.0.0.0/8"
              : ""
          }
          aria-describedby={hintId}
          className="flex-1 min-w-[150px] bg-transparent text-sm text-text-primary placeholder:text-text-secondary outline-none"
        />
      </div>

      {/* Live parse preview — appears as the user types a resolvable entry. */}
      {canAdd && (
        <button
          type="button"
          onMouseDown={(e) => e.preventDefault()}
          onClick={commitDraft}
          className={cn(
            "mt-1.5 w-full flex items-center gap-2.5 px-3 py-2 rounded-lg border text-left transition-colors",
            parsed.status === "any"
              ? "border-accent-yellow/40 bg-accent-yellow/[0.06] hover:bg-accent-yellow/10"
              : "border-border bg-card hover:bg-hover-subtle",
          )}
        >
          <span
            className={cn(
              "grid place-items-center w-5 h-5 rounded shrink-0 text-2xs font-bold",
              parsed.status === "any"
                ? "bg-accent-yellow/15 text-accent-yellow"
                : "bg-accent-green/15 text-accent-green",
            )}
          >
            {parsed.status === "any" ? "!" : "✓"}
          </span>
          <span className="font-mono text-sm text-text-primary">
            {parsed.value}
          </span>
          <span className="text-xs text-text-muted">{parsed.label}</span>
          <span className="ml-auto text-2xs text-text-muted border border-border rounded px-1.5 py-0.5">
            ⏎ add
          </span>
        </button>
      )}
      {parsed.status === "invalid" && (
        <div className="mt-1.5 flex items-center gap-2 px-3 py-2 rounded-lg border border-accent-red/40 bg-accent-red/[0.06]">
          <ExclamationTriangleIcon
            className="w-4 h-4 text-accent-red shrink-0"
            strokeWidth={2}
          />
          <span className="text-xs text-accent-red">{parsed.note}</span>
        </div>
      )}

      {/* Quick-add common private ranges. */}
      {draft === "" && (
        <div className="mt-2 flex flex-wrap gap-1.5">
          {SUGGESTIONS.filter((s) => !values.includes(s.value)).map((s) => (
            <button
              key={s.value}
              type="button"
              onClick={() => add(s.value)}
              className="text-2xs font-mono text-text-secondary bg-card border border-border rounded-md px-2 py-1 hover:border-primary hover:text-primary transition-colors"
            >
              {s.value}
              <span className="font-sans text-text-muted ml-1.5">{s.note}</span>
            </button>
          ))}
        </div>
      )}

      <FieldHint id={hintId}>{hint}</FieldHint>
    </div>
  );
}
