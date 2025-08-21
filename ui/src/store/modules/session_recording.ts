import { defineStore } from "pinia";
import { ref } from "vue";
import * as usersApi from "../api/users";
import { IUserPutSessionRecording } from "@/interfaces/IUser";

const useSessionRecordingStore = defineStore("sessionRecording", () => {
  const isEnabled = ref<boolean>(true);

  const setStatus = async (data: IUserPutSessionRecording) => {
    await usersApi.setSessionRecordStatus(data);
    isEnabled.value = data.status;
  };

  const getStatus = async () => {
    const res = await usersApi.getSessionRecordStatus();
    isEnabled.value = res.data;
  };

  return {
    isEnabled,
    setStatus,
    getStatus,
  };
});

export default useSessionRecordingStore;
