import { useEffect } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import { resolvePostLoginRedirect } from "@/utils/navigation";
import PendingDeviceCallout from "@/components/auth/PendingDeviceCallout";
import ScreenIntro from "@/components/layout/ScreenIntro";
import MfaCodeForm from "@/components/mfa/MfaCodeForm";

/**
 * The second factor at sign-in. Holds the partial token from the first step and exchanges it for
 * a full one, so a wrong code leaves the user still un-signed-in rather than signed out.
 */
export default function MfaLogin() {
  const mfaToken = useAuthStore((s) => s.mfaToken);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (!mfaToken) {
      void navigate("/login");
    }
  }, [mfaToken, navigate]);

  if (!mfaToken) {
    return null;
  }

  return (
    <>
      <ScreenIntro
        eyebrow="Two-factor"
        title="Two-Factor Authentication"
        lead="Enter the 6-digit code from your authenticator app to complete sign in."
      />

      <div className="mb-6 empty:hidden">
        <PendingDeviceCallout />
      </div>

      <MfaCodeForm
        links={[{ label: "Use a recovery code", to: "/mfa-recover" }]}
        onVerified={() =>
          void navigate(
            resolvePostLoginRedirect(new URLSearchParams(location.search)),
          )
        }
      />
    </>
  );
}
