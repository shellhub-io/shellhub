import type { SVGProps } from "react";

// Google's brand guidelines require the four-colour mark, so this one ignores currentColor.
export function GoogleIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" {...props}>
      <path
        fill="#4285F4"
        d="M23.52 12.273c0-.851-.076-1.67-.218-2.455H12v4.642h6.458a5.52 5.52 0 01-2.394 3.622v3.01h3.878c2.269-2.088 3.578-5.163 3.578-8.82z"
      />
      <path
        fill="#34A853"
        d="M12 24c3.24 0 5.956-1.075 7.942-2.908l-3.878-3.01c-1.075.72-2.45 1.145-4.064 1.145-3.125 0-5.77-2.11-6.715-4.947H1.276v3.109A11.995 11.995 0 0012 24z"
      />
      <path
        fill="#FBBC05"
        d="M5.285 14.28A7.213 7.213 0 014.909 12c0-.791.136-1.56.376-2.28V6.611H1.276A11.995 11.995 0 000 12c0 1.937.464 3.769 1.276 5.389l4.009-3.109z"
      />
      <path
        fill="#EA4335"
        d="M12 4.773c1.762 0 3.344.606 4.589 1.795l3.442-3.442C17.951 1.19 15.235 0 12 0 7.31 0 3.255 2.69 1.276 6.611l4.009 3.109C6.23 6.883 8.875 4.773 12 4.773z"
      />
    </svg>
  );
}
