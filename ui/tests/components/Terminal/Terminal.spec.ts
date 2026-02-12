import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper, DOMWrapper } from "@vue/test-utils";
import TerminalComponent from "@/components/Terminal/Terminal.vue";
import { mountComponent } from "@tests/utils/mount";
import { mockTerminalThemes } from "@tests/mocks/terminalTheme";
import { MessageKind } from "@/interfaces/ITerminal";
import useTerminalThemeStore from "@/store/modules/terminal_theme";
import { Terminal } from "@xterm/xterm";

vi.mock("@/utils/sshKeys", () => ({
  generateSignature: vi.fn(() => "mock-signature-response"),
}));

class MockWebSocket {
  public readyState: number = WebSocket.OPEN;

  public onopen: (() => void) | null = null;

  public onmessage: ((event: { data: string | Blob }) => void) | null = null;

  public onclose: (() => void) | null = null;

  send = vi.fn();

  close = vi.fn();
}

const MockWebSocketConstructor = vi.fn(() => new MockWebSocket()) as unknown as {
  new(url: string | URL, protocols?: string | string[]): MockWebSocket;
  readonly CONNECTING: 0;
  readonly OPEN: 1;
  readonly CLOSING: 2;
  readonly CLOSED: 3;
};

Object.defineProperty(MockWebSocketConstructor, "CONNECTING", { value: 0, writable: false });
Object.defineProperty(MockWebSocketConstructor, "OPEN", { value: 1, writable: false });
Object.defineProperty(MockWebSocketConstructor, "CLOSING", { value: 2, writable: false });
Object.defineProperty(MockWebSocketConstructor, "CLOSED", { value: 3, writable: false });

vi.stubGlobal("WebSocket", MockWebSocketConstructor);

const mockXtermInstance = {
  open: vi.fn(),
  focus: vi.fn(),
  write: vi.fn(),
  onData: vi.fn(),
  onResize: vi.fn(),
  loadAddon: vi.fn(),
  cols: 80,
  rows: 24,
  options: {} as Record<string, unknown>,
};

const mockFitAddon = {
  fit: vi.fn(),
};

vi.mock("@xterm/xterm", () => ({
  Terminal: vi.fn(() => mockXtermInstance),
}));

vi.mock("@xterm/addon-fit", () => ({
  FitAddon: vi.fn(() => mockFitAddon),
}));

describe("Terminal", () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalComponent>>;
  let dialog: DOMWrapper<Element>;
  let terminalThemeStore: ReturnType<typeof useTerminalThemeStore>;

  const mountWrapper = ({
    token = "test-token",
    deviceName = "test-device",
    privateKey = null as string | null,
    passphrase = undefined as string | undefined,
  } = {}) => {
    wrapper = mountComponent(TerminalComponent, {
      props: {
        modelValue: true,
        token,
        deviceName,
        privateKey,
        passphrase,
      },
      global: {
        stubs: {
          TerminalThemeDrawer: {
            name: "TerminalThemeDrawer",
            template: "<div data-test='theme-drawer-stub' />",
            props: ["modelValue", "showDrawer"],
            emits: ["update:selectedTheme", "update:fontSettings"],
          },
        },
      },
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          terminalTheme: {
            currentFontFamily: "Monospace",
            currentFontSize: 15,
            currentThemeName: "ShellHub Dark",
            terminalThemes: mockTerminalThemes,
          },
        },
      },
    });

    terminalThemeStore = useTerminalThemeStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
    mockXtermInstance.options = {};
    mockXtermInstance.onData.mockClear();
    mockXtermInstance.onResize.mockClear();
  });

  describe("Component rendering", () => {
    it("renders WindowDialog with correct props", () => {
      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.exists()).toBe(true);
      expect(windowDialog.props("title")).toBe("Terminal");
      expect(windowDialog.props("description")).toBe("Connected to test-device");
      expect(windowDialog.props("icon")).toBe("mdi-console");
      expect(windowDialog.props("showCloseButton")).toBe(true);
      expect(windowDialog.props("forceFullscreen")).toBe(true);
    });

    it("renders terminal container", () => {
      expect(dialog.find('[data-test="terminal-container"]').exists()).toBe(true);
    });

    it("renders theme toggle button", () => {
      const themeBtn = dialog.find('[data-test="theme-toggle-btn"]');
      expect(themeBtn.exists()).toBe(true);
    });
  });

  describe("xterm initialization", () => {
    it("initializes xterm with correct configuration", () => {
      expect(Terminal).toHaveBeenCalledWith({
        cursorBlink: true,
        fontFamily: "Monospace",
        fontSize: 15,
      });
    });

    it("loads theme on initialization", () => {
      expect(mockXtermInstance.options.theme).toBeDefined();
      expect(mockXtermInstance.options.theme).toEqual(mockTerminalThemes[0].colors);
    });

    it("loads FitAddon to xterm", () => {
      expect(mockXtermInstance.loadAddon).toHaveBeenCalledWith(mockFitAddon);
    });

    it("opens xterm in the terminal container", () => {
      const terminalContainer = dialog.find('[data-test="terminal-container"]').element;
      expect(mockXtermInstance.open).toHaveBeenCalledWith(terminalContainer);
    });

    it("focuses on the terminal after initialization", () => {
      expect(mockXtermInstance.focus).toHaveBeenCalled();
    });

    it("loads initial font settings from store", () => {
      expect(terminalThemeStore.loadInitialFont).toHaveBeenCalled();
    });

    it("loads themes from store", () => {
      expect(terminalThemeStore.loadThemes).toHaveBeenCalled();
    });
  });

  describe("WebSocket initialization", () => {
    it("initializes WebSocket with correct URL parameters", () => {
      const protocol = window.location.protocol === "http:" ? "ws" : "wss";
      expect(MockWebSocketConstructor).toHaveBeenCalledWith(
        `${protocol}://${window.location.host}/ws/ssh?token=test-token&cols=80&rows=24`,
      );
    });

    it("fits terminal before establishing WebSocket connection", () => {
      expect(mockFitAddon.fit).toHaveBeenCalled();
    });
  });

  describe("Terminal input and resize", () => {
    it("sends user input to WebSocket when connection is open", () => {
      const onDataHandler = vi.mocked(mockXtermInstance.onData).mock.calls[0][0];
      const mockWs = wrapper.vm.ws as unknown as MockWebSocket;
      mockWs.readyState = WebSocket.OPEN;

      onDataHandler("test input");

      expect(mockWs.send).toHaveBeenCalledWith(
        JSON.stringify({
          kind: MessageKind.Input,
          data: "test input",
        }),
      );
    });

    it("limits input data to 4096 characters", () => {
      const onDataHandler = vi.mocked(mockXtermInstance.onData).mock.calls[0][0];
      const mockWs = wrapper.vm.ws as unknown as MockWebSocket;
      mockWs.readyState = WebSocket.OPEN;

      const longInput = "a".repeat(5000);
      onDataHandler(longInput);

      const sentData = JSON.parse(mockWs.send.mock.calls[0][0] as string);
      expect(sentData.data).toHaveLength(4096);
    });

    it("does not send input when WebSocket is not open", () => {
      const mockWs = wrapper.vm.ws as unknown as MockWebSocket;

      mockWs.readyState = WebSocket.CLOSED;

      mockWs.send.mockClear();

      const onDataHandler = vi.mocked(mockXtermInstance.onData).mock.calls[0][0];
      onDataHandler("test input");

      expect(mockWs.send).not.toHaveBeenCalled();
    });

    it("sends resize message to WebSocket", () => {
      const onResizeHandler = vi.mocked(mockXtermInstance.onResize).mock.calls[0][0];
      const mockWs = wrapper.vm.ws as unknown as MockWebSocket;
      mockWs.readyState = WebSocket.OPEN;

      onResizeHandler({ cols: 100, rows: 30 });

      expect(mockWs.send).toHaveBeenCalledWith(
        JSON.stringify({
          kind: MessageKind.Resize,
          data: { cols: 100, rows: 30 },
        }),
      );
    });

    it("does not send resize when WebSocket is not open", () => {
      const mockWs = wrapper.vm.ws as unknown as MockWebSocket;

      mockWs.readyState = WebSocket.CLOSED;

      mockWs.send.mockClear();

      const onResizeHandler = vi.mocked(mockXtermInstance.onResize).mock.calls[0][0];
      onResizeHandler({ cols: 100, rows: 30 });

      expect(mockWs.send).not.toHaveBeenCalled();
    });
  });

  describe("WebSocket message handling", () => {
    it("writes Blob data to terminal for password-based login", async () => {
      const mockWs = wrapper.vm.ws as unknown as MockWebSocket;
      const testText = "terminal output";
      const mockBlob = new Blob();

      Object.defineProperty(mockBlob, "text", {
        value: vi.fn().mockResolvedValue(testText),
      });

      mockWs.onmessage?.({ data: mockBlob });
      await flushPromises();

      expect(mockXtermInstance.write).toHaveBeenCalledWith(testText);
    });

    it("handles error message from WebSocket", () => {
      const mockWs = wrapper.vm.ws as unknown as MockWebSocket;
      const errorMessage = JSON.stringify({
        kind: MessageKind.Error,
        data: "Authentication failed",
      });

      mockWs.onmessage?.({ data: errorMessage });

      expect(mockXtermInstance.write).toHaveBeenCalledWith("Authentication failed");
    });

    it("handles signature challenge for private key authentication", async () => {
      wrapper.unmount();
      document.body.innerHTML = "";

      mountWrapper({
        privateKey: "test-private-key",
        passphrase: "test-passphrase",
      });

      await flushPromises();

      const mockWs = wrapper.vm.ws as unknown as MockWebSocket;

      const challengeMessage = JSON.stringify({
        kind: MessageKind.Signature,
        data: Buffer.from("challenge").toString("base64"),
      });

      mockWs.onmessage?.({ data: challengeMessage });
      await flushPromises();

      expect(mockWs.send).toHaveBeenCalledWith(
        expect.stringContaining(`"kind":${MessageKind.Signature}`),
      );
    });

    it("does not handle signature challenge without private key", () => {
      const mockWs = wrapper.vm.ws as unknown as MockWebSocket;
      const challengeMessage = JSON.stringify({
        kind: MessageKind.Signature,
        data: Buffer.from("challenge").toString("base64"),
      });

      mockWs.onmessage?.({ data: challengeMessage });

      const sentMessages = mockWs.send.mock.calls.filter((call) => {
        const data = JSON.parse(call[0] as string);
        return data.kind === MessageKind.Signature;
      });
      expect(sentMessages).toHaveLength(0);
    });

    it("writes connection ended message on WebSocket close", () => {
      const mockWs = wrapper.vm.ws as unknown as MockWebSocket;

      mockWs.onclose?.();

      expect(mockXtermInstance.write).toHaveBeenCalledWith("\r\nConnection ended\r\n");
    });
  });

  describe("Theme management", () => {
    it("toggles theme drawer when button is clicked", async () => {
      const themeBtn = dialog.find('[data-test="theme-toggle-btn"]');

      await themeBtn.trigger("click");
      await flushPromises();

      expect(themeBtn.find("i").classes()).toContain("mdi-palette");

      await themeBtn.trigger("click");
      await flushPromises();

      expect(themeBtn.find("i").classes()).toContain("mdi-palette-outline");
    });

    it("applies theme when update:selectedTheme is emitted", async () => {
      const themeDrawer = wrapper.findComponent({ name: "TerminalThemeDrawer" });
      expect(themeDrawer.exists()).toBe(true);

      themeDrawer.vm.$emit("update:selectedTheme", mockTerminalThemes[1]);
      await flushPromises();

      expect(mockXtermInstance.options.theme).toEqual(mockTerminalThemes[1].colors);
    });

    it("applies font settings when update:fontSettings is emitted", async () => {
      const themeDrawer = wrapper.findComponent({ name: "TerminalThemeDrawer" });
      expect(themeDrawer.exists()).toBe(true);

      themeDrawer.vm.$emit("update:fontSettings", {
        fontFamily: "JetBrains Mono",
        fontSize: 18,
      });
      await flushPromises();

      expect(mockXtermInstance.options.fontSize).toBe(18);
      expect(mockXtermInstance.options.fontFamily).toBe("JetBrains Mono");
      expect(mockFitAddon.fit).toHaveBeenCalled();
    });
  });

  describe("Dialog close behavior", () => {
    it("closes dialog on double ESC press within 400ms", async () => {
      const escEvent = new KeyboardEvent("keyup", { key: "Escape", bubbles: true });

      document.dispatchEvent(escEvent);
      await flushPromises();
      const windowDialog1 = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog1.props("modelValue")).toBe(true);

      document.dispatchEvent(escEvent);
      await flushPromises();
      const windowDialog2 = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog2.props("modelValue")).toBe(false);
    });

    it("does not close dialog on single ESC press", async () => {
      const escEvent = new KeyboardEvent("keyup", { key: "Escape", bubbles: true });

      document.dispatchEvent(escEvent);
      await flushPromises();
      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("modelValue")).toBe(true);
    });

    it("does not close dialog on double ESC press with delay > 400ms", async () => {
      vi.useFakeTimers();

      const escEvent = new KeyboardEvent("keyup", { key: "Escape", bubbles: true });

      document.dispatchEvent(escEvent);
      vi.advanceTimersByTime(500);
      document.dispatchEvent(escEvent);
      await flushPromises();

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("modelValue")).toBe(true);

      vi.useRealTimers();
    });

    it("emits close event when dialog is closed", async () => {
      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });

      windowDialog.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("close")).toBeTruthy();
    });
  });

  describe("Lifecycle and cleanup", () => {
    it("closes WebSocket connection on component unmount if open", () => {
      const mockWs = wrapper.vm.ws as unknown as MockWebSocket;
      mockWs.readyState = WebSocket.OPEN;

      wrapper.unmount();

      expect(mockWs.close).toHaveBeenCalled();
    });

    it("does not call close on WebSocket if not open", () => {
      const mockWs = wrapper.vm.ws as unknown as MockWebSocket;

      mockWs.readyState = WebSocket.CLOSED;

      mockWs.close.mockClear();

      wrapper.unmount();

      expect(mockWs.close).not.toHaveBeenCalled();
    });
  });
});
