<template>
  <v-avatar
    :size="size"
    color="primary"
    class="border"
  >
    <v-img
      v-if="!avatarLoadingFailed"
      :src="avatarUrl"
      data-test="gravatar-image"
      @error="onImageError"
    />
    <v-icon
      v-else
      color="surface"
      data-test="gravatar-placeholder"
    >
      mdi-account
    </v-icon>
  </v-avatar>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from "vue";
import useAuthStore from "@/store/modules/auth";

const props = defineProps<{
  size: string | number;
  email?: string | null;
}>();

const authStore = useAuthStore();
const userEmail = computed(() => authStore.email);
const effectiveEmail = computed(() => props.email || userEmail.value);

const avatarLoadingFailed = ref(false);
const avatarUrl = ref("");

const generateGravatarUrl = async (email: string | null) => {
  if (!email) {
    avatarUrl.value = "";
    avatarLoadingFailed.value = true;
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
  effectiveEmail,
  async (newEmail) => {
    avatarLoadingFailed.value = false;
    await generateGravatarUrl(newEmail);
  },
  { immediate: true },
);

const onImageError = () => {
  avatarLoadingFailed.value = true;
};

onMounted(async () => {
  await generateGravatarUrl(effectiveEmail.value);
});
</script>

<style scoped>
.border {
  border: 1px solid var(--v-theme-border);
}
</style>
