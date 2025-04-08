<template>
  <div class="wrapper ma-0 pa-0 w-100 fill-height position-relative bg-v-theme-terminal" v-if="logs" ref="wrapper" />

  <v-card-actions
    class="text-h5 pa-3 d-flex justify-start ga-4 align-center"
  >
    <v-icon
      v-if="isPlaying"
      variant="text"
      icon="mdi-pause-circle"
      color="primary"
      rounded
      size="x-large"
      data-test="pause-icon"
      @click="pause"
    />
    <v-icon
      v-else
      variant="text"
      icon="mdi-play-circle"
      color="primary"
      rounded
      size="x-large"
      data-test="play-icon"
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
    />
  </v-card-actions>
</template>

<script setup lang="ts">
import * as AsciinemaPlayer from "asciinema-player";
import { onMounted, ref } from "vue";

const { logs } = defineProps<{
  logs: string | null;
}>();

const isPlaying = ref(true);
const wrapper = ref<HTMLDivElement | null>(null);
const player = ref<AsciinemaPlayer.AsciinemaPlayer | null>(null);
const currentTime = ref(0);
const duration = ref(0);
const formattedCurrentTime = ref("00:00:00");
const formattedDuration = ref("00:00:00");
const timeUpdaterId = ref<number>();

const formatTime = (time: number) => new Date(time * 1000).toISOString().slice(11, 19);

const clearCurrentTimeUpdater = () => {
  clearInterval(timeUpdaterId.value);
};

const startCurrentTimeUpdater = () => {
  clearCurrentTimeUpdater();
  timeUpdaterId.value = window.setInterval(() => {
    if (player.value) {
      const time = player.value.getCurrentTime();
      currentTime.value = time;
      formattedCurrentTime.value = formatTime(time);
    }
  }, 100);
};

const changePlaybackTime = (value: number) => {
  player.value.seek(value);
  currentTime.value = value;
  formattedCurrentTime.value = formatTime(value);
};

const getSessionRows = () => {
  const dimensionsLine = logs?.split("\n")[1] ?? ""; // returns a string in the format of `[0.123, "r", "80x24"]`
  const dimensions = JSON.parse(dimensionsLine)[2];
  const rows = dimensions.split("x")[1];
  return rows;
};

onMounted(() => {
  const playerOptions = {
    fit: "height",
    controls: false,
    autoplay: true,
    rows: getSessionRows(),
  };

  player.value = AsciinemaPlayer.create({ data: logs }, wrapper.value, playerOptions);

  player.value.addEventListener("playing", () => {
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
});

const play = () => {
  player.value.play();
  isPlaying.value = true;
};

const pause = () => {
  player.value.pause();
  isPlaying.value = false;
};
</script>

<style lang="scss" scoped>
:deep(.ap-wrapper) {
  background-color: #121314;
  justify-content: start;
}
</style>
