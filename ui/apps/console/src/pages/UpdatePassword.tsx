import { useState, FormEvent } from "react";
import { Link, useSearchParams, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { updateRecoverPassword } from "@/client";
import { updatePasswordResolver } from "./setup/updatePasswordResolver";
import type { UpdatePasswordFormValues } from "./setup/updatePasswordResolver";
import { FormPasswordField } from "@/components/common/fields/rhf";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";

/**
 * Sets a new password from a reset link. The token is in the query string and works once.
 */
export default function UpdatePassword() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const uid = searchParams.get("id") ?? "";
  const token = searchParams.get("token") ?? "";

  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const { control, handleSubmit, formState } =
    useForm<UpdatePasswordFormValues>({
      resolver: updatePasswordResolver,
      mode: "onTouched",
      defaultValues: { password: "", confirmPassword: "" },
    });

  const onSubmit = async (values: UpdatePasswordFormValues) => {
    setError("");
    setLoading(true);
    try {
      await updateRecoverPassword({
        path: { uid },
        body: { token, password: values.password },
        throwOnError: true,
      });
      void navigate("/login", {
        state: { notice: "Password updated successfully. Please sign in." },
      });
    } catch {
      setError(
        "Failed to update password. The link may have expired. Please request a new one.",
      );
    } finally {
      setLoading(false);
    }
  };

  const handleFormSubmit = (e: FormEvent) => {
    void handleSubmit(onSubmit)(e);
  };

  const backToSignIn = [{ label: "Back to sign in", to: "/login" }];

  if (!uid || !token) {
    return (
      <>
        <ScreenIntro
          eyebrow="Password recovery"
          title="Invalid reset link"
          lead="This password reset link is invalid or has expired."
        />
        <AuthActions
          primary={
            <Button as={Link} to="/forgot-password" size="lg" fullWidth>
              Request a new reset link
            </Button>
          }
          links={backToSignIn}
        />
      </>
    );
  }

  return (
    <>
      <ScreenIntro
        eyebrow="Password recovery"
        title="Reset your password"
        lead="Choose a new password for your account."
      />

      <form onSubmit={handleFormSubmit} className="space-y-6">
        {error && <Callout variant="error">{error}</Callout>}

        <FormPasswordField<UpdatePasswordFormValues>
          id="password"
          label="New Password"
          name="password"
          control={control}
          placeholder="••••••••"
          hint="5–32 characters"
          required
        />

        <FormPasswordField<UpdatePasswordFormValues>
          id="confirmPassword"
          label="Confirm Password"
          name="confirmPassword"
          control={control}
          placeholder="••••••••"
          required
        />

        <AuthActions
          primary={
            <Button
              variant="primary"
              size="lg"
              fullWidth
              type="submit"
              loading={loading}
              disabled={loading || !formState.isValid}
            >
              {loading ? "Updating..." : "Update Password"}
            </Button>
          }
          links={backToSignIn}
        />
      </form>
    </>
  );
}
