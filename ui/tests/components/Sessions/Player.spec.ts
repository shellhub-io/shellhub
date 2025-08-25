import { setActivePinia, createPinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { describe, beforeEach, vi, it, expect, afterEach } from "vitest";
import { nextTick } from "vue";
import { createVuetify } from "vuetify";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";
import Player from "@/components/Sessions/Player.vue";
import formatPlaybackTime from "@/utils/playerPlayback";

vi.mock("asciinema-player", () => ({
  create: vi.fn().mockReturnValue({
    play: vi.fn(),
    pause: vi.fn(),
    seek: vi.fn(),
    getCurrentTime: vi.fn(() => 10),
    getDuration: vi.fn(() => 100),
    addEventListener: vi.fn(),
    dispose: vi.fn(),
  }),
}));

type PlayerWrapper = VueWrapper<InstanceType<typeof Player>>;

describe("Asciinema Player", () => {
  let wrapper: PlayerWrapper;
  const vuetify = createVuetify();
  setActivePinia(createPinia());

  // eslint-disable-next-line vue/max-len
  const logsMock = "{\"version\": 2, \"width\": 80, \"height\": 24}\n[0.123, \"r\", \"80x24\"]\n[1.0, \"o\", \"Asciinema player test\"]\n[2.0, \"o\", \"logout\"]";

  beforeEach(async () => {
    wrapper = mount(Player, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
      props: {
        logs: logsMock,
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", async () => {
    expect(wrapper.find('[data-test="player-container"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="player-controls"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="pause-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="play-btn"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="playback-time"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="time-slider"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="speed-select"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="shortcuts-btn"]').exists()).toBe(true);
  });

  it("Creates player on mount", async () => {
    expect(wrapper.vm.player).toBeDefined();
  });

  it("Initializes with correct default values", () => {
    expect(wrapper.vm.isPlaying).toBe(true);
    expect(wrapper.vm.currentTime).toBe(0);
    expect(wrapper.vm.currentSpeed).toBe(1);
  });

  it("Shows pause button when player is playing", async () => {
    wrapper.vm.isPlaying = true;
    await nextTick();

    const pauseBtn = wrapper.find('[data-test="pause-btn"]');

    expect(pauseBtn.exists()).toBe(true);
    expect(wrapper.find('[data-test="play-btn"]').exists()).toBe(false);
  });

  it("Shows play button when player is paused", async () => {
    wrapper.vm.isPlaying = true;
    await nextTick();
    const pauseBtn = wrapper.find('[data-test="pause-btn"]');

    await pauseBtn.trigger("click");
    await nextTick();

    expect(wrapper.find('[data-test="pause-btn"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="play-btn"]').exists()).toBe(true);
  });

  it("Updates player state when play/pause is clicked", async () => {
    const pauseBtn = wrapper.find('[data-test="pause-btn"]');
    await pauseBtn.trigger("click");

    expect(wrapper.vm.player.pause).toHaveBeenCalled();
    expect(wrapper.vm.isPlaying).toBe(false);

    const playBtn = wrapper.find('[data-test="play-btn"]');
    await playBtn.trigger("click");

    expect(wrapper.vm.player.play).toHaveBeenCalled();
    expect(wrapper.vm.isPlaying).toBe(true);
  });

  it("Shows keyboard shortcuts dialog when button is clicked", async () => {
    const dialogBtn = wrapper.find('[data-test="shortcuts-btn"]');
    await dialogBtn.trigger("click");

    expect(wrapper.vm.showShortcutsDialog).toBe(true);
    expect(wrapper.vm.player.pause).toHaveBeenCalled();
  });

  it("Changes playback speed when speed selector is changed", async () => {
    const speedSelect = wrapper.findComponent({ name: "v-select" });
    await speedSelect.vm.$emit("update:modelValue", 2);

    expect(wrapper.vm.currentSpeed).toBe(2);
  });

  it("Formats time correctly", () => {
    expect(formatPlaybackTime(3661)).toBe("1:01:01"); // hh:mm:ss if session is longer than 1 hour
    expect(formatPlaybackTime(61)).toBe("01:01"); // mm:ss otherwise
    expect(formatPlaybackTime(59)).toBe("00:59"); // Less than 1 minute
    expect(formatPlaybackTime(0)).toBe("00:00"); // Zero time
    expect(formatPlaybackTime(90061)).toBe("25:01:01");
    expect(formatPlaybackTime(172800)).toBe("48:00:00");
    expect(formatPlaybackTime(363599)).toBe("100:59:59");
  });

  it("Updates current time when slider is moved", async () => {
    const newTime = 50;
    const slider = wrapper.findComponent({ name: "v-slider" });

    await slider.vm.$emit("update:modelValue", newTime);

    expect(wrapper.vm.player.seek).toHaveBeenCalledWith(newTime);
  });

  it("Pauses playback when slider interaction starts", async () => {
    const slider = wrapper.find('[data-test="time-slider"]');
    await slider.trigger("mousedown");

    expect(wrapper.vm.player.pause).toHaveBeenCalled();
  });

  it("Resumes playback when slider interaction ends", async () => {
    const slider = wrapper.find('[data-test="time-slider"]');
    await slider.trigger("mouseup");

    expect(wrapper.vm.player.play).toHaveBeenCalled();
  });

  it("Disposes player when component is unmounted", async () => {
    wrapper.unmount();
    expect(wrapper.vm.player.dispose).toHaveBeenCalled();
  });
});
