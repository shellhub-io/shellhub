<template>
  <div class="d-flex pa-0 align-center">
    <h1>Firewall Details</h1>
  </div>
  <v-card class="mt-2 pa-4" v-if="!firewallRuleIsEmpty">
    <v-card-text>
      <div>
        <div class="text-overline mt-3">
          <h3>id:</h3>
        </div>
        <div :data-test="firewallRule.id">
          <p>{{ firewallRule.id }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Tenant Id:</h3>
        </div>
        <div :data-test="firewallRule.tenant_id">
          <p>{{ firewallRule.tenant_id }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Priority:</h3>
        </div>
        <div :data-test="firewallRule.priority">
          <p>{{ firewallRule.priority }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Action:</h3>
        </div>
        <div :data-test="firewallRule.action">
          <p>{{ firewallRule.action }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Source Ip:</h3>
        </div>
        <div :data-test="firewallRule.source_ip">
          <p>{{ formatSourceIP(firewallRule.source_ip) }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Username:</h3>
        </div>
        <div :data-test="firewallRule.username">
          <p>{{ formatUsername(firewallRule.username) }}</p>
        </div>
      </div>

      <div v-if="firewallRule.filter">
        <div class="text-overline mt-3">
          <h3>Filter:</h3>
        </div>
        <div v-if="isHostname(firewallRule.filter)" :data-test="firewallRule.filter">
          {{ formatHostnameFilter(firewallRule.filter) }}
        </div>
        <div v-else :data-test="firewallRule.filter">
          <v-tooltip
            v-for="(tag, index) in firewallRule.filter.tags"
            :key="index"
            bottom
            :disabled="!showTag(tag)"
          >
            <template #activator="{ props }">
              <v-chip class="mr-1" density="compact" outlined v-bind="props">
                {{ displayOnlyTenCharacters(tag) }}
              </v-chip>
            </template>

            <span v-if="showTag(tag)">
              {{ tag }}
            </span>
          </v-tooltip>
        </div>
      </div>
    </v-card-text>
  </v-card>
  <v-card class="mt-2 pa-4" v-else>
    <p class="text-center">Something is wrong, try again !</p>
  </v-card>
</template>

<script lang="ts">
import { computed, ref, defineComponent, onMounted } from "vue";
import { useRoute } from "vue-router";
import { filterType, IFirewallRule } from "../interfaces/IFirewallRule";
import { INotificationsError } from "../interfaces/INotifications";
import { useStore } from "../store";
import showTag from "../hooks/tag";
import displayOnlyTenCharacters from "../hooks/string";

export default defineComponent({
  setup() {
    const store = useStore();
    const route = useRoute();

    const firewallRuleId = computed(() => route.params.id);
    const firewallRule = ref({} as IFirewallRule);

    onMounted(async () => {
      try {
        await store.dispatch("firewallRules/get", firewallRuleId.value);
        firewallRule.value = store.getters["firewallRules/get"];
      } catch {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.firewallRuleDetails);
      }
    });

    const firewallRuleIsEmpty = computed(() => store.getters["firewallRules/get"] && store.getters["firewallRules/get"].lenght === 0);

    const formatSourceIP = (ip: string) => (ip === ".*" ? "Any IP" : ip);

    const formatUsername = (username: string) => (username === ".*" ? "All users" : username);

    const formatHostnameFilter = (filter: filterType) => filter.hostname === ".*" ? "All devices" : filter.hostname;

    const isHostname = (filter: filterType) => Object.prototype.hasOwnProperty.call(filter, "hostname");

    return {
      firewallRule,
      firewallRuleIsEmpty,
      formatSourceIP,
      formatUsername,
      formatHostnameFilter,
      isHostname,
      showTag,
      displayOnlyTenCharacters,
    };
  },
});
</script>
S
