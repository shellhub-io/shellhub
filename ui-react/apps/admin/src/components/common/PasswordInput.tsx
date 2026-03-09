import { InputHTMLAttributes, forwardRef } from "react";
import { INPUT } from "@/utils/styles";

/**
 * Password input that uses `type="text"` + CSS masking to prevent
 * browsers from triggering credential management (save/autofill prompts)
 * on non-login forms.
 *
 * Use this for master passwords, vault passphrases, device passwords,
 * and any field where browser credential management should not interfere.
 *
 * For actual login/signup forms, use a regular `<input type="password">`
 * with proper `autoComplete` values (`current-password`, `new-password`).
 */
const PasswordInput = forwardRef<
  HTMLInputElement,
  Omit<InputHTMLAttributes<HTMLInputElement>, "type">
>(({ className, ...props }, ref) => (
  <input
    ref={ref}
    type="text"
    autoComplete="off"
    spellCheck={false}
    autoCapitalize="off"
    autoCorrect="off"
    className={className ?? INPUT}
    style={
      {
        WebkitTextSecurity: "disc",
        textSecurity: "disc",
      } as React.CSSProperties
    }
    {...props}
  />
));

PasswordInput.displayName = "PasswordInput";

export default PasswordInput;
