<template>
  <v-dialog
    v-model="dialog"
    transition="dialog-bottom-transition"
    width="650"
    height="800"
    persistent
  >
    <v-card color="background" data-test="card-dialog">
      <v-container>
        <v-row data-test="icon-crown">
          <v-col class="d-flex justify-center align-center">
            <div class="circle-one shadow d-flex justify-center align-center">
              <div class="circle-two shadow d-flex justify-center align-center">
                <v-icon color="success" class="green-inner-shadow" size="108">
                  mdi-crown-circle
                </v-icon>
              </div>
            </div>
          </v-col>
        </v-row>
        <v-row>
          <v-col class="pb-0">
            <h1 class="d-flex justify-center align-center text-center" data-test="upgrade-heading">
              Upgrade to have access to all features!
            </h1>
          </v-col>
        </v-row>
        <v-row>
          <v-col class="pt-1 pb-6">
            <p class="d-flex justify-center align-center text-grey text-center" data-test="upgrade-description">
              To use this feature, upgrade from the ShellHub Community Edition to one of our premium editions.
              Each edition of ShellHub offers its own set of features and benefits, making it easy to find the
              right solution for your needs.
            </p>
          </v-col>
        </v-row>
        <div v-if="items.length === 0">
          <v-row>
            <v-col class="d-flex align-center justify-center">
              <v-btn
                href="https://www.shellhub.io"
                color="primary"
                target="_blank"
                rel="noreferrer noopener"
                data-test="no-link-available-btn"
              >
                Checkout our website
              </v-btn>
            </v-col>
          </v-row>
        </div>
        <v-row v-else data-test="items-row">
          <v-col v-for="(item, i) in items" :key="i" :data-test="'item-' + i">
            <v-card class="bg-v-theme-surface border d-flex flex-column justify-space-between" height="100%" :data-test="'item-card-' + i">
              <v-card-title class="d-flex justify-center" :data-test="'item-title-' + i">
                <b>{{ item.title }}</b>
              </v-card-title>
              <v-card-text class="flex-grow-1" :data-test="'item-content-' + i">
                <v-row v-for="(feature, j) in item.features" :key="j" :data-test="'item-content-row-' + i + '-' + j">
                  <v-col class="d-flex align-center pb-2">
                    <v-icon>mdi-check-circle</v-icon>
                    <p class="ml-2">{{ feature }}</p>
                  </v-col>
                </v-row>
              </v-card-text>
              <v-card-actions class="d-flex justify-center" :data-test="'item-actions-' + i">
                <v-btn
                  block
                  color="primary"
                  variant="outlined"
                  :href="item.button.link"
                  target="_blank"
                  rel="noreferrer noopener"
                  :data-test="'pricing-btn-' + i">
                  {{ item.button.label }}
                </v-btn>
              </v-card-actions>
            </v-card>
          </v-col>
        </v-row>
        <v-card-actions data-test="card-actions">
          <v-spacer />
          <v-btn @click="close" class="mt-4" variant="text" data-test="close-btn">
            Close
          </v-btn>
        </v-card-actions>
      </v-container>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useStore } from "../../store";

const store = useStore();
const dialog = ref(false);
const close = () => {
  dialog.value = false;
  store.commit("users/setShowPaywall", false);
};
const items = computed(() => store.getters["users/getPremiumContent"]);

onMounted(() => {
  store.dispatch("users/getPremiumContent");
});
defineExpose({ dialog });
</script>

<style scoped>
.green-inner-shadow {
  filter: drop-shadow(0px 0px 20px rgba(43, 255, 10, 0.45));
}

.shadow {
  box-shadow: rgba(0, 0, 0, 0.35) 0px 15px 10px 0px;
  border-radius: 50%;
}

.circle-one {
  height: 15.625rem;
  width: 15.625rem;
  background: linear-gradient(180deg, rgba(0,0,0,0) 40%, rgba(76,175,80,0.1) 60%, rgba(76,175,80,0.75) 120%);
}

.circle-two {
  height: 12.5rem;
  width: 12.5rem;
  background: linear-gradient(180deg, rgba(0,0,0,0) 60%, rgba(76,175,80,0.1) 75%, rgba(76,175,80,0.75) 120%);
}
</style>
