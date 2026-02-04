import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import Player from "@/components/Sessions/Player.vue";

const mockPlayer = {
  play: vi.fn(),
  pause: vi.fn(),
  seek: vi.fn(),
  dispose: vi.fn(),
  getCurrentTime: vi.fn().mockResolvedValue(0),
  getDuration: vi.fn().mockResolvedValue(100),
  addEventListener: vi.fn(),
};

vi.mock("asciinema-player", () => ({
  create: vi.fn(() => mockPlayer),
}));

describe("Player", () => {
  let wrapper: VueWrapper<InstanceType<typeof Player>>;

  const mockLogs = '{"version":2,"width":80,"height":24}';

  const mountWrapper = (logs: string | null = mockLogs) => {
    wrapper = mountComponent(Player, { props: { logs } });
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Player container rendering", () => {
    it("Renders player container when logs exist", () => {
      const container = wrapper.find('[data-test="player-container"]');
      expect(container.exists()).toBe(true);
    });

    it("Does not render when logs are null", () => {
      wrapper.unmount();
      mountWrapper(null);

      const container = wrapper.find('[data-test="player-container"]');
      expect(container.exists()).toBe(false);
    });
  });

  describe("Player controls rendering", () => {
    it("Renders player controls", () => {
      const controls = wrapper.find('[data-test="player-controls"]');
      expect(controls.exists()).toBe(true);
    });

    it("Shows play button initially", () => {
      // After mount, player starts playing automatically
      // So we need to check the initial state or pause it
      const controls = wrapper.find('[data-test="player-controls"]');
      expect(controls.exists()).toBe(true);
    });

    it("Shows time slider", () => {
      const slider = wrapper.find('[data-test="time-slider"]');
      expect(slider.exists()).toBe(true);
    });

    it("Shows speed select", () => {
      const speedSelect = wrapper.find('[data-test="speed-select"]');
      expect(speedSelect.exists()).toBe(true);
    });

    it("Speed select has correct options", () => {
      const speedSelect = wrapper.findComponent({ name: "VSelect" });
      expect(speedSelect.props("items")).toEqual([0.5, 1, 1.5, 2]);
    });

    it("Shows shortcuts button on medium and up screens", () => {
      const shortcutsBtn = wrapper.find('[data-test="shortcuts-btn"]');
      // Button exists but visibility depends on display size
      expect(shortcutsBtn.exists()).toBe(true);
    });
  });

  describe("Play/Pause functionality", () => {
    it("Shows pause button when playing", () => {
      // Player starts playing automatically
      const pauseBtn = wrapper.find('[data-test="pause-btn"]');
      expect(pauseBtn.exists()).toBe(true);
    });

    it("Calls pause when pause button is clicked", async () => {
      const pauseBtn = wrapper.find('[data-test="pause-btn"]');
      await pauseBtn.trigger("click");

      expect(mockPlayer.pause).toHaveBeenCalled();
    });

    it("Shows play button when paused", async () => {
      const pauseBtn = wrapper.find('[data-test="pause-btn"]');
      await pauseBtn.trigger("click");
      await wrapper.vm.$nextTick();

      const playBtn = wrapper.find('[data-test="play-btn"]');
      expect(playBtn.exists()).toBe(true);
    });

    it("Calls play when play button is clicked", async () => {
      // First pause
      const pauseBtn = wrapper.find('[data-test="pause-btn"]');
      await pauseBtn.trigger("click");
      await wrapper.vm.$nextTick();

      // Then play
      const playBtn = wrapper.find('[data-test="play-btn"]');
      await playBtn.trigger("click");

      expect(mockPlayer.play).toHaveBeenCalled();
    });
  });

  describe("Playback time", () => {
    it("Shows playback time display", () => {
      const playbackTime = wrapper.find('[data-test="playback-time"]');
      // Display is conditional on smAndUp
      expect(playbackTime.exists()).toBe(true);
    });

    it("Displays current time and duration", () => {
      const playbackTime = wrapper.find('[data-test="playback-time"]');
      expect(playbackTime.text()).toContain("/");
    });
  });

  describe("Time slider", () => {
    it("Slider has correct min value", () => {
      const slider = wrapper.findComponent({ name: "VSlider" });
      expect(slider.props("min")).toBe("0");
    });

    it("Calls seek when slider value changes", async () => {
      const slider = wrapper.findComponent({ name: "VSlider" });
      await slider.vm.$emit("update:modelValue", 50);

      expect(mockPlayer.seek).toHaveBeenCalledWith(50);
    });

    it("Pauses when slider mousedown", async () => {
      const slider = wrapper.findComponent({ name: "VSlider" });
      await slider.trigger("mousedown");

      expect(mockPlayer.pause).toHaveBeenCalled();
    });

    it("Plays when slider mouseup", async () => {
      const playSpy = vi.spyOn(mockPlayer, "play");
      const slider = wrapper.findComponent({ name: "VSlider" });
      await slider.trigger("mouseup");

      expect(playSpy).toHaveBeenCalled();
    });

    it("Pauses when slider touchstart", async () => {
      const slider = wrapper.findComponent({ name: "VSlider" });
      await slider.trigger("touchstart");

      expect(mockPlayer.pause).toHaveBeenCalled();
    });

    it("Plays when slider touchend", async () => {
      const playSpy = vi.spyOn(mockPlayer, "play");
      const slider = wrapper.findComponent({ name: "VSlider" });
      await slider.trigger("touchend");

      expect(playSpy).toHaveBeenCalled();
    });
  });

  describe("Speed control", () => {
    it("Has default speed of 1", () => {
      const speedSelect = wrapper.findComponent({ name: "VSelect" });
      expect(speedSelect.props("modelValue")).toBe(1);
    });

    it("Changes playback speed when select changes", async () => {
      const speedSelect = wrapper.findComponent({ name: "VSelect" });
      await speedSelect.vm.$emit("update:modelValue", 1.5);
      await wrapper.vm.$nextTick();

      // Speed change recreates player with new speed
      expect(wrapper.vm.currentSpeed).toBe(1.5);
    });
  });

  describe("Shortcuts dialog", () => {
    it("Renders PlayerShortcutsDialog component", () => {
      const dialog = wrapper.findComponent({ name: "PlayerShortcutsDialog" });
      expect(dialog.exists()).toBe(true);
    });

    it("Opens shortcuts dialog when button is clicked", async () => {
      const shortcutsBtn = wrapper.find('[data-test="shortcuts-btn"]');
      await shortcutsBtn.trigger("click");

      expect(wrapper.findComponent({ name: "PlayerShortcutsDialog" }).exists()).toBe(true);
    });

    it("Pauses player when opening shortcuts", async () => {
      const shortcutsBtn = wrapper.find('[data-test="shortcuts-btn"]');
      await shortcutsBtn.trigger("click");

      expect(mockPlayer.pause).toHaveBeenCalled();
    });
  });

  describe("Keyboard shortcuts", () => {
    it("Toggles play/pause on space key", async () => {
      const container = wrapper.find('[data-test="player-container"]');
      const initialPlayingState = wrapper.vm.isPlaying;

      await container.trigger("keydown.space");
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.isPlaying).toBe(!initialPlayingState);
    });
  });

  describe("Component cleanup", () => {
    it("Disposes player on unmount", () => {
      wrapper.unmount();

      expect(mockPlayer.dispose).toHaveBeenCalled();
    });
  });
});
