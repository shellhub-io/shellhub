import { ref } from "vue";
import moment from "moment";

export default function useCountdown() {
  const countdown = ref("");

  let countdownInterval;

  function startCountdown(loginTimeoutEpoch: number) {
    clearInterval(countdownInterval);
    const endTime = moment.unix(loginTimeoutEpoch); // Convert to seconds
    countdownInterval = setInterval(() => {
      const diff = moment.duration(endTime.diff(moment()));
      if (diff.asSeconds() <= 0) {
        clearInterval(countdownInterval);
        countdown.value = "0 seconds";
      } else if (diff.asMinutes() < 1) {
        countdown.value = `${Math.floor(diff.asSeconds())} seconds`;
      } else {
        countdown.value = `${Math.floor(diff.asMinutes())} minutes`;
      }
    }, 1000);
  }

  return { startCountdown, countdown };
}
