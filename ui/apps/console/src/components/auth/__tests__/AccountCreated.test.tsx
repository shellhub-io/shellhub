import { describe, it, expect, beforeEach } from "vitest";
import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import { useSignUpStore } from "@/stores/signUpStore";
import { mockUserAuth } from "@/tests/factories";
import AccountCreated from "../AccountCreated";

beforeEach(() => {
  useAuthStore.setState({ token: null, userId: null, onboardingSurvey: null });
});

describe("AccountCreated", () => {
  it("signs an invited user in with the session the sign-up returned, onboarding survey included", () => {
    useSignUpStore.setState({
      signUpAuth: mockUserAuth({
        token: "signup-token",
        tenant: "tenant-abc",
        id: "user-new",
        onboarding_survey: "pending",
      }),
    });

    render(
      <MemoryRouter initialEntries={["/sign-up?sig=x"]}>
        <AccountCreated />
      </MemoryRouter>,
    );

    const auth = useAuthStore.getState();
    expect(auth.token).toBe("signup-token");
    expect(auth.tenant).toBe("tenant-abc");
    expect(auth.userId).toBe("user-new");
    expect(auth.onboardingSurvey).toBe("pending");
  });
});
