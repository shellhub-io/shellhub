<template>
  <fragment>
    <div class="d-flex pa-0 align-center">
      <h1>Public Keys</h1>

      <v-spacer />
      <v-spacer />

      <span @click="publicKeyCreateShow = !publicKeyCreateShow">
        <PublicKeyCreate
          :create-key="true"
          :show.sync="publicKeyCreateShow"
          data-test="publicKeyCreate-component"
          @update="refresh"
        />
      </span>
    </div>

    <v-card class="mt-2">
      <router-view
        v-if="hasPublickey"
      />

      <BoxMessagePublicKey
        v-if="showBoxMessage"
        type-message="publicKey"
        data-test="boxMessagePublicKey-component"
      />
    </v-card>
  </fragment>
</template>

<script>

import PublicKeyCreate from '@/components/public_key/KeyFormDialog';
import BoxMessagePublicKey from '@/components/box/BoxMessage';

export default {
  name: 'PublickeyComponent',

  components: {
    PublicKeyCreate,
    BoxMessagePublicKey,
  },

  data() {
    return {
      show: false,
      publicKeyCreateShow: false,
    };
  },

  computed: {
    hasPublickey() {
      return this.$store.getters['publickeys/getNumberPublicKeys'] > 0;
    },

    showBoxMessage() {
      return !this.hasPublickey && this.show;
    },
  },

  async created() {
    this.$store.dispatch('boxs/setStatus', true);
    this.$store.dispatch('publickeys/resetPagePerpage');

    await this.refresh();
    this.show = true;
  },

  methods: {
    async refresh() {
      try {
        await this.$store.dispatch('publickeys/refresh');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.publicKeyList);
      }
    },
  },
};
</script>
