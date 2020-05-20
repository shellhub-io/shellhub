<template>
  <fragment>
    <v-dialog
      v-model="dialog"
      max-width="290"
    >
      <v-card>
        <v-card-title class="headline">
          Are you sure?
        </v-card-title>

        <v-card-text>
          Your device is going to be deleted
        </v-card-text>
        
        <v-card-actions>
          <v-spacer />

          <v-btn
            color="green darken-1"
            text
            @click="dialog = false"
          >
            Cancel
          </v-btn>

          
          <v-btn
            color="green darken-1"
            text
            @click="remove(); dialog=false"
          >
            Ok
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
    dialog:{
      type: Boolean,
    }
  },

  methods:{
    async remove() {
      await this.$store.dispatch('devices/remove', this.uid);
      this.$router.push('/devices');
    },
  }

  
};
</script>