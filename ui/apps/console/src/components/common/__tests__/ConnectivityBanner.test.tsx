import { describe, it, expect, afterEach, beforeEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import ConnectivityBanner from "@/components/common/ConnectivityBanner";
import { useConnectivityStore } from "@/stores/connectivityStore";

beforeEach(() => useConnectivityStore.setState({ apiReachable: true }));
afterEach(cleanup);

describe("ConnectivityBanner", () => {
  it("is hidden when api is reachable", () => {
    useConnectivityStore.setState({ apiReachable: true });
    render(<ConnectivityBanner />);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("delegates to NoticeBanner: renders role=alert when api is unreachable", () => {
    useConnectivityStore.setState({ apiReachable: false });
    render(<ConnectivityBanner />);
    expect(screen.getByRole("alert")).toBeInTheDocument();
  });

  it("shows the unreachable message text when api is unreachable", () => {
    useConnectivityStore.setState({ apiReachable: false });
    render(<ConnectivityBanner />);
    expect(
      screen.getByText(/API unreachable — reconnecting automatically/),
    ).toBeInTheDocument();
  });
});
