import { useState } from "react";
import { EyeIcon, EyeSlashIcon } from "@heroicons/react/24/outline";
import { LABEL, INPUT } from "@/utils/styles";

interface PasswordInputProps {
  id: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  hint?: string;
}

export default function PasswordInput({
  id,
  value,
  onChange,
  placeholder,
  hint,
}: PasswordInputProps) {
  const [show, setShow] = useState(false);

  return (
    <div>
      <label className={LABEL} htmlFor={id}>
        Password
      </label>
      <div className="relative">
        <input
          id={id}
          type={show ? "text" : "password"}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          className={`${INPUT} pr-10`}
        />
        <button
          type="button"
          tabIndex={-1}
          onClick={() => setShow(!show)}
          className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary transition-colors"
          aria-label={show ? "Hide password" : "Show password"}
        >
          {show ? (
            <EyeSlashIcon className="w-4 h-4" />
          ) : (
            <EyeIcon className="w-4 h-4" />
          )}
        </button>
      </div>
      {hint && <p className="text-2xs text-text-muted mt-1.5">{hint}</p>}
    </div>
  );
}
