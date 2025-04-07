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
      @click="isPlaying = false"
    />
    <v-icon
      v-else
      variant="text"
      icon="mdi-play-circle"
      color="primary"
      rounded
      size="x-large"
      data-test="play-icon"
      @click="isPlaying = true"
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

    />
  </v-card-actions>
</template>

<script setup lang="ts">
import * as AsciinemaPlayer from "asciinema-player";
import { onMounted, ref, watchEffect } from "vue";

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

const formatTime = (time: number) => new Date(time * 1000).toISOString().slice(11, 19);

onMounted(() => {
  const playerOptions = {
    fit: "height",
    terminalFontSize: "1rem",
    controls: false,
    autoplay: true,
  };
  const timeIntervalId = ref<NodeJS.Timeout>();

  player.value = AsciinemaPlayer.create({ data: logs }, wrapper.value, playerOptions);

  player.value.addEventListener("playing", () => {
    console.log("playing");
    timeIntervalId.value = setInterval(() => {
      console.log("currentTime", player.value.getCurrentTime());
      currentTime.value = player.value.getCurrentTime();
      formattedCurrentTime.value = formatTime(currentTime.value);
    }, 100);
    duration.value = player.value.getDuration();
    formattedDuration.value = formatTime(duration.value);
  });

  player.value.addEventListener("ended", () => {
    isPlaying.value = false;
    clearInterval(timeIntervalId.value);
  });
});

watchEffect(() => isPlaying.value ? player.value?.play() : player.value?.pause());
</script>

<style lang="scss" scoped>
:deep(.ap-wrapper) {
  background-color: #121314;
  justify-content: start;
}
</style>
