<template>
  <v-dialog v-model="showDownloadDialog" max-width="500">
    <v-card>
      <v-card-title class="text-h5 px-5 py-3 bg-primary">Session Too Large</v-card-title>
      <v-divider />
      <v-card-text class="text-justify px-5 py-5">
        This session is very large ({{ getBlobSizeInMB() }} MB), and trying to play it in the browser may not work.<br />
        You can download it to your computer and try to play locally with
        <a href="https://docs.asciinema.org/manual/cli/quick-start/" target="_blank" rel="noopener noreferrer">Asciinema CLI</a>.
      </v-card-text>
      <v-card-actions class="d-flex justify-end py-4">
        <v-btn text @click="showDownloadDialog = false">Close</v-btn>
        <v-btn color="primary" @click="handleDownload()" :loading="downloading">Download</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";

const { sessionBlob } = defineProps<{
  sessionBlob: Blob | null;
}>();

const snackbar = useSnackbar();
const showDownloadDialog = defineModel<boolean>();
const downloading = ref(false);

const getBlobSizeInMB = () => {
  const sizeInMB = (sessionBlob?.size || 0) / (1000 * 1000);
  return sizeInMB.toFixed(2);
};

const downloadSessionFile = async () => {
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore: showSaveFilePicker is experimental
  const fileHandler = await window.showSaveFilePicker({
    suggestedName: "session.cast",
    startIn: "downloads",
  });

  const writable = await fileHandler.createWritable();
  await writable.write(sessionBlob);
  snackbar.showInfo("Downloading session...");

  await writable.close();
  snackbar.showSuccess("Session downloaded successfully.");
};

const handleDownload = async () => {
  try {
    downloading.value = true;
    await downloadSessionFile();
  } catch (error) {
    if (error instanceof Error && error.name === "AbortError") {
      snackbar.showInfo("Download cancelled.");
    } else {
      handleError(error);
      snackbar.showError("Failed to download the session.");
    }
  }

  downloading.value = false;
  showDownloadDialog.value = false;
};
</script>
