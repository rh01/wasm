import Go from './wasm_exec'

import Vue from 'vue'
import App from './App.vue'
import apiConfig from './api.config'
import Axios from 'axios'
import VueAxios from 'vue-axios'

Vue.use(VueAxios, Axios)
Axios.defaults.baseURL = apiConfig.baseURL
/* eslint no-undef: "off"*/
const go = new Go()
WebAssembly.instantiateStreaming(fetch("calc.wasm"), go.importObject)
  // .then(async (result) => {
  // await go.run(result.instance)
  .then((result) => {
    go.run(result.instance)
  //   console.log(waAdd(...Array("2", "1")))
    // log().then((result) => console.log(result))

    Vue.prototype.$go = {
      add: waAdd,
      sub: waSub,
      multi: waMulti,
      divi: waDivi,
      // log: log,
      log: myGoFunc,
      asyncOne: asyncOne,
      fetchHttp: fetchHttp,
      fetchMongoDocument: fetchMongoDocument
    }
  })

new Vue({
  render: h => h(App)
}).$mount('#app')
