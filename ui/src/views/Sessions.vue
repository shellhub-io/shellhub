<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row mb-2"
  >
    <h1>Sessions</h1>
  </div>
  <div>
    <SessionList v-if="hasSession" />

    <BoxMessage
      v-if="showBoxMessage"
      typeMessage="session"
      data-test="BoxMessageSession-component"
    />
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from "vue";
import BoxMessage from "../components/Box/BoxMessage.vue";
import { useStore } from "../store";
import SessionList from "../components/Sessions/SessionList.vue";
import { INotificationsError } from "../interfaces/INotifications";
import handleError from "@/utils/handleError";

export default defineComponent({
  setup() {
    const store = useStore();
    const show = ref(false);

    onMounted(async () => {
      try {
        store.dispatch("box/setStatus", true);
        store.dispatch("sessions/resetPagePerpage");

        await store.dispatch("sessions/refresh");
        show.value = true;
      } catch (error: unknown) {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.sessionList,
        );
        handleError(error);
      }
    });

    const hasSession = computed(
      () => store.getters["sessions/getNumberSessions"] > 0,
    );
    const showBoxMessage = computed(() => !hasSession.value && show.value);

    return {
      show,
      hasSession,
      showBoxMessage,
    };
  },
  components: { BoxMessage, SessionList },
});
</script>
