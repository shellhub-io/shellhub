<template>
  <v-container>
    <v-row
      align="center"
      justify="center"
      class="mt-4"
    >
      <v-col sm="8">
        <SettingOwnerInfo
          :is-owner="isOwner"
        />
        <div
          v-if="isOwner"
          data-test="webhookOperation"
        >
          <SettingWebhook />
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>

import SettingWebhook from '@/components/setting/SettingWebhook';
import SettingOwnerInfo from '@/components/setting/SettingOwnerInfo';

export default {
  name: 'SettingAPI',

  components: {
    SettingWebhook,
    SettingOwnerInfo,
  },

  data() {
    return {
    };
  },

  computed: {
    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },
  },
};

</script>
