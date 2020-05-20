<template>
  <fragment>
    <v-tooltip bottom>
      <template #activator="{ on }">
        <v-icon
          v-on="on"
          @click="dialog = !dialog"
        >
          delete
        </v-icon>
      </template>
      <span>Remove</span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card>
        <v-card-title class="headline">
          Are you sure?
        </v-card-title>

        <v-card-text>
          You are about to remove this device
        </v-card-text>
        
        <v-card-actions>
          <v-spacer />

          <v-btn
            color="primary"
            text
            @click="dialog=!dialog"
          >
            CANCEL
          </v-btn>

          <v-btn
            color="red darken-1"
            text
            @click="remove();"
          >
            REMOVE
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>
export default {
  name: 'DeviceDelete',

  props: {
    uid:{
      type: String,
      required: true
    },
    redirect:{
      type: Boolean,
    }
  },

  data() {
    return {
      dialog: false,
    };
  },

  methods:{
    async remove() {
      await this.$store.dispatch('devices/remove', this.uid);
      if(this.redirect){
        this.$router.push('/devices');
      }
    },
  }  
};
</script>

