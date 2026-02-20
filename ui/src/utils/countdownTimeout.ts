import { ref, onUnmounted } from "vue";
import moment from "moment";

export default function useCountdown() {
  const countdown = ref("");

  let countdownInterval: NodeJS.Timeout;

  function startCountdown(loginTimeoutEpoch: number) {
    clearInterval(countdownInterval);
    const endTime = moment.unix(loginTimeoutEpoch); // Convert to seconds
    countdownInterval = setInterval(() => {
      const diff = moment.duration(endTime.diff(moment()));
      if (diff.asSeconds() <= 0) {
        clearInterval(countdownInterval);
        countdown.value = "0 seconds";
      } else if (diff.asMinutes() < 1) {
        const seconds = Math.floor(diff.asSeconds());
        countdown.value = `${seconds} ${seconds === 1 ? "second" : "seconds"}`;
      } else {
        const minutes = Math.floor(diff.asMinutes());
        countdown.value = `${minutes} ${minutes === 1 ? "minute" : "minutes"}`;
      }
    }, 1000);
  }

  onUnmounted(() => clearInterval(countdownInterval));

  return { startCountdown, countdown };
}
