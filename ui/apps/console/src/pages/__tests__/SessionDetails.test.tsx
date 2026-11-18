import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import {
  mockNamespace,
  mockRecordingMeta,
  mockSession,
} from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { useTerminalStore } from "@/stores/terminalStore";
import { useRecordingsStore } from "@/stores/recordingsStore";
import SessionDetails from "../SessionDetails";

function ptyRequest(seat: number) {
  return {
    type: "pty-req",
    timestamp: "2024-01-01T00:00:00Z",
    data: { term: "xterm" },
    seat,
  };
}

function serveSession(seats: number[], timelineSeats = seats) {
  server.use(
    http.get("*/api/sessions/:uid", () =>
      HttpResponse.json(
        mockSession({
          uid: "session-1",
          recorded: true,
          events: {
            types: ["pty-req", "shell"],
            seats,
            first: "pty-req",
            items: timelineSeats.map(ptyRequest),
          },
        }),
      ),
    ),
  );
}

function serveRecordings(bodies: Record<string, string>) {
  server.use(
    http.get("*/api/sessions/:uid/records/:seat", ({ params }) => {
      const body = bodies[String(params.seat)];
      return body === undefined
        ? HttpResponse.json({}, { status: 404 })
        : HttpResponse.text(body);
    }),
  );
}

function renderDetails() {
  return render(
    <MemoryRouter initialEntries={["/sessions/session-1"]}>
      <Routes>
        <Route path="/sessions/:uid" element={<SessionDetails />} />
      </Routes>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

async function deleteTerminal(
  user: ReturnType<typeof userEvent.setup>,
  terminal: string,
) {
  await user.click(await screen.findByRole("tab", { name: terminal }));
  await user.click(screen.getByRole("button", { name: "Delete recording" }));
  await user.click(
    within(screen.getByRole("dialog")).getByRole("button", {
      name: "Delete recording",
    }),
  );
}

const openTabs = () => useTerminalStore.getState().recordings.map((r) => r.id);

beforeEach(() => {
  seedAuthStore();
  useTerminalStore.setState({ sessions: [], recordings: [] });
  useRecordingsStore.setState({ recordings: [] });
  server.use(
    http.get("*/api/namespaces/api-key", () => HttpResponse.json([])),
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace()),
    ),
  );
});

describe("SessionDetails recordings", () => {
  it("shows no terminal picker for a session with one terminal", async () => {
    serveSession([0]);
    renderDetails();

    await screen.findByRole("button", { name: "Play Recording" });
    expect(
      screen.queryByRole("tablist", { name: "Terminal" }),
    ).not.toBeInTheDocument();
  });

  it("plays the terminal that is picked", async () => {
    const user = userEvent.setup();
    serveSession([0, 1]);
    serveRecordings({ 0: "first", 1: "second" });
    renderDetails();

    await user.click(await screen.findByRole("tab", { name: "Terminal 2" }));
    await user.click(screen.getByRole("button", { name: "Play Recording" }));

    await waitFor(() => expect(openTabs()).toEqual(["session-1#1"]));
  });

  it("offers a terminal whose request fell past the timeline's cap", async () => {
    const user = userEvent.setup();
    serveSession([0, 1], [0]);
    serveRecordings({ 0: "first", 1: "second" });
    renderDetails();

    await user.click(await screen.findByRole("tab", { name: "Terminal 2" }));
    await user.click(screen.getByRole("button", { name: "Play Recording" }));

    await waitFor(() => expect(openTabs()).toEqual(["session-1#1"]));
  });

  it("shows a terminal the server has no recording for as deleted", async () => {
    const user = userEvent.setup();
    serveSession([0, 1]);
    serveRecordings({ 0: "first" });
    renderDetails();

    await user.click(await screen.findByRole("tab", { name: "Terminal 2" }));
    await user.click(screen.getByRole("button", { name: "Play Recording" }));

    expect(
      await screen.findByText("This terminal's recording was deleted"),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Play Recording" }),
    ).not.toBeInTheDocument();
  });

  it("shows a terminal's playback error only while that terminal is picked", async () => {
    const user = userEvent.setup();
    serveSession([0, 1]);
    server.use(
      http.get(
        "*/api/sessions/:uid/records/:seat",
        () => new HttpResponse(null, { status: 500 }),
      ),
    );
    renderDetails();

    await user.click(await screen.findByRole("tab", { name: "Terminal 2" }));
    await user.click(screen.getByRole("button", { name: "Play Recording" }));
    expect(
      await screen.findByText("Failed to load recording"),
    ).toBeInTheDocument();

    await user.click(screen.getByRole("tab", { name: "Terminal 1" }));
    expect(
      screen.queryByText("Failed to load recording"),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("tab", { name: "Terminal 2" }));
    expect(screen.getByText("Failed to load recording")).toBeInTheDocument();
  });

  it("asks for the play permission for a terminal the browser holds no copy of", async () => {
    const user = userEvent.setup();
    seedAuthStore({ role: "observer" });
    useRecordingsStore.setState({ recordings: [mockRecordingMeta()] });
    serveSession([0, 1]);
    renderDetails();
    const playRestricted = () =>
      screen
        .queryAllByTitle("You don't have permission to perform this action.")
        .some((wrapper) => within(wrapper).queryByText("Play Recording"));

    await screen.findByRole("tab", { name: "Terminal 1" });
    expect(playRestricted()).toBe(false);

    await user.click(screen.getByRole("tab", { name: "Terminal 2" }));
    expect(playRestricted()).toBe(true);
  });

  it("deletes only the terminal that is picked", async () => {
    const user = userEvent.setup();
    const deleted: string[] = [];
    serveSession([0, 1]);
    server.use(
      http.delete("*/api/sessions/:uid/records/:seat", ({ params }) => {
        deleted.push(String(params.seat));
        return new HttpResponse(null, { status: 200 });
      }),
    );
    renderDetails();

    await deleteTerminal(user, "Terminal 2");

    await waitFor(() => expect(deleted).toEqual(["1"]));
  });

  it("shows a terminal deleted from this page as deleted", async () => {
    const user = userEvent.setup();
    serveSession([0, 1]);
    server.use(
      http.delete(
        "*/api/sessions/:uid/records/:seat",
        () => new HttpResponse(null, { status: 200 }),
      ),
    );
    renderDetails();

    await deleteTerminal(user, "Terminal 2");

    expect(
      await screen.findByText("This terminal's recording was deleted"),
    ).toBeInTheDocument();
  });
});
