import Vue from 'vue'
import './plugins/vuetify'
import App from './App.vue'
import router from "./router";
import store from './store'

Vue.config.productionTip = false

Vue.use(require('vue-moment'));

new Vue({
  router,
  store,
  render: h => h(App),
}).$mount('#app')
