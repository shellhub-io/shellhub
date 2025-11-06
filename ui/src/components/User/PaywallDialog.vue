<template>
  <MessageDialog
    v-model="showDialog"
    title="Upgrade to have access to all features!"
    description="To use this feature, upgrade from the ShellHub Community Edition to one of our premium editions.
    Each edition of ShellHub offers its own set of features and benefits, making it easy to find the right solution for your needs."
    icon="mdi-crown-circle"
    icon-color="success"
    cancel-text="Close"
    cancel-data-test="close-btn"
    data-test="paywall-features-dialog"
    @close="close"
    @cancel="close"
  >
    <div
      v-if="premiumFeatures.length === 0"
      class="d-flex justify-center mt-6"
    >
      <v-btn
        href="https://www.shellhub.io"
        color="primary"
        target="_blank"
        rel="noreferrer noopener"
        data-test="no-link-available-btn"
      >
        Check out our website
      </v-btn>
    </div>

    <v-row
      v-else
      class="mt-6"
      data-test="items-row"
    >
      <v-col
        v-for="(item, i) in premiumFeatures"
        :key="i"
        :data-test="'item-' + i"
      >
        <v-card
          class="bg-v-theme-surface border d-flex flex-column justify-space-between"
          height="100%"
          :data-test="'item-card-' + i"
        >
          <v-card-title
            class="d-flex justify-center"
            :data-test="'item-title-' + i"
          >
            <b>{{ item.title }}</b>
          </v-card-title>

          <v-card-text
            class="flex-grow-1"
            :data-test="'item-content-' + i"
          >
            <v-row
              v-for="(feature, j) in item.features"
              :key="j"
              :data-test="'item-content-row-' + i + '-' + j"
            >
              <v-col class="d-flex align-center pb-2">
                <v-icon>mdi-check-circle</v-icon>
                <p class="ml-2 text-left">
                  {{ feature }}
                </p>
              </v-col>
            </v-row>
          </v-card-text>

          <v-card-actions
            class="d-flex justify-center"
            :data-test="'item-actions-' + i"
          >
            <v-btn
              block
              color="primary"
              variant="outlined"
              :href="item.button.link"
              target="_blank"
              rel="noreferrer noopener"
              :data-test="'pricing-btn-' + i"
            >
              {{ item.button.label }}
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </MessageDialog>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import useUsersStore from "@/store/modules/users";
import { IPremiumFeature } from "@/interfaces/IUser";

const usersStore = useUsersStore();
const showDialog = ref(false);
const premiumFeatures = ref<Array<IPremiumFeature>>([]);

const close = () => {
  showDialog.value = false;
  usersStore.showPaywall = false;
};

onMounted(async () => {
  premiumFeatures.value = await usersStore.getPremiumContent();
});

defineExpose({ showDialog });
</script>
