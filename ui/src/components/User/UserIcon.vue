<template>
  <v-avatar :size="size" color="primary" class="border">
    <v-img
      v-if="!avatarLoadingFailed"
      :src="avatarUrl"
      @error="onImageError"
      data-test="gravatar-image"
    />
    <v-icon v-else color="surface" data-test="gravatar-placeholder">mdi-account</v-icon>
  </v-avatar>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from "vue";
import { useStore } from "@/store";

interface Props {
  size: string | number;
}

defineProps<Props>();

const store = useStore();
const userEmail = computed(() => store.getters["auth/email"]);

const avatarLoadingFailed = ref(false);
const avatarUrl = ref("");

const generateGravatarUrl = async (email: string | null) => {
  if (!email) {
    avatarUrl.value = "";
    return;
  }
  const encoder = new TextEncoder();
  const data = encoder.encode(email);
  const hash = await crypto.subtle.digest("SHA-256", data);
  const digest = Array.from(new Uint8Array(hash))
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
  avatarUrl.value = `https://gravatar.com/avatar/${digest}?d=404`;
};

watch(
  () => userEmail.value,
  (newEmail) => {
    avatarLoadingFailed.value = false;
    generateGravatarUrl(newEmail);
  },
  { immediate: true },
);

const onImageError = () => {
  avatarLoadingFailed.value = true;
};

onMounted(() => {
  generateGravatarUrl(userEmail.value);
});
</script>

<style scoped>
.border {
  border: 1px solid var(--v-theme-border);
}
</style>
