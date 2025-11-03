import { setActivePinia, createPinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { describe, it, beforeEach, vi, expect } from "vitest";
import Terminal from "@/components/Terminal/Terminal.vue";
import useTerminalThemeStore from "@/store/modules/terminal_theme";

class MockWebSocket {
  public readyState: number = WebSocket.CONNECTING;

  public onopen: (() => void) | null = null;

  public onmessage: ((event: { data: string | Blob }) => void) | null = null;

  public onclose: (() => void) | null = null;

  send = vi.fn();

  close = vi.fn();
}

vi.stubGlobal("WebSocket", vi.fn(() => new MockWebSocket()));

vi.mock("@xterm/xterm", () => ({
  Terminal: vi.fn().mockImplementation(() => ({
    open: vi.fn(),
    focus: vi.fn(),
    write: vi.fn(),
    onData: vi.fn(),
    onResize: vi.fn(),
    loadAddon: vi.fn(),
    cols: 80,
    rows: 24,
    options: {},
  })),
}));

vi.mock("@xterm/addon-fit", () => ({
  FitAddon: vi.fn().mockImplementation(() => ({
    fit: vi.fn(),
  })),
}));

describe("Terminal.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof Terminal>>;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const terminalThemeStore = useTerminalThemeStore();
  terminalThemeStore.loadThemes = vi.fn().mockResolvedValue(true);
  terminalThemeStore.loadInitialFont = vi.fn().mockResolvedValue(true);
  terminalThemeStore.$patch({
    currentFontFamily: "Monospace",
    currentFontSize: 15,
    currentThemeName: "ShellHub Dark",
  });

  beforeEach(() => {
    wrapper = mount(Terminal, {
      global: {
        plugins: [vuetify],
        stubs: {
          TerminalThemeDrawer: {
            template: "<div />",
            props: ["modelValue", "showDrawer"],
            emits: ["update:selectedTheme", "update:fontSettings"],
          },
        },
      },
      props: {
        modelValue: true,
        deviceName: "test-device",
        token: "test-token",
      },
    });
  });

  it("renders the terminal container", () => {
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("initializes WebSocket with correct URL parameters", () => {
    const mockWsConstructor = vi.mocked(WebSocket);
    expect(mockWsConstructor).toHaveBeenCalledWith(
      expect.stringContaining("/ws/ssh?token=test-token&cols=80&rows=24"),
    );
  });

  it("closes WebSocket connection on component unmount", () => {
    const mockWs = wrapper.vm.ws as unknown as MockWebSocket;
    mockWs.readyState = WebSocket.OPEN;

    wrapper.unmount();
    expect(mockWs.close).toHaveBeenCalled();
  });

  it("initializes xterm with correct configuration", async () => {
    const { Terminal: MockTerminal } = await import("@xterm/xterm");
    expect(MockTerminal).toHaveBeenCalledWith({
      cursorBlink: true,
      fontFamily: "Monospace",
      fontSize: 15,
    });
  });

  it("loads the FitAddon to xterm", () => {
    const mockXterm = wrapper.vm.xterm;
    expect(mockXterm.loadAddon).toHaveBeenCalledWith(expect.any(Object));
  });

  it("opens xterm in the terminal container", () => {
    const mockXterm = wrapper.vm.xterm;
    const dialog = new DOMWrapper(document.body);
    const terminalContainer = dialog.find("[data-test='terminal-container']").element;
    expect(mockXterm.open).toHaveBeenCalledWith(terminalContainer);
  });

  it("focuses on the terminal after initialization", () => {
    const mockXterm = wrapper.vm.xterm;
    expect(mockXterm.focus).toHaveBeenCalled();
  });

  it("sends user input to WebSocket", () => {
    const mockXterm = wrapper.vm.xterm;
    const mockWs = wrapper.vm.ws as unknown as MockWebSocket;
    const onDataHandler = vi.mocked(mockXterm.onData).mock.calls[0][0];

    mockWs.readyState = WebSocket.OPEN;

    onDataHandler("test input");

    expect(mockWs.send).toHaveBeenCalledWith(
      expect.stringContaining('"kind":1'),
    );
    expect(mockWs.send).toHaveBeenCalledWith(
      expect.stringContaining('"data":"test input"'),
    );
  });

  it("writes Blob WebSocket data to the terminal", async () => {
    const mockXterm = wrapper.vm.xterm;
    const mockWs = wrapper.vm.ws as unknown as MockWebSocket;

    mockWs.onopen?.();

    const testText = "terminal output";

    const realBlob = new Blob();

    Object.defineProperty(realBlob, "text", {
      value: vi.fn().mockResolvedValue(testText),
    });

    mockWs.onmessage?.({ data: realBlob });

    await flushPromises();

    expect(mockXterm.write).toHaveBeenCalledWith(testText);
  });

  it("handles WebSocket close event", () => {
    const mockXterm = wrapper.vm.xterm;
    const mockWs = wrapper.vm.ws as unknown as MockWebSocket;

    mockWs.onclose?.();

    expect(mockXterm.write).toHaveBeenCalledWith("\r\nConnection ended\r\n");
  });

  it("Closes the terminal dialog on double ESC press", () => {
    const escEvent = new KeyboardEvent("keyup", { key: "Escape", bubbles: true });

    document.dispatchEvent(escEvent);
    expect(wrapper.vm.showDialog).toBe(true);

    document.dispatchEvent(escEvent);
    expect(wrapper.vm.showDialog).toBe(false);
  });
});
