import { useState, FormEvent } from "react";
import { Link, useSearchParams, useNavigate } from "react-router-dom";
import {
  ExclamationCircleIcon,
  LockClosedIcon,
} from "@heroicons/react/24/outline";
import { useForm } from "react-hook-form";
import { Button } from "@shellhub/design-system/primitives";
import { activateUser } from "@/client";
import { updatePasswordResolver } from "./setup/updatePasswordResolver";
import type { UpdatePasswordFormValues } from "./setup/updatePasswordResolver";
import { FormPasswordField } from "@/components/common/fields/rhf";

export default function ActivateAccount() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const id = searchParams.get("id") ?? "";
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
      await activateUser({
        path: { id },
        body: { token, password: values.password },
        throwOnError: true,
      });
      void navigate("/login", {
        state: { notice: "Account activated. Please sign in." },
      });
    } catch {
      setError(
        "Failed to activate the account. The link may have expired. Ask your administrator for a new one.",
      );
    } finally {
      setLoading(false);
    }
  };

  const handleFormSubmit = (e: FormEvent) => {
    void handleSubmit(onSubmit)(e);
  };

  if (!id || !token) {
    return (
      <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
        <div className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 text-center animate-fade-in">
          <ExclamationCircleIcon
            className="w-10 h-10 text-accent-red mx-auto mb-4"
            strokeWidth={1.5}
          />
          <p className="text-sm font-semibold text-text-primary mb-2">
            Invalid activation link
          </p>
          <p className="text-xs text-text-muted mb-6">
            This activation link is invalid or has expired.
          </p>
          <Link
            to="/login"
            className="text-xs text-primary hover:text-primary-400 transition-colors"
          >
            Back to login
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      {/* Hero */}
      <div className="text-center mb-12 animate-fade-in">
        <div className="animate-float mb-6 inline-block">
          <div className="w-20 h-20 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center shadow-lg shadow-primary/10">
            <LockClosedIcon
              className="w-10 h-10 text-primary"
              strokeWidth={1.2}
            />
          </div>
        </div>

        <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
          Account Activation
        </p>
        <h1 className="text-3xl font-bold text-text-primary mb-3">
          Activate your account
        </h1>
        <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
          Choose a password to finish setting up your account.
        </p>
      </div>

      {/* Card */}
      <div
        className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
        style={{ animationDelay: "200ms" }}
      >
        <form onSubmit={handleFormSubmit} className="space-y-5">
          {error && (
            <div
              role="alert"
              className="flex items-start gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down"
            >
              <ExclamationCircleIcon
                className="w-3.5 h-3.5 shrink-0 mt-0.5"
                strokeWidth={2}
              />
              {error}
            </div>
          )}

          <FormPasswordField<UpdatePasswordFormValues>
            id="password"
            label="Password"
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

          <Button
            variant="primary"
            size="lg"
            fullWidth
            type="submit"
            className="px-4"
            loading={loading}
            disabled={loading || !formState.isValid}
          >
            {loading ? "Activating..." : "Activate Account"}
          </Button>
        </form>
      </div>

      {/* Back to login */}
      <div className="mt-8 animate-fade-in" style={{ animationDelay: "600ms" }}>
        <Link
          to="/login"
          className="text-xs text-text-muted hover:text-text-secondary transition-colors"
        >
          &larr; Back to login
        </Link>
      </div>
    </div>
  );
}
