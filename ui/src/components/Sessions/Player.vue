<template>
  <div
    class="wrapper ma-0 pa-0 w-100 fill-height position-relative bg-v-theme-terminal"
    v-if="logs"
    ref="containerDiv"
    @keydown.space.prevent="isPlaying = !isPlaying"
    data-test="player-container"
  />

  <v-card-actions
    class="text-h5 px-3 py-2 d-flex ga-4 align-center"
    @click="changeFocusToPlayer"
    data-test="player-controls"
  >
    <v-btn
      v-if="isPlaying"
      class="bg-primary"
      rounded="circle"
      size="48"
      :ripple="false"
      icon="mdi-pause"
      data-test="pause-btn"
      @click="pause"
    />
    <v-btn
      v-else
      class="bg-primary"
      rounded="circle"
      size="48"
      :ripple="false"
      icon="mdi-play"
      data-test="play-btn"
      @click="play"
    />

    <span
      v-if="smAndUp"
      id="playback-time"
      class="text-medium-emphasis text-body-1"
      data-test="playback-time"
    >
      {{formattedCurrentTime}} / {{formattedDuration}}
    </span>

    <v-slider
      v-model="currentTime"
      class="ml-0 flex-grow-1 flex-shrink-0"
      min="0"
      :max="duration"
      aria-labelledby="playback-time"
      hide-details
      color="white"
      data-test="time-slider"
      @update:model-value="(value) => changePlaybackTime(value)"
      @mousedown="pause"
      @mouseup="play"
      @touchstart="pause"
      @touchend="play"
    />
    <v-select
      class="flex-grow-0 flex-shrink-0"
      :items="[0.5, 1, 1.5, 2]"
      v-model="currentSpeed"
      hide-details
      flat
      prepend-inner-icon="mdi-speedometer"
      data-test="speed-select"
      @click.stop
      @update:model-value="changePlaybackSpeed"
    />
    <v-btn
      v-if="mdAndUp"
      role="button"
      variant="text"
      icon="mdi-keyboard"
      rounded
      density="compact"
      size="x-large"
      data-test="shortcuts-btn"
      @click="openDialog"
    />
  </v-card-actions>

  <PlayerShortcutsDialog
    v-model:showDialog="showDialog"
  />
</template>

<script setup lang="ts">
import * as AsciinemaPlayer from "asciinema-player";
import { onMounted, onUnmounted, ref, watchEffect } from "vue";
import { useDisplay } from "vuetify";
import PlayerShortcutsDialog from "./PlayerShortcutsDialog.vue";

const { logs } = defineProps<{
  logs: string | null;
}>();

const emit = defineEmits(["close"]);

const containerDiv = ref<HTMLDivElement | null>(null);
const player = ref<AsciinemaPlayer.AsciinemaPlayer | null>(null);
const playerWrapper = ref<HTMLDivElement | null>(null);

const { smAndUp, mdAndUp } = useDisplay();
const showDialog = ref(false);

const isPlaying = ref(true);
const sessionEnded = ref(false);
const currentTime = ref(0);
const duration = ref(0);
const formattedCurrentTime = ref("00:00:00");
const formattedDuration = ref("00:00:00");
const timeUpdaterId = ref<number>();
const currentSpeed = ref(1);

const formatTime = (time: number) => new Date(time * 1000).toISOString().slice(time >= 3600 ? 11 : 14, 19);
const changeFocusToPlayer = () => { playerWrapper.value?.focus(); };

const getSessionDimensions = () => {
  const dimensionsLine = logs?.split("\n")[1] ?? ""; // returns a string in the format of `[0.123, "r", "80x24"]`
  const dimensions = JSON.parse(dimensionsLine)[2];
  const [cols, rows] = dimensions.split("x");
  return { cols, rows };
};

const getCurrentTime = () => {
  const time = player.value.getCurrentTime();
  currentTime.value = time;
  formattedCurrentTime.value = formatTime(time);
};

const getDuration = () => {
  duration.value = player.value.getDuration();
  formattedDuration.value = formatTime(duration.value);
};

const changePlaybackTime = (value: number) => {
  player.value.seek(value);
  getCurrentTime();
};

const clearCurrentTimeUpdater = () => {
  clearInterval(timeUpdaterId.value);
};

const startCurrentTimeUpdater = () => {
  clearCurrentTimeUpdater();
  timeUpdaterId.value = window.setInterval(getCurrentTime, 100);
};

const play = () => {
  player.value.play();
  isPlaying.value = true;
};

const pause = () => {
  player.value.pause();
  isPlaying.value = false;
};

const openDialog = () => {
  pause();
  showDialog.value = true;
};

const createPlayer = (startAt = 0) => {
  const playerOptions = {
    fit: "height",
    controls: false,
    ...getSessionDimensions(),
    speed: currentSpeed.value,
    startAt,
  };

  return AsciinemaPlayer.create({ data: logs }, containerDiv.value, playerOptions);
};

const setPlayerEventListeners = () => {
  player.value.addEventListener("playing", () => {
    sessionEnded.value = false;
    getCurrentTime();
    // clear to prevent multiple intervals when replaying
    clearCurrentTimeUpdater();
    startCurrentTimeUpdater();
    getDuration();
  });

  player.value.addEventListener("ended", () => {
    sessionEnded.value = true;
    isPlaying.value = false;
    clearCurrentTimeUpdater();
  });

  containerDiv.value?.addEventListener("keydown", (event: KeyboardEvent) => {
    if (event.key === "Escape") emit("close");
  });
};

const changePlaybackSpeed = () => {
  const startAt = sessionEnded.value ? 0 : currentTime.value;

  player.value.dispose();
  player.value = createPlayer(startAt);
  play();
  setPlayerEventListeners();
  playerWrapper.value = containerDiv.value?.querySelector(".ap-wrapper") as HTMLDivElement;
};

onMounted(() => {
  player.value = createPlayer();
  playerWrapper.value = containerDiv.value?.querySelector(".ap-wrapper") as HTMLDivElement;
  changeFocusToPlayer();
  play();
  setPlayerEventListeners();
});

onUnmounted(() => {
  clearCurrentTimeUpdater();
  player.value?.dispose();
});

watchEffect(() => !showDialog.value && changeFocusToPlayer());

defineExpose({ player });
</script>

<style lang="scss" scoped>
:deep(.ap-wrapper) {
  background-color: #121314;
  justify-content: start;
  max-height: calc(100vh - 4rem) !important;
}
</style>
