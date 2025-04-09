<template>
  <div
    class="wrapper ma-0 pa-0 w-100 fill-height position-relative bg-v-theme-terminal"
    v-if="logs"
    ref="containerDiv"
    @keydown.space.prevent="isPlaying = !isPlaying"
  />

  <v-card-actions
    class="text-h5 pa-3 d-flex ga-4 align-center"
    @click="changeFocusToPlayer()"
  >
    <v-btn
      v-if="isPlaying"
      class="bg-primary"
      rounded="circle"
      size="48"
      :ripple="false"
      icon="mdi-pause"
      data-test="pause-icon"
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

    <v-slider
      v-model="currentTime"
      class="ml-0 flex-grow-1 flex-shrink-0"
      min="0"
      :max="duration"
      :label="`${formattedCurrentTime} - ${formattedDuration}`"
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
      @update:model-value="changePlaybackSpeed()"
    />
    <v-btn
      role="button"
      variant="text"
      icon="mdi-keyboard"
      rounded
      size="x-large"
      data-test="keyboard-icon"
      @click="openDialog"
    />
  </v-card-actions>
  <v-dialog
    :fullscreen="false"
    v-model="showDialog"
  >
    <v-card title="Keyboard Shortcuts">
      <v-card-text>
        <div class="shortcut"><v-kbd>space</v-kbd>: pause / resume</div>
        <div class="shortcut"><v-kbd>f</v-kbd>: toggle fullscreen mode</div>
        <div class="shortcut"><v-kbd>←</v-kbd> / <v-kbd>→</v-kbd>: rewind / fast-forward by 5 seconds</div>
        <div class="shortcut"><v-kbd>Shift</v-kbd> + <v-kbd>←</v-kbd> / <v-kbd>→</v-kbd>: rewind / fast-forward by 10%</div>
        <div class="shortcut"><v-kbd>0</v-kbd>, <v-kbd>1</v-kbd>, <v-kbd>2</v-kbd> ... <v-kbd>9</v-kbd>: jump to 0%, 10%, 20% ... 90%</div>
        <div class="shortcut"><v-kbd>,</v-kbd>/<v-kbd>.</v-kbd>: step back / forward, a frame at a time (only when paused)</div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import * as AsciinemaPlayer from "asciinema-player";
import { onMounted, ref, watchEffect } from "vue";

const { logs } = defineProps<{
  logs: string | null;
}>();

const isPlaying = ref(true);
const showDialog = ref(false);
const containerDiv = ref<HTMLDivElement | null>(null);
const player = ref<AsciinemaPlayer.AsciinemaPlayer | null>(null);
const playerWrapper = ref<HTMLDivElement | null>(null);
const currentTime = ref(0);
const duration = ref(0);
const formattedCurrentTime = ref("00:00:00");
const formattedDuration = ref("00:00:00");
const timeUpdaterId = ref<number>();
const currentSpeed = ref(1);

const formatTime = (time: number) => new Date(time * 1000).toISOString().slice(11, 19);

const getCurrentTime = () => {
  const time = player.value.getCurrentTime();
  currentTime.value = time;
  formattedCurrentTime.value = formatTime(time);
};

const clearCurrentTimeUpdater = () => {
  clearInterval(timeUpdaterId.value);
};

const startCurrentTimeUpdater = () => {
  clearCurrentTimeUpdater();
  timeUpdaterId.value = window.setInterval(
    () => {
      getCurrentTime();
    },
    100,
  );
};

const changePlaybackTime = (value: number) => {
  player.value.seek(value);
  currentTime.value = value;
  formattedCurrentTime.value = formatTime(value);
};

const getSessionDimensions = () => {
  const dimensionsLine = logs?.split("\n")[1] ?? ""; // returns a string in the format of `[0.123, "r", "80x24"]`
  const dimensions = JSON.parse(dimensionsLine)[2];
  const [cols, rows] = dimensions.split("x");
  return { cols, rows };
};

const playerOptions = {
  fit: "height",
  controls: false,
  ...getSessionDimensions(),
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

const setPlayerEventListeners = () => {
  player.value.addEventListener("playing", () => {
    getCurrentTime();
    // clear to prevent multiple intervals when replaying
    clearInterval(timeUpdaterId.value);
    startCurrentTimeUpdater();
    duration.value = player.value.getDuration();
    formattedDuration.value = formatTime(duration.value);
  });

  player.value.addEventListener("ended", () => {
    isPlaying.value = false;
    clearCurrentTimeUpdater();
  });
};

const changePlaybackSpeed = () => {
  player.value.dispose();
  player.value = AsciinemaPlayer.create(
    { data: logs },
    containerDiv.value,
    { ...playerOptions, speed: currentSpeed.value, startAt: currentTime.value },
  );
  play();
  setPlayerEventListeners();
};

const changeFocusToPlayer = () => { playerWrapper.value?.focus(); };

onMounted(() => {
  player.value = AsciinemaPlayer.create({ data: logs }, containerDiv.value, playerOptions);

  playerWrapper.value = containerDiv.value?.querySelector(".ap-wrapper") as HTMLDivElement;
  changeFocusToPlayer();

  play();

  setPlayerEventListeners();
});

watchEffect(() => !showDialog.value && changeFocusToPlayer());
</script>

<style lang="scss" scoped>
:deep(.ap-wrapper) {
  background-color: #121314;
  justify-content: start;
}

.shortcut {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 1rem;

  .v-kbd {
    padding: .2rem .5rem;
    font-weight: 700;
  }
}
</style>
